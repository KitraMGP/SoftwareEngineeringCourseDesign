package admin

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"backend/internal/platform/httpx"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) GetOverview(ctx context.Context) (*Overview, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT
			(SELECT count(*) FROM users WHERE deleted_at IS NULL),
			(SELECT count(*) FROM users WHERE deleted_at IS NULL AND status = 'active'),
			(SELECT count(*) FROM knowledge_bases WHERE deleted_at IS NULL),
			(SELECT count(*) FROM documents WHERE deleted_at IS NULL),
			(SELECT count(*) FROM tasks WHERE status = 'pending'),
			(SELECT count(*) FROM tasks WHERE status = 'failed')
	`)

	var result Overview
	if err := row.Scan(
		&result.UserCount,
		&result.ActiveUserCount,
		&result.KnowledgeBaseCount,
		&result.DocumentCount,
		&result.PendingTaskCount,
		&result.FailedTaskCount,
	); err != nil {
		return nil, fmt.Errorf("get admin overview: %w", err)
	}

	return &result, nil
}

func (r *Repository) ListUsers(ctx context.Context, params ListUsersParams) (*ListUsersResult, error) {
	whereSQL, args := buildUserFilters(params.Keyword, params.Status, params.Role)

	countSQL := `SELECT count(*) FROM users u` + whereSQL
	var total int
	if err := r.pool.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("count admin users: %w", err)
	}

	args = append(args, params.Size, httpx.Offset(params.Page, params.Size))
	rows, err := r.pool.Query(ctx, `
		SELECT
			u.id, u.username, u.email, u.nickname, u.avatar_url, u.role, u.status, u.created_at, u.updated_at
		FROM users u
	`+whereSQL+`
		ORDER BY u.created_at DESC
		LIMIT $`+fmt.Sprint(len(args)-1)+` OFFSET $`+fmt.Sprint(len(args))+`
	`, args...)
	if err != nil {
		return nil, fmt.Errorf("list admin users: %w", err)
	}
	defer rows.Close()

	items := make([]User, 0, params.Size)
	for rows.Next() {
		var item User
		if err := rows.Scan(
			&item.ID,
			&item.Username,
			&item.Email,
			&item.Nickname,
			&item.AvatarURL,
			&item.Role,
			&item.Status,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan admin user: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate admin users: %w", err)
	}

	return &ListUsersResult{
		Items: items,
		Total: total,
		Page:  params.Page,
		Size:  params.Size,
	}, nil
}

func (r *Repository) UpdateUserStatus(ctx context.Context, userID uuid.UUID, status string) (*User, error) {
	row := r.pool.QueryRow(ctx, `
		UPDATE users
		SET status = $2
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id, username, email, nickname, avatar_url, role, status, created_at, updated_at
	`, userID, status)

	var item User
	if err := row.Scan(
		&item.ID,
		&item.Username,
		&item.Email,
		&item.Nickname,
		&item.AvatarURL,
		&item.Role,
		&item.Status,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, httpx.NotFound("user not found")
		}
		return nil, fmt.Errorf("update user status: %w", err)
	}

	return &item, nil
}

func (r *Repository) ListTasks(ctx context.Context, params ListTasksParams) (*ListTasksResult, error) {
	whereSQL, args := buildTaskFilters(params.Status, params.TaskType)

	countSQL := `SELECT count(*) FROM tasks t` + whereSQL
	var total int
	if err := r.pool.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("count admin tasks: %w", err)
	}

	args = append(args, params.Size, httpx.Offset(params.Page, params.Size))
	rows, err := r.pool.Query(ctx, `
		SELECT
			t.id,
			t.task_type,
			t.resource_type,
			t.resource_id,
			t.user_id,
			COALESCE(u.email, u.username) AS user_label,
			t.status,
			t.attempt_count,
			t.max_attempts,
			t.next_run_at,
			t.started_at,
			t.finished_at,
			t.error_code,
			t.error_message,
			t.created_at,
			t.updated_at
		FROM tasks t
		LEFT JOIN users u ON u.id = t.user_id
	`+whereSQL+`
		ORDER BY t.created_at DESC
		LIMIT $`+fmt.Sprint(len(args)-1)+` OFFSET $`+fmt.Sprint(len(args))+`
	`, args...)
	if err != nil {
		return nil, fmt.Errorf("list admin tasks: %w", err)
	}
	defer rows.Close()

	items := make([]Task, 0, params.Size)
	for rows.Next() {
		item, err := scanAdminTask(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate admin tasks: %w", err)
	}

	return &ListTasksResult{
		Items: items,
		Total: total,
		Page:  params.Page,
		Size:  params.Size,
	}, nil
}

func (r *Repository) RetryTask(ctx context.Context, taskID uuid.UUID) (*Task, error) {
	row := r.pool.QueryRow(ctx, `
		UPDATE tasks
		SET status = 'pending',
		    next_run_at = CURRENT_TIMESTAMP,
		    started_at = NULL,
		    finished_at = NULL,
		    error_code = NULL,
		    error_message = NULL
		WHERE id = $1
		  AND status = 'failed'
		RETURNING id, task_type, resource_type, resource_id, user_id, NULL::text, status, attempt_count, max_attempts, next_run_at, started_at, finished_at, error_code, error_message, created_at, updated_at
	`, taskID)

	task, err := scanAdminTask(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, httpx.NotFound("retryable task not found")
		}
		return nil, fmt.Errorf("retry admin task: %w", err)
	}

	return task, nil
}

func (r *Repository) ListProviderConfigs(ctx context.Context) ([]ProviderConfig, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, provider, base_url, default_chat_model, default_embedding_model, is_enabled,
		       (char_length(encrypted_api_key) > 0) AS has_api_key, created_at, updated_at
		FROM provider_configs
		ORDER BY provider ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("list provider configs: %w", err)
	}
	defer rows.Close()

	items := make([]ProviderConfig, 0)
	for rows.Next() {
		var item ProviderConfig
		if err := rows.Scan(
			&item.ID,
			&item.Provider,
			&item.BaseURL,
			&item.DefaultChatModel,
			&item.DefaultEmbeddingModel,
			&item.IsEnabled,
			&item.HasAPIKey,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan provider config: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate provider configs: %w", err)
	}
	return items, nil
}

func (r *Repository) ListSystemSettings(ctx context.Context) ([]SystemSetting, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT key, category, value, description, updated_at
		FROM system_settings
		ORDER BY category ASC, key ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("list system settings: %w", err)
	}
	defer rows.Close()

	items := make([]SystemSetting, 0)
	for rows.Next() {
		var (
			item        SystemSetting
			value       []byte
			description sql.NullString
		)
		if err := rows.Scan(&item.Key, &item.Category, &value, &description, &item.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan system setting: %w", err)
		}
		item.Value = json.RawMessage(value)
		if description.Valid {
			item.Description = &description.String
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate system settings: %w", err)
	}
	return items, nil
}

func (r *Repository) ListQuotaPolicies(ctx context.Context) ([]QuotaPolicy, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, scope_type, scope_id, daily_total_tokens_limit, storage_bytes_limit, document_count_limit, warn_ratio::text, created_at, updated_at
		FROM quota_policies
		ORDER BY scope_type ASC, created_at ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("list quota policies: %w", err)
	}
	defer rows.Close()

	items := make([]QuotaPolicy, 0)
	for rows.Next() {
		var (
			item    QuotaPolicy
			scopeID uuid.NullUUID
		)
		if err := rows.Scan(
			&item.ID,
			&item.ScopeType,
			&scopeID,
			&item.DailyTotalTokensLimit,
			&item.StorageBytesLimit,
			&item.DocumentCountLimit,
			&item.WarnRatio,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan quota policy: %w", err)
		}
		if scopeID.Valid {
			item.ScopeID = &scopeID.UUID
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate quota policies: %w", err)
	}
	return items, nil
}

func (r *Repository) ListAuditLogs(ctx context.Context, page, size int) (*ListAuditLogsResult, error) {
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT count(*) FROM audit_logs`).Scan(&total); err != nil {
		return nil, fmt.Errorf("count audit logs: %w", err)
	}

	rows, err := r.pool.Query(ctx, `
		SELECT
			a.id,
			a.actor_user_id,
			COALESCE(actor.email, actor.username) AS actor_label,
			a.actor_role,
			a.action,
			a.resource_type,
			a.resource_id,
			a.target_user_id,
			COALESCE(target_user.email, target_user.username) AS target_label,
			a.result,
			a.metadata,
			a.created_at
		FROM audit_logs a
		LEFT JOIN users actor ON actor.id = a.actor_user_id
		LEFT JOIN users target_user ON target_user.id = a.target_user_id
		ORDER BY a.created_at DESC
		LIMIT $1 OFFSET $2
	`, size, httpx.Offset(page, size))
	if err != nil {
		return nil, fmt.Errorf("list audit logs: %w", err)
	}
	defer rows.Close()

	items := make([]AuditLog, 0, size)
	for rows.Next() {
		var (
			item         AuditLog
			actorUserID  uuid.NullUUID
			resourceID   uuid.NullUUID
			targetUserID uuid.NullUUID
			actorLabel   sql.NullString
			resourceType sql.NullString
			targetLabel  sql.NullString
			metadata     []byte
		)

		if err := rows.Scan(
			&item.ID,
			&actorUserID,
			&actorLabel,
			&item.ActorRole,
			&item.Action,
			&resourceType,
			&resourceID,
			&targetUserID,
			&targetLabel,
			&item.Result,
			&metadata,
			&item.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan audit log: %w", err)
		}

		if actorUserID.Valid {
			item.ActorUserID = &actorUserID.UUID
		}
		if actorLabel.Valid {
			item.ActorLabel = &actorLabel.String
		}
		if resourceType.Valid {
			item.ResourceType = &resourceType.String
		}
		if resourceID.Valid {
			item.ResourceID = &resourceID.UUID
		}
		if targetUserID.Valid {
			item.TargetUserID = &targetUserID.UUID
		}
		if targetLabel.Valid {
			item.TargetLabel = &targetLabel.String
		}
		item.Metadata = json.RawMessage(metadata)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate audit logs: %w", err)
	}

	return &ListAuditLogsResult{
		Items: items,
		Total: total,
		Page:  page,
		Size:  size,
	}, nil
}

