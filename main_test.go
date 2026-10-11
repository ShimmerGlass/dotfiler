package main

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const e2eEnv = "DOTFILER_E2E_MAIN"

func TestMain(m *testing.M) {
	if os.Getenv(e2eEnv) == "1" {
		main()
		os.Exit(0)
	}
	os.Exit(m.Run())
}

type env struct {
	t    *testing.T
	home string
	w    string
	vars []string
}

func newEnv(t *testing.T) *env {
	t.Helper()
	root := t.TempDir()
	home := filepath.Join(root, "home")
	if err := os.MkdirAll(home, 0o755); err != nil {
		t.Fatal(err)
	}
	return &env{
		t:    t,
		home: home,
		w:    filepath.Join(root, "dots"),
		vars: []string{
			e2eEnv + "=1",
			"HOME=" + home,
			"GIT_CONFIG_GLOBAL=/dev/null",
			"GIT_CONFIG_NOSYSTEM=1",
			"GIT_AUTHOR_NAME=Test",
			"GIT_AUTHOR_EMAIL=test@example.com",
			"GIT_COMMITTER_NAME=Test",
			"GIT_COMMITTER_EMAIL=test@example.com",
			"GIT_CONFIG_COUNT=2",
			"GIT_CONFIG_KEY_0=init.defaultBranch",
			"GIT_CONFIG_VALUE_0=main",
			"GIT_CONFIG_KEY_1=pull.rebase",
			"GIT_CONFIG_VALUE_1=false",
		},
	}
}

func (e *env) runWith(extra []string, args ...string) (string, int) {
	e.t.Helper()
	cmd := exec.Command(os.Args[0], args...)
	cmd.Dir = e.home
	cmd.Env = append(append(os.Environ(), e.vars...), extra...)
	out, err := cmd.CombinedOutput()
	code := 0
	if err != nil {
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) {
			e.t.Fatalf("dotfiler %s: %v", strings.Join(args, " "), err)
		}
		code = exitErr.ExitCode()
	}
	return string(out), code
}

func (e *env) run(args ...string) (string, int) {
	e.t.Helper()
	return e.runWith(nil, append([]string{"-D", e.w}, args...)...)
}

func (e *env) ok(args ...string) string {
	e.t.Helper()
	out, code := e.run(args...)
	if code != 0 {
		e.t.Fatalf("dotfiler %s: exit %d\n%s", strings.Join(args, " "), code, out)
	}
	return out
}

func (e *env) fails(args ...string) string {
	e.t.Helper()
	out, code := e.run(args...)
	if code == 0 {
		e.t.Fatalf("dotfiler %s: expected failure, got exit 0\n%s", strings.Join(args, " "), out)
	}
	return out
}

func (e *env) read(path string) string {
	e.t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		e.t.Fatal(err)
	}
	return string(b)
}

func (e *env) write(path, content string) {
	e.t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		e.t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		e.t.Fatal(err)
	}
}

func (e *env) isSymlinkIntoW(path string) {
	e.t.Helper()
	fi, err := os.Lstat(path)
	if err != nil {
		e.t.Fatal(err)
	}
	if fi.Mode()&os.ModeSymlink == 0 {
		e.t.Fatalf("%s is not a symlink", path)
	}
	target, err := os.Readlink(path)
	if err != nil {
		e.t.Fatal(err)
	}
	if !strings.HasPrefix(target, e.w+string(os.PathSeparator)) {
		e.t.Fatalf("%s -> %s, not inside %s", path, target, e.w)
	}
}

func (e *env) isRegular(path string) {
	e.t.Helper()
	fi, err := os.Lstat(path)
	if err != nil {
		e.t.Fatal(err)
	}
	if !fi.Mode().IsRegular() && !fi.IsDir() {
		e.t.Fatalf("%s is not a regular file or directory: %v", path, fi.Mode())
	}
}

func (e *env) contains(out string, subs ...string) {
	e.t.Helper()
	for _, s := range subs {
		if !strings.Contains(out, s) {
			e.t.Errorf("output missing %q:\n%s", s, out)
		}
	}
}

