package sync

import (
	"os"
)

func Inited(path string) (bool, error) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func Init(path string) error {
	err := gitv(path, "init")
	if err != nil {
		return err
	}
	err = gitv(path, "add", "-A")
	if err != nil {
		return err
	}
	err = gitv(path, "commit", "-m", "Init")
	if err != nil {
		return err
	}

	return nil
}
