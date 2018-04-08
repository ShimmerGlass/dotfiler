package config

import (
	"fmt"
	"strings"

	"github.com/aestek/dotfiler/link"
	"github.com/aestek/dotfiler/path"
)

func parseLink(base, l string) (*link.Link, error) {
	link := &link.Link{}

	parts := strings.Split(l, ":")
	switch len(parts) {
	case 3:
		link.Templated = strings.Contains(parts[0], "T")
		link.From = parts[1]
		link.To = parts[2]
	case 2:
		link.From = parts[0]
		link.To = parts[1]
	default:
		return nil, fmt.Errorf("invalid link `%s`, expected [flags]:source:dest")
	}

	link.From = path.Expand(base, link.From)
	link.To = path.Expand(base, link.To)

	return link, nil
}

func linkString(base string, l *link.Link) string {
	flags := ""
	if l.Templated {
		flags += "T:"
	}
	return fmt.Sprintf("%s%s:%s",
		flags,
		path.Simple(base, l.From),
		path.Simple(base, l.To),
	)
}