func (r *Repository) CreateAuditLog(ctx context.Context, input CreateAuditLogInput) error {
	metadata := input.Metadata
	if strings.TrimSpace(metadata) == "" {
		metadata = "{}"
	}

	_, err := r.pool.Exec(ctx, `
		INSERT INTO audit_logs (
			actor_user_id,
			actor_role,
			action,
			resource_type,
			resource_id,
			target_user_id,
			result,
			metadata
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8::jsonb)
	`, nullableUUID(input.ActorUserID), input.ActorRole, input.Action, nullableStringPtr(input.ResourceType), nullableUUID(input.ResourceID), nullableUUID(input.TargetUserID), input.Result, metadata)
	if err != nil {
		return fmt.Errorf("create audit log: %w", err)
	}
	return nil
}

func buildUserFilters(keyword, status, role string) (string, []any) {
	conditions := []string{"u.deleted_at IS NULL"}
	args := make([]any, 0, 3)

	if trimmed := strings.TrimSpace(keyword); trimmed != "" {
		args = append(args, "%"+trimmed+"%")
		conditions = append(conditions, fmt.Sprintf("(u.username ILIKE $%d OR u.email ILIKE $%d OR COALESCE(u.nickname, '') ILIKE $%d)", len(args), len(args), len(args)))
	}
	if trimmed := strings.TrimSpace(status); trimmed != "" {
		args = append(args, trimmed)
		conditions = append(conditions, fmt.Sprintf("u.status = $%d", len(args)))
	}
	if trimmed := strings.TrimSpace(role); trimmed != "" {
		args = append(args, trimmed)
		conditions = append(conditions, fmt.Sprintf("u.role = $%d", len(args)))
	}

	return " WHERE " + strings.Join(conditions, " AND "), args
}

