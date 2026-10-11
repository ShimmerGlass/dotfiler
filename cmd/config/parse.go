package config

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/aestek/dotfiler/link"
	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
	yamlv3 "gopkg.in/yaml.v3"
)

type rawGroupConfig struct {
	Name  string   `yaml:"name"`
	Links []string `yaml:"links"`
}

// Parse load given files and parses them to generate a config
func Parse(cfgPath, localCfgPath string) (*Config, error) {
	absCfgPath, err := filepath.Abs(cfgPath)
	if err != nil {
		return nil, errors.Wrapf(err, "parse config %s", cfgPath)
	}

	base := filepath.Dir(absCfgPath)

	cfg := &Config{
		Workdir: base,
		raw:     map[*link.Link]string{},
	}

	err = parseProject(cfg, cfgPath)
	if err != nil {
		return nil, err
	}

	contents, err := ioutil.ReadFile(localCfgPath)
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

func parseProject(cfg *Config, cfgPath string) error {
	contents, err := ioutil.ReadFile(cfgPath)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return errors.Wrapf(err, "parse config %s", cfgPath)
	}

	doc := &yamlv3.Node{}
	err = yamlv3.Unmarshal(contents, doc)
	if err != nil {
		return errors.Wrapf(err, "parse config %s", cfgPath)
	}

	var groups []rawGroupConfig
	if len(doc.Content) > 0 {
		err = doc.Content[0].Decode(&groups)
		if err != nil {
			return errors.Wrapf(err, "parse config %s", cfgPath)
		}
		cfg.doc = doc
	}

	for _, g := range groups {
		gc := &Group{
			Name: g.Name,
		}

		for _, rl := range g.Links {
			link, err := parseLink(cfg.Workdir, rl)
			if err != nil {
				return err
			}

			cfg.raw[link] = rl
			gc.Links = append(gc.Links, link)
		}

		cfg.Groups = append(cfg.Groups, gc)
	}

	return nil
}
