package install

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

type Manager string

const (
	Pacman Manager = "pacman"
	Brew   Manager = "brew"
)

var ErrNoManager = errors.New("no supported package manager")

type Runner interface {
	Output(name string, args ...string) ([]byte, int, error)
	Run(name string, args ...string) error
}

type execRunner struct{}

func (execRunner) Output(name string, args ...string) ([]byte, int, error) {
	cmd := exec.Command(name, args...)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	err := cmd.Run()
	if err == nil {
		return stdout.Bytes(), 0, nil
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return stdout.Bytes(), exitErr.ExitCode(), nil
	}
	return nil, -1, err
}

func (execRunner) Run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

var (
	DefaultRunner Runner = execRunner{}
	goos                 = runtime.GOOS
	lookPath             = exec.LookPath
)

func Detect() (Manager, error) {
	switch goos {
	case "darwin":
		return Brew, nil
	case "linux":
		if _, err := lookPath("pacman"); err == nil {
			return Pacman, nil
		}
	}
	return "", ErrNoManager
}

func Names(m Manager, pkgs []Package) []string {
	seen := map[string]bool{}
	var res []string
	for _, p := range pkgs {
		n := p.For(m)
		if seen[n] {
			continue
		}
		seen[n] = true
		res = append(res, n)
	}
	return res
}

func Missing(r Runner, m Manager, names []string) ([]string, error) {
	if len(names) == 0 {
		return nil, nil
	}
	switch m {
	case Pacman:
		out, code, err := r.Output("pacman", append([]string{"-T"}, names...)...)
		if err != nil {
			return nil, fmt.Errorf("pacman -T: %w", err)
		}
		if code != 0 && code != 127 {
			return nil, fmt.Errorf("pacman -T: exit status %d", code)
		}
		return fields(out), nil
	case Brew:
		installed := map[string]bool{}
		for _, kind := range []string{"--formula", "--cask"} {
			out, code, err := r.Output("brew", "list", "-1", kind)
			if err != nil {
				return nil, fmt.Errorf("brew list: %w", err)
			}
			if code != 0 {
				return nil, fmt.Errorf("brew list %s: exit status %d", kind, code)
			}
			for _, n := range fields(out) {
				installed[n] = true
			}
		}
		var missing []string
		for _, n := range names {
			if !installed[n] {
				missing = append(missing, n)
			}
		}
		return missing, nil
	}
	return nil, ErrNoManager
}

func fields(out []byte) []string {
	var res []string
	for _, l := range strings.Split(string(out), "\n") {
		if l = strings.TrimSpace(l); l != "" {
			res = append(res, l)
		}
	}
	return res
}

func Install(r Runner, m Manager, names []string) error {
	if len(names) == 0 {
		return nil
	}
	switch m {
	case Pacman:
		return r.Run("sudo", append([]string{"pacman", "-S", "--needed"}, names...)...)
	case Brew:
		return r.Run("brew", append([]string{"install"}, names...)...)
	}
	return ErrNoManager
}

func InstallTool(r Runner, t Tool) error {
	return r.Run("go", "install", t.Spec())
}
