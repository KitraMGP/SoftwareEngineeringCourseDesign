package chat

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"

	"github.com/google/uuid"
)

var ErrGenerationInProgress = errors.New("message generation already in progress")

type streamManager struct {
	mu      sync.Mutex
	streams map[uuid.UUID]*activeStream
}

type activeStream struct {
	cancel        context.CancelFunc
	stopRequested atomic.Bool
}

type streamHandle struct {
	manager   *streamManager
	sessionID uuid.UUID
	stream    *activeStream
	ctx       context.Context
	once      sync.Once
}

func newStreamManager() *streamManager {
	return &streamManager{
		streams: make(map[uuid.UUID]*activeStream),
	}
}

func (m *streamManager) Acquire(parent context.Context, sessionID uuid.UUID) (*streamHandle, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.streams[sessionID]; exists {
		return nil, ErrGenerationInProgress
	}

	childCtx, cancel := context.WithCancel(parent)
	active := &activeStream{
		cancel: cancel,
	}
	m.streams[sessionID] = active
	return &streamHandle{
		manager:   m,
		sessionID: sessionID,
		stream:    active,
		ctx:       childCtx,
	}, nil
}

func (m *streamManager) Stop(sessionID uuid.UUID) bool {
	m.mu.Lock()
	active, ok := m.streams[sessionID]
	m.mu.Unlock()
	if !ok {
		return false
	}

	active.stopRequested.Store(true)
	active.cancel()
	return true
}

func (h *streamHandle) Context() context.Context {
	return h.ctx
}

func (h *streamHandle) StopRequested() bool {
	return h.stream.stopRequested.Load()
}

func (h *streamHandle) Release() {
	h.once.Do(func() {
		h.manager.mu.Lock()
		if current, ok := h.manager.streams[h.sessionID]; ok && current == h.stream {
			delete(h.manager.streams, h.sessionID)
		}
		h.manager.mu.Unlock()
		h.stream.cancel()
	})
}
