package install

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

type call struct {
	name string
	args []string
}

type fakeRunner struct {
	outputs map[string]fakeOut
	calls   []call
	runErr  error
}

type fakeOut struct {
	out  string
	code int
	err  error
}

func (f *fakeRunner) Output(name string, args ...string) ([]byte, int, error) {
	f.calls = append(f.calls, call{name, args})
	o := f.outputs[name+" "+strings.Join(args, " ")]
	return []byte(o.out), o.code, o.err
}

func (f *fakeRunner) Run(name string, args ...string) error {
	f.calls = append(f.calls, call{name, args})
	return f.runErr
}

func TestPacmanMissing(t *testing.T) {
	f := &fakeRunner{outputs: map[string]fakeOut{
		"pacman -T a b c": {out: "b\nc\n", code: 127},
	}}
	got, err := Missing(f, Pacman, []string{"a", "b", "c"})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, []string{"b", "c"}) {
		t.Errorf("missing = %v", got)
	}
}

func TestPacmanNoneMissing(t *testing.T) {
	f := &fakeRunner{outputs: map[string]fakeOut{"pacman -T a": {}}}
	got, err := Missing(f, Pacman, []string{"a"})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Errorf("missing = %v", got)
	}
}

func TestPacmanUnexpectedExit(t *testing.T) {
	f := &fakeRunner{outputs: map[string]fakeOut{"pacman -T a": {code: 1}}}
	if _, err := Missing(f, Pacman, []string{"a"}); err == nil {
		t.Error("expected error")
	}
}

func TestPacmanRunError(t *testing.T) {
	f := &fakeRunner{outputs: map[string]fakeOut{"pacman -T a": {err: errors.New("boom")}}}
	if _, err := Missing(f, Pacman, []string{"a"}); err == nil {
		t.Error("expected error")
	}
}

func TestBrewMissing(t *testing.T) {
	f := &fakeRunner{outputs: map[string]fakeOut{
		"brew list -1 --formula": {out: "a\nb\n"},
		"brew list -1 --cask":    {out: "c\n"},
	}}
	got, err := Missing(f, Brew, []string{"a", "c", "d", "e"})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, []string{"d", "e"}) {
		t.Errorf("missing = %v", got)
	}
}

func TestMissingEmptyNoCalls(t *testing.T) {
	f := &fakeRunner{}
	got, err := Missing(f, Pacman, nil)
	if err != nil || len(got) != 0 || len(f.calls) != 0 {
		t.Errorf("got %v, %v, calls %v", got, err, f.calls)
	}
}

func TestInstallCommands(t *testing.T) {
	oldLook := lookPath
	defer func() { lookPath = oldLook }()

	has := func(bin string) func(string) (string, error) {
		return func(name string) (string, error) {
			if name == bin {
				return "/usr/bin/" + name, nil
			}
			return "", errors.New("not found")
		}
	}

	cases := []struct {
		m    Manager
		look func(string) (string, error)
		want call
	}{
		{Pacman, has("none"), call{"sudo", []string{"pacman", "-S", "--needed", "a", "b"}}},
		{Pacman, has("trizen"), call{"trizen", []string{"-S", "--needed", "a", "b"}}},
		{Brew, has("trizen"), call{"brew", []string{"install", "a", "b"}}},
	}
	for _, c := range cases {
		lookPath = c.look
		f := &fakeRunner{}
		if err := Install(f, c.m, []string{"a", "b"}); err != nil {
			t.Fatal(err)
		}
		if len(f.calls) != 1 || !reflect.DeepEqual(f.calls[0], c.want) {
			t.Errorf("%s: calls = %v, want %v", c.m, f.calls, c.want)
		}
	}
}

func TestInstallEmptyNoCalls(t *testing.T) {
	f := &fakeRunner{}
	if err := Install(f, Pacman, nil); err != nil || len(f.calls) != 0 {
		t.Errorf("err %v calls %v", err, f.calls)
	}
}

func TestInstallError(t *testing.T) {
	f := &fakeRunner{runErr: errors.New("boom")}
	if err := Install(f, Brew, []string{"a"}); err == nil {
		t.Error("expected error")
	}
}

func TestNamesOverrides(t *testing.T) {
	pkgs := []Package{
		{Name: "zsh"},
		{Name: "fd", Pacman: "fd-p", Brew: "fd-b"},
		{Name: "zsh"},
		{Name: "git", Brew: "git-b"},
	}
	if got := Names(Pacman, pkgs); !reflect.DeepEqual(got, []string{"zsh", "fd-p", "git"}) {
		t.Errorf("pacman names = %v", got)
	}
	if got := Names(Brew, pkgs); !reflect.DeepEqual(got, []string{"zsh", "fd-b", "git-b"}) {
		t.Errorf("brew names = %v", got)
	}
}

func TestDetect(t *testing.T) {
	oldOS, oldLook := goos, lookPath
	defer func() { goos, lookPath = oldOS, oldLook }()

	found := func(string) (string, error) { return "/usr/bin/pacman", nil }
	notFound := func(string) (string, error) { return "", errors.New("not found") }

	cases := []struct {
		os      string
		look    func(string) (string, error)
		want    Manager
		wantErr bool
	}{
		{"darwin", notFound, Brew, false},
		{"linux", found, Pacman, false},
		{"linux", notFound, "", true},
		{"windows", found, "", true},
	}
	for _, c := range cases {
		goos, lookPath = c.os, c.look
		got, err := Detect()
		if got != c.want || (err != nil) != c.wantErr {
			t.Errorf("%s: got %q, %v", c.os, got, err)
		}
		if c.wantErr && !errors.Is(err, ErrNoManager) {
			t.Errorf("%s: err = %v", c.os, err)
		}
	}
}

func TestToolSpec(t *testing.T) {
	if got := (Tool{Go: "github.com/a/b"}).Spec(); got != "github.com/a/b@latest" {
		t.Errorf("spec = %q", got)
	}
	if got := (Tool{Go: "github.com/a/b@v1.2.3"}).Spec(); got != "github.com/a/b@v1.2.3" {
		t.Errorf("spec = %q", got)
	}
}

func TestInstallTool(t *testing.T) {
	f := &fakeRunner{}
	if err := InstallTool(f, Tool{Go: "github.com/a/b"}); err != nil {
		t.Fatal(err)
	}
	want := call{"go", []string{"install", "github.com/a/b@latest"}}
	if len(f.calls) != 1 || !reflect.DeepEqual(f.calls[0], want) {
		t.Errorf("calls = %v", f.calls)
	}
}
