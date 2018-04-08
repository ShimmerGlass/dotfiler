package config

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
)

type rawGroupConfig struct {
	Name  string   `yaml:"name"`
	Links []string `yaml:"links"`
}

func Parse(cfgPath, localCfgPath string) (*Config, error) {
	absCfgPath, err := filepath.Abs(cfgPath)
	if err != nil {
		return nil, errors.Wrapf(err, "parse config %s", cfgPath)
	}

	base := filepath.Dir(absCfgPath)

	cfg := &Config{
		Project: Project{
			Base: base,
		},
	}

	contents, err := ioutil.ReadFile(cfgPath)
	if os.IsNotExist(err) {
		return cfg, nil
	}
	if err != nil {
		return nil, errors.Wrapf(err, "parse config %s", cfgPath)
	}

	var groups []rawGroupConfig
	err = yaml.Unmarshal(contents, &groups)
	if err != nil {
		return nil, errors.Wrapf(err, "parse config %s", cfgPath)
	}

	for _, g := range groups {
		gc := &Group{
			Name: g.Name,
		}

		for _, rl := range g.Links {
			link, err := parseLink(base, rl)
			if err != nil {
				return nil, err
			}

			gc.Links = append(gc.Links, link)
		}

		cfg.Groups = append(cfg.Groups, gc)
	}

	contents, err = ioutil.ReadFile(localCfgPath)
	if os.IsNotExist(err) {
		return cfg, nil
	}
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(contents, &cfg.Local)
	if err != nil {
		return nil, errors.Wrap(err, "local config parse")
	}
	return cfg, nil
}
