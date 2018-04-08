package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/aestek/dotfiler/link"
	"github.com/aestek/dotfiler/path"
	"github.com/fatih/color"
)

type LinkWithStatus struct {
	*df.Link
	Status df.LinkStatus
}

type statusPrinter struct {
	links []LinkWithStatus
}

func (p *statusPrinter) Add(link *df.Link, status df.LinkStatus) {
	p.links = append(p.links, LinkWithStatus{Link: link, Status: status})
}

func (p *statusPrinter) Print(name string) {
	if name != "" {
		fmt.Printf("%s\n%s\n", name, strings.Repeat("-", len(name)))
	}

	sml, tml := 0, 0
	sources, targets := make([]string, len(p.links)), make([]string, len(p.links))
	for i, l := range p.links {
		sources[i] = path.Simple(basePath(), l.Source())
		targets[i] = path.Simple("/", l.To)
		if len(sources[i]) > sml {
			sml = len(sources[i])
		}
		if len(targets[i]) > tml {
			tml = len(targets[i])
		}
	}
	for i := 0; i < len(p.links); i++ {
		fmt.Print("* ")
		if p.links[i].Templated {
			fmt.Print("T ")
		} else {
			fmt.Print("  ")
		}
		color.New(fileColor(p.links[i].To)).Print(targets[i])
		fmt.Print(strings.Repeat(" ", tml-len(targets[i])))
		fmt.Print(" -> ")
		color.New(fileColor(p.links[i].From)).Print(sources[i])
		fmt.Print(strings.Repeat(" ", sml-len(sources[i])))
		fmt.Print(" : ")
		fmt.Println(formatStatus(p.links[i].Status))
	}

	fmt.Println()
}

func fileColor(path string) color.Attribute {
	s, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return color.FgHiRed
	}
	must(err)

	switch {
	case s.Mode()&os.ModeSymlink > 0:
		target, err := os.Readlink(path)
		must(err)
		_, err = os.Lstat(target)
		if os.IsNotExist(err) {
			return color.FgHiRed
		}
		must(err)
		return color.FgHiCyan

	case s.Mode()&os.ModeDir > 0:
		return color.FgHiBlue
	default:
		return color.FgWhite
	}
}

func formatStatus(s df.LinkStatus) string {
	switch s {
	case df.LinkStatusLinked:
		return color.GreenString("Linked")
	case df.LinkStatusSourceMiss:
		return color.RedString("Source miss")
	case df.LinkStatusTargetExists:
		return color.RedString("Target exists")
	case df.LinkStatusUnlinked:
		return color.YellowString("Unlinked")
	}
	return ""
}

func must(err error) {
	if err == nil {
		return
	}

	color.Red(err.Error() + "\n")
	os.Exit(1)
}

func fail(f string, s ...interface{}) {
	must(fmt.Errorf(f, s...))
}
