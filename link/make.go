package link

import (
	"fmt"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"strings"
)

func MakeLink(base, file string) (*Link, error) {
	if file == "" {
		return nil, fmt.Errorf("no file given")
	}

	if file[len(file)-1] == '/' {
		file = file[:len(file)-2]
	}

	usr, err := user.Current()
	if err != nil {
		return nil, err
	}

	from := path.Join(base, strings.TrimPrefix(file, usr.HomeDir))
	fromDir := filepath.Dir(from)
	fmt.Println(fromDir)
	err = os.MkdirAll(fromDir, 0755)
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
