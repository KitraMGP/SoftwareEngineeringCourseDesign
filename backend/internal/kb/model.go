package kb

import (
	"time"

	"github.com/google/uuid"
)

type KnowledgeBase struct {
	ID                  uuid.UUID  `json:"id"`
	UserID              uuid.UUID  `json:"user_id"`
	Name                string     `json:"name"`
	Description         *string    `json:"description,omitempty"`
	EmbeddingModel      string     `json:"embedding_model"`
	PromptTemplate      *string    `json:"prompt_template,omitempty"`
	RetrievalTopK       int        `json:"retrieval_top_k"`
	SimilarityThreshold *float64   `json:"similarity_threshold,omitempty"`
	LastIndexedAt       *time.Time `json:"last_indexed_at,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

type Document struct {
	ID               uuid.UUID  `json:"id"`
	KnowledgeBaseID  uuid.UUID  `json:"knowledge_base_id"`
	FileID           *uuid.UUID `json:"file_id,omitempty"`
	Title            *string    `json:"title,omitempty"`
	OriginalFilename *string    `json:"original_filename,omitempty"`
	Status           string     `json:"status"`
	ErrorMessage     *string    `json:"error_message,omitempty"`
	ContentLength    int        `json:"content_length"`
	ChunkCount       int        `json:"chunk_count"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type CreateKnowledgeBaseInput struct {
	Name                string
	Description         *string
	EmbeddingModel      string
	PromptTemplate      *string
	RetrievalTopK       int
	SimilarityThreshold *float64
}

type UpdateKnowledgeBaseInput struct {
	Name                string
	Description         *string
	PromptTemplate      *string
	RetrievalTopK       int
	SimilarityThreshold *float64
}

type ListKnowledgeBasesResult struct {
	Items []KnowledgeBase `json:"items"`
	Total int             `json:"total"`
	Page  int             `json:"page"`
	Size  int             `json:"size"`
}

type ListDocumentsResult struct {
	Items []Document `json:"items"`
	Total int        `json:"total"`
	Page  int        `json:"page"`
	Size  int        `json:"size"`
}

type UploadDocumentInput struct {
	Filename    string
	ContentType string
	Content     []byte
}

type UploadDocumentResult struct {
	DocumentID uuid.UUID `json:"document_id"`
	TaskID     uuid.UUID `json:"task_id"`
	Status     string    `json:"status"`
}

type CreateUploadedDocumentParams struct {
	UserID           uuid.UUID
	KnowledgeBaseID  uuid.UUID
	StorageProvider  string
	BucketName       string
	ObjectKey        string
	OriginalFilename string
	MIMEType         string
	SizeBytes        int64
	SHA256           string
	Title            *string
}

type DocumentIngestSource struct {
	DocumentID       uuid.UUID
	KnowledgeBaseID  uuid.UUID
	FileID           uuid.UUID
	BucketName       string
	ObjectKey        string
	OriginalFilename string
	MIMEType         string
}

type DocumentChunkInput struct {
	ChunkIndex int
	Content    string
	TokenCount int
	Embedding  string
}

type FileObjectRef struct {
	FileID         uuid.UUID
	BucketName     string
	ObjectKey      string
	WasSoftDeleted bool
}
