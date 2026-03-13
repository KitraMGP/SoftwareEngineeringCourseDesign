package model

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"backend/internal/platform/config"
)

type ChatMessage struct {
	Role    string
	Content string
}

type ChatRequest struct {
	Model    string
	Messages []ChatMessage
}

type ChatDelta struct {
	Content string
}

type StreamResult interface {
	Recv() (ChatDelta, error)
	Close() error
}

type EmbeddingRequest struct {
	Model string
	Texts []string
}

type EmbeddingResult struct {
	Vectors [][]float32
}

type ChatProvider interface {
	StreamChat(ctx context.Context, req ChatRequest) (StreamResult, error)
}

type EmbeddingProvider interface {
	Embed(ctx context.Context, req EmbeddingRequest) (EmbeddingResult, error)
}

const EmbeddingDimension = 1536

type ProviderErrorKind string

const (
	ProviderErrorMisconfigured ProviderErrorKind = "misconfigured"
	ProviderErrorUnavailable   ProviderErrorKind = "unavailable"
	ProviderErrorAuthFailed    ProviderErrorKind = "auth_failed"
	ProviderErrorRateLimited   ProviderErrorKind = "rate_limited"
	ProviderErrorBadRequest    ProviderErrorKind = "bad_request"
)

type ProviderError struct {
	Kind    ProviderErrorKind
	Message string
	Err     error
}

func (e *ProviderError) Error() string {
	return e.Message
}

func (e *ProviderError) Unwrap() error {
	return e.Err
}

func AsProviderError(err error) (*ProviderError, bool) {
	var providerErr *ProviderError
	if errors.As(err, &providerErr) {
		return providerErr, true
	}
	return nil, false
}

func NewEmbeddingProvider(cfg config.AIConfig) EmbeddingProvider {
	switch strings.ToLower(strings.TrimSpace(cfg.EmbeddingProvider)) {
	case "", "local_hash":
		return NewLocalHashEmbeddingProvider(EmbeddingDimension)
	case "openai_compatible":
		return NewOpenAICompatibleEmbeddingProvider(cfg.EmbeddingBaseURL, cfg.EmbeddingAPIKey, cfg.EmbeddingTimeout)
	default:
		return NewDisabledEmbeddingProvider(fmt.Sprintf("unsupported embedding provider %q", cfg.EmbeddingProvider))
	}
}

type DisabledEmbeddingProvider struct {
	reason string
}

func NewDisabledEmbeddingProvider(reason string) *DisabledEmbeddingProvider {
	return &DisabledEmbeddingProvider{reason: reason}
}

func (p *DisabledEmbeddingProvider) Embed(context.Context, EmbeddingRequest) (EmbeddingResult, error) {
	return EmbeddingResult{}, &ProviderError{
		Kind:    ProviderErrorMisconfigured,
		Message: p.reason,
	}
}
