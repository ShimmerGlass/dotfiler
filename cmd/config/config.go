package config

import (
	"github.com/aestek/dotfiler/link"
)

type Config struct {
	Project
	Local
}

type Project struct {
	Base   string
	Groups []*Group
}

type Group struct {
	Name  string
	Links []*link.Link
}

type Local struct {
	Vars    interface{} `yaml:"vars"`
	Exclude []string    `yaml:"exclude"`
}

func (c *Config) LinkCount() (i int) {
	for _, g := range c.Groups {
		for range g.Links {
			i++
		}
	}
	return
}

func (c *Config) Links() []*link.Link {
	res := []*link.Link{}
	for _, g := range c.Groups {
		for _, l := range g.Links {
			res = append(res, l)
		}
	}
	return res
}

func (c *Config) AddLink(groupName string, link *link.Link) {
	group := c.getGroup(groupName)
	group.Links = append(group.Links, link)
}

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

func (c *Config) Excluded(name string) bool {
	for _, g := range c.Exclude {
		if g == name {
			return true
		}
	}

	return false
}
