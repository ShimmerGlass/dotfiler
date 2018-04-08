package cmd

import (
	"github.com/aestek/dotfiler/link"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(unLinkCmd)
}

var unLinkCmd = &cobra.Command{
	Use:  "unlink all|<file or directory>",
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := getConfig()

		for _, file := range args {
			must(df.Unlink(basePath(), file))
			cfg.RemoveLink(file)
		}

		writeConfig(cfg)
	},
}
