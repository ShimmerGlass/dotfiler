package defaults

import (
	"bytes"
	"testing"
)

func TestFiles(t *testing.T) {
	cases := map[string]string{
		"dotfiler.yaml":       ProjectCfgHeader,
		"dotfiler_local.yaml": LocalCfgHeader,
	}
	for name, header := range cases {
		content, ok := Files[name]
		if !ok {
			t.Errorf("missing default file %s", name)
			continue
		}
		if !bytes.HasPrefix(content, []byte(header)) {
			t.Errorf("%s does not start with its header", name)
		}
	}
}
