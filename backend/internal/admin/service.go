package admin

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"backend/internal/account"
	"backend/internal/platform/auth"
	"backend/internal/platform/httpx"

	"github.com/google/uuid"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetOverview(ctx context.Context) (*Overview, error) {
	overview, err := s.repo.GetOverview(ctx)
	if err != nil {
		return nil, httpx.Internal("failed to load admin overview").WithErr(err)
	}
	return overview, nil
}

func (s *Service) ListUsers(ctx context.Context, page, size int, keyword, status, role string) (*ListUsersResult, error) {
	if status != "" && status != account.StatusActive && status != account.StatusFrozen {
		return nil, httpx.ValidationFailed(httpx.FieldError{Field: "status", Message: "invalid status"})
	}
	if role != "" && role != account.RoleUser && role != account.RoleAdmin {
		return nil, httpx.ValidationFailed(httpx.FieldError{Field: "role", Message: "invalid role"})
	}

	result, err := s.repo.ListUsers(ctx, ListUsersParams{
		Page:    page,
		Size:    size,
		Keyword: keyword,
		Status:  status,
		Role:    role,
	})
	if err != nil {
		return nil, httpx.Internal("failed to list admin users").WithErr(err)
	}
	return result, nil
}

func (s *Service) FreezeUser(ctx context.Context, actor auth.Principal, userID uuid.UUID) (*User, error) {
	return s.updateUserStatus(ctx, actor, userID, account.StatusFrozen, "admin.user.freeze")
}

func (s *Service) UnfreezeUser(ctx context.Context, actor auth.Principal, userID uuid.UUID) (*User, error) {
	return s.updateUserStatus(ctx, actor, userID, account.StatusActive, "admin.user.unfreeze")
}

func (s *Service) ListTasks(ctx context.Context, page, size int, status, taskType string) (*ListTasksResult, error) {
	result, err := s.repo.ListTasks(ctx, ListTasksParams{
		Page:     page,
		Size:     size,
		Status:   status,
		TaskType: taskType,
	})
	if err != nil {
		return nil, httpx.Internal("failed to list admin tasks").WithErr(err)
	}
	return result, nil
}

func (s *Service) RetryTask(ctx context.Context, actor auth.Principal, taskID uuid.UUID) (*Task, error) {
	task, err := s.repo.RetryTask(ctx, taskID)
	if err != nil {
		if appErr, ok := httpx.AsAppError(err); ok {
			return nil, appErr
		}
		return nil, httpx.Internal("failed to retry task").WithErr(err)
	}

	metadata, _ := json.Marshal(map[string]any{
		"task_id":   task.ID,
		"task_type": task.TaskType,
	})
	_ = s.repo.CreateAuditLog(ctx, CreateAuditLogInput{
		ActorUserID:  &actor.UserID,
		ActorRole:    actor.Role,
		Action:       "admin.task.retry",
		ResourceType: stringPtr("task"),
		ResourceID:   &task.ID,
		Result:       "success",
		Metadata:     string(metadata),
	})

	return task, nil
}

func (s *Service) ListProviderConfigs(ctx context.Context) ([]ProviderConfig, error) {
	items, err := s.repo.ListProviderConfigs(ctx)
	if err != nil {
		return nil, httpx.Internal("failed to list provider configs").WithErr(err)
	}
	return items, nil
}

func (s *Service) ListSystemSettings(ctx context.Context) ([]SystemSetting, error) {
	items, err := s.repo.ListSystemSettings(ctx)
	if err != nil {
		return nil, httpx.Internal("failed to list system settings").WithErr(err)
	}
	return items, nil
}

func (s *Service) ListQuotaPolicies(ctx context.Context) ([]QuotaPolicy, error) {
	items, err := s.repo.ListQuotaPolicies(ctx)
	if err != nil {
		return nil, httpx.Internal("failed to list quota policies").WithErr(err)
	}
	return items, nil
}

func (s *Service) ListAuditLogs(ctx context.Context, page, size int) (*ListAuditLogsResult, error) {
	result, err := s.repo.ListAuditLogs(ctx, page, size)
	if err != nil {
		return nil, httpx.Internal("failed to list audit logs").WithErr(err)
	}
	return result, nil
}

func (s *Service) updateUserStatus(ctx context.Context, actor auth.Principal, userID uuid.UUID, status, action string) (*User, error) {
	if actor.UserID == userID && status == account.StatusFrozen {
		return nil, httpx.Conflict("cannot freeze current admin account")
	}

	user, err := s.repo.UpdateUserStatus(ctx, userID, status)
	if err != nil {
		if appErr, ok := httpx.AsAppError(err); ok {
			return nil, appErr
		}
		return nil, httpx.Internal("failed to update user status").WithErr(err)
	}

	metadata, _ := json.Marshal(map[string]any{
		"user_id": user.ID,
		"status":  user.Status,
	})
	_ = s.repo.CreateAuditLog(ctx, CreateAuditLogInput{
		ActorUserID:  &actor.UserID,
		ActorRole:    actor.Role,
		Action:       action,
		ResourceType: stringPtr("user"),
		ResourceID:   &user.ID,
		TargetUserID: &user.ID,
		Result:       "success",
		Metadata:     string(metadata),
	})

	return user, nil
}

func stringPtr(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func formatMetadata(pairs map[string]any) string {
	if len(pairs) == 0 {
		return "{}"
	}
	data, err := json.Marshal(pairs)
	if err != nil {
		return fmt.Sprintf(`{"error":"%s"}`, err.Error())
	}
	return string(data)
}
