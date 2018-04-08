package cmd

import (
	"io/ioutil"
	"path/filepath"

	"github.com/aestek/dotfiler/cmd/defaults"
	"github.com/aestek/dotfiler/sync"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Init dotfiler",
	Run: func(cmd *cobra.Command, args []string) {
		base := basePath()

		if workdirExists() {
			fail("workir %s already exists", base)
		}

		for name, f := range defaults.Files {
			must(ioutil.WriteFile(filepath.Join(base, name), f, 0644))
			must(sync.Init(basePath()))
		}
	},
}
