package account

import (
	"context"
	"net/mail"
	"net/url"
	"regexp"
	"strings"
	"time"

	"backend/internal/platform/auth"
	"backend/internal/platform/httpx"

	"github.com/google/uuid"
)

var usernamePattern = regexp.MustCompile(`^[a-zA-Z0-9_]{3,50}$`)

type Service struct {
	repo   *Repository
	tokens *auth.TokenManager
}

func NewService(repo *Repository, tokens *auth.TokenManager) *Service {
	return &Service{
		repo:   repo,
		tokens: tokens,
	}
}

func (s *Service) Register(ctx context.Context, input RegisterInput) (string, error) {
	details := validateRegisterInput(input)
	if len(details) > 0 {
		return "", httpx.ValidationFailed(details...)
	}

	passwordHash, err := auth.HashPassword(strings.TrimSpace(input.Password))
	if err != nil {
		return "", httpx.Internal("failed to hash password").WithErr(err)
	}

	user, err := s.repo.CreateUser(ctx, CreateUserParams{
		Username:     strings.TrimSpace(input.Username),
		Email:        strings.TrimSpace(input.Email),
		PasswordHash: passwordHash,
	})
	if err != nil {
		if err == ErrDuplicateAccount {
			return "", httpx.Conflict("username or email already exists")
		}
		return "", httpx.Internal("failed to create user").WithErr(err)
	}

	return user.ID.String(), nil
}

func (s *Service) Login(ctx context.Context, input LoginInput, meta LoginMeta) (*LoginResult, error) {
	account := strings.TrimSpace(input.Account)
	password := strings.TrimSpace(input.Password)
	if account == "" || password == "" {
		return nil, httpx.ValidationFailed(
			httpx.FieldError{Field: "account", Message: "account is required"},
			httpx.FieldError{Field: "password", Message: "password is required"},
		)
	}

	user, err := s.repo.FindUserByAccount(ctx, account)
	if err != nil {
		if err == ErrNotFound {
			return nil, httpx.InvalidCredentials()
		}
		return nil, httpx.Internal("failed to load user").WithErr(err)
	}

	if user.Status == StatusFrozen {
		return nil, httpx.Unauthorized("account is frozen")
	}

	matched, err := auth.VerifyPassword(password, user.PasswordHash)
	if err != nil {
		return nil, httpx.Internal("failed to verify password").WithErr(err)
	}
	if !matched {
		return nil, httpx.InvalidCredentials()
	}

	rawRefreshToken, refreshHash, err := s.tokens.GenerateRefreshToken()
	if err != nil {
		return nil, httpx.Internal("failed to generate refresh token").WithErr(err)
	}

	expiresAt := time.Now().UTC().Add(s.tokens.RefreshTokenTTL())
	var session *UserSession

	if err := s.repo.WithTx(ctx, func(q dbQuerier) error {
		if err := s.repo.RevokeActiveSessionsByUserID(ctx, q, user.ID); err != nil {
			return err
		}

		session, err = s.repo.CreateSession(ctx, q, CreateSessionParams{
			UserID:           user.ID,
			RefreshTokenHash: refreshHash,
			DeviceLabel:      nilIfEmpty(meta.DeviceLabel),
			UserAgent:        nilIfEmpty(meta.UserAgent),
			IPAddress:        nilIfEmpty(meta.IPAddress),
			ExpiresAt:        expiresAt,
		})
		return err
	}); err != nil {
		return nil, httpx.Internal("failed to create login session").WithErr(err)
	}

	accessToken, err := s.tokens.GenerateAccessToken(user.ID, user.Role, session.ID)
	if err != nil {
		return nil, httpx.Internal("failed to generate access token").WithErr(err)
	}

	return &LoginResult{
		AccessToken:  accessToken.Token,
		ExpiresIn:    int64(time.Until(accessToken.ExpiresAt).Seconds()),
		RefreshToken: rawRefreshToken,
		RefreshUntil: expiresAt,
		User:         *user,
	}, nil
}

func (s *Service) Refresh(ctx context.Context, rawRefreshToken string) (*RefreshResult, error) {
	rawRefreshToken = strings.TrimSpace(rawRefreshToken)
	if rawRefreshToken == "" {
		return nil, httpx.Unauthorized("missing refresh token")
	}

	session, err := s.repo.GetSessionByRefreshHash(ctx, auth.HashToken(rawRefreshToken))
	if err != nil {
		if err == ErrNotFound {
			return nil, httpx.Unauthorized("refresh token is invalid or expired")
		}
		return nil, httpx.Internal("failed to load session").WithErr(err)
	}

	if session.User.Status == StatusFrozen {
		return nil, httpx.Unauthorized("account is frozen")
	}
	if session.Session.RevokedAt != nil {
		return nil, httpx.Unauthorized("refresh token is invalid or expired")
	}
	if time.Now().UTC().After(session.Session.ExpiresAt) {
		return nil, httpx.Unauthorized("refresh token is invalid or expired")
	}

	newRawRefreshToken, newRefreshHash, err := s.tokens.GenerateRefreshToken()
	if err != nil {
		return nil, httpx.Internal("failed to rotate refresh token").WithErr(err)
	}

	expiresAt := time.Now().UTC().Add(s.tokens.RefreshTokenTTL())
	if err := s.repo.WithTx(ctx, func(q dbQuerier) error {
		return s.repo.RotateSession(ctx, q, session.Session.ID, newRefreshHash, expiresAt)
	}); err != nil {
		return nil, httpx.Internal("failed to update session").WithErr(err)
	}

	accessToken, err := s.tokens.GenerateAccessToken(session.User.ID, session.User.Role, session.Session.ID)
	if err != nil {
		return nil, httpx.Internal("failed to generate access token").WithErr(err)
	}

	return &RefreshResult{
		AccessToken:  accessToken.Token,
		ExpiresIn:    int64(time.Until(accessToken.ExpiresAt).Seconds()),
		RefreshToken: newRawRefreshToken,
		RefreshUntil: expiresAt,
		User:         session.User,
	}, nil
}

