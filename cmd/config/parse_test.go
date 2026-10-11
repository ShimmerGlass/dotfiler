package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aestek/dotfiler/cmd/defaults"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestParseMissingFiles(t *testing.T) {
	dir := t.TempDir()
	cfg, err := Parse(filepath.Join(dir, "p.yaml"), filepath.Join(dir, "l.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Groups) != 0 || cfg.LinkCount() != 0 {
		t.Errorf("expected empty config, got %+v", cfg.Groups)
	}
	if cfg.Workdir != dir {
		t.Errorf("Workdir = %q, want %q", cfg.Workdir, dir)
	}
}

func TestParseWorkdirFromRelativePath(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	cfg, err := Parse("p.yaml", "l.yaml")
	if err != nil {
		t.Fatal(err)
	}
	want, err := filepath.EvalSymlinks(dir)
	if err != nil {
		t.Fatal(err)
	}
	got, err := filepath.EvalSymlinks(cfg.Workdir)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Errorf("Workdir = %q, want %q", got, want)
	}
}

func TestParseEmptyProjectFile(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "p.yaml")
	writeFile(t, p, "")
	cfg, err := Parse(p, filepath.Join(dir, "l.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Groups) != 0 {
		t.Errorf("expected no groups, got %+v", cfg.Groups)
	}
}

func TestParseGroups(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "p.yaml")
	writeFile(t, p, "- name: a\n  links:\n  - x:/etc/x\n  - T:y:/etc/y\n- name: b\n  links:\n  - z:/etc/z\n")
	cfg, err := Parse(p, filepath.Join(dir, "l.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Groups) != 2 || cfg.Groups[0].Name != "a" || cfg.Groups[1].Name != "b" {
		t.Fatalf("unexpected groups: %+v", cfg.Groups)
	}
	l := cfg.Groups[0].Links
	if len(l) != 2 || l[0].From != filepath.Join(dir, "x") || l[0].To != "/etc/x" || l[0].Templated {
		t.Errorf("unexpected first link: %+v", l[0])
	}
	if !l[1].Templated || l[1].From != filepath.Join(dir, "y") {
		t.Errorf("unexpected templated link: %+v", l[1])
	}
}

func TestParseErrors(t *testing.T) {
	cases := []struct {
		name, project, local string
	}{
		{"malformed yaml", "- name: [\n", ""},
		{"root not a sequence", "name: x\nlinks: []\n", ""},
		{"bad link", "- name: a\n  links:\n  - justone\n", ""},
		{"too many parts", "- name: a\n  links:\n  - a:b:c:d\n", ""},
		{"malformed local", "- name: a\n", "exclude: [\n"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dir := t.TempDir()
			p := filepath.Join(dir, "p.yaml")
			l := filepath.Join(dir, "l.yaml")
			writeFile(t, p, c.project)
			if c.local != "" {
				writeFile(t, l, c.local)
			}
			if _, err := Parse(p, l); err == nil {
				t.Error("expected error")
			}
		})
	}
}

func TestParseLocalVars(t *testing.T) {
	dir := t.TempDir()
	l := filepath.Join(dir, "l.yaml")
	writeFile(t, l, "exclude: [i3]\nvars:\n  email: foo@bar.com\n")
	cfg, err := Parse(filepath.Join(dir, "p.yaml"), l)
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.Excluded("i3") {
		t.Error("exclude not loaded")
	}
	vars, ok := cfg.Vars.(map[interface{}]interface{})
	if !ok {
		t.Fatalf("Vars has type %T", cfg.Vars)
	}
	if vars["email"] != "foo@bar.com" {
		t.Errorf("vars = %v", vars)
	}
}

func TestParseDefaultFiles(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "dotfiler.yaml")
	l := filepath.Join(dir, "dotfiler_local.yaml")
	writeFile(t, p, string(defaults.Files["dotfiler.yaml"]))
	writeFile(t, l, string(defaults.Files["dotfiler_local.yaml"]))
	cfg, err := Parse(p, l)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Groups) != 1 || cfg.Groups[0].Name != "default" || cfg.LinkCount() != 0 {
		t.Errorf("unexpected default config: %+v", cfg.Groups)
	}
	if len(cfg.Exclude) != 0 {
		t.Errorf("unexpected exclude: %v", cfg.Exclude)
	}
}
