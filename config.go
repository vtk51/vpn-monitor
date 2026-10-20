package main

import (
	"fmt"
	"strconv"
	"sync/atomic"
	"time"
)

type liveConfig struct{ ms atomic.Int64 }

func newLiveConfig() *liveConfig             { c := &liveConfig{}; c.ms.Store(2000); return c }
func (c *liveConfig) timeout() time.Duration { return time.Duration(c.ms.Load()) * time.Millisecond }
func (c *liveConfig) apply(s string) error {
	n, e := strconv.Atoi(s)
	if e != nil || n < 1 || n > 60000 {
		return fmt.Errorf("Введите целое число от 1 до 60000 мс")
	}
	c.ms.Store(int64(n))
	return nil
}
