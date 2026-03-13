package kb

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"path"
	"path/filepath"
	"strings"
	"time"

	"backend/internal/platform/httpx"
	"backend/internal/platform/storage"
	"backend/internal/task"

	"github.com/google/uuid"
)

type Service struct {
	repo           *Repository
	taskService    *task.Service
	storage        storage.Service
	maxUploadBytes int64
}

func NewService(repo *Repository, taskService *task.Service, storageService storage.Service, maxUploadBytes int64) *Service {
	return &Service{
		repo:           repo,
		taskService:    taskService,
		storage:        storageService,
		maxUploadBytes: maxUploadBytes,
	}
}

func (s *Service) MaxUploadBytes() int64 {
	return s.maxUploadBytes
}

func (s *Service) CreateKnowledgeBase(ctx context.Context, userID uuid.UUID, input CreateKnowledgeBaseInput) (*KnowledgeBase, error) {
	if details := validateCreateInput(input); len(details) > 0 {
		return nil, httpx.ValidationFailed(details...)
	}
	return s.repo.Create(ctx, userID, input)
}

func (s *Service) ListKnowledgeBases(ctx context.Context, userID uuid.UUID, page, size int, keyword string) (*ListKnowledgeBasesResult, error) {
	return s.repo.List(ctx, userID, page, size, keyword)
}

func (s *Service) GetKnowledgeBase(ctx context.Context, userID, knowledgeBaseID uuid.UUID) (*KnowledgeBase, error) {
	kb, err := s.repo.Get(ctx, userID, knowledgeBaseID)
	if err != nil {
		if err == ErrNotFound {
			return nil, httpx.NotFound("knowledge base not found")
		}
		return nil, httpx.Internal("failed to load knowledge base").WithErr(err)
	}
	return kb, nil
}

func (s *Service) UpdateKnowledgeBase(ctx context.Context, userID, knowledgeBaseID uuid.UUID, input UpdateKnowledgeBaseInput) (*KnowledgeBase, error) {
	if details := validateUpdateInput(input); len(details) > 0 {
		return nil, httpx.ValidationFailed(details...)
	}

	kb, err := s.repo.Update(ctx, userID, knowledgeBaseID, input)
	if err != nil {
		if err == ErrNotFound {
			return nil, httpx.NotFound("knowledge base not found")
		}
		return nil, httpx.Internal("failed to update knowledge base").WithErr(err)
	}
	return kb, nil
}

func (s *Service) DeleteKnowledgeBase(ctx context.Context, userID, knowledgeBaseID uuid.UUID) error {
	if err := s.repo.Delete(ctx, userID, knowledgeBaseID); err != nil {
		if err == ErrNotFound {
			return httpx.NotFound("knowledge base not found")
		}
		return httpx.Internal("failed to delete knowledge base").WithErr(err)
	}

	if _, err := s.taskService.CreateCleanupTask(ctx, &userID, task.ResourceTypeKnowledgeBase, knowledgeBaseID, map[string]any{
		"knowledge_base_id": knowledgeBaseID,
	}); err != nil {
		return err
	}
	return nil
}

func (s *Service) ListDocuments(ctx context.Context, userID, knowledgeBaseID uuid.UUID, page, size int) (*ListDocumentsResult, error) {
	return s.repo.ListDocuments(ctx, userID, knowledgeBaseID, page, size)
}

func (s *Service) UploadDocument(ctx context.Context, userID, knowledgeBaseID uuid.UUID, input UploadDocumentInput) (*UploadDocumentResult, error) {
	if _, err := s.GetKnowledgeBase(ctx, userID, knowledgeBaseID); err != nil {
		return nil, err
	}

	contentSize := int64(len(input.Content))
	if contentSize == 0 {
		return nil, httpx.ValidationFailed(httpx.FieldError{Field: "file", Message: "file is required"})
	}
	if contentSize > s.maxUploadBytes {
		return nil, httpx.FileTooLarge()
	}

	mimeType, err := NormalizeUploadMetadata(input.Filename, input.ContentType)
	if err != nil {
		return nil, httpx.UnsupportedFileType()
	}

	checksum := sha256.Sum256(input.Content)
	sha256Hex := hex.EncodeToString(checksum[:])

	exists, err := s.repo.HasDuplicateDocument(ctx, knowledgeBaseID, sha256Hex)
	if err != nil {
		return nil, httpx.Internal("failed to check duplicate document").WithErr(err)
	}
	if exists {
		return nil, httpx.DuplicateDocument()
	}

	objectKey := buildObjectKey(knowledgeBaseID, input.Filename)
	uploaded, err := s.storage.PutObject(ctx, objectKey, bytes.NewReader(input.Content), contentSize, mimeType)
	if err != nil {
		return nil, httpx.StorageError("failed to store uploaded file").WithErr(err)
	}

	result, err := s.repo.CreateUploadedDocument(ctx, CreateUploadedDocumentParams{
		UserID:           userID,
		KnowledgeBaseID:  knowledgeBaseID,
		StorageProvider:  uploaded.Provider,
		BucketName:       uploaded.Bucket,
		ObjectKey:        uploaded.ObjectKey,
		OriginalFilename: sanitizeFilename(input.Filename),
		MIMEType:         mimeType,
		SizeBytes:        uploaded.SizeBytes,
		SHA256:           sha256Hex,
		Title:            defaultDocumentTitle(input.Filename),
	})
	if err != nil {
		_ = s.storage.DeleteObject(ctx, uploaded.Bucket, uploaded.ObjectKey)
		return nil, httpx.TaskDispatchFailed("failed to create document records").WithErr(err)
	}

	return result, nil
}

