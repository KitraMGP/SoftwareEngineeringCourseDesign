package account

import (
	"time"

	"github.com/google/uuid"
)

const (
	RoleUser  = "user"
	RoleAdmin = "admin"

	StatusActive = "active"
	StatusFrozen = "frozen"
)

type User struct {
	ID           uuid.UUID  `json:"id"`
	Username     string     `json:"username"`
	Email        string     `json:"email"`
	Nickname     *string    `json:"nickname,omitempty"`
	AvatarURL    *string    `json:"avatar_url,omitempty"`
	Role         string     `json:"role"`
	Status       string     `json:"status"`
	PasswordHash string     `json:"-"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"-"`
}

type UserSession struct {
	ID               uuid.UUID  `json:"id"`
	UserID           uuid.UUID  `json:"user_id"`
	RefreshTokenHash string     `json:"-"`
	DeviceLabel      *string    `json:"device_label,omitempty"`
	UserAgent        *string    `json:"user_agent,omitempty"`
	IPAddress        *string    `json:"ip_address,omitempty"`
	ExpiresAt        time.Time  `json:"expires_at"`
	LastActiveAt     *time.Time `json:"last_active_at,omitempty"`
	RevokedAt        *time.Time `json:"revoked_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type AuthenticatedSession struct {
	User    User
	Session UserSession
}

type RegisterInput struct {
	Username string
	Email    string
	Password string
}

type LoginInput struct {
	Account  string
	Password string
}

type LoginMeta struct {
	UserAgent   string
	IPAddress   string
	DeviceLabel string
}

type UpdateProfileInput struct {
	Nickname  *string
	AvatarURL *string
}

type ChangePasswordInput struct {
	OldPassword string
	NewPassword string
}

type LoginResult struct {
	AccessToken  string
	ExpiresIn    int64
	RefreshToken string
	RefreshUntil time.Time
	User         User
}

type RefreshResult struct {
	AccessToken  string
	ExpiresIn    int64
	RefreshToken string
	RefreshUntil time.Time
	User         User
}
