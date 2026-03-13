package account

import (
	"net/http"
	"time"

	"backend/internal/platform/auth"
	"backend/internal/platform/config"
	"backend/internal/platform/httpx"
)

type Handler struct {
	service *Service
	authCfg config.AuthConfig
}

type registerRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRequest struct {
	Account  string `json:"account"`
	Password string `json:"password"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type updateProfileRequest struct {
	Nickname  *string `json:"nickname"`
	AvatarURL *string `json:"avatar_url"`
}

type changePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

func NewHandler(service *Service, authCfg config.AuthConfig) *Handler {
	return &Handler{
		service: service,
		authCfg: authCfg,
	}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) error {
	var req registerRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		return err
	}

	userID, err := h.service.Register(r.Context(), RegisterInput{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		return err
	}

	httpx.Success(w, http.StatusOK, map[string]any{
		"user_id": userID,
	})
	return nil
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) error {
	var req loginRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		return err
	}

	result, err := h.service.Login(r.Context(), LoginInput{
		Account:  req.Account,
		Password: req.Password,
	}, LoginMeta{
		UserAgent: r.UserAgent(),
		IPAddress: httpx.ClientIP(r),
	})
	if err != nil {
		return err
	}

	h.writeRefreshCookie(w, result.RefreshToken, result.RefreshUntil)
	httpx.Success(w, http.StatusOK, map[string]any{
		"access_token": result.AccessToken,
		"expires_in":   result.ExpiresIn,
		"user": map[string]any{
			"id":         result.User.ID,
			"username":   result.User.Username,
			"email":      result.User.Email,
			"nickname":   result.User.Nickname,
			"avatar_url": result.User.AvatarURL,
			"role":       result.User.Role,
			"status":     result.User.Status,
		},
	})
	return nil
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) error {
	rawRefreshToken := h.readRefreshToken(r)
	if rawRefreshToken == "" {
		var req refreshRequest
		if err := httpx.DecodeJSON(r, &req); err == nil {
			rawRefreshToken = req.RefreshToken
		}
	}

	result, err := h.service.Refresh(r.Context(), rawRefreshToken)
	if err != nil {
		return err
	}

	h.writeRefreshCookie(w, result.RefreshToken, result.RefreshUntil)
	httpx.Success(w, http.StatusOK, map[string]any{
		"access_token": result.AccessToken,
		"expires_in":   result.ExpiresIn,
	})
	return nil
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) error {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		return httpx.Unauthorized("missing auth context")
	}

	if err := h.service.Logout(r.Context(), principal.SessionID); err != nil {
		return err
	}

	h.clearRefreshCookie(w)
	httpx.Success(w, http.StatusOK, nil)
	return nil
}

func (h *Handler) GetCurrentUser(w http.ResponseWriter, r *http.Request) error {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		return httpx.Unauthorized("missing auth context")
	}

	user, err := h.service.GetCurrentUser(r.Context(), principal.UserID)
	if err != nil {
		return err
	}

	httpx.Success(w, http.StatusOK, map[string]any{
		"id":         user.ID,
		"username":   user.Username,
		"email":      user.Email,
		"nickname":   user.Nickname,
		"avatar_url": user.AvatarURL,
		"role":       user.Role,
		"status":     user.Status,
		"created_at": user.CreatedAt,
		"updated_at": user.UpdatedAt,
	})
	return nil
}

func (h *Handler) UpdateCurrentUser(w http.ResponseWriter, r *http.Request) error {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		return httpx.Unauthorized("missing auth context")
	}

	var req updateProfileRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		return err
	}

	if err := h.service.UpdateCurrentUser(r.Context(), principal.UserID, UpdateProfileInput{
		Nickname:  req.Nickname,
		AvatarURL: req.AvatarURL,
	}); err != nil {
		return err
	}

	httpx.Success(w, http.StatusOK, nil)
	return nil
}

func (h *Handler) ChangePassword(w http.ResponseWriter, r *http.Request) error {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		return httpx.Unauthorized("missing auth context")
	}

	var req changePasswordRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		return err
	}

	if err := h.service.ChangePassword(r.Context(), principal.UserID, ChangePasswordInput{
		OldPassword: req.OldPassword,
		NewPassword: req.NewPassword,
	}); err != nil {
		return err
	}

	h.clearRefreshCookie(w)
	httpx.Success(w, http.StatusOK, nil)
	return nil
}

func (h *Handler) readRefreshToken(r *http.Request) string {
	cookie, err := r.Cookie(h.authCfg.RefreshCookieName)
	if err != nil {
		return ""
	}
	return cookie.Value
}

func (h *Handler) writeRefreshCookie(w http.ResponseWriter, token string, expiresAt time.Time) {
	cookie := &http.Cookie{
		Name:     h.authCfg.RefreshCookieName,
		Value:    token,
		Path:     h.authCfg.RefreshCookiePath,
		HttpOnly: true,
		Secure:   h.authCfg.RefreshCookieSecure,
		SameSite: http.SameSiteLaxMode,
		Expires:  expiresAt,
	}
	if h.authCfg.RefreshCookieDomain != "" {
		cookie.Domain = h.authCfg.RefreshCookieDomain
	}
	http.SetCookie(w, cookie)
}

func (h *Handler) clearRefreshCookie(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:     h.authCfg.RefreshCookieName,
		Value:    "",
		Path:     h.authCfg.RefreshCookiePath,
		HttpOnly: true,
		Secure:   h.authCfg.RefreshCookieSecure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	}
	if h.authCfg.RefreshCookieDomain != "" {
		cookie.Domain = h.authCfg.RefreshCookieDomain
	}
	http.SetCookie(w, cookie)
}
