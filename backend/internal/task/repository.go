package task

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

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

func (r *Repository) Create(ctx context.Context, input CreateTaskInput) (*Task, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO tasks (
			task_type,
			resource_type,
			resource_id,
			user_id,
			status,
			payload,
			max_attempts
		)
		VALUES ($1, $2, $3, $4, 'pending', $5::jsonb, $6)
		RETURNING id, task_type, resource_type, resource_id, user_id, status, attempt_count, max_attempts, next_run_at, started_at, finished_at, error_code, error_message, created_at, updated_at
	`, input.TaskType, input.ResourceType, nullableUUID(input.ResourceID), nullableUUID(input.UserID), input.Payload, input.MaxAttempts)

	task, err := scanTask(row)
	if err != nil {
		return nil, fmt.Errorf("create task: %w", err)
	}
	return task, nil
}

func (r *Repository) HasActiveTask(ctx context.Context, taskType string, resourceID uuid.UUID) (bool, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM tasks
			WHERE task_type = $1
			  AND resource_id = $2
			  AND status IN ('pending', 'running')
		)
	`, taskType, resourceID)

	var exists bool
	if err := row.Scan(&exists); err != nil {
		return false, fmt.Errorf("check active task: %w", err)
	}
	return exists, nil
}

func (r *Repository) CountRunnable(ctx context.Context) (int, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT count(*)
		FROM tasks
		WHERE status = 'pending'
		  AND (next_run_at IS NULL OR next_run_at <= CURRENT_TIMESTAMP)
	`)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("count runnable tasks: %w", err)
	}
	return count, nil
}

func (r *Repository) ClaimNextRunnable(ctx context.Context) (*Task, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin task claim transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	row := tx.QueryRow(ctx, `
		SELECT id
		FROM tasks
		WHERE status = 'pending'
		  AND (next_run_at IS NULL OR next_run_at <= CURRENT_TIMESTAMP)
		ORDER BY created_at ASC
		FOR UPDATE SKIP LOCKED
		LIMIT 1
	`)

	var taskID uuid.UUID
	if err := row.Scan(&taskID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			if err := tx.Commit(ctx); err != nil {
				return nil, fmt.Errorf("commit empty task claim transaction: %w", err)
			}
			return nil, nil
		}
		return nil, fmt.Errorf("select runnable task: %w", err)
	}

	row = tx.QueryRow(ctx, `
		UPDATE tasks
		SET status = 'running',
		    attempt_count = attempt_count + 1,
		    started_at = CURRENT_TIMESTAMP,
		    finished_at = NULL,
		    error_code = NULL,
		    error_message = NULL
		WHERE id = $1
		RETURNING id, task_type, resource_type, resource_id, user_id, status, attempt_count, max_attempts, next_run_at, started_at, finished_at, error_code, error_message, created_at, updated_at
	`, taskID)

	task, err := scanTask(row)
	if err != nil {
		return nil, fmt.Errorf("claim runnable task: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit task claim transaction: %w", err)
	}
	return task, nil
}

func (r *Repository) MarkSucceeded(ctx context.Context, taskID uuid.UUID, result string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE tasks
		SET status = 'succeeded',
		    result = COALESCE($2::jsonb, '{}'::jsonb),
		    next_run_at = NULL,
		    finished_at = CURRENT_TIMESTAMP,
		    error_code = NULL,
		    error_message = NULL
		WHERE id = $1
	`, taskID, emptyJSON(result))
	if err != nil {
		return fmt.Errorf("mark task succeeded: %w", err)
	}
	return nil
}

func (r *Repository) MarkRetryPending(ctx context.Context, taskID uuid.UUID, errorCode, errorMessage string, nextRunAt time.Time) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE tasks
		SET status = 'pending',
		    next_run_at = $2,
		    finished_at = NULL,
		    error_code = $3,
		    error_message = $4
		WHERE id = $1
	`, taskID, nextRunAt, nullableString(errorCode), nullableString(errorMessage))
	if err != nil {
		return fmt.Errorf("mark task pending for retry: %w", err)
	}
	return nil
}

func (r *Repository) MarkFailed(ctx context.Context, taskID uuid.UUID, errorCode, errorMessage string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE tasks
		SET status = 'failed',
		    next_run_at = NULL,
		    finished_at = CURRENT_TIMESTAMP,
		    error_code = $2,
		    error_message = $3
		WHERE id = $1
	`, taskID, nullableString(errorCode), nullableString(errorMessage))
	if err != nil {
		return fmt.Errorf("mark task failed: %w", err)
	}
	return nil
}

func scanTask(row pgx.Row) (*Task, error) {
	var (
		task         Task
		resourceID   uuid.NullUUID
		userID       uuid.NullUUID
		nextRunAt    sql.NullTime
		startedAt    sql.NullTime
		finishedAt   sql.NullTime
		errorCode    sql.NullString
		errorMessage sql.NullString
	)

	if err := row.Scan(
		&task.ID,
		&task.TaskType,
		&task.ResourceType,
		&resourceID,
		&userID,
		&task.Status,
		&task.AttemptCount,
		&task.MaxAttempts,
		&nextRunAt,
		&startedAt,
		&finishedAt,
		&errorCode,
		&errorMessage,
		&task.CreatedAt,
		&task.UpdatedAt,
	); err != nil {
		return nil, err
	}

	task.ResourceID = nullUUIDPtr(resourceID)
	task.UserID = nullUUIDPtr(userID)
	task.NextRunAt = nullTimePtr(nextRunAt)
	task.StartedAt = nullTimePtr(startedAt)
	task.FinishedAt = nullTimePtr(finishedAt)
	task.ErrorCode = nullStringPtr(errorCode)
	task.ErrorMessage = nullStringPtr(errorMessage)
	return &task, nil
}

func emptyJSON(value string) string {
	if value == "" {
		return "{}"
	}
	return value
}

func nullableUUID(id *uuid.UUID) any {
	if id == nil {
		return nil
	}
	return *id
}

func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func nullUUIDPtr(value uuid.NullUUID) *uuid.UUID {
	if !value.Valid {
		return nil
	}
	v := value.UUID
	return &v
}

func nullTimePtr(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}
	v := value.Time
	return &v
}

func nullStringPtr(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}
	v := value.String
	return &v
}
