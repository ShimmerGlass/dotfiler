package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
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

		must(os.Mkdir(base, 0755))

		for name, f := range defaults.Files {
			path := filepath.Join(base, name)
			fmt.Println("write", path)
			must(ioutil.WriteFile(path, f, 0644))
		}

		must(sync.Init(basePath()))
	},
}
