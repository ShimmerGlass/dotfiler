package link

import (
	"fmt"
	"os"
	"strings"

	"github.com/aestek/dotfiler/tmpl"
)

func Unlink(base, file string) error {
	s, err := os.Lstat(file)
	exists := true
	if os.IsNotExist(err) {
		exists = false
	} else if err != nil {
		return err
	}

	if s.Mode()&os.ModeSymlink == 0 {
		return fmt.Errorf("%s is not a symbolic link", file)
	}

	target, err := os.Readlink(file)
	if err != nil {
		return err
	}

	if !strings.HasPrefix(target, base) {
		return fmt.Errorf("%s is not in %s", target, base)
	}

	fmt.Println("Removing symlink", file)
	err = os.Remove(file)
	if err != nil {
		return err
	}

	if !exists {
		return nil
	}

	if s.Mode()&os.ModeDir > 0 {
		err := tmpl.Clean(file)
		if err != nil {
			return err
		}
	}

	fmt.Println("Moving", target, "to", file)
	return os.Rename(target, file)
}
