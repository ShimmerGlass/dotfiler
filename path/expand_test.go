package path

import (
	"os/user"
	"testing"
)

func TestAbs(t *testing.T) {
	usr, err := user.Current()
	if err != nil {
		t.Skip(err)
	}
	home := usr.HomeDir
	t.Setenv("DOTFILER_TEST_VAR", "val")

	cases := []struct {
		name, base, p, want string
	}{
		{"empty", "/b", "", ""},
		{"tilde", "/b", "~", home},
		{"tilde subpath", "/b", "~/x", home + "/x"},
		{"relative", "/b", "x/y", "/b/x/y"},
		{"relative dot", "/b", "./x", "/b/x"},
		{"absolute", "/b", "/etc/x", "/etc/x"},
		{"env mid segment", "/b", "git/$DOTFILER_TEST_VAR", "/b/git/val"},
		{"env in absolute", "/b", "/etc/$DOTFILER_TEST_VAR/x", "/etc/val/x"},
		{"multiple env", "/b", "$DOTFILER_TEST_VAR/$DOTFILER_TEST_VAR", "/b/val/val"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := Abs(c.base, c.p); got != c.want {
				t.Errorf("Abs(%q, %q) = %q, want %q", c.base, c.p, got, c.want)
			}
		})
	}
}
