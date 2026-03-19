package providerconfig

import (
	"context"
	"fmt"
	"strings"
	"time"

	"backend/internal/platform/config"
	"backend/internal/platform/httpx"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	APIKeySourceMissing     = "missing"
	APIKeySourceEnvironment = "environment"
	APIKeySourceDatabase    = "database"
	defaultProviderName     = "deepseek"
)

type StoredConfig struct {
	ID                    uuid.UUID
	Provider              string
	BaseURL               string
	DefaultChatModel      string
	DefaultEmbeddingModel string
	EncryptedAPIKey       string
	IsEnabled             bool
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

type ViewConfig struct {
	ID                    *uuid.UUID
	Provider              string
	BaseURL               string
	DefaultChatModel      string
	DefaultEmbeddingModel string
	IsEnabled             bool
	HasAPIKey             bool
	APIKeySource          string
	CreatedAt             *time.Time
	UpdatedAt             *time.Time
}

type Repository struct {
	pool *pgxpool.Pool
}

type Manager struct {
	repo *Repository
	cfg  config.AIConfig
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func NewManager(repo *Repository, cfg config.AIConfig) *Manager {
	return &Manager{repo: repo, cfg: cfg}
}

func (r *Repository) List(ctx context.Context) ([]StoredConfig, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, provider, base_url, default_chat_model, default_embedding_model, encrypted_api_key, is_enabled, created_at, updated_at
		FROM provider_configs
		ORDER BY provider ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("list provider configs: %w", err)
	}
	defer rows.Close()

	items := make([]StoredConfig, 0)
	for rows.Next() {
		var item StoredConfig
		if err := rows.Scan(
			&item.ID,
			&item.Provider,
			&item.BaseURL,
			&item.DefaultChatModel,
			&item.DefaultEmbeddingModel,
			&item.EncryptedAPIKey,
			&item.IsEnabled,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan provider config: %w", err)
		}
		items = append(items, normalizeStoredConfig(item))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate provider configs: %w", err)
	}
	return items, nil
}

func (r *Repository) GetByProvider(ctx context.Context, provider string) (*StoredConfig, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, provider, base_url, default_chat_model, default_embedding_model, encrypted_api_key, is_enabled, created_at, updated_at
		FROM provider_configs
		WHERE lower(provider) = lower($1)
	`, provider)

	var item StoredConfig
	if err := row.Scan(
		&item.ID,
		&item.Provider,
		&item.BaseURL,
		&item.DefaultChatModel,
		&item.DefaultEmbeddingModel,
		&item.EncryptedAPIKey,
		&item.IsEnabled,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get provider config: %w", err)
	}

	normalized := normalizeStoredConfig(item)
	return &normalized, nil
}

func (r *Repository) Upsert(ctx context.Context, item StoredConfig) (*StoredConfig, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO provider_configs (
			provider,
			base_url,
			default_chat_model,
			default_embedding_model,
			encrypted_api_key,
			is_enabled
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (provider) DO UPDATE
		SET
			base_url = EXCLUDED.base_url,
			default_chat_model = EXCLUDED.default_chat_model,
			default_embedding_model = EXCLUDED.default_embedding_model,
			encrypted_api_key = EXCLUDED.encrypted_api_key,
			is_enabled = EXCLUDED.is_enabled
		RETURNING id, provider, base_url, default_chat_model, default_embedding_model, encrypted_api_key, is_enabled, created_at, updated_at
	`, item.Provider, item.BaseURL, item.DefaultChatModel, item.DefaultEmbeddingModel, item.EncryptedAPIKey, item.IsEnabled)

	var saved StoredConfig
	if err := row.Scan(
		&saved.ID,
		&saved.Provider,
		&saved.BaseURL,
		&saved.DefaultChatModel,
		&saved.DefaultEmbeddingModel,
		&saved.EncryptedAPIKey,
		&saved.IsEnabled,
		&saved.CreatedAt,
		&saved.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("upsert provider config: %w", err)
	}

	normalized := normalizeStoredConfig(saved)
	return &normalized, nil
}

