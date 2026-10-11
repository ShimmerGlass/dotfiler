package cmd

import (
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(editCmd)
}

func editorCommand() []string {
	for _, env := range []string{"VISUAL", "EDITOR"} {
		if f := strings.Fields(os.Getenv(env)); len(f) > 0 {
			return f
		}
	}
	if _, err := exec.LookPath("editor"); err == nil {
		return []string{"editor"}
	}
	return []string{"vi"}
}

var editCmd = &cobra.Command{
	Use:   "edit [file]",
	Short: "Edit dotfiler configuration.",
	Long: `Edit dotfiler configuration.

[file] can be :
* main (~/.dotfiles/dotfiler.yaml) assumed if not given
* local (~/.dotfiles/dotfiler_local.yaml)
* a file ~/.dotfiles/
`,
	Example: `dotfiler edit
dotfiler edit local
dotfiler edit .gitconfig
`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		file := configPath()
		if len(args) > 0 {
			switch args[0] {
			case "main":
			case "local":
				file = localConfigPath()

			default:
				file = args[0]
			}
		}

		editor := editorCommand()
		c := exec.Command(editor[0], append(editor[1:], file)...)
		c.Dir = basePath()
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		must(c.Run())
	},
}
