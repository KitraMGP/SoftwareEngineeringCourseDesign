package admin

import (
	"net/http"

	"backend/internal/platform/auth"
	"backend/internal/platform/httpx"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Handler struct {
	service *Service
}

type updateProviderAPIKeyRequest struct {
	APIKey string `json:"api_key"`
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/overview", httpx.Adapt(h.GetOverview))
	r.Get("/users", httpx.Adapt(h.ListUsers))
	r.Post("/users", httpx.Adapt(h.notImplemented("admin user creation")))
	r.Put("/users/{userId}", httpx.Adapt(h.notImplemented("admin user update")))
	r.Put("/users/{userId}/password", httpx.Adapt(h.notImplemented("admin password reset")))
	r.Post("/users/{userId}/freeze", httpx.Adapt(h.FreezeUser))
	r.Post("/users/{userId}/unfreeze", httpx.Adapt(h.UnfreezeUser))
	r.Get("/provider-configs", httpx.Adapt(h.ListProviderConfigs))
	r.Put("/provider-configs/{provider}", httpx.Adapt(h.UpdateProviderAPIKey))
	r.Get("/settings", httpx.Adapt(h.ListSystemSettings))
	r.Put("/settings", httpx.Adapt(h.notImplemented("system settings management")))
	r.Get("/tasks", httpx.Adapt(h.ListTasks))
	r.Post("/tasks/{taskId}/retry", httpx.Adapt(h.RetryTask))
	r.Get("/audit-logs", httpx.Adapt(h.ListAuditLogs))
	r.Get("/quota-policies", httpx.Adapt(h.ListQuotaPolicies))
	r.Put("/quota-policies/{policyId}", httpx.Adapt(h.notImplemented("quota policy management")))
	r.Get("/usage/users/{userId}", httpx.Adapt(h.notImplemented("usage query")))
}

func (h *Handler) GetOverview(w http.ResponseWriter, r *http.Request) error {
	result, err := h.service.GetOverview(r.Context())
	if err != nil {
		return err
	}
	httpx.Success(w, http.StatusOK, result)
	return nil
}

func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) error {
	page, size := httpx.ParsePageSize(r, 20, 100)
	result, err := h.service.ListUsers(
		r.Context(),
		page,
		size,
		r.URL.Query().Get("keyword"),
		r.URL.Query().Get("status"),
		r.URL.Query().Get("role"),
	)
	if err != nil {
		return err
	}
	httpx.Success(w, http.StatusOK, result)
	return nil
}

func (h *Handler) FreezeUser(w http.ResponseWriter, r *http.Request) error {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		return httpx.Unauthorized("missing auth context")
	}

	userID, err := uuid.Parse(chi.URLParam(r, "userId"))
	if err != nil {
		return httpx.BadRequest("invalid user id")
	}

	result, err := h.service.FreezeUser(r.Context(), principal, userID)
	if err != nil {
		return err
	}
	httpx.Success(w, http.StatusOK, result)
	return nil
}

func (h *Handler) UnfreezeUser(w http.ResponseWriter, r *http.Request) error {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		return httpx.Unauthorized("missing auth context")
	}

	userID, err := uuid.Parse(chi.URLParam(r, "userId"))
	if err != nil {
		return httpx.BadRequest("invalid user id")
	}

	result, err := h.service.UnfreezeUser(r.Context(), principal, userID)
	if err != nil {
		return err
	}
	httpx.Success(w, http.StatusOK, result)
	return nil
}

func (h *Handler) ListTasks(w http.ResponseWriter, r *http.Request) error {
	page, size := httpx.ParsePageSize(r, 20, 100)
	result, err := h.service.ListTasks(
		r.Context(),
		page,
		size,
		r.URL.Query().Get("status"),
		r.URL.Query().Get("task_type"),
	)
	if err != nil {
		return err
	}
	httpx.Success(w, http.StatusOK, result)
	return nil
}

func (h *Handler) RetryTask(w http.ResponseWriter, r *http.Request) error {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		return httpx.Unauthorized("missing auth context")
	}

	taskID, err := uuid.Parse(chi.URLParam(r, "taskId"))
	if err != nil {
		return httpx.BadRequest("invalid task id")
	}

	result, err := h.service.RetryTask(r.Context(), principal, taskID)
	if err != nil {
		return err
	}
	httpx.Success(w, http.StatusOK, result)
	return nil
}

func (h *Handler) ListProviderConfigs(w http.ResponseWriter, r *http.Request) error {
	items, err := h.service.ListProviderConfigs(r.Context())
	if err != nil {
		return err
	}
	httpx.Success(w, http.StatusOK, map[string]any{"items": items})
	return nil
}

func (h *Handler) UpdateProviderAPIKey(w http.ResponseWriter, r *http.Request) error {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		return httpx.Unauthorized("missing auth context")
	}

	var req updateProviderAPIKeyRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		return err
	}

	item, err := h.service.UpdateProviderAPIKey(r.Context(), principal, chi.URLParam(r, "provider"), req.APIKey)
	if err != nil {
		return err
	}

	httpx.Success(w, http.StatusOK, item)
	return nil
}

func (h *Handler) ListSystemSettings(w http.ResponseWriter, r *http.Request) error {
	items, err := h.service.ListSystemSettings(r.Context())
	if err != nil {
		return err
	}
	httpx.Success(w, http.StatusOK, map[string]any{"items": items})
	return nil
}

func (h *Handler) ListQuotaPolicies(w http.ResponseWriter, r *http.Request) error {
	items, err := h.service.ListQuotaPolicies(r.Context())
	if err != nil {
		return err
	}
	httpx.Success(w, http.StatusOK, map[string]any{"items": items})
	return nil
}

func (h *Handler) ListAuditLogs(w http.ResponseWriter, r *http.Request) error {
	page, size := httpx.ParsePageSize(r, 20, 100)
	result, err := h.service.ListAuditLogs(r.Context(), page, size)
	if err != nil {
		return err
	}
	httpx.Success(w, http.StatusOK, result)
	return nil
}

func (h *Handler) notImplemented(feature string) func(http.ResponseWriter, *http.Request) error {
	return func(w http.ResponseWriter, r *http.Request) error {
		return httpx.FeatureNotReady(feature + " will be implemented in a later phase")
	}
}
