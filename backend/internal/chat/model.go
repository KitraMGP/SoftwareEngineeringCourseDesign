package chat

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID                uuid.UUID  `json:"id"`
	UserID            uuid.UUID  `json:"user_id"`
	Name              *string    `json:"name,omitempty"`
	Model             string     `json:"model"`
	KnowledgeBaseID   *uuid.UUID `json:"knowledge_base_id,omitempty"`
	KnowledgeBaseName *string    `json:"knowledge_base_name,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

type Message struct {
	ID               uuid.UUID  `json:"id"`
	SessionID        uuid.UUID  `json:"session_id"`
	Role             string     `json:"role"`
	ReplyToMessageID *uuid.UUID `json:"reply_to_message_id,omitempty"`
	Content          string     `json:"content"`
	Status           string     `json:"status"`
	ModelUsed        *string    `json:"model_used,omitempty"`
	Grounded         bool       `json:"grounded"`
	PromptTokens     int        `json:"prompt_tokens"`
	CompletionTokens int        `json:"completion_tokens"`
	TotalTokens      int        `json:"total_tokens"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type CreateSessionInput struct {
	Name            *string
	Model           string
	KnowledgeBaseID *uuid.UUID
}

type CreateMessageInput struct {
	ID               uuid.UUID
	Role             string
	ReplyToMessageID *uuid.UUID
	Content          string
	Status           string
	ModelUsed        *string
	Grounded         bool
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

type ListSessionsResult struct {
	Items []Session `json:"items"`
	Total int       `json:"total"`
	Page  int       `json:"page"`
	Size  int       `json:"size"`
}

type SessionDetail struct {
	Session  Session   `json:"session"`
	Messages []Message `json:"messages"`
}

type StreamMeta struct {
	MessageID uuid.UUID `json:"message_id"`
	Grounded  bool      `json:"grounded"`
	Model     string    `json:"model"`
}
