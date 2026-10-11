package link

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aestek/dotfiler/tmpl"
)

func write(t *testing.T, p, c string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(c), 0644); err != nil {
		t.Fatal(err)
	}
}

func readlink(t *testing.T, p string) string {
	t.Helper()
	got, err := os.Readlink(p)
	if err != nil {
		t.Fatal(err)
	}
	return got
}

func TestSource(t *testing.T) {
	l := Link{From: "/a/b", To: "/c"}
	if l.Source() != "/a/b" {
		t.Fatalf("Source = %q", l.Source())
	}
	l.Templated = true
	if l.Source() != "/a/b"+tmpl.Ext {
		t.Fatalf("Source = %q", l.Source())
	}
}

func TestString(t *testing.T) {
	s := Link{From: "/a/from", To: "/b/to"}.String()
	if !strings.Contains(s, "from") || !strings.Contains(s, "to") {
		t.Fatalf("String = %q", s)
	}
}

func TestStatus(t *testing.T) {
	_, dir := setup(t)
	from := filepath.Join(dir, "from")
	to := filepath.Join(dir, "to")
	l := Link{From: from, To: to}

	check := func(want LinkStatus) {
		t.Helper()
		got, err := l.Status()
		if err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Fatalf("status = %v, want %v", got, want)
		}
	}

	check(LinkStatusSourceMiss)

	write(t, from, "x")
	check(LinkStatusUnlinked)

	write(t, to, "y")
	check(LinkStatusTargetExists)

	os.Remove(to)
	other := filepath.Join(dir, "other")
	write(t, other, "o")
	if err := os.Symlink(other, to); err != nil {
		t.Fatal(err)
	}
	check(LinkStatusUnlinked)

	os.Remove(to)
	if err := os.Symlink(from, to); err != nil {
		t.Fatal(err)
	}
	check(LinkStatusLinked)
}

func TestStatusTemplated(t *testing.T) {
	_, dir := setup(t)
	from := filepath.Join(dir, "from")
	to := filepath.Join(dir, "to")
	write(t, from, "x")
	l := Link{From: from, To: to, Templated: true}

	if err := os.Symlink(from, to); err != nil {
		t.Fatal(err)
	}
	if s, _ := l.Status(); s != LinkStatusUnlinked {
		t.Fatalf("status = %v, want Unlinked", s)
	}
	os.Remove(to)
	if err := os.Symlink(from+tmpl.Ext, to); err != nil {
		t.Fatal(err)
	}
	if s, err := l.Status(); err != nil || s != LinkStatusLinked {
		t.Fatalf("status = %v, %v", s, err)
	}
}

func TestUpdateSourceMiss(t *testing.T) {
	_, dir := setup(t)
	to := filepath.Join(dir, "sub", "to")
	l := Link{From: filepath.Join(dir, "missing"), To: to}

	s, err := l.Update(nil)
	if err != nil || s != LinkStatusSourceMiss {
		t.Fatalf("status = %v, err = %v", s, err)
	}
	if _, err := os.Lstat(filepath.Dir(to)); !os.IsNotExist(err) {
		t.Fatalf("parent dir created: %v", err)
	}
}

func TestUpdateFreshCreatesParents(t *testing.T) {
	_, dir := setup(t)
	from := filepath.Join(dir, "from")
	to := filepath.Join(dir, "a", "b", "to")
	write(t, from, "x")
	l := Link{From: from, To: to}

	s, err := l.Update(nil)
	if err != nil || s != LinkStatusLinked {
		t.Fatalf("status = %v, err = %v", s, err)
	}
	if readlink(t, to) != from {
		t.Fatalf("target = %q", readlink(t, to))
	}
	if s, _ := l.Status(); s != LinkStatusLinked {
		t.Fatalf("status = %v", s)
	}
}

