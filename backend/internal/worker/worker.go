package worker

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"strings"
	"time"

	"backend/internal/kb"
	"backend/internal/model"
	"backend/internal/platform/httpx"
	"backend/internal/platform/storage"
	"backend/internal/task"

	"github.com/google/uuid"
)

type Worker struct {
	logger       *slog.Logger
	taskService  *task.Service
	kbRepo       *kb.Repository
	storage      storage.Service
	embedder     model.EmbeddingProvider
	embedTimeout time.Duration
	pollInterval time.Duration
}

type processingError struct {
	Code      string
	Message   string
	Retryable bool
}

func (e *processingError) Error() string {
	return e.Message
}

func New(logger *slog.Logger, taskService *task.Service, kbRepo *kb.Repository, storageService storage.Service, embedder model.EmbeddingProvider, embedTimeout, pollInterval time.Duration) *Worker {
	if embedTimeout <= 0 {
		embedTimeout = 20 * time.Second
	}
	return &Worker{
		logger:       logger,
		taskService:  taskService,
		kbRepo:       kbRepo,
		storage:      storageService,
		embedder:     embedder,
		embedTimeout: embedTimeout,
		pollInterval: pollInterval,
	}
}

func (w *Worker) Run(ctx context.Context) error {
	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	w.logger.Info("worker started", slog.Duration("poll_interval", w.pollInterval))

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("worker stopped")
			return nil
		case <-ticker.C:
			if err := w.runBatch(ctx); err != nil {
				w.logger.Error("worker batch failed", slog.Any("error", err))
			}
		}
	}
}

func (w *Worker) runBatch(ctx context.Context) error {
	processed := 0
	for {
		taskItem, err := w.taskService.ClaimNextRunnableTask(ctx)
		if err != nil {
			return err
		}
		if taskItem == nil {
			if processed == 0 {
				count, err := w.taskService.CountRunnableTasks(ctx)
				if err != nil {
					return err
				}
				w.logger.Info("worker heartbeat", slog.Int("runnable_tasks", count))
			}
			return nil
		}

		processed++
		if err := w.processTask(ctx, taskItem); err != nil {
			w.logger.Error(
				"task execution failed",
				slog.String("task_id", taskItem.ID.String()),
				slog.String("task_type", taskItem.TaskType),
				slog.Any("error", err),
			)
		}
	}
}

func (w *Worker) processTask(ctx context.Context, taskItem *task.Task) error {
	switch taskItem.TaskType {
	case task.TaskTypeDocumentIngest:
		return w.processDocumentIngest(ctx, taskItem)
	case task.TaskTypeKnowledgeBaseReindex:
		return w.processKnowledgeBaseReindex(ctx, taskItem)
	case task.TaskTypeResourceCleanup:
		return w.processResourceCleanup(ctx, taskItem)
	default:
		return w.failTask(ctx, taskItem, &processingError{
			Code:      "unsupported_task_type",
			Message:   "unsupported task type",
			Retryable: false,
		})
	}
}

func (w *Worker) processDocumentIngest(ctx context.Context, taskItem *task.Task) error {
	documentID, ok := resourceID(taskItem)
	if !ok {
		return w.failTask(ctx, taskItem, &processingError{
			Code:      "missing_resource_id",
			Message:   "document ingest task missing document id",
			Retryable: false,
		})
	}

	source, err := w.kbRepo.GetDocumentIngestSource(ctx, documentID)
	if err != nil {
		return w.retryOrFailDocumentIngest(ctx, taskItem, documentID, classifyError(err))
	}
	if source == nil {
		return w.taskService.MarkTaskSucceeded(ctx, taskItem.ID, map[string]any{
			"skipped": true,
			"reason":  "document_missing_or_deleted",
		})
	}

	if err := w.kbRepo.MarkDocumentProcessing(ctx, documentID); err != nil && !errors.Is(err, kb.ErrNotFound) {
		return w.retryOrFailDocumentIngest(ctx, taskItem, documentID, classifyError(err))
	}

	reader, err := w.storage.OpenObject(ctx, source.BucketName, source.ObjectKey)
	if err != nil {
		return w.retryOrFailDocumentIngest(ctx, taskItem, documentID, classifyStorageError(err))
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return w.retryOrFailDocumentIngest(ctx, taskItem, documentID, classifyError(err))
	}

	content, err := kb.ParseDocumentContent(source.MIMEType, source.OriginalFilename, data)
	if err != nil {
		return w.retryOrFailDocumentIngest(ctx, taskItem, documentID, classifyError(err))
	}
	if strings.TrimSpace(content) == "" {
		return w.retryOrFailDocumentIngest(ctx, taskItem, documentID, &processingError{
			Code:      "empty_document",
			Message:   "document content is empty",
			Retryable: false,
		})
	}

	chunks := kb.BuildDocumentChunks(content)
	if len(chunks) == 0 {
		return w.retryOrFailDocumentIngest(ctx, taskItem, documentID, &processingError{
			Code:      "empty_chunks",
			Message:   "document content could not be chunked",
			Retryable: false,
		})
	}

	embeddingModel, err := w.kbRepo.GetEmbeddingModelForKnowledgeBase(ctx, source.KnowledgeBaseID)
	if err != nil {
		return w.retryOrFailDocumentIngest(ctx, taskItem, documentID, classifyError(err))
	}
	if err := w.embedChunks(ctx, embeddingModel, chunks); err != nil {
		return w.retryOrFailDocumentIngest(ctx, taskItem, documentID, classifyEmbeddingError(err))
	}

	if err := w.kbRepo.ReplaceDocumentContent(ctx, documentID, content, chunks); err != nil {
		return w.retryOrFailDocumentIngest(ctx, taskItem, documentID, classifyError(err))
	}

	return w.taskService.MarkTaskSucceeded(ctx, taskItem.ID, map[string]any{
		"document_id":    documentID,
		"knowledge_base": source.KnowledgeBaseID,
		"chunk_count":    len(chunks),
		"content_length": len([]rune(content)),
	})
}

