package cmd

import (
	"errors"
	"os"
	"os/exec"
	"strings"

	"github.com/aestek/dotfiler/sync"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(gitCmd)
}

func extractDirFlag(args []string) []string {
	for len(args) > 0 {
		a := args[0]
		switch {
		case (a == "-D" || a == "--dir") && len(args) > 1:
			customDirectory = args[1]
			args = args[2:]
		case strings.HasPrefix(a, "--dir="):
			customDirectory = strings.TrimPrefix(a, "--dir=")
			args = args[1:]
		case strings.HasPrefix(a, "-D") && len(a) > 2:
			customDirectory = strings.TrimPrefix(strings.TrimPrefix(a, "-D"), "=")
			args = args[1:]
		default:
			return args
		}
	}
	return args
}

var gitCmd = &cobra.Command{
	Use:                "git",
	Short:              "Run a git command in dotfiles directory",
	Example:            "dotfiler git status",
	DisableFlagParsing: true,
	Run: func(cmd *cobra.Command, args []string) {
		args = extractDirFlag(args)
		ensureWorkdir()

		err := sync.Git(basePath(), args...)
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			os.Exit(exitErr.ExitCode())
		}
		must(err)
	},
}