func (m *Manager) List(ctx context.Context) ([]ViewConfig, error) {
	stored, err := m.repo.GetByProvider(ctx, defaultProviderName)
	if err != nil {
		return nil, err
	}
	return []ViewConfig{m.toViewConfig(defaultProviderName, stored)}, nil
}

func (m *Manager) UpdateAPIKey(ctx context.Context, provider, apiKey string) (*ViewConfig, error) {
	normalizedProvider := normalizeProvider(provider)
	if normalizedProvider == "" {
		return nil, httpx.ValidationFailed(httpx.FieldError{Field: "provider", Message: "provider is required"})
	}
	if normalizedProvider != defaultProviderName {
		return nil, httpx.ValidationFailed(httpx.FieldError{Field: "provider", Message: "only DeepSeek is supported"})
	}

	trimmedAPIKey := strings.TrimSpace(apiKey)
	if trimmedAPIKey == "" {
		return nil, httpx.ValidationFailed(httpx.FieldError{Field: "api_key", Message: "api_key is required"})
	}

	current, err := m.repo.GetByProvider(ctx, normalizedProvider)
	if err != nil {
		return nil, err
	}

	defaults := m.defaultsForProvider(normalizedProvider)
	item := StoredConfig{
		Provider:              normalizedProvider,
		BaseURL:               coalesceString(currentValue(current, func(cfg StoredConfig) string { return cfg.BaseURL }), defaults.baseURL),
		DefaultChatModel:      coalesceString(currentValue(current, func(cfg StoredConfig) string { return cfg.DefaultChatModel }), defaults.defaultChatModel),
		DefaultEmbeddingModel: coalesceString(currentValue(current, func(cfg StoredConfig) string { return cfg.DefaultEmbeddingModel }), defaults.defaultEmbeddingModel),
		EncryptedAPIKey:       trimmedAPIKey,
		IsEnabled:             currentBool(current, defaults.isEnabled),
	}

	saved, err := m.repo.Upsert(ctx, item)
	if err != nil {
		return nil, err
	}

	view := m.toViewConfig(normalizedProvider, saved)
	return &view, nil
}

func (m *Manager) ResolveChatAPIKey(ctx context.Context) (string, error) {
	return m.resolveAPIKey(ctx, defaultProviderName, strings.TrimSpace(m.cfg.APIKey))
}

func (m *Manager) ResolveEmbeddingAPIKey(ctx context.Context) (string, error) {
	provider := normalizeProvider(m.cfg.EmbeddingProvider)
	if provider == "" || provider == "local_hash" {
		return strings.TrimSpace(m.cfg.EmbeddingAPIKey), nil
	}
	return m.resolveAPIKey(ctx, provider, strings.TrimSpace(m.cfg.EmbeddingAPIKey))
}

func (m *Manager) resolveAPIKey(ctx context.Context, provider, fallback string) (string, error) {
	if strings.TrimSpace(fallback) != "" {
		return strings.TrimSpace(fallback), nil
	}

	normalizedProvider := normalizeProvider(provider)
	if normalizedProvider == "" {
		return "", nil
	}

	item, err := m.repo.GetByProvider(ctx, normalizedProvider)
	if err != nil {
		return "", err
	}
	if item != nil && strings.TrimSpace(item.EncryptedAPIKey) != "" {
		return strings.TrimSpace(item.EncryptedAPIKey), nil
	}
	return strings.TrimSpace(fallback), nil
}

type providerDefaults struct {
	baseURL               string
	defaultChatModel      string
	defaultEmbeddingModel string
	environmentAPIKey     string
	isEnabled             bool
}

