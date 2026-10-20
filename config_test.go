package main

import (
	"sync"
	"testing"
	"time"
)

func TestLiveTimeout(t *testing.T) {
	c := newLiveConfig()
	if c.timeout() != 2*time.Second {
		t.Fatal("default")
	}
	if err := c.apply("137"); err != nil {
		t.Fatal(err)
	}
	if c.timeout() != 137*time.Millisecond {
		t.Fatal("next probe must see new timeout")
	}
}
func TestTimeoutValidation(t *testing.T) {
	c := newLiveConfig()
	for _, s := range []string{"", "0", "60001", "-1", "abc", "1.5"} {
		if c.apply(s) == nil {
			t.Fatalf("accepted %q", s)
		}
		if c.timeout() != 2*time.Second {
			t.Fatal("invalid edit changed config")
		}
	}
	for _, s := range []string{"1", "60000"} {
		if err := c.apply(s); err != nil {
			t.Fatal(err)
		}
	}
}
func TestConcurrentTimeout(t *testing.T) {
	c := newLiveConfig()
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				if err := c.apply("97"); err != nil {
					t.Error(err)
				}
				if c.timeout() < time.Millisecond {
					t.Error("bad read")
				}
			}
		}()
	}
	wg.Wait()
}
