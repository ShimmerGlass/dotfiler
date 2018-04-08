package cmd

import (
	"os"
	"os/user"
	"path/filepath"

	"github.com/aestek/dotfiler/cmd/config"
	"github.com/aestek/dotfiler/path"
)

var workdirName = ".dotfiles"
var cfgName = "dotfiler.yaml"
var localCfgName = "dotfiler_local.yaml"

func workdirExists() bool {
	_, err := os.Stat(basePath())
	if os.IsNotExist(err) {
		return false
	}
	must(err)
	return true
}

func ensureWorkdir() {
	if !workdirExists() {
		fail("%s does not exist. See `dotfiler install` to setup dotfiler\n", path.Simple("/", basePath()))
	}
}

func configPath() string {
	return filepath.Join(basePath(), cfgName)
}

func localConfigPath() string {
	return filepath.Join(basePath(), localCfgName)
}

func basePath() string {
	if customDirectory != "" {
		abs, err := filepath.Abs(customDirectory)
		must(err)
		return abs
	}
	usr, err := user.Current()
	must(err)
	return filepath.Join(usr.HomeDir, workdirName)
}

func getConfig() *config.Config {
	ensureWorkdir()
	c, err := config.Parse(configPath(), localConfigPath())
	must(err)
	return c
}

func writeConfig(c *config.Config) {
	must(config.Write(c, configPath(), localConfigPath()))
}
