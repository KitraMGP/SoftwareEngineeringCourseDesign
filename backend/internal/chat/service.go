package chat

import (
	"context"
	"net/http"
	"strings"

	"backend/internal/platform/httpx"

	"github.com/google/uuid"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateSession(ctx context.Context, userID uuid.UUID, input CreateSessionInput) (*Session, error) {
	model := strings.TrimSpace(input.Model)
	if model == "" {
		return nil, httpx.ValidationFailed(httpx.FieldError{Field: "model", Message: "model is required"})
	}

	session, err := s.repo.Create(ctx, userID, CreateSessionInput{
		Name:            input.Name,
		Model:           model,
		KnowledgeBaseID: input.KnowledgeBaseID,
	})
	if err != nil {
		if err == ErrNotFound && input.KnowledgeBaseID != nil {
			return nil, httpx.NotFound("knowledge base not found")
		}
		return nil, httpx.Internal("failed to create session").WithErr(err)
	}
	return session, nil
}

func (s *Service) ListSessions(ctx context.Context, userID uuid.UUID, page, size int, keyword string) (*ListSessionsResult, error) {
	return s.repo.List(ctx, userID, page, size, keyword)
}

func (s *Service) GetSessionDetail(ctx context.Context, userID, sessionID uuid.UUID) (*SessionDetail, error) {
	detail, err := s.repo.GetDetail(ctx, userID, sessionID)
	if err != nil {
		if err == ErrNotFound {
			return nil, httpx.NotFound("session not found")
		}
		return nil, httpx.Internal("failed to load session detail").WithErr(err)
	}
	return detail, nil
}

func (s *Service) DeleteSession(ctx context.Context, userID, sessionID uuid.UUID) error {
	if err := s.repo.Delete(ctx, userID, sessionID); err != nil {
		if err == ErrNotFound {
			return httpx.NotFound("session not found")
		}
		return httpx.Internal("failed to delete session").WithErr(err)
	}
	return nil
}

func StreamNotImplemented(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("event: error\ndata: {\"code\":50100,\"message\":\"streaming chat will be implemented in the next phase\"}\n\n"))
}