func TestUpdateAlreadyLinkedNoop(t *testing.T) {
	_, dir := setup(t)
	from := filepath.Join(dir, "from")
	to := filepath.Join(dir, "to")
	write(t, from, "x")
	l := Link{From: from, To: to}
	if _, err := l.Update(nil); err != nil {
		t.Fatal(err)
	}
	before, err := os.Lstat(to)
	if err != nil {
		t.Fatal(err)
	}

	s, err := l.Update(nil)
	if err != nil || s != LinkStatusLinked {
		t.Fatalf("status = %v, err = %v", s, err)
	}
	after, err := os.Lstat(to)
	if err != nil {
		t.Fatal(err)
	}
	if !os.SameFile(before, after) || readlink(t, to) != from {
		t.Fatal("symlink changed")
	}
}

func TestUpdateReplacesStaleSymlink(t *testing.T) {
	_, dir := setup(t)
	from := filepath.Join(dir, "from")
	to := filepath.Join(dir, "to")
	other := filepath.Join(dir, "other")
	write(t, from, "x")
	write(t, other, "o")
	if err := os.Symlink(other, to); err != nil {
		t.Fatal(err)
	}
	l := Link{From: from, To: to}

	s, err := l.Update(nil)
	if err != nil || s != LinkStatusLinked {
		t.Fatalf("status = %v, err = %v", s, err)
	}
	if readlink(t, to) != from {
		t.Fatalf("target = %q", readlink(t, to))
	}
	if c, _ := os.ReadFile(other); string(c) != "o" {
		t.Fatalf("other modified: %q", c)
	}
}

func TestUpdateRegularFileErrors(t *testing.T) {
	_, dir := setup(t)
	from := filepath.Join(dir, "from")
	to := filepath.Join(dir, "to")
	write(t, from, "x")
	write(t, to, "precious")
	l := Link{From: from, To: to}

	if _, err := l.Update(nil); err == nil {
		t.Fatal("expected error")
	}
	s, err := os.Lstat(to)
	if err != nil || !s.Mode().IsRegular() {
		t.Fatalf("to changed: %v %v", s, err)
	}
	if c, _ := os.ReadFile(to); string(c) != "precious" {
		t.Fatalf("content = %q", c)
	}
}

func TestUpdateTemplated(t *testing.T) {
	_, dir := setup(t)
	from := filepath.Join(dir, "from")
	to := filepath.Join(dir, "to")
	write(t, from, "v={{.v}}")
	l := Link{From: from, To: to, Templated: true}

	s, err := l.Update(map[string]string{"v": "1"})
	if err != nil || s != LinkStatusLinked {
		t.Fatalf("status = %v, err = %v", s, err)
	}
	if readlink(t, to) != from+tmpl.Ext {
		t.Fatalf("target = %q", readlink(t, to))
	}
	if c, _ := os.ReadFile(from + tmpl.Ext); string(c) != "v=1" {
		t.Fatalf("rendered = %q", c)
	}
	if s, _ := l.Status(); s != LinkStatusLinked {
		t.Fatalf("status = %v", s)
	}

	s, err = l.Update(map[string]string{"v": "2"})
	if err != nil || s != LinkStatusLinked {
		t.Fatalf("status = %v, err = %v", s, err)
	}
	if c, _ := os.ReadFile(to); string(c) != "v=2" {
		t.Fatalf("rendered = %q", c)
	}
}

func TestMakeLinkErrors(t *testing.T) {
	base, dir := setup(t)

	if _, err := MakeLink(base, ""); err == nil {
		t.Error("empty file: expected error")
	}

	inside := filepath.Join(base, "f")
	write(t, inside, "x")
	if _, err := MakeLink(base, inside); err == nil {
		t.Error("file inside base: expected error")
	}

	if _, err := MakeLink(base, filepath.Join(dir, "nope")); err == nil {
		t.Error("nonexistent file: expected error")
	}
}

