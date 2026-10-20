package main

import (
	"fmt"
	"strconv"
	"sync"
	"time"
)

type liveConfig struct {
	mu         sync.RWMutex
	value      settings
	generation uint64
}

func newLiveConfig() *liveConfig { return &liveConfig{value: defaultSettings()} }
func (c *liveConfig) snapshot() (settings, uint64) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.value, c.generation
}
func (c *liveConfig) timeout() time.Duration {
	v, _ := c.snapshot()
	return time.Duration(v.TimeoutMS) * time.Millisecond
}
func parseTimeout(s string) (int, error) {
	n, e := strconv.Atoi(s)
	if e != nil || n < 1 || n > 60000 {
		return 0, fmt.Errorf("Введите целое число от 1 до 60000 мс")
	}
	return n, nil
}
func (c *liveConfig) apply(s string) error {
	n, e := parseTimeout(s)
	if e != nil {
		return e
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.value.TimeoutMS = n
	c.generation++
	return nil
}

// Lock order: configuration then monitor state. Save before publishing.
func (c *liveConfig) commit(path string, next settings, s *monitorState) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if e := saveSettings(path, next); e != nil {
		return e
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	old := c.value
	c.value = next
	c.generation++
	s.config = next
	for i := range s.nodes {
		if old.identity(i) != next.identity(i) {
			s.nodes[i] = nodeState{}
			s.rows[i] = "Ожидание проверки…"
		}
	}
	if old.Targets[3] != next.Targets[3] {
		s.samples = nil
		s.youtubeOutages = 0
		s.youtubeOutageActive = false
		s.appendEvent(time.Now().Format("15:04:05") + "   Узел 4 изменён: " + next.identity(3) + " · график и счётчик сброшены")
	}
	s.appendEvent(time.Now().Format("15:04:05") + "   Настройки сохранены · со следующего цикла")
	s.revision++
	return nil
}
func (c *liveConfig) publish(g uint64, s *monitorState, i int, ok bool, elapsed time.Duration, at time.Time, line string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if g != c.generation {
		return false
	}
	s.record(i, ok, elapsed, at)
	s.row(i, line)
	return true
}
