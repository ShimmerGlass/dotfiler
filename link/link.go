package df

import (
	"fmt"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"strings"

	"github.com/aestek/dotfiler/tmpl"
	"github.com/pkg/errors"
)

type LinkStatus int

const (
	LinkStatusLinked LinkStatus = iota
	LinkStatusUnlinked
	LinkStatusTargetExists
	LinkStatusSourceMiss
)

type Link struct {
	From      string
	To        string
	Templated bool
}

func (l Link) Source() string {
	if !l.Templated {
		return l.From
	}
	s, err := os.Stat(l.From)
	if err != nil {
		return l.From
	}
	if s.Mode()&os.ModeDir > 0 {
		return l.From
	}
	return l.From + tmpl.Ext
}

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

func (l Link) String() string {
	return fmt.Sprintf("<%s -> %s>", l.From, l.To)
}

func (l Link) Status() (LinkStatus, error) {
	s, err := os.Stat(l.From)
	if os.IsNotExist(err) {
		return LinkStatusSourceMiss, nil
	}
	s, err = os.Lstat(l.To)
	if os.IsNotExist(err) {
		return LinkStatusUnlinked, nil
	}
	if err != nil {
		return LinkStatusUnlinked, errors.Wrapf(err, "link %s status", l)
	}
	if s.Mode()&os.ModeSymlink == 0 {
		return LinkStatusTargetExists, nil
	}
	target, err := os.Readlink(l.To)
	if err != nil {
		return LinkStatusUnlinked, errors.Wrapf(err, "link %s status", l)
	}
	if l.Source() != target {
		return LinkStatusTargetExists, nil
	}
	return LinkStatusLinked, nil
}

func (l Link) Ensure(vars interface{}) (LinkStatus, error) {
	status, err := l.Status()
	if err != nil {
		return status, err
	}

	if status == LinkStatusSourceMiss {
		return status, nil
	}

	if status == LinkStatusTargetExists {
		s, err := os.Lstat(l.To)
		if err != nil {
			return status, errors.Wrapf(err, "link %s ensure", l)
		}

		if s.Mode()&os.ModeSymlink > 0 {
			fmt.Println("rm", l.To)
			err := os.Remove(l.To)
			if err != nil {
				return status, errors.Wrapf(err, "link %s ensure", l)
			}
		}
	}

	source := l.From
	if l.Templated {
		name, err := tmpl.Build(l.From, vars)
		if err != nil {
			return status, err
		}
		source = name
	}

	if status == LinkStatusLinked {
		return status, nil
	}

	fmt.Println("ln", "-s", source, l.To)
	err = os.Symlink(source, l.To)
	if err != nil {
		return LinkStatusUnlinked, errors.Wrapf(err, "link %s ensure", l)
	}

	return LinkStatusLinked, nil
}

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

	fmt.Println("rm", file)
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

	fmt.Println("mv", target, file)
	return os.Rename(target, file)
}
