package cmd

import (
	"path/filepath"

	"github.com/aestek/dotfiler/cmd/config"
	"github.com/aestek/dotfiler/link"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(unLinkCmd)
}

func unlink(cfg *config.Config, file string) bool {
	err := link.Unlink(basePath(), file)
	if err != nil {
		color.Red(err.Error())
		return false
	}
	cfg.RemoveLink(file)
	return true
}

var unLinkCmd = &cobra.Command{
	Use:  "unlink all|<file or directory>",
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := getConfig()
		ok := true

		if args[0] == "all" {
			for _, l := range cfg.Links() {
				status, err := l.Status()
				if err != nil {
					color.Red(err.Error())
					ok = false
					continue
				}
				if status != link.LinkStatusLinked {
					color.Yellow("Skipping %s: not linked", l)
					continue
				}
				ok = unlink(cfg, l.To) && ok
			}
		} else {
			for _, file := range args {
				abs, err := filepath.Abs(file)
				if err != nil {
					color.Red(err.Error())
					ok = false
					continue
				}
				ok = unlink(cfg, abs) && ok
			}
		}

		writeConfig(cfg)
		if !ok {
			fail("some links could not be removed")
		}
	},
}
