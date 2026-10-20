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
		statuses := []string{}
		for _, v := range r {
			statuses = append(statuses, v.name+"="+boolText(v.ok))
		}
		return fmt.Sprintf("[%s] ВНИМАНИЕ: Сбой %s! %s | Неответ: %s", stamp, r[3].name, strings.Join(statuses, ", "), strings.Join(failed, ", ")), true
	}
	times := []string{}
	for _, v := range r {
		times = append(times, fmt.Sprintf("%s=%.2fs", v.name, v.elapsed.Seconds()))
	}
	return fmt.Sprintf("[%s] Все в норме. Времена ответа: %s", stamp, strings.Join(times, ", ")), false
}
