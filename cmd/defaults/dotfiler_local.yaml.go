package defaults

const LocalCfgHeader = `# Dotfiler local configuration file
#
# Contains settings specific to this machine.
#
# exclude: ignore links in the given groups.
#     exclude: ["i3"] # no i3 config on OSX
#
# vars: vars to be interpolated in config files using golang templates.
#     vars:
#       email: foo@bar.com
#       name: Foo
#
#     ~/.gitconfig:
#
#     [user]
#       email = {{.email}}
#       name = {{.name}}

`

func init() {
	addFile(
		"dotfiler_local.yaml",
		LocalCfgHeader+`exclude: []
vars:
`)
}
