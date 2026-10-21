package main

import "testing"

func TestCompactLayoutKeepsJournalBelowFold(t *testing.T) {
	c, e := compactDimensions(false), compactDimensions(true)
	if c.Width != 380 || c.Height != 310 {
		t.Fatalf("collapsed: %+v", c)
	}
	if e.Width != c.Width || e.Height != c.Height+184 {
		t.Fatalf("expanded: %+v", e)
	}
	for _, r := range compactControls() {
		if r.X < 0 || r.Y < 0 || r.X+r.W > 380 || r.Y+r.H > c.Height {
			t.Fatalf("control outside compact: %+v", r)
		}
	}
	rs := compactControls()
	for i, a := range rs {
		for j, b := range rs {
			if i < j && a.X < b.X+b.W && b.X < a.X+a.W && a.Y < b.Y+b.H && b.Y < a.Y+a.H {
				t.Fatalf("overlap: %+v %+v", a, b)
			}
		}
	}
}
