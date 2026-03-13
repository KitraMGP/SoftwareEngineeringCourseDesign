package chat

import (
	"net/http"
	"time"

	"backend/internal/platform/auth"
	"backend/internal/platform/httpx"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Handler struct {
	service           *Service
	heartbeatInterval time.Duration
}

type createSessionRequest struct {
	Name            *string `json:"name"`
	Model           string  `json:"model"`
	KnowledgeBaseID *string `json:"knowledge_base_id"`
}

type sendMessageRequest struct {
	Content string `json:"content"`
}

func NewHandler(service *Service, heartbeatInterval time.Duration) *Handler {
	return &Handler{
		service:           service,
		heartbeatInterval: heartbeatInterval,
	}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/sessions", func(r chi.Router) {
		r.Get("/", httpx.Adapt(h.ListSessions))
		r.Post("/", httpx.Adapt(h.CreateSession))
		r.Get("/{sessionId}", httpx.Adapt(h.GetSessionDetail))
		r.Delete("/{sessionId}", httpx.Adapt(h.DeleteSession))
		r.Post("/{sessionId}/messages", httpx.Adapt(h.SendMessage))
		r.Post("/{sessionId}/messages/{messageId}/regenerate", httpx.Adapt(h.RegenerateMessage))
		r.Post("/{sessionId}/stream/stop", httpx.Adapt(h.StopStream))
	})
}

func (h *Handler) CreateSession(w http.ResponseWriter, r *http.Request) error {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		return httpx.Unauthorized("missing auth context")
	}

	var req createSessionRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		return err
	}

	var knowledgeBaseID *uuid.UUID
	if req.KnowledgeBaseID != nil && *req.KnowledgeBaseID != "" {
		parsed, err := uuid.Parse(*req.KnowledgeBaseID)
		if err != nil {
			return httpx.ValidationFailed(httpx.FieldError{Field: "knowledge_base_id", Message: "knowledge_base_id must be a valid UUID"})
		}
		knowledgeBaseID = &parsed
	}

	session, err := h.service.CreateSession(r.Context(), principal.UserID, CreateSessionInput{
		Name:            req.Name,
		Model:           req.Model,
		KnowledgeBaseID: knowledgeBaseID,
	})
	if err != nil {
		return err
	}

	httpx.Success(w, http.StatusOK, map[string]any{
		"session_id": session.ID,
	})
	return nil
}

func (h *Handler) ListSessions(w http.ResponseWriter, r *http.Request) error {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		return httpx.Unauthorized("missing auth context")
	}

	page, size := httpx.ParsePageSize(r, 20, 100)
	result, err := h.service.ListSessions(r.Context(), principal.UserID, page, size, r.URL.Query().Get("keyword"))
	if err != nil {
		return httpx.Internal("failed to list sessions").WithErr(err)
	}

	httpx.Success(w, http.StatusOK, result)
	return nil
}

func (h *Handler) GetSessionDetail(w http.ResponseWriter, r *http.Request) error {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		return httpx.Unauthorized("missing auth context")
	}

	sessionID, err := uuid.Parse(chi.URLParam(r, "sessionId"))
	if err != nil {
		return httpx.BadRequest("invalid session id")
	}

	detail, err := h.service.GetSessionDetail(r.Context(), principal.UserID, sessionID)
	if err != nil {
		return err
	}

	httpx.Success(w, http.StatusOK, detail)
	return nil
}

func (h *Handler) DeleteSession(w http.ResponseWriter, r *http.Request) error {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		return httpx.Unauthorized("missing auth context")
	}

	sessionID, err := uuid.Parse(chi.URLParam(r, "sessionId"))
	if err != nil {
		return httpx.BadRequest("invalid session id")
	}

	if err := h.service.DeleteSession(r.Context(), principal.UserID, sessionID); err != nil {
		return err
	}

	httpx.Success(w, http.StatusOK, nil)
	return nil
}

func (h *Handler) SendMessage(w http.ResponseWriter, r *http.Request) error {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		return httpx.Unauthorized("missing auth context")
	}

	sessionID, err := uuid.Parse(chi.URLParam(r, "sessionId"))
	if err != nil {
		return httpx.BadRequest("invalid session id")
	}

	var req sendMessageRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		return err
	}

	stream, err := NewSSEWriter(w, h.heartbeatInterval)
	if err != nil {
		return httpx.Internal("failed to initialize sse stream").WithErr(err)
	}

	return h.service.SendMessageStream(r.Context(), principal.UserID, sessionID, req.Content, stream)
}

func (h *Handler) RegenerateMessage(w http.ResponseWriter, r *http.Request) error {
	return httpx.FeatureNotReady("message regeneration will be implemented in the next phase")
}

func (h *Handler) StopStream(w http.ResponseWriter, r *http.Request) error {
	return httpx.FeatureNotReady("stream stopping will be implemented in the next phase")
}
