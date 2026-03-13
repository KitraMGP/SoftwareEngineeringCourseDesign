package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var ErrInvalidToken = errors.New("invalid token")

type Principal struct {
	UserID    uuid.UUID
	Role      string
	SessionID uuid.UUID
}

type AccessClaims struct {
	Role      string `json:"role"`
	SessionID string `json:"session_id"`
	jwt.RegisteredClaims
}

type AccessTokenResult struct {
	Token     string
	ExpiresAt time.Time
}

type TokenManager struct {
	secret          []byte
	issuer          string
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

func NewTokenManager(secret, issuer string, accessTokenTTL, refreshTokenTTL time.Duration) *TokenManager {
	return &TokenManager{
		secret:          []byte(secret),
		issuer:          issuer,
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
	}
}

func (m *TokenManager) GenerateAccessToken(userID uuid.UUID, role string, sessionID uuid.UUID) (AccessTokenResult, error) {
	now := time.Now().UTC()
	expiresAt := now.Add(m.accessTokenTTL)

	claims := AccessClaims{
		Role:      role,
		SessionID: sessionID.String(),
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			Issuer:    m.issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
	if err != nil {
		return AccessTokenResult{}, fmt.Errorf("sign jwt: %w", err)
	}

	return AccessTokenResult{
		Token:     token,
		ExpiresAt: expiresAt,
	}, nil
}

func (m *TokenManager) ParseAccessToken(raw string) (Principal, error) {
	token, err := jwt.ParseWithClaims(raw, &AccessClaims{}, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, ErrInvalidToken
		}
		return m.secret, nil
	}, jwt.WithIssuer(m.issuer))
	if err != nil {
		return Principal{}, ErrInvalidToken
	}

	claims, ok := token.Claims.(*AccessClaims)
	if !ok || !token.Valid {
		return Principal{}, ErrInvalidToken
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return Principal{}, ErrInvalidToken
	}

	sessionID, err := uuid.Parse(claims.SessionID)
	if err != nil {
		return Principal{}, ErrInvalidToken
	}

	return Principal{
		UserID:    userID,
		Role:      claims.Role,
		SessionID: sessionID,
	}, nil
}

func (m *TokenManager) GenerateRefreshToken() (raw string, hash string, err error) {
	bytes := make([]byte, 32)
	if _, err = rand.Read(bytes); err != nil {
		return "", "", fmt.Errorf("generate refresh token: %w", err)
	}

	raw = hex.EncodeToString(bytes)
	return raw, HashToken(raw), nil
}

func (m *TokenManager) RefreshTokenTTL() time.Duration {
	return m.refreshTokenTTL
}

func HashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
