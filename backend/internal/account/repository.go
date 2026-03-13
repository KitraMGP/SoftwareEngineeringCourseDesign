package account

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound         = errors.New("account not found")
	ErrDuplicateAccount = errors.New("duplicate account")
)

type dbQuerier interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, arguments ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, arguments ...any) pgx.Row
}

type Repository struct {
	pool *pgxpool.Pool
}

type CreateUserParams struct {
	Username     string
	Email        string
	PasswordHash string
}

type CreateSessionParams struct {
	UserID           uuid.UUID
	RefreshTokenHash string
	DeviceLabel      *string
	UserAgent        *string
	IPAddress        *string
	ExpiresAt        time.Time
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) WithTx(ctx context.Context, fn func(q dbQuerier) error) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if err := fn(tx); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

func (r *Repository) CreateUser(ctx context.Context, params CreateUserParams) (*User, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO users (username, email, password_hash, role, status)
		VALUES ($1, $2, $3, 'user', 'active')
		RETURNING id, username, email, nickname, avatar_url, role, status, password_hash, created_at, updated_at, deleted_at
	`, params.Username, params.Email, params.PasswordHash)

	user, err := scanUser(row)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrDuplicateAccount
		}
		return nil, fmt.Errorf("create user: %w", err)
	}
	return user, nil
}

func (r *Repository) FindUserByAccount(ctx context.Context, account string) (*User, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, username, email, nickname, avatar_url, role, status, password_hash, created_at, updated_at, deleted_at
		FROM users
		WHERE deleted_at IS NULL AND (lower(username) = lower($1) OR lower(email) = lower($1))
		LIMIT 1
	`, account)

	user, err := scanUser(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("find user by account: %w", err)
	}
	return user, nil
}

func (r *Repository) FindUserByID(ctx context.Context, userID uuid.UUID) (*User, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, username, email, nickname, avatar_url, role, status, password_hash, created_at, updated_at, deleted_at
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
	`, userID)

	user, err := scanUser(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("find user by id: %w", err)
	}
	return user, nil
}

func (r *Repository) UpdateUserProfile(ctx context.Context, userID uuid.UUID, nickname, avatarURL *string) (*User, error) {
	row := r.pool.QueryRow(ctx, `
		UPDATE users
		SET nickname = $2, avatar_url = $3
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id, username, email, nickname, avatar_url, role, status, password_hash, created_at, updated_at, deleted_at
	`, userID, nickname, avatarURL)

	user, err := scanUser(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("update user profile: %w", err)
	}
	return user, nil
}

func (r *Repository) UpdateUserPassword(ctx context.Context, q dbQuerier, userID uuid.UUID, passwordHash string) error {
	result, err := q.Exec(ctx, `
		UPDATE users
		SET password_hash = $2
		WHERE id = $1 AND deleted_at IS NULL
	`, userID, passwordHash)
	if err != nil {
		return fmt.Errorf("update user password: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) RevokeActiveSessionsByUserID(ctx context.Context, q dbQuerier, userID uuid.UUID) error {
	_, err := q.Exec(ctx, `
		UPDATE user_sessions
		SET revoked_at = CURRENT_TIMESTAMP
		WHERE user_id = $1 AND revoked_at IS NULL
	`, userID)
	if err != nil {
		return fmt.Errorf("revoke active sessions: %w", err)
	}
	return nil
}

func (r *Repository) CreateSession(ctx context.Context, q dbQuerier, params CreateSessionParams) (*UserSession, error) {
	row := q.QueryRow(ctx, `
		INSERT INTO user_sessions (
			user_id,
			refresh_token_hash,
			device_label,
			user_agent,
			ip_address,
			expires_at,
			last_active_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP)
		RETURNING id, user_id, refresh_token_hash, device_label, user_agent, ip_address, expires_at, last_active_at, revoked_at, created_at, updated_at
	`, params.UserID, params.RefreshTokenHash, params.DeviceLabel, params.UserAgent, params.IPAddress, params.ExpiresAt)

	session, err := scanSession(row)
	if err != nil {
		return nil, fmt.Errorf("create user session: %w", err)
	}
	return session, nil
}

func (r *Repository) GetSessionByRefreshHash(ctx context.Context, refreshTokenHash string) (*AuthenticatedSession, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT
			u.id, u.username, u.email, u.nickname, u.avatar_url, u.role, u.status, u.password_hash, u.created_at, u.updated_at, u.deleted_at,
			s.id, s.user_id, s.refresh_token_hash, s.device_label, s.user_agent, s.ip_address, s.expires_at, s.last_active_at, s.revoked_at, s.created_at, s.updated_at
		FROM user_sessions s
		INNER JOIN users u ON u.id = s.user_id
		WHERE s.refresh_token_hash = $1 AND u.deleted_at IS NULL
		LIMIT 1
	`, refreshTokenHash)

	session, err := scanAuthenticatedSession(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get session by refresh token hash: %w", err)
	}
	return session, nil
}

