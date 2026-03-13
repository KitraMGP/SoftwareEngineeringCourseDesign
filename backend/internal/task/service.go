package task

import (
	"context"
	"encoding/json"
	"time"

	"backend/internal/platform/httpx"

	"github.com/google/uuid"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateDocumentIngestTask(ctx context.Context, userID *uuid.UUID, knowledgeBaseID, documentID uuid.UUID) (*Task, error) {
	exists, err := s.repo.HasActiveTask(ctx, TaskTypeDocumentIngest, documentID)
	if err != nil {
		return nil, httpx.TaskDispatchFailed("failed to check existing document ingest task").WithErr(err)
	}
	if exists {
		return nil, httpx.Conflict("document already has an active ingest task")
	}

	payload, err := json.Marshal(map[string]any{
		"knowledge_base_id": knowledgeBaseID,
		"document_id":       documentID,
	})
	if err != nil {
		return nil, httpx.TaskDispatchFailed("failed to build document ingest task payload").WithErr(err)
	}

	task, err := s.repo.Create(ctx, CreateTaskInput{
		TaskType:     TaskTypeDocumentIngest,
		ResourceType: ResourceTypeDocument,
		ResourceID:   &documentID,
		UserID:       userID,
		Payload:      string(payload),
		MaxAttempts:  3,
	})
	if err != nil {
		return nil, httpx.TaskDispatchFailed("failed to create document ingest task").WithErr(err)
	}
	return task, nil
}

func (s *Service) CreateKnowledgeBaseReindexTask(ctx context.Context, userID, knowledgeBaseID uuid.UUID) (*Task, error) {
	exists, err := s.repo.HasActiveTask(ctx, TaskTypeKnowledgeBaseReindex, knowledgeBaseID)
	if err != nil {
		return nil, httpx.TaskDispatchFailed("failed to check existing reindex task").WithErr(err)
	}
	if exists {
		return nil, httpx.Conflict("knowledge base already has an active reindex task")
	}

	payload, err := json.Marshal(map[string]any{
		"knowledge_base_id": knowledgeBaseID,
	})
	if err != nil {
		return nil, httpx.TaskDispatchFailed("failed to build reindex task payload").WithErr(err)
	}

	task, err := s.repo.Create(ctx, CreateTaskInput{
		TaskType:     TaskTypeKnowledgeBaseReindex,
		ResourceType: ResourceTypeKnowledgeBase,
		ResourceID:   &knowledgeBaseID,
		UserID:       &userID,
		Payload:      string(payload),
		MaxAttempts:  3,
	})
	if err != nil {
		return nil, httpx.TaskDispatchFailed("failed to create knowledge base reindex task").WithErr(err)
	}
	return task, nil
}

func (s *Service) CreateCleanupTask(ctx context.Context, userID *uuid.UUID, resourceType string, resourceID uuid.UUID, payload map[string]any) (*Task, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, httpx.TaskDispatchFailed("failed to build cleanup task payload").WithErr(err)
	}

	task, err := s.repo.Create(ctx, CreateTaskInput{
		TaskType:     TaskTypeResourceCleanup,
		ResourceType: resourceType,
		ResourceID:   &resourceID,
		UserID:       userID,
		Payload:      string(payloadBytes),
		MaxAttempts:  3,
	})
	if err != nil {
		return nil, httpx.TaskDispatchFailed("failed to create cleanup task").WithErr(err)
	}
	return task, nil
}

func (s *Service) CountRunnableTasks(ctx context.Context) (int, error) {
	return s.repo.CountRunnable(ctx)
}

func (s *Service) ClaimNextRunnableTask(ctx context.Context) (*Task, error) {
	return s.repo.ClaimNextRunnable(ctx)
}

func (s *Service) MarkTaskSucceeded(ctx context.Context, taskID uuid.UUID, result map[string]any) error {
	resultBytes, err := json.Marshal(result)
	if err != nil {
		return err
	}
	return s.repo.MarkSucceeded(ctx, taskID, string(resultBytes))
}

func (s *Service) MarkTaskRetryPending(ctx context.Context, taskID uuid.UUID, errorCode, errorMessage string, nextRunAt time.Time) error {
	return s.repo.MarkRetryPending(ctx, taskID, errorCode, errorMessage, nextRunAt)
}

func (s *Service) MarkTaskFailed(ctx context.Context, taskID uuid.UUID, errorCode, errorMessage string) error {
	return s.repo.MarkFailed(ctx, taskID, errorCode, errorMessage)
}
