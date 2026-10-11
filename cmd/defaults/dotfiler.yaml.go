package defaults

const ProjectCfgHeader = `# Dotfiler configuration file
#
# This file lists symlinks managed by dotfiler.
# Links are organised in groups, groups can be excluded on a particular
# machine in local.yaml.
#
# Links are written in the form <source>:<destination>, as in ln -s <source> <destination>
# <source> and <destination> can contain ~ or env vars ($USER)
# <source> is relative to ~/.dotfiles
#
# Groups can also declare system packages and Go tools installed by 'dotfiler up'.
# packages: a name (same for all managers) or a mapping with per-manager overrides
#   (pacman on linux, installed with trizen when available for AUR support, brew on macOS)
# tools: go: <module>[@version], installed with go install (latest by default)
#
# Example :
#
# - name: vim
#   packages:
#   - vim
#   - name: fd
#     brew: fd
#     pacman: fd
#   tools:
#   - go: github.com/aestek/ps1
#   links:
#   - .vimrc:~/.vimrc

`

func init() {
	addFile(
		"dotfiler.yaml",
		ProjectCfgHeader+`- name: default
  links:
`)
}
