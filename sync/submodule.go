package sync

import (
	"os"
	"path/filepath"
	"strings"
)

func OutdatedSubmodules(path string) ([]string, error) {
	_, err := os.Stat(filepath.Join(path, ".gitmodules"))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	out, err := gitOutput(path, "submodule", "status")
	if err != nil {
		return nil, err
	}

	var res []string
	for _, line := range strings.Split(out, "\n") {
		if line == "" || (line[0] != '+' && line[0] != '-') {
			continue
		}
		fields := strings.Fields(line[1:])
		if len(fields) < 2 {
			continue
		}
		res = append(res, fields[1])
	}
	return res, nil
}

func UpdateSubmodules(path string) error {
	outdated, err := OutdatedSubmodules(path)
	if err != nil || len(outdated) == 0 {
		return err
	}
	return gitv(path, append([]string{"submodule", "update", "--init", "--recursive", "--"}, outdated...)...)
}
