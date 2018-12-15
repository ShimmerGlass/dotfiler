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

func Build(source string, vars interface{}) (string, error) {
	s, err := os.Stat(source)
	if err != nil {
		return "", err
	}
	if s.Mode()&os.ModeDir > 0 {
		files, err := ioutil.ReadDir(source)
		if err != nil {
			return "", err
		}

		for _, f := range files {
			p := filepath.Join(source, f.Name())
			_, err := Build(p, vars)
			if err != nil {
				return "", err
			}
		}
		return source, nil
	}

	contents, err := ioutil.ReadFile(source)
	if err != nil {
		return "", errors.Wrapf(err, "build %s", source)
	}

	dest := source + Ext

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