func (m *Manager) defaultsForProvider(provider string) providerDefaults {
	normalizedProvider := normalizeProvider(provider)
	defaults := providerDefaults{
		isEnabled: true,
	}

	if normalizedProvider == normalizeProvider(m.cfg.Provider) {
		defaults.baseURL = strings.TrimSpace(m.cfg.BaseURL)
		defaults.defaultChatModel = strings.TrimSpace(m.cfg.DefaultChatModel)
		defaults.environmentAPIKey = strings.TrimSpace(m.cfg.APIKey)
	}
	if normalizedProvider == defaultProviderName {
		if defaults.baseURL == "" {
			defaults.baseURL = strings.TrimSpace(m.cfg.BaseURL)
		}
		if defaults.defaultChatModel == "" {
			defaults.defaultChatModel = strings.TrimSpace(m.cfg.DefaultChatModel)
		}
		if defaults.environmentAPIKey == "" {
			defaults.environmentAPIKey = strings.TrimSpace(m.cfg.APIKey)
		}
	}

	if normalizedProvider == normalizeProvider(m.cfg.EmbeddingProvider) {
		if defaults.baseURL == "" {
			defaults.baseURL = strings.TrimSpace(m.cfg.EmbeddingBaseURL)
		}
		if defaults.environmentAPIKey == "" {
			defaults.environmentAPIKey = strings.TrimSpace(m.cfg.EmbeddingAPIKey)
		}
	}

	return defaults
}

func (m *Manager) toViewConfig(provider string, stored *StoredConfig) ViewConfig {
	defaults := m.defaultsForProvider(provider)
	view := ViewConfig{
		Provider:              provider,
		BaseURL:               defaults.baseURL,
		DefaultChatModel:      defaults.defaultChatModel,
		DefaultEmbeddingModel: defaults.defaultEmbeddingModel,
		IsEnabled:             defaults.isEnabled,
		APIKeySource:          APIKeySourceMissing,
	}

	if stored != nil {
		view.ID = &stored.ID
		view.CreatedAt = &stored.CreatedAt
		view.UpdatedAt = &stored.UpdatedAt
		if view.BaseURL == "" && strings.TrimSpace(stored.BaseURL) != "" {
			view.BaseURL = strings.TrimSpace(stored.BaseURL)
		}
		if view.DefaultChatModel == "" && strings.TrimSpace(stored.DefaultChatModel) != "" {
			view.DefaultChatModel = strings.TrimSpace(stored.DefaultChatModel)
		}
		if view.DefaultEmbeddingModel == "" && strings.TrimSpace(stored.DefaultEmbeddingModel) != "" {
			view.DefaultEmbeddingModel = strings.TrimSpace(stored.DefaultEmbeddingModel)
		}
		view.IsEnabled = stored.IsEnabled
	}

	if strings.TrimSpace(defaults.environmentAPIKey) != "" {
		view.HasAPIKey = true
		view.APIKeySource = APIKeySourceEnvironment
		return view
	}

	if stored != nil {
		if strings.TrimSpace(stored.EncryptedAPIKey) != "" {
			view.HasAPIKey = true
			view.APIKeySource = APIKeySourceDatabase
			return view
		}
	}

	return view
}

func normalizeStoredConfig(item StoredConfig) StoredConfig {
	item.Provider = normalizeProvider(item.Provider)
	item.BaseURL = strings.TrimSpace(item.BaseURL)
	item.DefaultChatModel = strings.TrimSpace(item.DefaultChatModel)
	item.DefaultEmbeddingModel = strings.TrimSpace(item.DefaultEmbeddingModel)
	item.EncryptedAPIKey = strings.TrimSpace(item.EncryptedAPIKey)
	return item
}

func normalizeProvider(provider string) string {
	return strings.ToLower(strings.TrimSpace(provider))
}

func coalesceString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func currentValue(current *StoredConfig, getter func(StoredConfig) string) string {
	if current == nil {
		return ""
	}
	return getter(*current)
}

func currentBool(current *StoredConfig, fallback bool) bool {
	if current == nil {
		return fallback
	}
	return current.IsEnabled
}
