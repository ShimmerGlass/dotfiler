package config

import (
	"github.com/aestek/dotfiler/link"
)

// Config sotres both local and project config
type Config struct {
	Project
	Local
	Workdir string
}

// Project contains the links configuration of the project
type Project struct {
	Groups []*Group
}

// Group bundles links to be toggled on and off
type Group struct {
	Name  string
	Links []*link.Link
}

// Local contains the settings relative to this system
type Local struct {
	Vars    interface{} `yaml:"vars"`
	Exclude []string    `yaml:"exclude"`
}

// LinkCount returns the number of configured links
func (c *Config) LinkCount() (i int) {
	for _, g := range c.Groups {
		for range g.Links {
			i++
		}
	}
	return
}

// Links is a helper to retreive links from each group
func (c *Config) Links() []*link.Link {
	res := []*link.Link{}
	for _, g := range c.Groups {
		for _, l := range g.Links {
			res = append(res, l)
		}
	}
	return res
}

// AddLink adds a new link.
// The group is created if is does not exist
func (c *Config) AddLink(groupName string, link *link.Link) {
	group := c.getGroup(groupName)
	group.Links = append(group.Links, link)
}

// RemoveLink removes a link
func (c *Config) RemoveLink(target string) {
	for _, g := range c.Groups {
		for i, l := range g.Links {
			if l.To == target {
				g.Links = append(g.Links[:i], g.Links[i+1:]...)
				return
			}
		}
	}
}

// Excluded reports if the group is excluded by local configuration
func (c *Config) Excluded(name string) bool {
	for _, g := range c.Exclude {
		if g == name {
			return true
		}
	}

	return false
}
