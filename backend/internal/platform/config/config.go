package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	App      AppConfig
	HTTP     HTTPConfig
	Database DatabaseConfig
	Auth     AuthConfig
	Task     TaskConfig
	Storage  StorageConfig
}

type AppConfig struct {
	Name string
	Env  string
}

type HTTPConfig struct {
	Addr         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type DatabaseConfig struct {
	DSN             string
	MaxConns        int32
	MinConns        int32
	MaxConnIdleTime time.Duration
	MaxConnLifetime time.Duration
}

type AuthConfig struct {
	Issuer              string
	JWTSecret           string
	AccessTokenTTL      time.Duration
	RefreshTokenTTL     time.Duration
	RefreshCookieName   string
	RefreshCookiePath   string
	RefreshCookieDomain string
	RefreshCookieSecure bool
}

type TaskConfig struct {
	PollInterval time.Duration
}

type StorageConfig struct {
	Provider       string
	Bucket         string
	LocalRoot      string
	MaxUploadBytes int64
}

func Load() (Config, error) {
	cfg := Config{
		App: AppConfig{
			Name: getEnv("APP_NAME", "private-kb-qa"),
			Env:  getEnv("APP_ENV", "development"),
		},
		HTTP: HTTPConfig{
			Addr:         getEnv("HTTP_ADDR", ":8080"),
			ReadTimeout:  getDurationEnv("HTTP_READ_TIMEOUT", 10*time.Second),
			WriteTimeout: getDurationEnv("HTTP_WRITE_TIMEOUT", 60*time.Second),
			IdleTimeout:  getDurationEnv("HTTP_IDLE_TIMEOUT", 60*time.Second),
		},
		Database: DatabaseConfig{
			DSN:             strings.TrimSpace(os.Getenv("DATABASE_DSN")),
			MaxConns:        int32(getIntEnv("DATABASE_MAX_CONNS", 10)),
			MinConns:        int32(getIntEnv("DATABASE_MIN_CONNS", 2)),
			MaxConnIdleTime: getDurationEnv("DATABASE_MAX_CONN_IDLE_TIME", 15*time.Minute),
			MaxConnLifetime: getDurationEnv("DATABASE_MAX_CONN_LIFETIME", time.Hour),
		},
		Auth: AuthConfig{
			Issuer:              getEnv("AUTH_ISSUER", "private-kb-qa"),
			JWTSecret:           strings.TrimSpace(os.Getenv("AUTH_JWT_SECRET")),
			AccessTokenTTL:      getDurationEnv("AUTH_ACCESS_TOKEN_TTL", 30*time.Minute),
			RefreshTokenTTL:     getDurationEnv("AUTH_REFRESH_TOKEN_TTL", 7*24*time.Hour),
			RefreshCookieName:   getEnv("AUTH_REFRESH_COOKIE_NAME", "refresh_token"),
			RefreshCookiePath:   getEnv("AUTH_REFRESH_COOKIE_PATH", "/api/v1/auth"),
			RefreshCookieDomain: strings.TrimSpace(os.Getenv("AUTH_REFRESH_COOKIE_DOMAIN")),
			RefreshCookieSecure: getBoolEnv("AUTH_REFRESH_COOKIE_SECURE", false),
		},
		Task: TaskConfig{
			PollInterval: getDurationEnv("TASK_POLL_INTERVAL", 5*time.Second),
		},
		Storage: StorageConfig{
			Provider:       getEnv("STORAGE_PROVIDER", "local_fs"),
			Bucket:         getEnv("STORAGE_BUCKET", "local-dev"),
			LocalRoot:      getEnv("STORAGE_LOCAL_ROOT", "./data/storage"),
			MaxUploadBytes: getInt64Env("STORAGE_MAX_UPLOAD_BYTES", 20*1024*1024),
		},
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c Config) Validate() error {
	var errs []error

	if c.Database.DSN == "" {
		errs = append(errs, errors.New("DATABASE_DSN is required"))
	}
	if c.Auth.JWTSecret == "" {
		errs = append(errs, errors.New("AUTH_JWT_SECRET is required"))
	}
	if c.Auth.AccessTokenTTL <= 0 {
		errs = append(errs, errors.New("AUTH_ACCESS_TOKEN_TTL must be positive"))
	}
	if c.Auth.RefreshTokenTTL <= 0 {
		errs = append(errs, errors.New("AUTH_REFRESH_TOKEN_TTL must be positive"))
	}
	if c.HTTP.Addr == "" {
		errs = append(errs, errors.New("HTTP_ADDR is required"))
	}
	if strings.TrimSpace(c.Storage.Provider) == "" {
		errs = append(errs, errors.New("STORAGE_PROVIDER is required"))
	}
	if strings.TrimSpace(c.Storage.Bucket) == "" {
		errs = append(errs, errors.New("STORAGE_BUCKET is required"))
	}
	if strings.TrimSpace(c.Storage.LocalRoot) == "" {
		errs = append(errs, errors.New("STORAGE_LOCAL_ROOT is required"))
	}
	if c.Storage.MaxUploadBytes <= 0 {
		errs = append(errs, errors.New("STORAGE_MAX_UPLOAD_BYTES must be positive"))
	}

	if len(errs) == 0 {
		return nil
	}

	parts := make([]string, 0, len(errs))
	for _, err := range errs {
		parts = append(parts, err.Error())
	}
	return fmt.Errorf("invalid config: %s", strings.Join(parts, "; "))
}

func getEnv(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok && strings.TrimSpace(value) != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	value, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(value) == "" {
		return defaultValue
	}

	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return defaultValue
	}
	return parsed
}

func getInt64Env(key string, defaultValue int64) int64 {
	value, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(value) == "" {
		return defaultValue
	}

	parsed, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	if err != nil {
		return defaultValue
	}
	return parsed
}

func getBoolEnv(key string, defaultValue bool) bool {
	value, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(value) == "" {
		return defaultValue
	}

	parsed, err := strconv.ParseBool(strings.TrimSpace(value))
	if err != nil {
		return defaultValue
	}
	return parsed
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	value, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(value) == "" {
		return defaultValue
	}

	parsed, err := time.ParseDuration(strings.TrimSpace(value))
	if err != nil {
		return defaultValue
	}
	return parsed
}
