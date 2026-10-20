//go:build windows

package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

var targets = []struct{ name, ip string }{{"Debian", "192.168.1.107"}, {"Router", "192.168.1.100"}, {"Yandex", "77.88.44.242"}, {"YouTube", "192.178.24.110"}, {"DNS", "192.168.1.107:53"}}

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
		ts := time.Now()
		results := make([]result, 0, 5)
		for i, target := range targets {
			if !s.waitRunning(ctx) {
				return
			}
			timeout := c.timeout()
			start := time.Now()
			var ok bool
			if i == 4 {
				ok = checkDNS(ctx, target.ip, timeout)
			} else {
				ok = p.ping(target.ip, timeout)
			}
			if ctx.Err() != nil {
				return
			}
			elapsed := time.Since(start)
			status := "НЕТ ОТВЕТА"
			if ok {
				status = "OK"
			}
			s.record(i, ok, elapsed, time.Now())
			s.row(i, fmt.Sprintf("%s  %s  |  %s  |  %.0f мс  |  таймаут %d мс  |  %s", target.name, target.ip, status, float64(elapsed.Microseconds())/1000, timeout.Milliseconds(), time.Now().Format("15:04:05")))
			results = append(results, result{target.name, ok, elapsed})
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
				s.event("Ошибка записи лога: " + err.Error())
				return
			}
			if err = log.Sync(); err != nil {
				s.event("Ошибка синхронизации лога: " + err.Error())
				return
			}
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Second):
		}
	}
}
