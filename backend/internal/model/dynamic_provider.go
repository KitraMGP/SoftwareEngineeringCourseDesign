package model

import (
	"context"

	"backend/internal/platform/config"
)

type EmbeddingAPIKeyResolver func(context.Context) (string, error)

type DynamicEmbeddingProvider struct {
	baseCfg       config.AIConfig
	resolveAPIKey EmbeddingAPIKeyResolver
}

func NewDynamicEmbeddingProvider(baseCfg config.AIConfig, resolveAPIKey EmbeddingAPIKeyResolver) *DynamicEmbeddingProvider {
	return &DynamicEmbeddingProvider{
		baseCfg:       baseCfg,
		resolveAPIKey: resolveAPIKey,
	}
}

func (p *DynamicEmbeddingProvider) Embed(ctx context.Context, req EmbeddingRequest) (EmbeddingResult, error) {
	cfg := p.baseCfg
	if p.resolveAPIKey != nil {
		apiKey, err := p.resolveAPIKey(ctx)
		if err != nil {
			return EmbeddingResult{}, &ProviderError{
				Kind:    ProviderErrorUnavailable,
				Message: "failed to load embedding provider configuration",
				Err:     err,
			}
		}
		cfg.EmbeddingAPIKey = apiKey
	}

	return NewEmbeddingProvider(cfg).Embed(ctx, req)
}
