package sync

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func Git(wd string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = wd
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	os.Stdout.Sync()
	os.Stderr.Sync()
	return err
}

func gitv(wd string, args ...string) error {
	fmt.Println("git", strings.Join(args, " "))
	return Git(wd, args...)
}

func gitOutput(wd string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = wd
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	return strings.TrimSpace(string(out)), err
}
