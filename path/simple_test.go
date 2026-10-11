package path

import (
	"os/user"
	"testing"
)

func TestHasDirPrefix(t *testing.T) {
	cases := []struct {
		p, dir string
		want   bool
	}{
		{"/home/u", "/home/u", true},
		{"/home/u/x", "/home/u", true},
		{"/home/user2/x", "/home/u", false},
		{"/etc", "/", true},
	}
	for _, c := range cases {
		if got := HasDirPrefix(c.p, c.dir); got != c.want {
			t.Errorf("HasDirPrefix(%q, %q) = %v, want %v", c.p, c.dir, got, c.want)
		}
	}
}

func TestSimple(t *testing.T) {
	usr, err := user.Current()
	if err != nil {
		t.Skip(err)
	}
	home := usr.HomeDir

	cases := []struct {
		base, p, want string
	}{
		{"/w", "/w/a/b", "./a/b"},
		{"/w", "/w", "."},
		{"/w", "/w2/a", "/w2/a"},
		{"/", home + "/x", "~/x"},
		{"/", home + "suffix/x", home + "suffix/x"},
		{"/", "/etc/x", "/etc/x"},
		{"/w", home, "~"},
		{"/w", "/w2", "/w2"},
	}
	for _, c := range cases {
		if got := Simple(c.base, c.p); got != c.want {
			t.Errorf("Simple(%q, %q) = %q, want %q", c.base, c.p, got, c.want)
		}
	}
}
