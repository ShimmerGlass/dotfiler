package config

import (
	"os/user"
	"reflect"
	"testing"

	"github.com/aestek/dotfiler/link"
)

func TestParseLink(t *testing.T) {
	usr, err := user.Current()
	if err != nil {
		t.Skip(err)
	}
	home := usr.HomeDir

	cases := []struct {
		name, in string
		want     *link.Link
	}{
		{"two parts", ".vimrc:~/.vimrc", &link.Link{From: "/base/.vimrc", To: home + "/.vimrc"}},
		{"templated", "T:git:~/.gitconfig", &link.Link{From: "/base/git", To: home + "/.gitconfig", Templated: true}},
		{"absolute", "/src/a:/etc/a", &link.Link{From: "/src/a", To: "/etc/a"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := parseLink("/base", c.in)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("parseLink(%q) = %+v, want %+v", c.in, got, c.want)
			}
		})
	}
}

func TestParseLinkInvalid(t *testing.T) {
	for _, in := range []string{"", "onlyone", "a:b:c:d"} {
		if _, err := parseLink("/base", in); err == nil {
			t.Errorf("parseLink(%q): expected error", in)
		}
	}
}

func TestLinkString(t *testing.T) {
	usr, err := user.Current()
	if err != nil {
		t.Skip(err)
	}
	home := usr.HomeDir

	cases := []struct {
		name string
		l    *link.Link
		want string
	}{
		{"plain", &link.Link{From: "/base/a", To: "/etc/a"}, "./a:/etc/a"},
		{"home target", &link.Link{From: "/base/a", To: home + "/.a"}, "./a:~/.a"},
		{"templated", &link.Link{From: "/base/a", To: "/etc/a", Templated: true}, "T:./a:/etc/a"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := linkString("/base", c.l); got != c.want {
				t.Errorf("linkString = %q, want %q", got, c.want)
			}
		})
	}
}

func TestLinkRoundTrip(t *testing.T) {
	usr, err := user.Current()
	if err != nil {
		t.Skip(err)
	}
	links := []*link.Link{
		{From: "/base/a", To: "/etc/a"},
		{From: "/base/sub/a", To: usr.HomeDir + "/.a", Templated: true},
		{From: "/elsewhere/a", To: "/etc/b"},
	}
	for _, l := range links {
		got, err := parseLink("/base", linkString("/base", l))
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, l) {
			t.Errorf("round trip = %+v, want %+v", got, l)
		}
	}
}
