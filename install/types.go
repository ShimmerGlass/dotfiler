package install

import (
	"fmt"
	"strings"

	yaml "gopkg.in/yaml.v3"
)

type Package struct {
	Name   string `yaml:"name"`
	Pacman string `yaml:"pacman"`
	Brew   string `yaml:"brew"`
}

func (p *Package) UnmarshalYAML(n *yaml.Node) error {
	switch n.Kind {
	case yaml.ScalarNode:
		var s string
		if err := n.Decode(&s); err != nil {
			return err
		}
		if s == "" {
			return fmt.Errorf("line %d: empty package name", n.Line)
		}
		*p = Package{Name: s}
		return nil
	case yaml.MappingNode:
		type plain Package
		var v plain
		if err := n.Decode(&v); err != nil {
			return err
		}
		if v.Name == "" {
			return fmt.Errorf("line %d: package requires a name", n.Line)
		}
		*p = Package(v)
		return nil
	}
	return fmt.Errorf("line %d: package must be a string or a mapping", n.Line)
}

func (p Package) For(m Manager) string {
	switch m {
	case Pacman:
		if p.Pacman != "" {
			return p.Pacman
		}
	case Brew:
		if p.Brew != "" {
			return p.Brew
		}
	}
	return p.Name
}

type Tool struct {
	Go string `yaml:"go"`
}

func (t *Tool) UnmarshalYAML(n *yaml.Node) error {
	if n.Kind != yaml.MappingNode {
		return fmt.Errorf("line %d: tool must be a mapping like {go: <module>}", n.Line)
	}
	type plain Tool
	var v plain
	if err := n.Decode(&v); err != nil {
		return err
	}
	if strings.TrimSpace(v.Go) == "" {
		return fmt.Errorf("line %d: tool requires a non-empty go module", n.Line)
	}
	*t = Tool(v)
	return nil
}

func (t Tool) Spec() string {
	if strings.Contains(t.Go, "@") {
		return t.Go
	}
	return t.Go + "@latest"
}
