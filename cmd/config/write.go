package config

import (
	"bytes"
	"fmt"
	"io/ioutil"

	"github.com/aestek/dotfiler/cmd/defaults"
	"github.com/aestek/dotfiler/link"

	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v3"
)

// Write stores the project config content on disk at given path
func Write(c *Config, cfgPath string) error {
	doc := c.doc
	header := false
	if doc == nil || len(doc.Content) == 0 || doc.Content[0].Kind != yaml.SequenceNode {
		doc = &yaml.Node{
			Kind:    yaml.DocumentNode,
			Content: []*yaml.Node{{Kind: yaml.SequenceNode}},
		}
		header = true
	}

	root := doc.Content[0]
	for _, g := range c.Groups {
		c.syncGroup(root, g)
	}

	buf := &bytes.Buffer{}
	if header {
		buf.WriteString(defaults.ProjectCfgHeader)
	}
	enc := yaml.NewEncoder(buf)
	enc.SetIndent(2)
	err := enc.Encode(doc)
	if err != nil {
		return errors.Wrap(err, "config write")
	}
	err = enc.Close()
	if err != nil {
		return errors.Wrap(err, "config write")
	}

	fmt.Println("write config to", cfgPath)
	err = ioutil.WriteFile(cfgPath, buf.Bytes(), 0664)
	if err != nil {
		return errors.Wrap(err, "config write")
	}

	c.doc = doc
	return nil
}

func (c *Config) linkRaw(l *link.Link) string {
	if r, ok := c.raw[l]; ok {
		return r
	}
	return linkString(c.Workdir, l)
}

func (c *Config) syncGroup(root *yaml.Node, g *Group) {
	groupNode := findGroupNode(root, g.Name)
	if groupNode == nil {
		groupNode = &yaml.Node{
			Kind: yaml.MappingNode,
			Content: []*yaml.Node{
				{Kind: yaml.ScalarNode, Value: "name"},
				{Kind: yaml.ScalarNode, Value: g.Name},
			},
		}
		root.Content = append(root.Content, groupNode)
	}

	linksNode := mappingValue(groupNode, "links")
	if linksNode == nil {
		linksNode = &yaml.Node{}
		groupNode.Content = append(groupNode.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Value: "links"},
			linksNode,
		)
	}
	if linksNode.Kind != yaml.SequenceNode {
		*linksNode = yaml.Node{Kind: yaml.SequenceNode, HeadComment: linksNode.HeadComment, LineComment: linksNode.LineComment}
	}

	existing := map[string][]*yaml.Node{}
	for _, n := range linksNode.Content {
		existing[n.Value] = append(existing[n.Value], n)
	}

	content := make([]*yaml.Node, 0, len(g.Links))
	for _, l := range g.Links {
		r := c.linkRaw(l)
		if nodes := existing[r]; len(nodes) > 0 {
			content = append(content, nodes[0])
			existing[r] = nodes[1:]
			continue
		}
		content = append(content, &yaml.Node{Kind: yaml.ScalarNode, Value: r})
	}
	linksNode.Content = content
}

func findGroupNode(root *yaml.Node, name string) *yaml.Node {
	for _, n := range root.Content {
		if n.Kind != yaml.MappingNode {
			continue
		}
		if v := mappingValue(n, "name"); v != nil && v.Value == name {
			return n
		}
	}
	return nil
}

func mappingValue(m *yaml.Node, key string) *yaml.Node {
	for i := 0; i+1 < len(m.Content); i += 2 {
		if m.Content[i].Value == key {
			return m.Content[i+1]
		}
	}
	return nil
}

func (c *Config) getGroup(name string) *Group {
	for _, g := range c.Groups {
		if g.Name == name {
			return g
		}
	}

	g := &Group{
		Name: name,
	}

	c.Groups = append(c.Groups, g)

	return g
}
