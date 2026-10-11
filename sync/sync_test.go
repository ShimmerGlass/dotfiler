package sync

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func isolate(t *testing.T) {
	t.Helper()
	t.Setenv("GIT_CONFIG_GLOBAL", "/dev/null")
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	t.Setenv("GIT_AUTHOR_NAME", "Test")
	t.Setenv("GIT_AUTHOR_EMAIL", "test@example.com")
	t.Setenv("GIT_COMMITTER_NAME", "Test")
	t.Setenv("GIT_COMMITTER_EMAIL", "test@example.com")
	t.Setenv("GIT_CONFIG_COUNT", "3")
	t.Setenv("GIT_CONFIG_KEY_0", "init.defaultBranch")
	t.Setenv("GIT_CONFIG_VALUE_0", "main")
	t.Setenv("GIT_CONFIG_KEY_1", "pull.rebase")
	t.Setenv("GIT_CONFIG_VALUE_1", "false")
	t.Setenv("GIT_CONFIG_KEY_2", "protocol.file.allow")
	t.Setenv("GIT_CONFIG_VALUE_2", "always")
}

func run(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
	return strings.TrimSpace(string(out))
}

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func newBare(t *testing.T) string {
	t.Helper()
	bare := filepath.Join(t.TempDir(), "remote.git")
	run(t, t.TempDir(), "init", "--bare", bare)
	return bare
}

func newRepo(t *testing.T, bare string) string {
	t.Helper()
	dir := t.TempDir()
	if err := Init(dir); err != nil {
		t.Fatal(err)
	}
	if bare != "" {
		if err := SetRemote(dir, bare); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func TestInit(t *testing.T) {
	isolate(t)
	dir := t.TempDir()
	if err := Init(dir); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, ".git")); err != nil {
		t.Fatalf(".git missing: %v", err)
	}
}

func TestSetRemote(t *testing.T) {
	isolate(t)
	dir := newRepo(t, "")

	if err := SetRemote(dir, "/first"); err != nil {
		t.Fatal(err)
	}
	if got := run(t, dir, "remote", "get-url", "origin"); got != "/first" {
		t.Fatalf("origin = %q, want /first", got)
	}

	if err := SetRemote(dir, "/second"); err != nil {
		t.Fatal(err)
	}
	if got := run(t, dir, "remote", "get-url", "origin"); got != "/second" {
		t.Fatalf("origin = %q, want /second", got)
	}
	if got := run(t, dir, "remote"); got != "origin" {
		t.Fatalf("remotes = %q, want only origin", got)
	}
}

func TestSyncFirstPush(t *testing.T) {
	isolate(t)
	bare := newBare(t)
	dir := newRepo(t, bare)
	write(t, filepath.Join(dir, "a.txt"), "a")

	if err := Sync(dir, "first"); err != nil {
		t.Fatal(err)
	}
	if got := run(t, bare, "log", "--format=%s", "main"); got != "first" {
		t.Fatalf("remote log = %q, want first", got)
	}
}

func TestSyncNoChanges(t *testing.T) {
	isolate(t)
	bare := newBare(t)
	dir := newRepo(t, bare)
	write(t, filepath.Join(dir, "a.txt"), "a")

	if err := Sync(dir, "first"); err != nil {
		t.Fatal(err)
	}
	before := run(t, dir, "rev-parse", "HEAD")
	if err := Sync(dir, "second"); err != nil {
		t.Fatal(err)
	}
	if after := run(t, dir, "rev-parse", "HEAD"); after != before {
		t.Fatalf("HEAD moved from %s to %s", before, after)
	}
	if got := run(t, bare, "rev-list", "--count", "main"); got != "1" {
		t.Fatalf("remote commits = %s, want 1", got)
	}
}

func TestSyncPullsRemoteChanges(t *testing.T) {
	isolate(t)
	bare := newBare(t)
	dir := newRepo(t, bare)
	write(t, filepath.Join(dir, "a.txt"), "a")
	if err := Sync(dir, "first"); err != nil {
		t.Fatal(err)
	}

	other := filepath.Join(t.TempDir(), "other")
	run(t, t.TempDir(), "clone", bare, other)
	write(t, filepath.Join(other, "b.txt"), "b")
	run(t, other, "add", "-A")
	run(t, other, "commit", "-m", "from other")
	run(t, other, "push", "origin", "main")

	write(t, filepath.Join(dir, "c.txt"), "c")
	if err := Sync(dir, "local"); err != nil {
		t.Fatal(err)
	}

	log := run(t, bare, "log", "--format=%s", "main")
	for _, msg := range []string{"first", "from other", "local"} {
		if !strings.Contains(log, msg) {
			t.Errorf("remote log missing %q:\n%s", msg, log)
		}
	}
	for _, f := range []string{"a.txt", "b.txt", "c.txt"} {
		if _, err := os.Stat(filepath.Join(dir, f)); err != nil {
			t.Errorf("local repo missing %s: %v", f, err)
		}
	}
	if local, remote := run(t, dir, "rev-parse", "HEAD"), run(t, bare, "rev-parse", "main"); local != remote {
		t.Errorf("local HEAD %s != remote %s", local, remote)
	}
}

func TestSyncNoRemote(t *testing.T) {
	isolate(t)
	dir := newRepo(t, "")
	write(t, filepath.Join(dir, "a.txt"), "a")
	if err := Sync(dir, "first"); err == nil {
		t.Fatal("expected error without origin")
	}
}

func TestGitExitError(t *testing.T) {
	isolate(t)
	dir := newRepo(t, "")
	err := Git(dir, "definitely-not-a-command")
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("err = %v, want *exec.ExitError", err)
	}
	if exitErr.ExitCode() == 0 {
		t.Fatal("exit code 0")
	}
}

func TestGitSuccess(t *testing.T) {
	isolate(t)
	dir := newRepo(t, "")
	if err := Git(dir, "status"); err != nil {
		t.Fatal(err)
	}
}
