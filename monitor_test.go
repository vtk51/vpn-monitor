package main

import (
	"strings"
	"testing"
	"time"
)

func TestFailureMessage(t *testing.T) {
	r := []result{{"A", false, 2 * time.Second}, {"Router", true, time.Millisecond}, {"Ya", true, time.Millisecond}, {"YouTube", false, 3 * time.Second}, {"DNS", false, time.Second}}
	msg, failure := formatResult(time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC), r)
	if !failure || !strings.Contains(msg, "A=False, Router=True, Ya=True, YouTube=False, DNS=False") || !strings.Contains(msg, "Неответ: A=2.00s, YouTube=3.00s, DNS=1.00s") {
		t.Fatalf("unexpected: %v %s", failure, msg)
	}
	r[3].ok = true
	_, failure = formatResult(time.Now(), r)
	if failure {
		t.Fatal("must only alert on YouTube ping failure")
	}
}
