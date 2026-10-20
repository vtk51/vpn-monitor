package main

import (
	"encoding/json"
	"testing"
	"time"
)

func outageCount(t *testing.T, s *monitorState) uint64 {
	t.Helper()
	b, err := json.Marshal(s.view(time.Now()))
	if err != nil {
		t.Fatal(err)
	}
	var v map[string]json.RawMessage
	if err := json.Unmarshal(b, &v); err != nil {
		t.Fatal(err)
	}
	raw, ok := v["YouTubeOutages"]
	if !ok {
		t.Fatal("dashboard snapshot missing YouTubeOutages counter")
	}
	var count uint64
	if err := json.Unmarshal(raw, &count); err != nil {
		t.Fatal(err)
	}
	return count
}

func TestYouTubeConsecutiveFailures(t *testing.T) {
	s := newMonitorState()
	for i := 0; i < 4; i++ {
		s.record(3, false, time.Millisecond, time.Now())
	}
	if n := outageCount(t, s); n != 1 {
		t.Fatalf("consecutive failures count = %d, want 1 episode", n)
	}
}

func TestYouTubeRecoveryThenFailure(t *testing.T) {
	s := newMonitorState()
	for _, ok := range []bool{false, false, true, true, false, false} {
		s.record(3, ok, time.Millisecond, time.Now())
	}
	if n := outageCount(t, s); n != 2 {
		t.Fatalf("two outage episodes count = %d, want 2", n)
	}
	if n := outageCount(t, newMonitorState()); n != 0 {
		t.Fatalf("new launch count = %d, want 0", n)
	}
}

func TestOtherHostFailuresNotCounted(t *testing.T) {
	s := newMonitorState()
	for _, i := range []int{0, 1, 2, 4} {
		s.record(i, false, time.Millisecond, time.Now())
	}
	s.record(3, true, time.Millisecond, time.Now())
	if n := outageCount(t, s); n != 0 {
		t.Fatalf("other host failures count = %d, want 0", n)
	}
}

func TestYouTubePauseDoesNotCount(t *testing.T) {
	s := newMonitorState()
	s.togglePause()
	// An in-flight real probe may finish during pause; display it but do not count it.
	s.record(3, false, time.Millisecond, time.Now())
	if n := outageCount(t, s); n != 0 {
		t.Fatalf("paused failure count = %d, want 0", n)
	}
	s.togglePause()
	s.record(3, false, time.Millisecond, time.Now())
	if n := outageCount(t, s); n != 1 {
		t.Fatalf("resumed failure count = %d, want 1", n)
	}
	s.togglePause()
	s.record(3, true, time.Millisecond, time.Now())
	s.togglePause()
	s.record(3, false, time.Millisecond, time.Now())
	if n := outageCount(t, s); n != 1 {
		t.Fatalf("pause must not rearm an active episode, got %d", n)
	}
	s.record(3, true, time.Millisecond, time.Now())
	s.record(3, false, time.Millisecond, time.Now())
	if n := outageCount(t, s); n != 2 {
		t.Fatalf("recovery after pause must rearm, got %d", n)
	}
}

func TestYouTubeFirstFailure(t *testing.T) {
	s := newMonitorState()
	if n := outageCount(t, s); n != 0 {
		t.Fatalf("initial count = %d, want 0", n)
	}
	s.record(3, false, time.Millisecond, time.Now())
	if n := outageCount(t, s); n != 1 {
		t.Fatalf("first failed sample count = %d, want 1", n)
	}
}
