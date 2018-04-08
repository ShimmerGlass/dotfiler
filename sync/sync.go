package sync

func Sync(path string, commitMsg string) error {
	err := gitv(path, "add", "-A")
	if err != nil {
		return err
	}
	err = gitv(path, "commit", "-m", commitMsg)
	if err != nil {
		return err
	}
	err = gitv(path, "pull", "--commit", "origin", "master")
	if err != nil {
		return err
	}
	err = gitv(path, "push", "origin", "master")
	if err != nil {
		return err
	}

	return nil
}
