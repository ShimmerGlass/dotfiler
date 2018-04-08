package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(statusCmd)
}

var statusCmd = &cobra.Command{
	Use:     "status",
	Aliases: []string{"st"},
	Short:   "Report links status",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := getConfig()

		if cfg.LinkCount() == 0 {
			fmt.Println("No links confgured, you can add ones by running the link command :")
			fmt.Println(linkCmd.Help())
			os.Exit(0)
		}

		for _, group := range cfg.Groups {
			p := &statusPrinter{}

			for _, link := range group.Links {
				status, err := link.Status()
				if err != nil {
					fmt.Println(link, err)
					continue
				}

				p.Add(link, status)
			}

			title := group.Name
			if cfg.Excluded(group.Name) {
				title += " (Excluded)"
			}

			p.Print(title)
		}

	},
}
