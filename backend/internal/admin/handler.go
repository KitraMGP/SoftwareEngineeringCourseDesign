package admin

import (
	"net/http"

	"backend/internal/platform/httpx"

	"github.com/go-chi/chi/v5"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/users", httpx.Adapt(h.notImplemented("admin user management")))
	r.Post("/users", httpx.Adapt(h.notImplemented("admin user creation")))
	r.Put("/users/{userId}", httpx.Adapt(h.notImplemented("admin user update")))
	r.Put("/users/{userId}/password", httpx.Adapt(h.notImplemented("admin password reset")))
	r.Post("/users/{userId}/freeze", httpx.Adapt(h.notImplemented("admin freeze user")))
	r.Post("/users/{userId}/unfreeze", httpx.Adapt(h.notImplemented("admin unfreeze user")))
	r.Get("/provider-configs", httpx.Adapt(h.notImplemented("provider config management")))
	r.Put("/provider-configs/{provider}", httpx.Adapt(h.notImplemented("provider config management")))
	r.Get("/settings", httpx.Adapt(h.notImplemented("system settings management")))
	r.Put("/settings", httpx.Adapt(h.notImplemented("system settings management")))
	r.Get("/tasks", httpx.Adapt(h.notImplemented("task management")))
	r.Post("/tasks/{taskId}/retry", httpx.Adapt(h.notImplemented("task retry")))
	r.Get("/audit-logs", httpx.Adapt(h.notImplemented("audit log query")))
	r.Get("/quota-policies", httpx.Adapt(h.notImplemented("quota policy management")))
	r.Put("/quota-policies/{policyId}", httpx.Adapt(h.notImplemented("quota policy management")))
	r.Get("/usage/users/{userId}", httpx.Adapt(h.notImplemented("usage query")))
}

func (h *Handler) notImplemented(feature string) func(http.ResponseWriter, *http.Request) error {
	return func(w http.ResponseWriter, r *http.Request) error {
		return httpx.FeatureNotReady(feature + " will be implemented in a later phase")
	}
}
