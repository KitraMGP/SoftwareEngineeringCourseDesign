package storage

import (
	"fmt"

	"backend/internal/platform/config"
)

func NewFromConfig(cfg config.StorageConfig) (Service, error) {
	switch cfg.Provider {
	case ProviderLocalFS:
		return NewLocalService(cfg.Provider, cfg.Bucket, cfg.LocalRoot)
	default:
		return nil, fmt.Errorf("unsupported storage provider %q", cfg.Provider)
	}
}
