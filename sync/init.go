package sync

func SetRemote(path string, remote string) error {
	return gitv(path, "remote", "add", "origin", remote)
}

func Init(path string) error {
	return gitv(path, "init")
}
