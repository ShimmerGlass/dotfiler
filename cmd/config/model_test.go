package config

import (
	"reflect"
	"testing"

	"github.com/aestek/dotfiler/link"
)

func newTestConfig() (*Config, []*link.Link) {
	ls := []*link.Link{
		{From: "/s/a", To: "/t/a"},
		{From: "/s/b", To: "/t/b"},
		{From: "/s/c", To: "/t/c"},
	}
	c := &Config{Project: Project{Groups: []*Group{
		{Name: "g1", Links: []*link.Link{ls[0], ls[1]}},
		{Name: "g2", Links: []*link.Link{ls[2]}},
	}}}
	return c, ls
}

func TestLinkCountAndLinks(t *testing.T) {
	empty := &Config{}
	if empty.LinkCount() != 0 || len(empty.Links()) != 0 {
		t.Error("empty config should have no links")
	}

	c, ls := newTestConfig()
	if c.LinkCount() != 3 {
		t.Errorf("LinkCount = %d, want 3", c.LinkCount())
	}
	if got := c.Links(); !reflect.DeepEqual(got, ls) {
		t.Errorf("Links = %v, want %v", got, ls)
	}
}

func TestAddLink(t *testing.T) {
	c, _ := newTestConfig()
	existing := &link.Link{From: "/s/d", To: "/t/d"}
	c.AddLink("g2", existing)
	if len(c.Groups) != 2 || len(c.Groups[1].Links) != 2 || c.Groups[1].Links[1] != existing {
		t.Errorf("link not appended to existing group: %+v", c.Groups)
	}

	created := &link.Link{From: "/s/e", To: "/t/e"}
	c.AddLink("g3", created)
	if len(c.Groups) != 3 || c.Groups[2].Name != "g3" || len(c.Groups[2].Links) != 1 || c.Groups[2].Links[0] != created {
		t.Errorf("group not created: %+v", c.Groups)
	}
	if c.LinkCount() != 5 {
		t.Errorf("LinkCount = %d, want 5", c.LinkCount())
	}
}

func TestFindLink(t *testing.T) {
	c, ls := newTestConfig()
	if got := c.FindLink("/t/c"); got != ls[2] {
		t.Errorf("FindLink hit = %v, want %v", got, ls[2])
	}
	if got := c.FindLink("/t/nope"); got != nil {
		t.Errorf("FindLink miss = %v, want nil", got)
	}
}

func TestRemoveLink(t *testing.T) {
	c, ls := newTestConfig()
	c.RemoveLink("/t/a")
	if got := c.Links(); !reflect.DeepEqual(got, ls[1:]) {
		t.Errorf("Links after remove = %v, want %v", got, ls[1:])
	}

	c.RemoveLink("/t/nope")
	if c.LinkCount() != 2 {
		t.Errorf("miss removed something, LinkCount = %d", c.LinkCount())
	}

	c.RemoveLink("/t/c")
	if got := c.Links(); !reflect.DeepEqual(got, ls[1:2]) {
		t.Errorf("Links = %v, want %v", got, ls[1:2])
	}
}

func TestExcluded(t *testing.T) {
	c := &Config{Local: Local{Exclude: []string{"i3", "mac"}}}
	if !c.Excluded("i3") || !c.Excluded("mac") {
		t.Error("expected excluded")
	}
	if c.Excluded("vim") || c.Excluded("") {
		t.Error("expected not excluded")
	}
	if (&Config{}).Excluded("i3") {
		t.Error("empty config excludes nothing")
	}
}
