package link

import (
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"testing"

	dpath "github.com/aestek/dotfiler/path"
)

func repoPath(t *testing.T, base, file string) string {
	usr, err := user.Current()
	if err != nil {
		t.Fatal(err)
	}
	if dpath.HasDirPrefix(file, usr.HomeDir) {
		file = strings.TrimPrefix(file, usr.HomeDir)
	}
	return filepath.Join(base, file)
}

func setup(t *testing.T) (base, dir string) {
	dir = t.TempDir()
	base = filepath.Join(dir, "dots")
	if err := os.Mkdir(base, 0755); err != nil {
		t.Fatal(err)
	}
	return base, dir
}

func TestMakeLinkTwiceKeepsContent(t *testing.T) {
	base, dir := setup(t)
	file := filepath.Join(dir, "rc")
	if err := os.WriteFile(file, []byte("precious"), 0644); err != nil {
		t.Fatal(err)
	}

	l, err := MakeLink(base, file)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := l.Update(nil); err != nil {
		t.Fatal(err)
	}

	if _, err := MakeLink(base, file); err == nil {
		t.Fatal("expected error when linking a symlink")
	}

	c, err := os.ReadFile(l.From)
	if err != nil || string(c) != "precious" {
		t.Fatalf("content = %q, %v", c, err)
	}
}

func TestMakeLinkRefusesExistingDestination(t *testing.T) {
	base, dir := setup(t)
	file := filepath.Join(dir, "rc")
	if err := os.WriteFile(file, []byte("local"), 0644); err != nil {
		t.Fatal(err)
	}
	dest := repoPath(t, base, file)
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dest, []byte("repo"), 0644); err != nil {
		t.Fatal(err)
	}

	if _, err := MakeLink(base, file); err == nil {
		t.Fatal("expected error")
	}
	c, _ := os.ReadFile(dest)
	if string(c) != "repo" {
		t.Fatalf("repo copy overwritten: %q", c)
	}
}

func TestMakeLinkTrailingSlash(t *testing.T) {
	base, dir := setup(t)
	d := filepath.Join(dir, "mydir")
	if err := os.Mkdir(d, 0755); err != nil {
		t.Fatal(err)
	}

	l, err := MakeLink(base, d+"/")
	if err != nil {
		t.Fatal(err)
	}
	if l.To != d {
		t.Fatalf("To = %q, want %q", l.To, d)
	}
}

func TestUnlinkMissingTarget(t *testing.T) {
	base, dir := setup(t)
	if err := Unlink(base, filepath.Join(dir, "nope")); err == nil {
		t.Fatal("expected error")
	}
}

func TestUnlinkRoundTrip(t *testing.T) {
	base, dir := setup(t)
	file := filepath.Join(dir, "rc")
	if err := os.WriteFile(file, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	l, err := MakeLink(base, file)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := l.Update(nil); err != nil {
		t.Fatal(err)
	}

	if err := Unlink(base, file); err != nil {
		t.Fatal(err)
	}
	s, err := os.Lstat(file)
	if err != nil || !s.Mode().IsRegular() {
		t.Fatalf("expected regular file back, got %v %v", s, err)
	}
}

func TestUpdateMkdirFailure(t *testing.T) {
	base, dir := setup(t)
	src := filepath.Join(base, "src")
	if err := os.WriteFile(src, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	blocker := filepath.Join(dir, "blocker")
	if err := os.WriteFile(blocker, nil, 0644); err != nil {
		t.Fatal(err)
	}

	l := Link{From: src, To: filepath.Join(blocker, "sub", "rc")}
	if _, err := l.Update(nil); err == nil {
		t.Fatal("expected error")
	}
}