func buildTaskFilters(status, taskType string) (string, []any) {
	conditions := make([]string, 0, 2)
	args := make([]any, 0, 2)

	if trimmed := strings.TrimSpace(status); trimmed != "" {
		args = append(args, trimmed)
		conditions = append(conditions, fmt.Sprintf("t.status = $%d", len(args)))
	}
	if trimmed := strings.TrimSpace(taskType); trimmed != "" {
		args = append(args, trimmed)
		conditions = append(conditions, fmt.Sprintf("t.task_type = $%d", len(args)))
	}

	if len(conditions) == 0 {
		return "", args
	}

	return " WHERE " + strings.Join(conditions, " AND "), args
}

func scanAdminTask(row pgx.Row) (*Task, error) {
	var (
		item         Task
		resourceID   uuid.NullUUID
		userID       uuid.NullUUID
		userLabel    sql.NullString
		nextRunAt    sql.NullTime
		startedAt    sql.NullTime
		finishedAt   sql.NullTime
		errorCode    sql.NullString
		errorMessage sql.NullString
	)

	if err := row.Scan(
		&item.ID,
		&item.TaskType,
		&item.ResourceType,
		&resourceID,
		&userID,
		&userLabel,
		&item.Status,
		&item.AttemptCount,
		&item.MaxAttempts,
		&nextRunAt,
		&startedAt,
		&finishedAt,
		&errorCode,
		&errorMessage,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		return nil, err
	}

	if resourceID.Valid {
		item.ResourceID = &resourceID.UUID
	}
	if userID.Valid {
		item.UserID = &userID.UUID
	}
	if userLabel.Valid {
		item.UserLabel = &userLabel.String
	}
	if nextRunAt.Valid {
		item.NextRunAt = &nextRunAt.Time
	}
	if startedAt.Valid {
		item.StartedAt = &startedAt.Time
	}
	if finishedAt.Valid {
		item.FinishedAt = &finishedAt.Time
	}
	if errorCode.Valid {
		item.ErrorCode = &errorCode.String
	}
	if errorMessage.Valid {
		item.ErrorMessage = &errorMessage.String
	}

	return &item, nil
}

func nullableUUID(id *uuid.UUID) any {
	if id == nil {
		return nil
	}
	return *id
}

func nullableStringPtr(value *string) any {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return trimmed
}
