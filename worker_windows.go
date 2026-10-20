//go:build windows

package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func monitor(ctx context.Context, c *liveConfig, s *monitorState) {
	exe, err := os.Executable()
	if err != nil {
		s.event("Ошибка: " + err.Error())
		return
	}
	log, err := os.OpenFile(filepath.Join(filepath.Dir(exe), "vpn-diag.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		s.event("Ошибка открытия лога: " + err.Error())
		return
	}
	defer log.Close()
	p, err := newPinger()
	if err != nil {
		s.event("Ошибка ICMP: " + err.Error())
		return
	}
	defer p.close()
	s.event("Мониторинг запущен. Лог сбоев: " + log.Name())
	for ctx.Err() == nil {
		if !s.waitRunning(ctx) {
			return
		}
		cfg, generation := c.snapshot()
		ts := time.Now()
		results := make([]result, 0, 5)
		for i := 0; i < 5; i++ {
			target := cfg.node(i)
			if !s.waitRunning(ctx) {
				return
			}
			timeout := time.Duration(cfg.TimeoutMS) * time.Millisecond
			start := time.Now()
			var ok bool
			if i == 4 {
				ok = checkDNSDomain(ctx, target.IP, cfg.DNSDomain, timeout)
			} else {
				ok = p.ping(target.IP, timeout)
			}
			if ctx.Err() != nil {
				return
			}
			elapsed := time.Since(start)
			status := "НЕТ ОТВЕТА"
			if ok {
				status = "OK"
			}
			line := fmt.Sprintf("%s | %s | %.0f мс | таймаут %d мс | %s", cfg.identity(i), status, float64(elapsed.Microseconds())/1000, timeout.Milliseconds(), time.Now().Format("15:04:05"))
			if !c.publish(generation, s, i, ok, elapsed, time.Now(), line) {
				break
			}
			results = append(results, result{cfg.identity(i), ok, elapsed})
		}
		if len(results) != 5 {
			continue
		}
		// Guard cycle journal and disk log against settings publication too.
		c.mu.RLock()
		if generation != c.generation {
			c.mu.RUnlock()
			continue
		}
		line, failed := formatResult(ts, results)
		available := 0
		for _, r := range results {
			if r.ok {
				available++
			}
		}
		s.finishCycle(fmt.Sprintf("Цикл завершён · доступны %d из 5 узлов", available))
		if failed {
			if _, err = fmt.Fprintln(log, line); err != nil {
				c.mu.RUnlock()
				s.event("Ошибка записи лога: " + err.Error())
				return
			}
			if err = log.Sync(); err != nil {
				c.mu.RUnlock()
				s.event("Ошибка синхронизации лога: " + err.Error())
				return
			}
		}
		c.mu.RUnlock()
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Second):
		}
	}
}
