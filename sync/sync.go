package sync

import (
	"fmt"
	"os/exec"
)

func Sync(path string, commitMsg string) error {
	err := gitv(path, "add", "-A")
	if err != nil {
		return err
	}

	changes, err := gitOutput(path, "status", "--porcelain")
	if err != nil {
		return err
	}
	if changes != "" {
		err = gitv(path, "commit", "-m", commitMsg)
		if err != nil {
			return err
		}
	} else {
		fmt.Println("nothing to commit")
	}

	branch, err := gitOutput(path, "symbolic-ref", "--short", "HEAD")
	if err != nil {
		return err
	}

	if remoteHasBranch(path, branch) {
		err = gitv(path, "pull", "--commit", "origin", branch)
		if err != nil {
			return err
		}
	}
	return gitv(path, "push", "origin", branch)
}

func remoteHasBranch(path, branch string) bool {
	cmd := exec.Command("git", "ls-remote", "--exit-code", "--heads", "origin", branch)
	cmd.Dir = path
	return cmd.Run() == nil
}
