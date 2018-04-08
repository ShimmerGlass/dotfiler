package defaults

import (
	"github.com/aestek/dotfiler/tmpl"
)

func init() {
	addFile(
		".gitignore",
		`# local dotfiler configuration, specific to this machine
/dotfiler_local.yaml
# compiled files
*`+tmpl.Ext+`
`)
}
