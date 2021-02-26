package tmpl

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"

	"github.com/aestek/dotfiler/path"
	"github.com/pkg/errors"
)

const Ext = ".cpld"

func Build(source, dest string, vars interface{}) (string, error) {
	s, err := os.Stat(source)
	if err != nil {
		return "", err
	}
	if s.Mode()&os.ModeDir > 0 {
		err := os.MkdirAll(dest, 0o755)
		if err != nil {
			return "", err
		}

		files, err := ioutil.ReadDir(source)
		if err != nil {
			return "", err
		}

		existing := map[string]bool{}

		for _, f := range files {
			_, err := Build(
				filepath.Join(source, f.Name()),
				filepath.Join(dest, f.Name()),
				vars,
			)
			if err != nil {
				return "", err
			}
			existing[f.Name()] = true
		}

		files, err = ioutil.ReadDir(dest)
		if err != nil {
			return "", err
		}

		for _, f := range files {
			if !existing[f.Name()] {
				p := filepath.Join(dest, f.Name())
				fmt.Println("Removing", path.Simple("/", p))
				err := os.Remove(p)
				if err != nil {
					return "", err
				}
			}
		}

		return dest, nil
	}

	contents, err := ioutil.ReadFile(source)
	if err != nil {
		return "", errors.Wrapf(err, "build %s", source)
	}

	fmt.Println("Build", path.Simple("/", source), "into", path.Simple("/", dest))

	tmpl := template.New("t")
	tmpl, err = tmpl.Parse(string(contents))
	if err != nil {
		return "", errors.Wrapf(err, "build %s", source)
	}

	out, err := os.OpenFile(dest, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		return "", errors.Wrapf(err, "build %s", source)
	}

	err = tmpl.Execute(out, vars)
	if err != nil {
		return "", errors.Wrapf(err, "build %s", source)
	}

	err = out.Close()
	if err != nil {
		return "", errors.Wrapf(err, "build %s", source)
	}

	return dest, nil
}