func (s *Service) Logout(ctx context.Context, sessionID uuid.UUID) error {
	if err := s.repo.RevokeSession(ctx, sessionID); err != nil {
		if err == ErrNotFound {
			return nil
		}
		return httpx.Internal("failed to revoke session").WithErr(err)
	}
	return nil
}

func (s *Service) GetCurrentUser(ctx context.Context, userID uuid.UUID) (*User, error) {
	user, err := s.repo.FindUserByID(ctx, userID)
	if err != nil {
		if err == ErrNotFound {
			return nil, httpx.NotFound("user not found")
		}
		return nil, httpx.Internal("failed to load current user").WithErr(err)
	}
	return user, nil
}

func (s *Service) UpdateCurrentUser(ctx context.Context, userID uuid.UUID, input UpdateProfileInput) error {
	details := validateProfileInput(input)
	if len(details) > 0 {
		return httpx.ValidationFailed(details...)
	}

	if _, err := s.repo.UpdateUserProfile(ctx, userID, normalizeOptionalString(input.Nickname), normalizeOptionalString(input.AvatarURL)); err != nil {
		if err == ErrNotFound {
			return httpx.NotFound("user not found")
		}
		return httpx.Internal("failed to update current user").WithErr(err)
	}
	return nil
}

func (s *Service) ChangePassword(ctx context.Context, userID uuid.UUID, input ChangePasswordInput) error {
	user, err := s.repo.FindUserByID(ctx, userID)
	if err != nil {
		if err == ErrNotFound {
			return httpx.NotFound("user not found")
		}
		return httpx.Internal("failed to load current user").WithErr(err)
	}

	matched, err := auth.VerifyPassword(strings.TrimSpace(input.OldPassword), user.PasswordHash)
	if err != nil {
		return httpx.Internal("failed to verify old password").WithErr(err)
	}
	if !matched {
		return httpx.BadRequest("old password is incorrect")
	}

	if err := validatePassword(user.Username, strings.TrimSpace(input.NewPassword)); err != nil {
		return err
	}

	newHash, err := auth.HashPassword(strings.TrimSpace(input.NewPassword))
	if err != nil {
		return httpx.Internal("failed to hash new password").WithErr(err)
	}

	if err := s.repo.WithTx(ctx, func(q dbQuerier) error {
		if err := s.repo.UpdateUserPassword(ctx, q, userID, newHash); err != nil {
			return err
		}
		return s.repo.RevokeActiveSessionsByUserID(ctx, q, userID)
	}); err != nil {
		return httpx.Internal("failed to change password").WithErr(err)
	}
	return nil
}

func validateRegisterInput(input RegisterInput) []httpx.FieldError {
	var details []httpx.FieldError

	username := strings.TrimSpace(input.Username)
	email := strings.TrimSpace(input.Email)
	password := strings.TrimSpace(input.Password)

	if username == "" {
		details = append(details, httpx.FieldError{Field: "username", Message: "username is required"})
	} else if !usernamePattern.MatchString(username) {
		details = append(details, httpx.FieldError{Field: "username", Message: "username must be 3-50 chars using letters, numbers or underscore"})
	}

	if email == "" {
		details = append(details, httpx.FieldError{Field: "email", Message: "email is required"})
	} else if _, err := mail.ParseAddress(email); err != nil {
		details = append(details, httpx.FieldError{Field: "email", Message: "email is invalid"})
	}

	if err := validatePassword(username, password); err != nil {
		appErr, ok := httpx.AsAppError(err)
		if ok {
			details = append(details, appErr.Details...)
		}
	}

	return details
}

func validateProfileInput(input UpdateProfileInput) []httpx.FieldError {
	var details []httpx.FieldError

	if input.Nickname != nil && len(strings.TrimSpace(*input.Nickname)) > 100 {
		details = append(details, httpx.FieldError{Field: "nickname", Message: "nickname must be <= 100 characters"})
	}

	if input.AvatarURL != nil && strings.TrimSpace(*input.AvatarURL) != "" {
		if _, err := url.ParseRequestURI(strings.TrimSpace(*input.AvatarURL)); err != nil {
			details = append(details, httpx.FieldError{Field: "avatar_url", Message: "avatar_url must be a valid URI"})
		}
	}

	return details
}

func validatePassword(username, password string) error {
	var details []httpx.FieldError

	if len(password) < 8 {
		details = append(details, httpx.FieldError{Field: "password", Message: "password must be at least 8 characters"})
	}
	if !strings.ContainsAny(password, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ") {
		details = append(details, httpx.FieldError{Field: "password", Message: "password must contain letters"})
	}
	if !strings.ContainsAny(password, "0123456789") {
		details = append(details, httpx.FieldError{Field: "password", Message: "password must contain digits"})
	}
	if username != "" && strings.EqualFold(strings.TrimSpace(username), password) {
		details = append(details, httpx.FieldError{Field: "password", Message: "password must not match username"})
	}

	if len(details) > 0 {
		return httpx.ValidationFailed(details...)
	}
	return nil
}

func normalizeOptionalString(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func nilIfEmpty(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
