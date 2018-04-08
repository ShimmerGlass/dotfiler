package cmd

import (
	"github.com/aestek/dotfiler/cmd/config"
	"github.com/aestek/dotfiler/link"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(unLinkCmd)
}

func unlink(cfg *config.Config, file string) {
	must(link.Unlink(basePath(), file))
	cfg.RemoveLink(file)
}

var unLinkCmd = &cobra.Command{
	Use:  "unlink all|<file or directory>",
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := getConfig()
		defer writeConfig(cfg)

		if args[0] == "all" {
			for _, l := range cfg.Links() {
				unlink(cfg, l.To)
			}
			return
		}

		for _, file := range args {
			unlink(cfg, file)
		}
	},
}
