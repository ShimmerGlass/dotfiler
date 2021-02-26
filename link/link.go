package link

import (
	"fmt"
	"os"
	"path"

	dpath "github.com/aestek/dotfiler/path"
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
	return l.From + tmpl.Ext
}

func (l Link) String() string {
	return fmt.Sprintf("<%s -> %s>", dpath.Simple("/", l.From), dpath.Simple("/", l.To))
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
		return LinkStatusUnlinked, nil
	}
	return LinkStatusLinked, nil
}

func (l Link) Update(vars interface{}) (LinkStatus, error) {
	status, err := l.Status()
	if err != nil {
		return status, err
	}

	if status == LinkStatusSourceMiss {
		return status, nil
	}

	if status == LinkStatusTargetExists || status == LinkStatusUnlinked {
		s, err := os.Lstat(l.To)
		if os.IsNotExist(err) {
			goto Install
		}
		if err != nil {
			return status, errors.Wrapf(err, "link %s update", l)
		}

		if s.Mode()&os.ModeSymlink > 0 {
			fmt.Println("Removing symlink", l.To)
			err := os.Remove(l.To)
			if err != nil {
				return status, errors.Wrapf(err, "link %s update", l)
			}
		}
	}

Install:

	source := l.From
	if l.Templated {
		name, err := tmpl.Build(l.From, l.From+tmpl.Ext, vars)
		if err != nil {
			return status, err
		}
		source = name
	}

	if status == LinkStatusLinked {
		return status, nil
	}

	dir := path.Dir(l.To)
	_, err = os.Stat(dir)
	if os.IsNotExist(err) {
		fmt.Println("Creating directory", dpath.Simple("/", dir))
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			errors.Wrapf(err, "link %s update", l)
		}
	}

	fmt.Println("Symlinking", dpath.Simple("/", source), "to", dpath.Simple("/", l.To))
	err = os.Symlink(source, l.To)
	if err != nil {
		return LinkStatusUnlinked, errors.Wrapf(err, "link %s update", l)
	}

	return LinkStatusLinked, nil
}
