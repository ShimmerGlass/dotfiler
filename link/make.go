package link

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	dpath "github.com/aestek/dotfiler/path"
)

func MakeLink(base, file string) (*Link, error) {
	if file == "" {
		return nil, fmt.Errorf("no file given")
	}

	file = filepath.Clean(file)

	if dpath.HasDirPrefix(file, base) {
		return nil, fmt.Errorf("%s is inside %s", file, base)
	}

	s, err := os.Lstat(file)
	if err != nil {
		return nil, err
	}
	if s.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("%s is a symbolic link", file)
	}

	usr, err := user.Current()
	if err != nil {
		return nil, err
	}

	rel := file
	if dpath.HasDirPrefix(file, usr.HomeDir) {
		rel = strings.TrimPrefix(file, usr.HomeDir)
	}
	from := filepath.Join(base, rel)

	_, err = os.Lstat(from)
	if err == nil {
		return nil, fmt.Errorf("%s already exists", from)
	}
	if !os.IsNotExist(err) {
		return nil, err
	}

	err = os.MkdirAll(filepath.Dir(from), 0755)
	if err != nil {
		return nil, err
	}

	err = os.Rename(file, from)
	if err != nil {
		return nil, err
	}

	link := &Link{
		From: from,
		To:   file,
	}

	return link, nil
}
