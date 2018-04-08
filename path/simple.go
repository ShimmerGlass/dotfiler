package path

import (
	"os/user"
	"path"
	"strings"
)

// Simple returns a simple version of p
func Simple(base, p string) string {
	usr, _ := user.Current()
	if strings.HasPrefix(p, base) && base != "/" {
		p = "./" + strings.TrimPrefix(p, base+"/")
	}

	if strings.HasPrefix(p, usr.HomeDir) {
		p = strings.TrimPrefix(p, usr.HomeDir)
		p = path.Join("~", p)
	}
	return p
}
