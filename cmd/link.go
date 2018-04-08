package cmd

import (
	"path/filepath"

	"github.com/aestek/dotfiler/link"
	"github.com/spf13/cobra"
)

var groupName string

func init() {
	linkCmd.Flags().StringVarP(&groupName, "group", "g", "default", "Name of the group to add the link to (see cfg.yaml)")
	RootCmd.AddCommand(linkCmd)
}

var linkCmd = &cobra.Command{
	Use:     "link <file or directory>",
	Short:   "Link a file with dotfiler",
	Long:    "Moves <file> to the dotfiles directory, and creates a symlink to the new location",
	Example: "dotfiler link ~/.vimrc",
	Args:    cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := getConfig()

		for _, file := range args {
			var err error
			if !filepath.IsAbs(file) {
				file, err = filepath.Abs(file)
				must(err)
			}
			link, err := link.MakeLink(cfg.Workdir, file)
			must(err)
			cfg.AddLink(groupName, link)
			writeConfig(cfg)

			status, err := link.Update(cfg.Vars)
			must(err)

			p := &statusPrinter{}
			p.Add(link, status)
			p.Print("")
		}
	},
}
