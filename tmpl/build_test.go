package tmpl

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuildFailureKeepsPreviousOutput(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "rc")
	dest := src + Ext

	if err := os.WriteFile(src, []byte("v={{.v}}"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := Build(src, dest, map[string]string{"v": "1"}); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(src, []byte("{{.v.missing.deeper}}"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := Build(src, dest, map[string]string{"v": "1"}); err == nil {
		t.Fatal("expected error")
	}

	c, _ := os.ReadFile(dest)
	if string(c) != "v=1" {
		t.Fatalf("dest = %q", c)
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 2 {
		t.Fatalf("leftover temp files: %v", entries)
	}
}

func TestBuildPreservesMode(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "script")
	if err := os.WriteFile(src, []byte("#!/bin/sh"), 0750); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(src, 0750); err != nil {
		t.Fatal(err)
	}

	dest, err := Build(src, src+Ext, nil)
	if err != nil {
		t.Fatal(err)
	}
	s, _ := os.Stat(dest)
	if s.Mode().Perm() != 0750 {
		t.Fatalf("mode = %v", s.Mode().Perm())
	}
}

func TestBuildDirRemovesStaleSubdir(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "d")
	dest := src + Ext
	if err := os.MkdirAll(filepath.Join(dest, "stale", "inner"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(src, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "f"), []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}

	if _, err := Build(src, dest, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dest, "stale")); !os.IsNotExist(err) {
		t.Fatalf("stale dir not removed: %v", err)
	}
}

func buildOne(t *testing.T, content string, vars interface{}) string {
	t.Helper()
	dir := t.TempDir()
	src := filepath.Join(dir, "rc")
	if err := os.WriteFile(src, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	dest, err := Build(src, src+Ext, vars)
	if err != nil {
		t.Fatal(err)
	}
	if dest != src+Ext {
		t.Fatalf("dest = %q", dest)
	}
	c, err := os.ReadFile(dest)
	if err != nil {
		t.Fatal(err)
	}
	return string(c)
}

func TestBuildStringMapVars(t *testing.T) {
	if got := buildOne(t, "hi {{.name}}", map[string]string{"name": "bob"}); got != "hi bob" {
		t.Fatalf("got %q", got)
	}
}

func TestBuildInterfaceMapVars(t *testing.T) {
	if got := buildOne(t, "hi {{.name}}", map[interface{}]interface{}{"name": "bob"}); got != "hi bob" {
		t.Fatalf("got %q", got)
	}
}

func TestBuildReadFile(t *testing.T) {
	dir := t.TempDir()
	secret := filepath.Join(dir, "secret")
	if err := os.WriteFile(secret, []byte("  s3cret\n\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if got := buildOne(t, `[{{read_file "`+secret+`"}}]`, nil); got != "[s3cret]" {
		t.Fatalf("got %q", got)
	}
	if got := buildOne(t, `[{{read_file "`+filepath.Join(dir, "missing")+`"}}]`, nil); got != "[]" {
		t.Fatalf("got %q", got)
	}
}

func TestBuildParseError(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "rc")
	dest := src + Ext
	if err := os.WriteFile(src, []byte("{{.a"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := Build(src, dest, nil); err == nil {
		t.Fatal("expected error")
	}
	if _, err := os.Stat(dest); !os.IsNotExist(err) {
		t.Fatalf("dest created: %v", err)
	}
}

func TestBuildMissingSource(t *testing.T) {
	dir := t.TempDir()
	if _, err := Build(filepath.Join(dir, "nope"), filepath.Join(dir, "out"), nil); err == nil {
		t.Fatal("expected error")
	}
}

func TestBuildDirRecursive(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "d")
	dest := src + Ext
	if err := os.MkdirAll(filepath.Join(src, "a", "b"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "top"), []byte("{{.v}}-top"), 0644); err != nil {
		t.Fatal(err)
	}
	deep := filepath.Join(src, "a", "b", "run")
	if err := os.WriteFile(deep, []byte("{{.v}}-deep"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(deep, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(dest, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dest, "stale"), []byte("old"), 0644); err != nil {
		t.Fatal(err)
	}

	got, err := Build(src, dest, map[string]string{"v": "1"})
	if err != nil || got != dest {
		t.Fatalf("got %q, %v", got, err)
	}
	if c, _ := os.ReadFile(filepath.Join(dest, "top")); string(c) != "1-top" {
		t.Fatalf("top = %q", c)
	}
	out := filepath.Join(dest, "a", "b", "run")
	if c, _ := os.ReadFile(out); string(c) != "1-deep" {
		t.Fatalf("deep = %q", c)
	}
	s, err := os.Stat(out)
	if err != nil || s.Mode().Perm() != 0700 {
		t.Fatalf("mode = %v, %v", s, err)
	}
	if _, err := os.Stat(filepath.Join(dest, "stale")); !os.IsNotExist(err) {
		t.Fatalf("stale file not removed: %v", err)
	}
}

func TestBuildOverwritesExistingDest(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "rc")
	dest := src + Ext
	if err := os.WriteFile(src, []byte("new"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dest, []byte("old"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := Build(src, dest, nil); err != nil {
		t.Fatal(err)
	}
	if c, _ := os.ReadFile(dest); string(c) != "new" {
		t.Fatalf("dest = %q", c)
	}
}
