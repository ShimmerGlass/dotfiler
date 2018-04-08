package tmpl

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func Clean(dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, Ext) {
			fmt.Println("rm", path)
			return os.Remove(path)
		}
		return nil
	})
}
