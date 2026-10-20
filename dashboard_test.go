package main

import (
	"context"
	"testing"
	"time"
)

func TestPauseGate(t *testing.T) {
	s := newMonitorState()
	s.togglePause()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan bool, 1)
	go func() { done <- s.waitRunning(ctx) }()
	select {
	case <-done:
		t.Fatal("paused gate released")
	case <-time.After(50 * time.Millisecond):
	}
	s.togglePause()
	select {
	case ok := <-done:
		if !ok {
			t.Fatal("resume failed")
		}
	case <-time.After(time.Second):
		t.Fatal("resume blocked")
	}
	s.togglePause()
	cancel()
	if s.waitRunning(ctx) {
		t.Fatal("cancel must unblock false")
	}
}
func TestRollingHistory(t *testing.T) {
	s := newMonitorState()
	now := time.Now()
	s.record(3, true, 10*time.Millisecond, now.Add(-61*time.Second))
	s.record(3, false, time.Second, now.Add(-30*time.Second))
	s.record(3, true, 20*time.Millisecond, now)
	v := s.view(now)
	if len(v.Samples) != 2 || v.Samples[0].OK || !v.Samples[1].OK {
		t.Fatalf("history must prune and retain gaps: %+v", v)
	}
	s.finishCycle("normal")
	_, a, _ := s.snapshot()
	s.finishCycle("normal")
	_, b, _ := s.snapshot()
	if len(a) != len(b) {
		t.Fatal("routine cycles leaked")
	}
	s.toggleVerbose()
	s.finishCycle("normal")
	_, b, _ = s.snapshot()
	if len(b) != len(a)+1 {
		t.Fatal("verbose cycle missing")
	}
}
func TestTransitionFiltering(t *testing.T) {
	s := newMonitorState()
	now := time.Now()
	s.record(0, true, time.Millisecond, now)
	_, a, _ := s.snapshot()
	s.record(0, true, time.Millisecond, now)
	_, b, _ := s.snapshot()
	if len(a) != 1 || len(b) != 1 {
		t.Fatalf("normal checks must not repeat: %v %v", a, b)
	}
	s.record(0, false, time.Second, now)
	s.record(0, false, time.Second, now)
	s.record(0, true, time.Millisecond, now)
	_, b, _ = s.snapshot()
	if len(b) != 3 {
		t.Fatalf("expected initial, outage, recovery: %v", b)
	}
}
