package sync

import (
	"os/exec"
)

func SetRemote(path string, remote string) error {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	cmd.Dir = path
	if cmd.Run() == nil {
		return gitv(path, "remote", "set-url", "origin", remote)
	}
	return gitv(path, "remote", "add", "origin", remote)
}

func Init(path string) error {
	return gitv(path, "init")
}
