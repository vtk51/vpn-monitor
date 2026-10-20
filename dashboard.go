package main

import (
	"context"
	"fmt"
	"time"
)

func (s *monitorState) togglePause() {
	s.mu.Lock()
	s.paused = !s.paused
	p := s.paused
	s.revision++
	s.mu.Unlock()
	if p {
		s.event(time.Now().Format("15:04:05") + "   Мониторинг приостановлен")
	} else {
		s.event(time.Now().Format("15:04:05") + "   Мониторинг продолжен")
	}
}
func (s *monitorState) waitRunning(ctx context.Context) bool {
	for {
		if ctx.Err() != nil {
			return false
		}
		s.mu.Lock()
		p := s.paused
		s.mu.Unlock()
		if !p {
			return true
		}
		select {
		case <-ctx.Done():
			return false
		case <-time.After(25 * time.Millisecond):
		}
	}
}

type dashboardView struct {
	Config          settings
	YouTubeOutages  uint64
	Nodes           [5]nodeState
	Samples         []nodeState
	Journal         []string
	Paused, Verbose bool
	Cycles          uint64
}

func (s *monitorState) prune(now time.Time) {
	i := 0
	for i < len(s.samples) && s.samples[i].At.Before(now.Add(-time.Minute)) {
		i++
	}
	s.samples = append([]nodeState(nil), s.samples[i:]...)
}
func (s *monitorState) view(now time.Time) dashboardView {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.prune(now)
	return dashboardView{Config: s.config, YouTubeOutages: s.youtubeOutages, Nodes: s.nodes, Samples: append([]nodeState(nil), s.samples...), Journal: append([]string(nil), s.journal...), Paused: s.paused, Verbose: s.verbose, Cycles: s.cycles}
}
func (s *monitorState) finishCycle(line string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cycles++
	if s.verbose {
		s.appendEvent(time.Now().Format("15:04:05") + "   Проверка   ·   " + line)
	}
	s.revision++
}
func (s *monitorState) toggleVerbose() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.verbose = !s.verbose
	s.revision++
}

type nodeState struct {
	Seen, OK bool
	Elapsed  time.Duration
	At       time.Time
}

func (s *monitorState) record(i int, ok bool, elapsed time.Duration, at time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	old := s.nodes[i]
	s.nodes[i] = nodeState{true, ok, elapsed, at}
	if i == 3 {
		// Pause can overlap an in-flight probe: show its result, but leave
		// both the counter and episode latch untouched until running again.
		if !s.paused {
			if !ok && !s.youtubeOutageActive {
				s.youtubeOutages++
			}
			s.youtubeOutageActive = !ok
		}
		s.samples = append(s.samples, s.nodes[i])
		s.prune(at)
	}
	if !old.Seen || old.OK != ok {
		label := "Доступен"
		if !ok {
			label = "Нет ответа"
		} else if old.Seen {
			label = "Восстановлено"
		}
		s.appendEvent(fmt.Sprintf("%s   %s   ·   %s", at.Format("15:04:05"), label, s.config.identity(i)))
	}
	s.rows[i] = fmt.Sprintf("%s | %v | %.2f мс", s.config.identity(i), ok, float64(elapsed.Microseconds())/1000)
	s.revision++
}
func (s *monitorState) appendEvent(line string) {
	s.journal = append(s.journal, line)
	if len(s.journal) > 200 {
		s.journal = append([]string(nil), s.journal[len(s.journal)-200:]...)
	}
}