func TestMakeLinkMovesFile(t *testing.T) {
	base, dir := setup(t)
	file := filepath.Join(dir, "sub", "rc")
	write(t, file, "content")

	l, err := MakeLink(base, file)
	if err != nil {
		t.Fatal(err)
	}
	want := repoPath(t, base, file)
	if l.From != want || l.To != file {
		t.Fatalf("link = %+v, want From %q To %q", l, want, file)
	}
	if c, _ := os.ReadFile(want); string(c) != "content" {
		t.Fatalf("moved content = %q", c)
	}
	if _, err := os.Lstat(file); !os.IsNotExist(err) {
		t.Fatalf("original still exists: %v", err)
	}
	if _, err := l.Update(nil); err != nil {
		t.Fatal(err)
	}
	if c, _ := os.ReadFile(file); string(c) != "content" {
		t.Fatalf("content via link = %q", c)
	}
}

func TestUnlinkRegularFile(t *testing.T) {
	base, dir := setup(t)
	file := filepath.Join(dir, "rc")
	write(t, file, "keep")

	if err := Unlink(base, file); err == nil {
		t.Fatal("expected error")
	}
	if c, _ := os.ReadFile(file); string(c) != "keep" {
		t.Fatalf("content = %q", c)
	}
}

func TestUnlinkTargetOutsideBase(t *testing.T) {
	base, dir := setup(t)
	target := filepath.Join(dir, "elsewhere")
	link := filepath.Join(dir, "lnk")
	write(t, target, "x")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}

	if err := Unlink(base, link); err == nil {
		t.Fatal("expected error")
	}
	if readlink(t, link) != target {
		t.Fatal("symlink modified")
	}
	if _, err := os.Stat(target); err != nil {
		t.Fatal(err)
	}
}

func TestUnlinkBasePrefixSibling(t *testing.T) {
	base, dir := setup(t)
	target := filepath.Join(dir, "dots2", "f")
	link := filepath.Join(dir, "lnk")
	write(t, target, "x")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}

	if err := Unlink(base, link); err == nil {
		t.Fatal("expected error")
	}
	if readlink(t, link) != target {
		t.Fatal("symlink modified")
	}
	if _, err := os.Stat(target); err != nil {
		t.Fatal(err)
	}
}

func TestUnlinkRelativeTarget(t *testing.T) {
	base, dir := setup(t)
	write(t, filepath.Join(base, "f"), "content")
	link := filepath.Join(dir, "lnk")
	if err := os.Symlink(filepath.Join("dots", "f"), link); err != nil {
		t.Fatal(err)
	}

	if err := Unlink(base, link); err != nil {
		t.Fatal(err)
	}
	s, err := os.Lstat(link)
	if err != nil || !s.Mode().IsRegular() {
		t.Fatalf("expected regular file, got %v %v", s, err)
	}
	if c, _ := os.ReadFile(link); string(c) != "content" {
		t.Fatalf("content = %q", c)
	}
	if _, err := os.Lstat(filepath.Join(base, "f")); !os.IsNotExist(err) {
		t.Fatalf("repo file still exists: %v", err)
	}
}

func TestUnlinkDirectoryRoundTrip(t *testing.T) {
	base, dir := setup(t)
	d := filepath.Join(dir, "conf")
	write(t, filepath.Join(d, "inner", "f"), "deep")
	write(t, filepath.Join(d, "top"), "top")

	l, err := MakeLink(base, d)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := l.Update(nil); err != nil {
		t.Fatal(err)
	}
	if s, err := os.Lstat(d); err != nil || s.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("expected symlink, got %v %v", s, err)
	}

	if err := Unlink(base, d); err != nil {
		t.Fatal(err)
	}
	s, err := os.Lstat(d)
	if err != nil || !s.IsDir() {
		t.Fatalf("expected dir, got %v %v", s, err)
	}
	if c, _ := os.ReadFile(filepath.Join(d, "inner", "f")); string(c) != "deep" {
		t.Fatalf("inner = %q", c)
	}
	if c, _ := os.ReadFile(filepath.Join(d, "top")); string(c) != "top" {
		t.Fatalf("top = %q", c)
	}
}
