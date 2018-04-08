package path

import (
	"os"
	"os/user"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

var envRxp *regexp.Regexp

func init() {
	envRxp = regexp.MustCompile("\\$\\w+")
}

// Abs returns the absolte path
func Abs(base, p string) string {
	if p == "" {
		return p
	}
	if strings.HasPrefix(p, "~") {
		usr, _ := user.Current()
		p = path.Join(usr.HomeDir, p[1:])
	}

	if !filepath.IsAbs(p) {
		p = filepath.Join(base, p)
	}

	ei := envRxp.FindAllStringIndex(p, -1)
	for i := len(ei) - 1; i >= 0; i-- {
		s, e := ei[i][0], ei[i][1]
		name := p[s+1 : e]
		value := os.Getenv(name)
		p = p[:s] + value + p[e:]
	}

	return p
}
