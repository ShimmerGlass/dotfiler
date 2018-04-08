package config

import (
	"fmt"
	"io/ioutil"

	"github.com/aestek/dotfiler/cmd/defaults"

	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
)

// Write stores the config content on disk at given paths
func Write(c *Config, cfgPath, localCfgPath string) error {
	cfg := []rawGroupConfig{}

	for _, g := range c.Groups {
		group := rawGroupConfig{
			Name: g.Name,
		}

		for _, l := range g.Links {
			group.Links = append(group.Links, linkString(c.Workdir, l))
		}

		cfg = append(cfg, group)
	}

	raw, err := yaml.Marshal(cfg)
	if err != nil {
		return errors.Wrap(err, "config write")
	}
	raw = append([]byte(defaults.ProjectCfgHeader), raw...)

	fmt.Println("write config to", cfgPath)
	err = ioutil.WriteFile(cfgPath, raw, 0664)
	if err != nil {
		return errors.Wrap(err, "config write")
	}

	raw, err = yaml.Marshal(c.Local)
	if err != nil {
		return errors.Wrap(err, "config write")
	}
	raw = append([]byte(defaults.LocalCfgHeader), raw...)

	fmt.Println("write local config to", localCfgPath)
	err = ioutil.WriteFile(localCfgPath, raw, 0664)
	if err != nil {
		return errors.Wrap(err, "config write")
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
