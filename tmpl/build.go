package tmpl

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
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
				err := os.RemoveAll(p)
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
	tmpl.Funcs(template.FuncMap{
		"read_file": func(path string) string {
			c, err := os.ReadFile(path)
			if err != nil {
				return ""
			}
			return strings.TrimSpace(string(c))
		},
	})
	tmpl, err = tmpl.Parse(string(contents))
	if err != nil {
		return "", errors.Wrapf(err, "build %s", source)
	}

	out, err := os.CreateTemp(filepath.Dir(dest), "."+filepath.Base(dest)+".tmp*")
	if err != nil {
		return "", errors.Wrapf(err, "build %s", source)
	}
	defer os.Remove(out.Name())

	err = tmpl.Execute(out, vars)
	if err != nil {
		out.Close()
		return "", errors.Wrapf(err, "build %s", source)
	}

	err = out.Chmod(s.Mode().Perm())
	if err != nil {
		out.Close()
		return "", errors.Wrapf(err, "build %s", source)
	}

	err = out.Close()
	if err != nil {
		return "", errors.Wrapf(err, "build %s", source)
	}

	err = os.Rename(out.Name(), dest)
	if err != nil {
		return "", errors.Wrapf(err, "build %s", source)
	}

	return dest, nil
}
