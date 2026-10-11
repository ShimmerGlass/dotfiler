package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aestek/dotfiler/link"
	"github.com/fatih/color"
)

func TestEditorCommand(t *testing.T) {
	cases := []struct {
		name           string
		visual, editor string
		want           []string
	}{
		{"visual wins", "nano", "vim", []string{"nano"}},
		{"editor only", "", "vim", []string{"vim"}},
		{"editor with args", "", "code -w", []string{"code", "-w"}},
		{"visual with args", "code -w --new", "vim", []string{"code", "-w", "--new"}},
		{"blank visual falls through", "   ", "vim", []string{"vim"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Setenv("VISUAL", c.visual)
			t.Setenv("EDITOR", c.editor)
			got := editorCommand()
			if strings.Join(got, "\x00") != strings.Join(c.want, "\x00") {
				t.Errorf("got %q, want %q", got, c.want)
			}
		})
	}

	t.Run("none set", func(t *testing.T) {
		t.Setenv("VISUAL", "")
		t.Setenv("EDITOR", "")
		got := editorCommand()
		if len(got) == 0 || got[0] == "" {
			t.Fatalf("got %q, want non-empty", got)
		}
	})
}

func TestFormatStatus(t *testing.T) {
	cases := []struct {
		status link.LinkStatus
		want   string
	}{
		{link.LinkStatusLinked, "Linked"},
		{link.LinkStatusSourceMiss, "Source miss"},
		{link.LinkStatusTargetExists, "Target exists"},
		{link.LinkStatusUnlinked, "Unlinked"},
	}
	for _, c := range cases {
		if got := formatStatus(c.status); !strings.Contains(got, c.want) {
			t.Errorf("formatStatus(%d) = %q, want it to contain %q", c.status, got, c.want)
		}
	}
	if got := formatStatus(link.LinkStatus(99)); got != "" {
		t.Errorf("formatStatus(unknown) = %q, want empty", got)
	}
}

func TestFileColor(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "file")
	sub := filepath.Join(dir, "sub")
	good := filepath.Join(dir, "good")
	dangling := filepath.Join(dir, "dangling")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(file, good); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(dir, "nowhere"), dangling); err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name string
		path string
		want color.Attribute
	}{
		{"missing", filepath.Join(dir, "missing"), color.FgHiRed},
		{"file", file, color.FgWhite},
		{"dir", sub, color.FgHiBlue},
		{"symlink", good, color.FgHiCyan},
		{"dangling", dangling, color.FgHiRed},
	}
	for _, c := range cases {
		if got := fileColor(c.path); got != c.want {
			t.Errorf("%s: fileColor = %v, want %v", c.name, got, c.want)
		}
	}
}

func TestPaths(t *testing.T) {
	saved := customDirectory
	defer func() { customDirectory = saved }()

	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	customDirectory = filepath.Join("rel", "dots")
	want := filepath.Join(wd, "rel", "dots")
	if got := basePath(); got != want {
		t.Errorf("basePath = %q, want %q", got, want)
	}
	if got := configPath(); got != filepath.Join(want, "dotfiler.yaml") {
		t.Errorf("configPath = %q", got)
	}
	if got := localConfigPath(); got != filepath.Join(want, "dotfiler_local.yaml") {
		t.Errorf("localConfigPath = %q", got)
	}

	abs := t.TempDir()
	customDirectory = abs
	if got := basePath(); got != abs {
		t.Errorf("basePath = %q, want %q", got, abs)
	}
}

func TestWorkdirExists(t *testing.T) {
	saved := customDirectory
	defer func() { customDirectory = saved }()

	dir := t.TempDir()
	customDirectory = filepath.Join(dir, "missing")
	if workdirExists() {
		t.Error("workdirExists true for missing dir")
	}
	customDirectory = dir
	if !workdirExists() {
		t.Error("workdirExists false for existing dir")
	}
}