func (r *Repository) RotateSession(ctx context.Context, q dbQuerier, sessionID uuid.UUID, refreshTokenHash string, expiresAt time.Time) error {
	result, err := q.Exec(ctx, `
		UPDATE user_sessions
		SET refresh_token_hash = $2,
		    expires_at = $3,
		    last_active_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND revoked_at IS NULL
	`, sessionID, refreshTokenHash, expiresAt)
	if err != nil {
		return fmt.Errorf("rotate session: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) RevokeSession(ctx context.Context, sessionID uuid.UUID) error {
	result, err := r.pool.Exec(ctx, `
		UPDATE user_sessions
		SET revoked_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND revoked_at IS NULL
	`, sessionID)
	if err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func scanUser(row pgx.Row) (*User, error) {
	var (
		user      User
		nickname  sql.NullString
		avatarURL sql.NullString
		deletedAt sql.NullTime
	)

	if err := row.Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&nickname,
		&avatarURL,
		&user.Role,
		&user.Status,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
		&deletedAt,
	); err != nil {
		return nil, err
	}

	user.Nickname = stringPtr(nickname)
	user.AvatarURL = stringPtr(avatarURL)
	user.DeletedAt = timePtr(deletedAt)
	return &user, nil
}

func scanSession(row pgx.Row) (*UserSession, error) {
	var (
		session      UserSession
		deviceLabel  sql.NullString
		userAgent    sql.NullString
		ipAddress    sql.NullString
		lastActiveAt sql.NullTime
		revokedAt    sql.NullTime
	)

	if err := row.Scan(
		&session.ID,
		&session.UserID,
		&session.RefreshTokenHash,
		&deviceLabel,
		&userAgent,
		&ipAddress,
		&session.ExpiresAt,
		&lastActiveAt,
		&revokedAt,
		&session.CreatedAt,
		&session.UpdatedAt,
	); err != nil {
		return nil, err
	}

	session.DeviceLabel = stringPtr(deviceLabel)
	session.UserAgent = stringPtr(userAgent)
	session.IPAddress = stringPtr(ipAddress)
	session.LastActiveAt = timePtr(lastActiveAt)
	session.RevokedAt = timePtr(revokedAt)
	return &session, nil
}

func scanAuthenticatedSession(row pgx.Row) (*AuthenticatedSession, error) {
	var (
		user      User
		nickname  sql.NullString
		avatarURL sql.NullString
		deletedAt sql.NullTime

		session      UserSession
		deviceLabel  sql.NullString
		userAgent    sql.NullString
		ipAddress    sql.NullString
		lastActiveAt sql.NullTime
		revokedAt    sql.NullTime
	)

	if err := row.Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&nickname,
		&avatarURL,
		&user.Role,
		&user.Status,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
		&deletedAt,
		&session.ID,
		&session.UserID,
		&session.RefreshTokenHash,
		&deviceLabel,
		&userAgent,
		&ipAddress,
		&session.ExpiresAt,
		&lastActiveAt,
		&revokedAt,
		&session.CreatedAt,
		&session.UpdatedAt,
	); err != nil {
		return nil, err
	}

	user.Nickname = stringPtr(nickname)
	user.AvatarURL = stringPtr(avatarURL)
	user.DeletedAt = timePtr(deletedAt)

	session.DeviceLabel = stringPtr(deviceLabel)
	session.UserAgent = stringPtr(userAgent)
	session.IPAddress = stringPtr(ipAddress)
	session.LastActiveAt = timePtr(lastActiveAt)
	session.RevokedAt = timePtr(revokedAt)

	return &AuthenticatedSession{
		User:    user,
		Session: session,
	}, nil
}

func stringPtr(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}
	v := value.String
	return &v
}

func timePtr(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}
	v := value.Time
	return &v
}
