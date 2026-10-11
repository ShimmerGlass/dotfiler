package link

import (
	"fmt"
	"os"
	"path/filepath"

	dpath "github.com/aestek/dotfiler/path"
)

func Unlink(base, file string) error {
	s, err := os.Lstat(file)
	if os.IsNotExist(err) {
		return fmt.Errorf("%s does not exist", file)
	}
	if err != nil {
		return err
	}

	if s.Mode()&os.ModeSymlink == 0 {
		return fmt.Errorf("%s is not a symbolic link", file)
	}

	target, err := os.Readlink(file)
	if err != nil {
		return err
	}
	if !filepath.IsAbs(target) {
		target = filepath.Join(filepath.Dir(file), target)
	}

	if !dpath.HasDirPrefix(target, base) {
		return fmt.Errorf("%s is not in %s", target, base)
	}

	fmt.Println("Removing symlink", file)
	err = os.Remove(file)
	if err != nil {
		return err
	}

	fmt.Println("Moving", target, "to", file)
	return os.Rename(target, file)
}
