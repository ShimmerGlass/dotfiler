package cmd

import (
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(editCmd)
}

var editCmd = &cobra.Command{
	Use:   "edit [file]",
	Short: "Edit dotfiler configuration.",
	Long: `Edit dotfiler configuration.

[file] can be main (~/.dofiles/dotfiler.yaml) or local (~/.dotfiles/dotfiler_local.yaml). main assumed if not given.
`,
	Example: "dotfiler git status",
	Args:    cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		file := configPath()
		if len(args) > 0 {
			switch args[0] {
			case "main":
				// noop
			case "local":
				file = localConfigPath()

			default:
				fail("Unknown configuration file %s, expected \"main\" or \"local\"", args[0])
			}
		}

		c := exec.Command("editor", file)
		c.Dir = basePath()
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		c.Run()
	},
}
