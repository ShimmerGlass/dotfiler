package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aestek/dotfiler/cmd/defaults"
	"github.com/aestek/dotfiler/link"
)

func TestWriteFreshConfig(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "dotfiler.yaml")
	missing := filepath.Join(dir, "missing.yaml")

	cfg, err := Parse(p, missing)
	if err != nil {
		t.Fatal(err)
	}
	cfg.AddLink("default", &link.Link{From: filepath.Join(dir, "a"), To: "/etc/a"})
	cfg.AddLink("default", &link.Link{From: filepath.Join(dir, "t"), To: "/etc/t", Templated: true})
	if err := Write(cfg, p); err != nil {
		t.Fatal(err)
	}

	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	out := string(b)
	if !strings.HasPrefix(out, defaults.ProjectCfgHeader) {
		t.Errorf("missing header:\n%s", out)
	}
	if !strings.Contains(out, "T:./t:/etc/t") {
		t.Errorf("templated link not written with T flag:\n%s", out)
	}

	cfg2, err := Parse(p, missing)
	if err != nil {
		t.Fatal(err)
	}
	got := cfg2.Links()
	if len(got) != 2 {
		t.Fatalf("LinkCount = %d\n%s", len(got), out)
	}
	for i, want := range cfg.Links() {
		if *got[i] != *want {
			t.Errorf("link %d = %+v, want %+v", i, *got[i], *want)
		}
	}
}

func TestWriteEmptyFileGetsHeader(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "dotfiler.yaml")
	writeFile(t, p, "")
	cfg, err := Parse(p, filepath.Join(dir, "missing.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	cfg.AddLink("g", &link.Link{From: filepath.Join(dir, "a"), To: "/etc/a"})
	if err := Write(cfg, p); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(p)
	if !strings.HasPrefix(string(b), defaults.ProjectCfgHeader) {
		t.Errorf("missing header:\n%s", b)
	}
}

func TestWriteReplacesNonSequenceLinks(t *testing.T) {
	for name, links := range map[string]string{
		"null": "links:\n",
	} {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			p := filepath.Join(dir, "dotfiler.yaml")
			missing := filepath.Join(dir, "missing.yaml")
			writeFile(t, p, "- name: g\n  "+links)
			cfg, err := Parse(p, missing)
			if err != nil {
				t.Fatal(err)
			}
			cfg.AddLink("g", &link.Link{From: filepath.Join(dir, "a"), To: "/etc/a"})
			if err := Write(cfg, p); err != nil {
				t.Fatal(err)
			}
			cfg2, err := Parse(p, missing)
			if err != nil {
				b, _ := os.ReadFile(p)
				t.Fatalf("%v\n%s", err, b)
			}
			if cfg2.LinkCount() != 1 || cfg2.Links()[0].To != "/etc/a" {
				b, _ := os.ReadFile(p)
				t.Errorf("unexpected config:\n%s", b)
			}
		})
	}
}

func TestWriteDuplicateRawEntries(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "dotfiler.yaml")
	missing := filepath.Join(dir, "missing.yaml")
	writeFile(t, p, "- name: g\n  links:\n  - a:/etc/x\n  - a:/etc/x\n")

	cfg, err := Parse(p, missing)
	if err != nil {
		t.Fatal(err)
	}
	if err := Write(cfg, p); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(p)
	if n := strings.Count(string(b), "a:/etc/x"); n != 2 {
		t.Errorf("expected both entries preserved, got %d:\n%s", n, b)
	}

	cfg.RemoveLink("/etc/x")
	if err := Write(cfg, p); err != nil {
		t.Fatal(err)
	}
	b, _ = os.ReadFile(p)
	if n := strings.Count(string(b), "a:/etc/x"); n != 1 {
		t.Errorf("expected one entry after removal, got %d:\n%s", n, b)
	}
}
