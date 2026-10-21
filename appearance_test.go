package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestTransparencyPersistenceAndBackwardCompatibility(t *testing.T) {
	p := filepath.Join(t.TempDir(), "vpn-monitor.json")
	c := defaultSettings()
	if c.TransparencyPercent != 0 {
		t.Fatal("default not opaque")
	}
	b, _ := json.Marshal(c)
	var old map[string]interface{}
	json.Unmarshal(b, &old)
	delete(old, "transparency_percent")
	b, _ = json.Marshal(old)
	os.WriteFile(p, b, 0600)
	got, err := loadSettings(p)
	if err != nil || got != c {
		t.Fatal(got, err)
	}
	c.TransparencyPercent = 70
	if err = saveSettings(p, c); err != nil {
		t.Fatal(err)
	}
	got, err = loadSettings(p)
	if err != nil || got != c {
		t.Fatal(got, err)
	}
	for _, n := range []int{-1, 71} {
		c.TransparencyPercent = n
		if c.validate() == nil {
			t.Fatalf("accepted %d", n)
		}
	}
	for _, n := range []int{0, 1, 69, 70} {
		c.TransparencyPercent = n
		if err = c.validate(); err != nil {
			t.Fatal(err)
		}
	}
}

func TestAppearanceCommitPreservesHistoryAndFailurePreservesRuntime(t *testing.T) {
	c, s := newLiveConfig(), newMonitorState()
	s.record(3, false, time.Millisecond, time.Now())
	before := s.view(time.Now())
	next, _ := c.snapshot()
	next.TransparencyPercent = 40
	if c.commit(filepath.Join(t.TempDir(), "missing", "x.json"), next, s) == nil {
		t.Fatal("save succeeded")
	}
	if got, _ := c.snapshot(); got.TransparencyPercent != 0 {
		t.Fatal("failed save published")
	}
	if err := c.commit(filepath.Join(t.TempDir(), "x.json"), next, s); err != nil {
		t.Fatal(err)
	}
	after := s.view(time.Now())
	if len(after.Samples) != len(before.Samples) || after.YouTubeOutages != before.YouTubeOutages || after.Nodes != before.Nodes {
		t.Fatal("appearance reset monitoring")
	}
}
