package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aestek/dotfiler/link"
)

func TestWritePreservesRawAndComments(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "dotfiler.yaml")
	localPath := filepath.Join(dir, "local.yaml")

	orig := `# my header

- name: default
  # keep me
  links:
  - .vimrc:~/.vimrc # vim
  - T:git/$USER:~/.gitconfig
  - gone:~/gone
`
	local := "# local comment\nexclude: [x]\n"
	if err := os.WriteFile(cfgPath, []byte(orig), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(localPath, []byte(local), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Parse(cfgPath, localPath)
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.Excluded("x") {
		t.Fatal("local config not loaded")
	}

	cfg.RemoveLink(cfg.Groups[0].Links[2].To)
	cfg.AddLink("default", &link.Link{From: filepath.Join(dir, "new"), To: "/etc/new"})
	cfg.AddLink("other", &link.Link{From: filepath.Join(dir, "o"), To: "/etc/o"})

	if err := Write(cfg, cfgPath); err != nil {
		t.Fatal(err)
	}

	b, _ := os.ReadFile(cfgPath)
	out := string(b)
	for _, want := range []string{
		"# my header",
		"# keep me",
		".vimrc:~/.vimrc # vim",
		"T:git/$USER:~/.gitconfig",
		"./new:/etc/new",
		"name: other",
		"./o:/etc/o",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
	if strings.Contains(out, "gone") {
		t.Errorf("removed link still present:\n%s", out)
	}

	lb, _ := os.ReadFile(localPath)
	if string(lb) != local {
		t.Errorf("local config rewritten: %q", lb)
	}

	cfg2, err := Parse(cfgPath, localPath)
	if err != nil {
		t.Fatal(err)
	}
	if cfg2.LinkCount() != 4 {
		t.Fatalf("LinkCount = %d", cfg2.LinkCount())
	}
}

func TestWriteDefaultFileWithNullLinks(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "dotfiler.yaml")
	if err := os.WriteFile(cfgPath, []byte("# hdr\n- name: default\n  links:\n"), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Parse(cfgPath, filepath.Join(dir, "missing.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	cfg.AddLink("default", &link.Link{From: filepath.Join(dir, "a"), To: "/etc/a"})
	if err := Write(cfg, cfgPath); err != nil {
		t.Fatal(err)
	}

	cfg2, err := Parse(cfgPath, filepath.Join(dir, "missing.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if cfg2.LinkCount() != 1 || cfg2.Groups[0].Links[0].To != "/etc/a" {
		b, _ := os.ReadFile(cfgPath)
		t.Fatalf("unexpected config:\n%s", b)
	}
}

func TestLocalLoadedWithoutProjectConfig(t *testing.T) {
	dir := t.TempDir()
	localPath := filepath.Join(dir, "local.yaml")
	if err := os.WriteFile(localPath, []byte("exclude: [x]\n"), 0644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Parse(filepath.Join(dir, "missing.yaml"), localPath)
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.Excluded("x") {
		t.Fatal("local config ignored")
	}
}