func (w *Worker) processKnowledgeBaseReindex(ctx context.Context, taskItem *task.Task) error {
	knowledgeBaseID, ok := resourceID(taskItem)
	if !ok {
		return w.failTask(ctx, taskItem, &processingError{
			Code:      "missing_resource_id",
			Message:   "reindex task missing knowledge base id",
			Retryable: false,
		})
	}

	documentIDs, err := w.kbRepo.ListDocumentIDsForKnowledgeBase(ctx, knowledgeBaseID)
	if err != nil {
		return w.failTask(ctx, taskItem, classifyError(err))
	}

	dispatched := 0
	skipped := 0
	for _, documentID := range documentIDs {
		if _, err := w.taskService.CreateDocumentIngestTask(ctx, taskItem.UserID, knowledgeBaseID, documentID); err != nil {
			if appErr, ok := httpx.AsAppError(err); ok && appErr.Code == httpx.CodeConflict {
				skipped++
				continue
			}
			return w.failTask(ctx, taskItem, classifyError(err))
		}
		dispatched++
	}

	return w.taskService.MarkTaskSucceeded(ctx, taskItem.ID, map[string]any{
		"knowledge_base_id": knowledgeBaseID,
		"dispatched_count":  dispatched,
		"skipped_count":     skipped,
	})
}

func (w *Worker) processResourceCleanup(ctx context.Context, taskItem *task.Task) error {
	resourceID, ok := resourceID(taskItem)
	if !ok {
		return w.failTask(ctx, taskItem, &processingError{
			Code:      "missing_resource_id",
			Message:   "cleanup task missing resource id",
			Retryable: false,
		})
	}

	var (
		refs []kb.FileObjectRef
		err  error
	)
	switch taskItem.ResourceType {
	case task.ResourceTypeDocument:
		refs, err = w.kbRepo.CleanupDocumentResource(ctx, resourceID)
	case task.ResourceTypeKnowledgeBase:
		refs, err = w.kbRepo.CleanupKnowledgeBaseResources(ctx, resourceID)
	default:
		return w.failTask(ctx, taskItem, &processingError{
			Code:      "unsupported_cleanup_resource",
			Message:   "unsupported cleanup resource type",
			Retryable: false,
		})
	}
	if err != nil {
		return w.failTask(ctx, taskItem, classifyError(err))
	}

	for _, ref := range refs {
		if err := w.storage.DeleteObject(ctx, ref.BucketName, ref.ObjectKey); err != nil {
			return w.failTask(ctx, taskItem, classifyStorageError(err))
		}
	}

	return w.taskService.MarkTaskSucceeded(ctx, taskItem.ID, map[string]any{
		"resource_type":   taskItem.ResourceType,
		"resource_id":     resourceID,
		"deleted_objects": len(refs),
	})
}