func TestEndToEnd(t *testing.T) {
	e := newEnv(t)
	cfg := filepath.Join(e.w, "dotfiler.yaml")
	local := filepath.Join(e.w, "dotfiler_local.yaml")
	rc := filepath.Join(e.home, "x", "rc")
	dir := filepath.Join(e.home, "x", "d")
	tplOut := filepath.Join(e.home, "x", "tplout")

	steps := []struct {
		name string
		fn   func(t *testing.T)
	}{
		{"init", func(t *testing.T) {
			e.ok("init")
			for _, n := range []string{"dotfiler.yaml", "dotfiler_local.yaml", ".git"} {
				if _, err := os.Stat(filepath.Join(e.w, n)); err != nil {
					t.Errorf("%s missing: %v", n, err)
				}
			}
			e.fails("init")
		}},
		{"status without links", func(t *testing.T) {
			e.contains(e.ok("status"), "No links")
		}},
		{"link file", func(t *testing.T) {
			e.write(rc, "rc content\n")
			e.ok("link", rc)
			e.isSymlinkIntoW(rc)
			if got := e.read(rc); got != "rc content\n" {
				t.Errorf("content = %q", got)
			}
			y := e.read(cfg)
			e.contains(y, "# Dotfiler configuration file", "name: default", "rc")
		}},
		{"link twice", func(t *testing.T) {
			out := e.fails("link", rc)
			e.contains(out, "already linked")
		}},
		{"link dir in group", func(t *testing.T) {
			e.write(filepath.Join(dir, "f"), "in dir\n")
			e.ok("-g", "other", "link", dir)
			e.isSymlinkIntoW(dir)
			e.contains(e.read(cfg), "name: other")
			if got := e.read(filepath.Join(dir, "f")); got != "in dir\n" {
				t.Errorf("content = %q", got)
			}
		}},
		{"status", func(t *testing.T) {
			out := e.ok("status")
			e.contains(out, "Linked", "default", "other")
		}},
		{"update restores symlink", func(t *testing.T) {
			if err := os.Remove(rc); err != nil {
				t.Fatal(err)
			}
			e.contains(e.ok("status"), "Unlinked")
			e.ok("update")
			e.isSymlinkIntoW(rc)
			if got := e.read(rc); got != "rc content\n" {
				t.Errorf("content = %q", got)
			}
		}},
		{"templated update", func(t *testing.T) {
			e.write(filepath.Join(e.w, "tpl", "rc"), "hello {{.name}}\n")
			y := e.read(cfg)
			y += "- name: tmpl\n  links:\n    - T:tpl/rc:" + tplOut + "\n"
			e.write(cfg, y)
			e.write(local, "exclude: []\nvars:\n  name: Foo\n")
			e.ok("update")
			e.isSymlinkIntoW(tplOut)
			if got := e.read(tplOut); got != "hello Foo\n" {
				t.Errorf("rendered = %q", got)
			}
			e.contains(e.ok("status"), "tmpl")

			e.write(local, "exclude: []\nvars:\n  name: Bar\n")
			e.ok("update")
			if got := e.read(tplOut); got != "hello Bar\n" {
				t.Errorf("re-rendered = %q", got)
			}
		}},
		{"excluded group", func(t *testing.T) {
			e.write(local, "exclude: [\"other\"]\nvars:\n  name: Foo\n")
			if err := os.Remove(dir); err != nil {
				t.Fatal(err)
			}
			e.contains(e.ok("status"), "Excluded")
			e.ok("update")
			if _, err := os.Lstat(dir); !os.IsNotExist(err) {
				t.Fatalf("excluded link recreated: %v", err)
			}
			e.write(local, "exclude: []\nvars:\n  name: Foo\n")
			e.ok("update")
			e.isSymlinkIntoW(dir)
		}},
		{"unlink", func(t *testing.T) {
			e.ok("unlink", rc)
			e.isRegular(rc)
			if got := e.read(rc); got != "rc content\n" {
				t.Errorf("content = %q", got)
			}
			y := e.read(cfg)
			if strings.Contains(y, filepath.Join("x", "rc")) {
				t.Errorf("link still in yaml:\n%s", y)
			}
			e.fails("unlink", rc)
			e.fails("unlink", filepath.Join(e.home, "x", "nothing"))
		}},
		{"unlink all", func(t *testing.T) {
			e.ok("unlink", "all")
			e.isRegular(dir)
			e.isRegular(tplOut)
			if got := e.read(filepath.Join(dir, "f")); got != "in dir\n" {
				t.Errorf("dir content = %q", got)
			}
			if got := e.read(tplOut); got != "hello Foo\n" {
				t.Errorf("tpl content = %q", got)
			}
			e.contains(e.ok("status"), "No links")
		}},
		{"edit", func(t *testing.T) {
			outFile := filepath.Join(e.home, "editor_args")
			script := filepath.Join(e.home, "fake-editor.sh")
			e.write(script, "#!/bin/sh\necho \"$@\" > \"$EDITOR_OUT\"\n")
			if err := os.Chmod(script, 0o755); err != nil {
				t.Fatal(err)
			}
			extra := []string{"VISUAL=", "EDITOR=" + script + " --flag", "EDITOR_OUT=" + outFile}

			if _, code := e.runWith(extra, "-D", e.w, "edit", "local"); code != 0 {
				t.Fatalf("edit local exit %d", code)
			}
			if got := strings.TrimSpace(e.read(outFile)); got != "--flag "+local {
				t.Errorf("editor args = %q, want %q", got, "--flag "+local)
			}

			if _, code := e.runWith(extra, "-D", e.w, "edit"); code != 0 {
				t.Fatalf("edit exit %d", code)
			}
			if got := strings.TrimSpace(e.read(outFile)); got != "--flag "+cfg {
				t.Errorf("editor args = %q, want %q", got, "--flag "+cfg)
			}
		}},
		{"git passthrough", func(t *testing.T) {
			e.ok("git", "status")
			out, code := e.run("git", "not-a-command")
			if code == 0 {
				t.Errorf("git not-a-command: exit 0\n%s", out)
			}
			if _, code := e.runWith(nil, "git", "-D", e.w, "status"); code != 0 {
				t.Errorf("git -D status: exit %d", code)
			}
		}},
		{"remote and sync", func(t *testing.T) {
			bare := filepath.Join(t.TempDir(), "remote.git")
			if out, err := exec.Command("git", "init", "--bare", bare).CombinedOutput(); err != nil {
				t.Fatalf("%v\n%s", err, out)
			}
			e.ok("remote", bare)
			e.ok("sync", "e2e", "commit")

			cmd := exec.Command("git", "log", "--format=%s", "main")
			cmd.Dir = bare
			cmd.Env = append(os.Environ(), e.vars...)
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("%v\n%s", err, out)
			}
			if got := strings.TrimSpace(string(out)); got != "e2e commit" {
				t.Errorf("remote log = %q", got)
			}
		}},
	}

	for _, s := range steps {
		if !t.Run(s.name, s.fn) {
			t.FailNow()
		}
	}
}

func TestNoWorkdir(t *testing.T) {
	e := newEnv(t)
	for _, args := range [][]string{{"status"}, {"update"}, {"sync"}, {"git", "status"}, {"remote", "x"}} {
		if out := e.fails(args...); !strings.Contains(out, "does not exist") {
			t.Errorf("%v: output %q", args, out)
		}
	}
}