func (s *Service) GetDocument(ctx context.Context, userID, knowledgeBaseID, documentID uuid.UUID) (*Document, error) {
	document, err := s.repo.GetDocument(ctx, userID, knowledgeBaseID, documentID)
	if err != nil {
		if err == ErrNotFound {
			return nil, httpx.NotFound("document not found")
		}
		return nil, httpx.Internal("failed to load document").WithErr(err)
	}
	return document, nil
}

func (s *Service) DeleteDocument(ctx context.Context, userID, knowledgeBaseID, documentID uuid.UUID) error {
	if err := s.repo.DeleteDocument(ctx, userID, knowledgeBaseID, documentID); err != nil {
		if err == ErrNotFound {
			return httpx.NotFound("document not found")
		}
		return httpx.Internal("failed to delete document").WithErr(err)
	}

	if _, err := s.taskService.CreateCleanupTask(ctx, &userID, task.ResourceTypeDocument, documentID, map[string]any{
		"knowledge_base_id": knowledgeBaseID,
		"document_id":       documentID,
	}); err != nil {
		return err
	}
	return nil
}

func (s *Service) ReindexKnowledgeBase(ctx context.Context, userID, knowledgeBaseID uuid.UUID) (*task.Task, error) {
	if _, err := s.GetKnowledgeBase(ctx, userID, knowledgeBaseID); err != nil {
		return nil, err
	}
	return s.taskService.CreateKnowledgeBaseReindexTask(ctx, userID, knowledgeBaseID)
}

func validateCreateInput(input CreateKnowledgeBaseInput) []httpx.FieldError {
	var details []httpx.FieldError

	if strings.TrimSpace(input.Name) == "" {
		details = append(details, httpx.FieldError{Field: "name", Message: "name is required"})
	}
	if strings.TrimSpace(input.EmbeddingModel) == "" {
		details = append(details, httpx.FieldError{Field: "embedding_model", Message: "embedding_model is required"})
	}
	if input.RetrievalTopK <= 0 {
		details = append(details, httpx.FieldError{Field: "retrieval_top_k", Message: "retrieval_top_k must be positive"})
	}
	return details
}

func validateUpdateInput(input UpdateKnowledgeBaseInput) []httpx.FieldError {
	var details []httpx.FieldError

	if strings.TrimSpace(input.Name) == "" {
		details = append(details, httpx.FieldError{Field: "name", Message: "name is required"})
	}
	if input.RetrievalTopK <= 0 {
		details = append(details, httpx.FieldError{Field: "retrieval_top_k", Message: "retrieval_top_k must be positive"})
	}
	return details
}

func buildObjectKey(knowledgeBaseID uuid.UUID, filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	return path.Join(
		"knowledge-bases",
		knowledgeBaseID.String(),
		time.Now().UTC().Format("2006/01/02"),
		uuid.NewString()+ext,
	)
}

func defaultDocumentTitle(filename string) *string {
	name := strings.TrimSpace(strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename)))
	if name == "" || name == "." || name == string(filepath.Separator) {
		return nil
	}
	if len([]rune(name)) > 500 {
		runes := []rune(name)
		name = string(runes[:500])
	}
	return &name
}

func sanitizeFilename(filename string) string {
	name := strings.TrimSpace(filepath.Base(filename))
	if name == "" || name == "." || name == string(filepath.Separator) {
		return "document"
	}
	return name
}
