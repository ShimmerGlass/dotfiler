package cmd

import (
	"reflect"
	"testing"
)

func TestExtractDirFlag(t *testing.T) {
	saved := customDirectory
	defer func() { customDirectory = saved }()

	cases := []struct {
		args    []string
		dir     string
		remains []string
	}{
		{[]string{"-D", "/x", "log", "--oneline"}, "/x", []string{"log", "--oneline"}},
		{[]string{"--dir=/y", "status"}, "/y", []string{"status"}},
		{[]string{"-D/z", "diff"}, "/z", []string{"diff"}},
		{[]string{"--dir", "/x", "status"}, "/x", []string{"status"}},
		{[]string{"-D"}, "unchanged", []string{"-D"}},
		{[]string{"--dir"}, "unchanged", []string{"--dir"}},
		{[]string{"-D", "/a", "--dir=/b", "log"}, "/b", []string{"log"}},
		{[]string{"commit", "-m", "-D"}, "unchanged", []string{"commit", "-m", "-D"}},
	}
	for _, c := range cases {
		customDirectory = "unchanged"
		got := extractDirFlag(c.args)
		if !reflect.DeepEqual(got, c.remains) || customDirectory != c.dir {
			t.Errorf("extractDirFlag(%v) = %v dir=%q, want %v dir=%q", c.args, got, customDirectory, c.remains, c.dir)
		}
	}
}
