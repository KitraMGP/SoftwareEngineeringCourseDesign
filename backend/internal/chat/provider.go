package chat

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"backend/internal/platform/config"
)

type ProviderMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ProviderRequest struct {
	Model       string
	Messages    []ProviderMessage
	Temperature float64
}

type Usage struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

type CompletionResult struct {
	Model        string
	FinishReason string
	Usage        Usage
}

type Provider interface {
	StreamChat(ctx context.Context, req ProviderRequest, onDelta func(string) error) (*CompletionResult, error)
}

type ProviderErrorKind string

const (
	ProviderErrorMisconfigured ProviderErrorKind = "misconfigured"
	ProviderErrorUnavailable   ProviderErrorKind = "unavailable"
	ProviderErrorAuthFailed    ProviderErrorKind = "auth_failed"
	ProviderErrorRateLimited   ProviderErrorKind = "rate_limited"
	ProviderErrorPromptTooLong ProviderErrorKind = "prompt_too_long"
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

func NewProvider(cfg config.AIConfig) Provider {
	switch strings.ToLower(strings.TrimSpace(cfg.Provider)) {
	case "", "deepseek":
		if strings.TrimSpace(cfg.APIKey) == "" {
			return &DisabledProvider{reason: "DeepSeek API key is not configured"}
		}
		return NewDeepSeekProvider(cfg.BaseURL, cfg.APIKey)
	default:
		return &DisabledProvider{reason: fmt.Sprintf("unsupported AI provider %q", cfg.Provider)}
	}
}

type DisabledProvider struct {
	reason string
}

func (p *DisabledProvider) StreamChat(context.Context, ProviderRequest, func(string) error) (*CompletionResult, error) {
	return nil, &ProviderError{
		Kind:    ProviderErrorMisconfigured,
		Message: p.reason,
	}
}
