package main

import (
	"os"

	"github.com/aestek/dotfiler/cmd"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
