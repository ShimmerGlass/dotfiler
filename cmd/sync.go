package cmd

import (
	"strings"

	"github.com/aestek/dotfiler/sync"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(syncCmd)
}

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync dotfiles with remote repository",
	Run: func(cmd *cobra.Command, args []string) {
		ensureWorkdir()

		msg := "Sync"
		if len(args) > 0 {
			msg = strings.Join(args, " ")
		}

		must(sync.Sync(basePath(), msg))
	},
}
