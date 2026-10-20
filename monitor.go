package main

import (
	"fmt"
	"strings"
	"time"
)

type result struct {
	name    string
	ok      bool
	elapsed time.Duration
}

func boolText(b bool) string {
	if b {
		return "True"
	}
	return "False"
}
func formatResult(ts time.Time, r []result) (string, bool) {
	stamp := ts.Format("2006-01-02 15:04:05")
	if !r[3].ok {
		failed := []string{}
		for _, v := range r {
			if !v.ok {
				failed = append(failed, fmt.Sprintf("%s=%.2fs", v.name, v.elapsed.Seconds()))
			}
		}
		return fmt.Sprintf("[%s] ВНИМАНИЕ: Сбой YouTube (192.178.24.110)! Debian(107)=%s, Router=%s, Ya=%s, DNS=%s | Неответ: %s", stamp, boolText(r[0].ok), boolText(r[1].ok), boolText(r[2].ok), boolText(r[4].ok), strings.Join(failed, ", ")), true
	}
	times := []string{}
	for _, v := range r {
		times = append(times, fmt.Sprintf("%s=%.2fs", v.name, v.elapsed.Seconds()))
	}
	return fmt.Sprintf("[%s] Все в норме. Времена ответа: %s", stamp, strings.Join(times, ", ")), false
}
