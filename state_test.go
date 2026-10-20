package main

import (
	"fmt"
	"testing"
)

func TestBoundedJournal(t *testing.T) {
	s := newMonitorState()
	for i := 0; i < 500; i++ {
		s.event(fmt.Sprint(i))
	}
	_, lines, _ := s.snapshot()
	if len(lines) != 200 || lines[0] != "300" || lines[199] != "499" {
		t.Fatalf("journal unbounded or wrong: %d", len(lines))
	}
}
