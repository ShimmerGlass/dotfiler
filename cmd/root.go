package cmd

import (
	"github.com/spf13/cobra"
)

var customDirectory string

func init() {
	RootCmd.PersistentFlags().StringVarP(&customDirectory, "dir", "D", basePath(), "dotfiler directory")
}

// RootCmd to be ran
var RootCmd = &cobra.Command{
	Use:  "dotfiler",
	Long: "Manage dotfiles",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}
