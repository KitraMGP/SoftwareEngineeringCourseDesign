package admin

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Overview struct {
	UserCount          int `json:"user_count"`
	ActiveUserCount    int `json:"active_user_count"`
	KnowledgeBaseCount int `json:"knowledge_base_count"`
	DocumentCount      int `json:"document_count"`
	PendingTaskCount   int `json:"pending_task_count"`
	FailedTaskCount    int `json:"failed_task_count"`
}

type User struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Nickname  *string   `json:"nickname,omitempty"`
	AvatarURL *string   `json:"avatar_url,omitempty"`
	Role      string    `json:"role"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ListUsersResult struct {
	Items []User `json:"items"`
	Total int    `json:"total"`
	Page  int    `json:"page"`
	Size  int    `json:"size"`
}

type ListUsersParams struct {
	Page    int
	Size    int
	Keyword string
	Status  string
	Role    string
}

type Task struct {
	ID           uuid.UUID  `json:"id"`
	TaskType     string     `json:"task_type"`
	ResourceType string     `json:"resource_type"`
	ResourceID   *uuid.UUID `json:"resource_id,omitempty"`
	UserID       *uuid.UUID `json:"user_id,omitempty"`
	UserLabel    *string    `json:"user_label,omitempty"`
	Status       string     `json:"status"`
	AttemptCount int        `json:"attempt_count"`
	MaxAttempts  int        `json:"max_attempts"`
	NextRunAt    *time.Time `json:"next_run_at,omitempty"`
	StartedAt    *time.Time `json:"started_at,omitempty"`
	FinishedAt   *time.Time `json:"finished_at,omitempty"`
	ErrorCode    *string    `json:"error_code,omitempty"`
	ErrorMessage *string    `json:"error_message,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type ListTasksResult struct {
	Items []Task `json:"items"`
	Total int    `json:"total"`
	Page  int    `json:"page"`
	Size  int    `json:"size"`
}

type ListTasksParams struct {
	Page     int
	Size     int
	Status   string
	TaskType string
}

type ProviderConfig struct {
	ID                    *uuid.UUID `json:"id,omitempty"`
	Provider              string     `json:"provider"`
	BaseURL               string     `json:"base_url"`
	DefaultChatModel      string     `json:"default_chat_model"`
	DefaultEmbeddingModel string     `json:"default_embedding_model"`
	IsEnabled             bool       `json:"is_enabled"`
	HasAPIKey             bool       `json:"has_api_key"`
	APIKeySource          string     `json:"api_key_source"`
	CreatedAt             *time.Time `json:"created_at,omitempty"`
	UpdatedAt             *time.Time `json:"updated_at,omitempty"`
}

type SystemSetting struct {
	Key         string          `json:"key"`
	Category    string          `json:"category"`
	Value       json.RawMessage `json:"value"`
	Description *string         `json:"description,omitempty"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type QuotaPolicy struct {
	ID                    uuid.UUID  `json:"id"`
	ScopeType             string     `json:"scope_type"`
	ScopeID               *uuid.UUID `json:"scope_id,omitempty"`
	DailyTotalTokensLimit int64      `json:"daily_total_tokens_limit"`
	StorageBytesLimit     int64      `json:"storage_bytes_limit"`
	DocumentCountLimit    int        `json:"document_count_limit"`
	WarnRatio             string     `json:"warn_ratio"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}

type AuditLog struct {
	ID           int64           `json:"id"`
	ActorUserID  *uuid.UUID      `json:"actor_user_id,omitempty"`
	ActorLabel   *string         `json:"actor_label,omitempty"`
	ActorRole    string          `json:"actor_role"`
	Action       string          `json:"action"`
	ResourceType *string         `json:"resource_type,omitempty"`
	ResourceID   *uuid.UUID      `json:"resource_id,omitempty"`
	TargetUserID *uuid.UUID      `json:"target_user_id,omitempty"`
	TargetLabel  *string         `json:"target_label,omitempty"`
	Result       string          `json:"result"`
	Metadata     json.RawMessage `json:"metadata"`
	CreatedAt    time.Time       `json:"created_at"`
}

type ListAuditLogsResult struct {
	Items []AuditLog `json:"items"`
	Total int        `json:"total"`
	Page  int        `json:"page"`
	Size  int        `json:"size"`
}

type CreateAuditLogInput struct {
	ActorUserID  *uuid.UUID
	ActorRole    string
	Action       string
	ResourceType *string
	ResourceID   *uuid.UUID
	TargetUserID *uuid.UUID
	Result       string
	Metadata     string
}
