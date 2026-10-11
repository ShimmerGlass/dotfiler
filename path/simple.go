package path

import (
	"os/user"
	"path"
	"strings"
)

func HasDirPrefix(p, dir string) bool {
	if dir == "/" {
		return strings.HasPrefix(p, "/")
	}
	return p == dir || strings.HasPrefix(p, dir+"/")
}

// Simple returns a simple version of p
func Simple(base, p string) string {
	if base != "/" && HasDirPrefix(p, base) {
		if p == base {
			return "."
		}
		return "./" + strings.TrimPrefix(p, base+"/")
	}

	usr, err := user.Current()
	if err != nil {
		return p
	}
	if HasDirPrefix(p, usr.HomeDir) {
		p = path.Join("~", strings.TrimPrefix(p, usr.HomeDir))
	}
	return p
}
