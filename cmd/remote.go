package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(remoteCmd)
}

var remoteCmd = &cobra.Command{
	Use:   "remote",
	Short: "Init dotfiler",
	Run: func(cmd *cobra.Command, args []string) {
	},
}
