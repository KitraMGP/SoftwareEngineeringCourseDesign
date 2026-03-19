package chat

import (
	"context"

	"backend/internal/platform/config"
)

type APIKeyResolver func(context.Context) (string, error)

type DynamicProvider struct {
	baseCfg       config.AIConfig
	resolveAPIKey APIKeyResolver
}

func NewDynamicProvider(baseCfg config.AIConfig, resolveAPIKey APIKeyResolver) *DynamicProvider {
	return &DynamicProvider{
		baseCfg:       baseCfg,
		resolveAPIKey: resolveAPIKey,
	}
}

func (p *DynamicProvider) StreamChat(ctx context.Context, req ProviderRequest, onDelta func(string) error) (*CompletionResult, error) {
	cfg := p.baseCfg
	if p.resolveAPIKey != nil {
		apiKey, err := p.resolveAPIKey(ctx)
		if err != nil {
			return nil, &ProviderError{
				Kind:    ProviderErrorUnavailable,
				Message: "failed to load chat provider configuration",
				Err:     err,
			}
		}
		cfg.APIKey = apiKey
	}

	return NewProvider(cfg).StreamChat(ctx, req, onDelta)
}
