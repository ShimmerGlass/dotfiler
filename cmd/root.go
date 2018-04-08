package cmd

import (
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:  "dotfiler",
	Long: "Manage dotfiles",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}
