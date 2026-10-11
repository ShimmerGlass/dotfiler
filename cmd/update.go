package cmd

import (
	"fmt"
	"strings"

	"github.com/aestek/dotfiler/cmd/config"
	"github.com/aestek/dotfiler/install"
	"github.com/aestek/dotfiler/sync"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(updateCmd)
}

var updateCmd = &cobra.Command{
	Use:     "update",
	Short:   "Update links based on configuration files",
	Aliases: []string{"up"},
	Run: func(cmd *cobra.Command, args []string) {
		cfg := getConfig()
		must(sync.UpdateSubmodules(cfg.Workdir))

		failed := installRequirements(cfg)

		for _, group := range cfg.Groups {
			if cfg.Excluded(group.Name) {
				continue
			}

			p := &statusPrinter{}
			for _, link := range group.Links {
				status, err := link.Update(cfg.Vars)
				if err != nil {
					color.Red(err.Error())
					continue
				}

				p.Add(link, status)
			}
			p.Print(group.Name)
		}

		if failed {
			fail("some packages or tools failed to install")
		}
	},
}

func installRequirements(cfg *config.Config) bool {
	var pkgs []install.Package
	var tools []install.Tool
	for _, group := range cfg.Groups {
		if cfg.Excluded(group.Name) {
			continue
		}
		pkgs = append(pkgs, group.Packages...)
		tools = append(tools, group.Tools...)
	}

	failed := false
	r := install.DefaultRunner

	if len(pkgs) > 0 {
		if err := installPackages(r, pkgs); err != nil {
			color.Red(err.Error())
			failed = true
		}
	}

	for _, t := range tools {
		fmt.Printf("go install %s\n", t.Spec())
		if err := install.InstallTool(r, t); err != nil {
			color.Red("install tool %s: %s", t.Go, err)
			failed = true
		}
	}

	return failed
}

func installPackages(r install.Runner, pkgs []install.Package) error {
	m, err := install.Detect()
	if err != nil {
		return err
	}
	missing, err := install.Missing(r, m, install.Names(m, pkgs))
	if err != nil {
		return err
	}
	if len(missing) == 0 {
		fmt.Println("All packages are installed")
		return nil
	}
	fmt.Printf("Installing packages with %s: %s\n", m, strings.Join(missing, " "))
	if err := install.Install(r, m, missing); err != nil {
		return fmt.Errorf("install packages: %w", err)
	}
	return nil
}
