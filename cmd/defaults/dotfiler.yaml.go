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
# Example :
#
# - name: vim
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
