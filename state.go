package main

import "sync"

type monitorState struct {
	youtubeOutages      uint64
	youtubeOutageActive bool
	paused              bool
	verbose             bool
	cycles              uint64
	samples             []nodeState
	nodes               [5]nodeState
	mu                  sync.Mutex
	rows                [5]string
	journal             []string
	revision            uint64
}

func newMonitorState() *monitorState {
	s := &monitorState{}
	for i := range s.rows {
		s.rows[i] = "Ожидание проверки…"
	}
	return s
}
func (s *monitorState) event(line string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.journal = append(s.journal, line)
	if len(s.journal) > 200 {
		s.journal = append([]string(nil), s.journal[len(s.journal)-200:]...)
	}
	s.revision++
}
func (s *monitorState) row(i int, line string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rows[i] = line
	s.revision++
}
func (s *monitorState) snapshot() ([5]string, []string, uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.rows, append([]string(nil), s.journal...), s.revision
}
