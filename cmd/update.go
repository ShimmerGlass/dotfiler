package cmd

import (
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(updateCmd)
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update links based on configuration files",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := getConfig()

		for _, group := range cfg.Groups {
			if cfg.Excluded(group.Name) {
				continue
			}

			p := &statusPrinter{}
			for _, link := range group.Links {
				status, err := link.Ensure(cfg.Vars)
				if err != nil {
					color.Red(err.Error())
					continue
				}

				p.Add(link, status)
			}
			p.Print(group.Name)
		}
	},
}
