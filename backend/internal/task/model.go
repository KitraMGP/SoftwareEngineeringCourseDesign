package task

import (
	"time"

	"github.com/google/uuid"
)

const (
	TaskTypeDocumentIngest       = "document_ingest"
	TaskTypeKnowledgeBaseReindex = "knowledge_base_reindex"
	TaskTypeResourceCleanup      = "resource_cleanup"

	ResourceTypeDocument      = "document"
	ResourceTypeKnowledgeBase = "knowledge_base"
	ResourceTypeSession       = "session"
	ResourceTypeFile          = "file"
	ResourceTypeSystem        = "system"

	TaskStatusPending   = "pending"
	TaskStatusRunning   = "running"
	TaskStatusSucceeded = "succeeded"
	TaskStatusFailed    = "failed"
	TaskStatusCancelled = "cancelled"
)

type Task struct {
	ID           uuid.UUID  `json:"id"`
	TaskType     string     `json:"task_type"`
	ResourceType string     `json:"resource_type"`
	ResourceID   *uuid.UUID `json:"resource_id,omitempty"`
	UserID       *uuid.UUID `json:"user_id,omitempty"`
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

type CreateTaskInput struct {
	TaskType     string
	ResourceType string
	ResourceID   *uuid.UUID
	UserID       *uuid.UUID
	Payload      string
	MaxAttempts  int
}