func (w *Worker) retryOrFailDocumentIngest(ctx context.Context, taskItem *task.Task, documentID uuid.UUID, err *processingError) error {
	if err == nil {
		return nil
	}

	if err.Retryable && taskItem.AttemptCount < taskItem.MaxAttempts {
		nextRunAt := time.Now().UTC().Add(retryDelay(taskItem.AttemptCount))
		taskErr := w.taskService.MarkTaskRetryPending(ctx, taskItem.ID, err.Code, err.Message, nextRunAt)
		docErr := w.kbRepo.MarkDocumentPending(ctx, documentID, err.Message)
		return errors.Join(taskErr, docErr)
	}

	taskErr := w.taskService.MarkTaskFailed(ctx, taskItem.ID, err.Code, err.Message)
	docErr := w.kbRepo.MarkDocumentFailed(ctx, documentID, err.Message)
	return errors.Join(taskErr, docErr)
}

func (w *Worker) failTask(ctx context.Context, taskItem *task.Task, err *processingError) error {
	if err == nil {
		return nil
	}

	if err.Retryable && taskItem.AttemptCount < taskItem.MaxAttempts {
		return w.taskService.MarkTaskRetryPending(ctx, taskItem.ID, err.Code, err.Message, time.Now().UTC().Add(retryDelay(taskItem.AttemptCount)))
	}
	return w.taskService.MarkTaskFailed(ctx, taskItem.ID, err.Code, err.Message)
}

func classifyStorageError(err error) *processingError {
	if err == nil {
		return nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return &processingError{
			Code:      "object_not_found",
			Message:   "stored object not found",
			Retryable: false,
		}
	}
	return &processingError{
		Code:      "storage_error",
		Message:   err.Error(),
		Retryable: true,
	}
}

func classifyError(err error) *processingError {
	if err == nil {
		return nil
	}
	if errors.Is(err, kb.ErrNotFound) {
		return &processingError{
			Code:      "resource_not_found",
			Message:   "knowledge base or document not found",
			Retryable: false,
		}
	}

	var ingestErr *kb.IngestError
	if errors.As(err, &ingestErr) {
		return &processingError{
			Code:      "document_ingest_error",
			Message:   ingestErr.Message,
			Retryable: ingestErr.Retryable,
		}
	}

	return &processingError{
		Code:      "worker_error",
		Message:   err.Error(),
		Retryable: true,
	}
}

func classifyEmbeddingError(err error) *processingError {
	if err == nil {
		return nil
	}

	providerErr, ok := model.AsProviderError(err)
	if !ok {
		return classifyError(err)
	}

	result := &processingError{
		Code:      "embedding_provider_error",
		Message:   providerErr.Message,
		Retryable: true,
	}
	switch providerErr.Kind {
	case model.ProviderErrorMisconfigured, model.ProviderErrorAuthFailed, model.ProviderErrorBadRequest:
		result.Retryable = false
	case model.ProviderErrorRateLimited, model.ProviderErrorUnavailable:
		result.Retryable = true
	}
	return result
}

func (w *Worker) embedChunks(ctx context.Context, embeddingModel string, chunks []kb.DocumentChunkInput) error {
	if len(chunks) == 0 {
		return nil
	}

	if w.embedder == nil {
		return &model.ProviderError{
			Kind:    model.ProviderErrorMisconfigured,
			Message: "embedding provider is not configured",
		}
	}

	const batchSize = 16
	for start := 0; start < len(chunks); start += batchSize {
		end := start + batchSize
		if end > len(chunks) {
			end = len(chunks)
		}

		texts := make([]string, 0, end-start)
		for idx := start; idx < end; idx++ {
			texts = append(texts, chunks[idx].Content)
		}

		batchCtx, cancel := context.WithTimeout(ctx, w.embedTimeout)
		result, err := w.embedder.Embed(batchCtx, model.EmbeddingRequest{
			Model: embeddingModel,
			Texts: texts,
		})
		cancel()
		if err != nil {
			return err
		}
		if len(result.Vectors) != len(texts) {
			return &model.ProviderError{
				Kind:    model.ProviderErrorUnavailable,
				Message: "embedding provider returned an unexpected batch size",
			}
		}

		for offset, vector := range result.Vectors {
			if len(vector) != model.EmbeddingDimension {
				return &model.ProviderError{
					Kind:    model.ProviderErrorBadRequest,
					Message: "embedding dimension mismatch",
				}
			}
			chunks[start+offset].Embedding = model.FormatVector(vector)
		}
	}
	return nil
}

func retryDelay(attempt int) time.Duration {
	if attempt < 1 {
		attempt = 1
	}
	return time.Duration(attempt*attempt) * 15 * time.Second
}

func resourceID(taskItem *task.Task) (uuid.UUID, bool) {
	if taskItem.ResourceID == nil {
		return uuid.UUID{}, false
	}
	return *taskItem.ResourceID, true
}
