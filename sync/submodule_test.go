package sync

import (
	"path/filepath"
	"reflect"
	"testing"
)

func commitAll(t *testing.T, dir, msg string) string {
	t.Helper()
	run(t, dir, "add", "-A")
	run(t, dir, "commit", "-m", msg)
	return run(t, dir, "rev-parse", "HEAD")
}

func superWithSubmodule(t *testing.T) (super, sub string) {
	t.Helper()
	sub = filepath.Join(t.TempDir(), "sub")
	run(t, t.TempDir(), "init", sub)
	write(t, filepath.Join(sub, "f"), "v1")
	commitAll(t, sub, "v1")

	super = newRepo(t, "")
	run(t, super, "submodule", "add", sub, "mod")
	commitAll(t, super, "add submodule")
	return super, sub
}

func bumpSubmodule(t *testing.T, super, sub string) string {
	t.Helper()
	write(t, filepath.Join(sub, "f"), "v2")
	sha := commitAll(t, sub, "v2")
	run(t, filepath.Join(super, "mod"), "pull", "origin", "main")
	commitAll(t, super, "bump submodule")
	return sha
}

func TestOutdatedSubmodulesNoGitmodules(t *testing.T) {
	isolate(t)
	dir := newRepo(t, "")
	got, err := OutdatedSubmodules(dir)
	if err != nil || got != nil {
		t.Fatalf("got %v, %v", got, err)
	}
	if err := UpdateSubmodules(dir); err != nil {
		t.Fatal(err)
	}
}

func TestOutdatedSubmodulesUpToDate(t *testing.T) {
	isolate(t)
	super, _ := superWithSubmodule(t)
	got, err := OutdatedSubmodules(super)
	if err != nil || len(got) != 0 {
		t.Fatalf("got %v, %v", got, err)
	}
}

func TestUpdateSubmodulesRefChanged(t *testing.T) {
	isolate(t)
	super, sub := superWithSubmodule(t)

	clone := filepath.Join(t.TempDir(), "clone")
	run(t, t.TempDir(), "clone", "--recurse-submodules", super, clone)

	want := bumpSubmodule(t, super, sub)
	run(t, clone, "pull", "origin", "main")

	got, err := OutdatedSubmodules(clone)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, []string{"mod"}) {
		t.Fatalf("outdated = %v, want [mod]", got)
	}

	if err := UpdateSubmodules(clone); err != nil {
		t.Fatal(err)
	}
	if head := run(t, filepath.Join(clone, "mod"), "rev-parse", "HEAD"); head != want {
		t.Fatalf("submodule HEAD = %s, want %s", head, want)
	}
	if got, err := OutdatedSubmodules(clone); err != nil || len(got) != 0 {
		t.Fatalf("still outdated: %v, %v", got, err)
	}
}

func TestUpdateSubmodulesInitializesMissing(t *testing.T) {
	isolate(t)
	super, _ := superWithSubmodule(t)

	clone := filepath.Join(t.TempDir(), "clone")
	run(t, t.TempDir(), "clone", super, clone)

	got, err := OutdatedSubmodules(clone)
	if err != nil || !reflect.DeepEqual(got, []string{"mod"}) {
		t.Fatalf("outdated = %v, %v", got, err)
	}
	if err := UpdateSubmodules(clone); err != nil {
		t.Fatal(err)
	}
	if head, want := run(t, filepath.Join(clone, "mod"), "rev-parse", "HEAD"), run(t, filepath.Join(super, "mod"), "rev-parse", "HEAD"); head != want {
		t.Fatalf("submodule HEAD = %s, want %s", head, want)
	}
}
