package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSettingsRoundTrip(t *testing.T) {
	p := filepath.Join(t.TempDir(), "vpn-monitor.json")
	c, e := loadSettings(p)
	if e != nil || c != defaultSettings() {
		t.Fatal(c, e)
	}
	c.Targets[0] = targetConfig{"Local", "127.0.0.1"}
	c.TimeoutMS = 137
	if e = saveSettings(p, c); e != nil {
		t.Fatal(e)
	}
	got, e := loadSettings(p)
	if e != nil || got != c {
		t.Fatal(got, e)
	}
}
func TestSettingsValidationAndCorruption(t *testing.T) {
	for _, edit := range []func(*settings){func(c *settings) { c.Targets[2].IP = "::1" }, func(c *settings) { c.Targets[0].Name = "\n" }, func(c *settings) { c.DNSServer = "localhost:53" }, func(c *settings) { c.DNSServer = "127.0.0.1:0" }, func(c *settings) { c.DNSDomain = "https://youtube.com" }, func(c *settings) { c.DNSDomain = "127.0.0.1" }, func(c *settings) { c.TimeoutMS = 0 }, func(c *settings) { c.Version = 2 }} {
		c := defaultSettings()
		edit(&c)
		if c.validate() == nil {
			t.Fatal("invalid accepted", c)
		}
	}
	p := filepath.Join(t.TempDir(), "vpn-monitor.json")
	os.WriteFile(p, []byte("{bad"), 0600)
	c, e := loadSettings(p)
	if e == nil || c != defaultSettings() {
		t.Fatal(c, e)
	}
	b, _ := os.ReadFile(p)
	if string(b) != "{bad" {
		t.Fatal("corruption overwritten")
	}
}
func TestSettingsGenerationAndSaveFailure(t *testing.T) {
	c := newLiveConfig()
	s := newMonitorState()
	old, g := c.snapshot()
	s.record(3, false, time.Millisecond, time.Now())
	next := old
	next.Targets[3] = targetConfig{"Local fourth", "127.0.0.1"}
	if e := c.commit(filepath.Join(t.TempDir(), "missing", "x.json"), next, s); e == nil {
		t.Fatal("save succeeded")
	}
	if got, _ := c.snapshot(); got != old {
		t.Fatal("failed save applied")
	}
	p := filepath.Join(t.TempDir(), "vpn-monitor.json")
	if e := c.commit(p, next, s); e != nil {
		t.Fatal(e)
	}
	if c.publish(g, s, 3, false, time.Millisecond, time.Now(), "stale") {
		t.Fatal("stale generation accepted")
	}
	v := s.view(time.Now())
	if len(v.Samples) != 0 || v.YouTubeOutages != 0 || v.Nodes[3].Seen {
		t.Fatal("identity not reset", v)
	}
	_, g2 := c.snapshot()
	if !c.publish(g2, s, 3, true, time.Millisecond, time.Now(), "current") {
		t.Fatal("fresh rejected")
	}
	next.TimeoutMS = 123
	if e := c.commit(p, next, s); e != nil {
		t.Fatal(e)
	}
	v = s.view(time.Now())
	if len(v.Samples) != 1 {
		t.Fatal("timeout reset graph")
	}
}
