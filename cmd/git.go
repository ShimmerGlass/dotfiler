package cmd

import (
	"github.com/aestek/dotfiler/sync"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(gitCmd)
}

var gitCmd = &cobra.Command{
	Use:     "git",
	Short:   "Run a git command in dotfiles directory",
	Example: "dotfiler git status",
	Run: func(cmd *cobra.Command, args []string) {
		ensureWorkdir()

		sync.Git(basePath(), args...)
	},
}
