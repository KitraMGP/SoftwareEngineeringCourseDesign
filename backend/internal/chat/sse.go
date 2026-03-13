package chat

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type StreamWriter interface {
	Start() error
	SendMeta(meta StreamMeta) error
	SendDelta(content string) error
	SendError(code int, message string) error
	SendDone(finishReason string) error
	Close()
}

type SSEWriter struct {
	w                 http.ResponseWriter
	flusher           http.Flusher
	heartbeatInterval time.Duration

	mu      sync.Mutex
	started bool
	stopCh  chan struct{}
	doneCh  chan struct{}
}

func NewSSEWriter(w http.ResponseWriter, heartbeatInterval time.Duration) (*SSEWriter, error) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil, fmt.Errorf("response writer does not support flushing")
	}

	return &SSEWriter{
		w:                 w,
		flusher:           flusher,
		heartbeatInterval: heartbeatInterval,
	}, nil
}

func (s *SSEWriter) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.started {
		return nil
	}

	s.w.Header().Set("Content-Type", "text/event-stream")
	s.w.Header().Set("Cache-Control", "no-cache")
	s.w.Header().Set("Connection", "keep-alive")
	s.w.WriteHeader(http.StatusOK)
	s.flusher.Flush()

	s.started = true
	if s.heartbeatInterval > 0 {
		s.stopCh = make(chan struct{})
		s.doneCh = make(chan struct{})
		go s.heartbeatLoop(s.stopCh, s.doneCh)
	}

	return nil
}

func (s *SSEWriter) SendMeta(meta StreamMeta) error {
	return s.writeEvent("meta", meta)
}

func (s *SSEWriter) SendDelta(content string) error {
	return s.writeEvent("delta", map[string]any{
		"content": content,
	})
}

func (s *SSEWriter) SendError(code int, message string) error {
	return s.writeEvent("error", map[string]any{
		"code":    code,
		"message": message,
	})
}

func (s *SSEWriter) SendDone(finishReason string) error {
	return s.writeEvent("done", map[string]any{
		"finish_reason": finishReason,
	})
}

func (s *SSEWriter) Close() {
	s.mu.Lock()
	if !s.started || s.stopCh == nil {
		s.mu.Unlock()
		return
	}

	stopCh := s.stopCh
	doneCh := s.doneCh
	s.stopCh = nil
	s.doneCh = nil
	s.mu.Unlock()

	close(stopCh)
	<-doneCh
}

func (s *SSEWriter) heartbeatLoop(stopCh <-chan struct{}, doneCh chan struct{}) {
	ticker := time.NewTicker(s.heartbeatInterval)
	defer ticker.Stop()
	defer close(doneCh)

	for {
		select {
		case <-ticker.C:
			if err := s.writeComment("ping"); err != nil {
				return
			}
		case <-stopCh:
			return
		}
	}
}

func (s *SSEWriter) writeEvent(event string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	_, err = fmt.Fprintf(s.w, "event: %s\ndata: %s\n\n", event, data)
	if err != nil {
		return err
	}
	s.flusher.Flush()
	return nil
}

func (s *SSEWriter) writeComment(comment string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := fmt.Fprintf(s.w, ": %s\n\n", comment)
	if err != nil {
		return err
	}
	s.flusher.Flush()
	return nil
}
