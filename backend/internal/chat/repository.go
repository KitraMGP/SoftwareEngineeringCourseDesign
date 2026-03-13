package chat

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("session not found")

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, userID uuid.UUID, input CreateSessionInput) (*Session, error) {
	row := r.pool.QueryRow(ctx, `
		WITH kb_guard AS (
			SELECT 1
			WHERE $4::uuid IS NULL
			   OR EXISTS (
				   SELECT 1
				   FROM knowledge_bases
				   WHERE id = $4
				     AND user_id = $1
				     AND deleted_at IS NULL
			   )
		)
		INSERT INTO sessions (user_id, name, model, knowledge_base_id)
		SELECT $1, $2, $3, $4
		FROM kb_guard
		RETURNING id, user_id, name, model, knowledge_base_id, created_at, updated_at
	`, userID, normalizeOptionalString(input.Name), input.Model, nullableUUID(input.KnowledgeBaseID))

	session, err := scanSession(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("create session: %w", err)
	}
	return session, nil
}

func (r *Repository) List(ctx context.Context, userID uuid.UUID, page, size int, keyword string) (*ListSessionsResult, error) {
	pattern := "%"
	if trimmed := strings.TrimSpace(keyword); trimmed != "" {
		pattern = "%" + trimmed + "%"
	}

	row := r.pool.QueryRow(ctx, `
		SELECT count(*)
		FROM sessions
		WHERE user_id = $1
		  AND deleted_at IS NULL
		  AND ($2 = '%' OR coalesce(name, '') ILIKE $2)
	`, userID, pattern)

	var total int
	if err := row.Scan(&total); err != nil {
		return nil, fmt.Errorf("count sessions: %w", err)
	}

	rows, err := r.pool.Query(ctx, `
		SELECT s.id, s.user_id, s.name, s.model, s.knowledge_base_id, kb.name, s.created_at, s.updated_at
		FROM sessions s
		LEFT JOIN knowledge_bases kb ON kb.id = s.knowledge_base_id
		WHERE s.user_id = $1
		  AND s.deleted_at IS NULL
		  AND ($2 = '%' OR coalesce(s.name, '') ILIKE $2)
		ORDER BY s.updated_at DESC
		LIMIT $3 OFFSET $4
	`, userID, pattern, size, (page-1)*size)
	if err != nil {
		return nil, fmt.Errorf("list sessions: %w", err)
	}
	defer rows.Close()

	items := make([]Session, 0, size)
	for rows.Next() {
		session, err := scanSessionWithKB(rows)
		if err != nil {
			return nil, fmt.Errorf("scan session: %w", err)
		}
		items = append(items, *session)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("iterate sessions: %w", rows.Err())
	}

	return &ListSessionsResult{
		Items: items,
		Total: total,
		Page:  page,
		Size:  size,
	}, nil
}

func (r *Repository) GetDetail(ctx context.Context, userID, sessionID uuid.UUID) (*SessionDetail, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT s.id, s.user_id, s.name, s.model, s.knowledge_base_id, kb.name, s.created_at, s.updated_at
		FROM sessions s
		LEFT JOIN knowledge_bases kb ON kb.id = s.knowledge_base_id
		WHERE s.id = $1
		  AND s.user_id = $2
		  AND s.deleted_at IS NULL
	`, sessionID, userID)

	session, err := scanSessionWithKB(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get session detail: %w", err)
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, session_id, role, reply_to_message_id, content, status, model_used, grounded, prompt_tokens, completion_tokens, total_tokens, created_at, updated_at
		FROM messages
		WHERE session_id = $1
		ORDER BY created_at ASC
	`, sessionID)
	if err != nil {
		return nil, fmt.Errorf("list session messages: %w", err)
	}
	defer rows.Close()

	messages := make([]Message, 0)
	for rows.Next() {
		message, err := scanMessage(rows)
		if err != nil {
			return nil, fmt.Errorf("scan message: %w", err)
		}
		messages = append(messages, *message)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("iterate messages: %w", rows.Err())
	}

	if err := r.loadMessageCitations(ctx, messages); err != nil {
		return nil, fmt.Errorf("load message citations: %w", err)
	}

	return &SessionDetail{
		Session:  *session,
		Messages: messages,
	}, nil
}

func (r *Repository) Delete(ctx context.Context, userID, sessionID uuid.UUID) error {
	result, err := r.pool.Exec(ctx, `
		UPDATE sessions
		SET deleted_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
	`, sessionID, userID)
	if err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) CreateMessage(ctx context.Context, userID, sessionID uuid.UUID, input CreateMessageInput) (*Message, error) {
	messageID := input.ID
	if messageID == uuid.Nil {
		messageID = uuid.New()
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin create message tx: %w", err)
	}
	defer tx.Rollback(ctx)

	row := tx.QueryRow(ctx, `
		UPDATE sessions
		SET updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		  AND user_id = $2
		  AND deleted_at IS NULL
		RETURNING id
	`, sessionID, userID)

	var guardedSessionID uuid.UUID
	if err := row.Scan(&guardedSessionID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("guard session before create message: %w", err)
	}

	messageRow := tx.QueryRow(ctx, `
		INSERT INTO messages (
			id,
			session_id,
			role,
			reply_to_message_id,
			content,
			status,
			model_used,
			grounded,
			prompt_tokens,
			completion_tokens,
			total_tokens
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, session_id, role, reply_to_message_id, content, status, model_used, grounded, prompt_tokens, completion_tokens, total_tokens, created_at, updated_at
	`, messageID, guardedSessionID, input.Role, nullableUUID(input.ReplyToMessageID), input.Content, input.Status, normalizeOptionalString(input.ModelUsed), input.Grounded, input.PromptTokens, input.CompletionTokens, input.TotalTokens)

	message, err := scanMessage(messageRow)
	if err != nil {
		return nil, fmt.Errorf("insert message: %w", err)
	}

	if len(input.Citations) > 0 {
		var batch pgx.Batch
		for _, citation := range input.Citations {
			batch.Queue(`
				INSERT INTO message_citations (
					message_id,
					document_chunk_id,
					document_id,
					knowledge_base_id,
					rank_no
				)
				VALUES ($1, $2, $3, $4, $5)
			`, message.ID, citation.DocumentChunkID, citation.DocumentID, citation.KnowledgeBaseID, citation.RankNo)
		}

		results := tx.SendBatch(ctx, &batch)
		for range input.Citations {
			if _, err := results.Exec(); err != nil {
				results.Close()
				return nil, fmt.Errorf("insert message citation: %w", err)
			}
		}
		if err := results.Close(); err != nil {
			return nil, fmt.Errorf("close citation batch: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit create message tx: %w", err)
	}
	return message, nil
}

func (r *Repository) UpdateMessage(ctx context.Context, userID, sessionID, messageID uuid.UUID, input UpdateMessageInput) (*Message, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin update message tx: %w", err)
	}
	defer tx.Rollback(ctx)

	row := tx.QueryRow(ctx, `
		UPDATE sessions
		SET updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		  AND user_id = $2
		  AND deleted_at IS NULL
		RETURNING id
	`, sessionID, userID)

	var guardedSessionID uuid.UUID
	if err := row.Scan(&guardedSessionID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("guard session before update message: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		DELETE FROM message_citations
		WHERE message_id = $1
	`, messageID); err != nil {
		return nil, fmt.Errorf("delete existing message citations: %w", err)
	}

	messageRow := tx.QueryRow(ctx, `
		UPDATE messages
		SET content = $4,
		    status = $5,
		    model_used = $6,
		    grounded = $7,
		    prompt_tokens = $8,
		    completion_tokens = $9,
		    total_tokens = $10,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		  AND session_id = $2
		  AND EXISTS (
			  SELECT 1
			  FROM sessions
			  WHERE id = $2
			    AND user_id = $3
			    AND deleted_at IS NULL
		  )
		RETURNING id, session_id, role, reply_to_message_id, content, status, model_used, grounded, prompt_tokens, completion_tokens, total_tokens, created_at, updated_at
	`, messageID, guardedSessionID, userID, input.Content, input.Status, normalizeOptionalString(input.ModelUsed), input.Grounded, input.PromptTokens, input.CompletionTokens, input.TotalTokens)

	message, err := scanMessage(messageRow)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("update message: %w", err)
	}

	if len(input.Citations) > 0 {
		var batch pgx.Batch
		for _, citation := range input.Citations {
			batch.Queue(`
				INSERT INTO message_citations (
					message_id,
					document_chunk_id,
					document_id,
					knowledge_base_id,
					rank_no
				)
				VALUES ($1, $2, $3, $4, $5)
			`, message.ID, citation.DocumentChunkID, citation.DocumentID, citation.KnowledgeBaseID, citation.RankNo)
		}

		results := tx.SendBatch(ctx, &batch)
		for range input.Citations {
			if _, err := results.Exec(); err != nil {
				results.Close()
				return nil, fmt.Errorf("insert regenerated message citation: %w", err)
			}
		}
		if err := results.Close(); err != nil {
			return nil, fmt.Errorf("close regenerated citation batch: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit update message tx: %w", err)
	}
	return message, nil
}

func (r *Repository) loadMessageCitations(ctx context.Context, messages []Message) error {
	messageIDs := make([]uuid.UUID, 0, len(messages))
	messageIndex := make(map[uuid.UUID]int, len(messages))
	for idx := range messages {
		messageIDs = append(messageIDs, messages[idx].ID)
		messageIndex[messages[idx].ID] = idx
	}
	if len(messageIDs) == 0 {
		return nil
	}

	rows, err := r.pool.Query(ctx, `
		SELECT
			mc.id,
			mc.message_id,
			mc.document_chunk_id,
			mc.document_id,
			mc.knowledge_base_id,
			mc.rank_no,
			COALESCE(NULLIF(d.title, ''), f.original_filename, 'Untitled Document') AS document_name,
			dc.source_page
		FROM message_citations mc
		INNER JOIN document_chunks dc ON dc.id = mc.document_chunk_id
		INNER JOIN documents d ON d.id = mc.document_id
		LEFT JOIN files f ON f.id = d.file_id
		WHERE mc.message_id = ANY($1::uuid[])
		ORDER BY mc.message_id ASC, mc.rank_no ASC
	`, messageIDs)
	if err != nil {
		return fmt.Errorf("query citations: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			messageID  uuid.UUID
			citation   MessageCitation
			sourcePage sql.NullInt32
		)
		if err := rows.Scan(
			&citation.ID,
			&messageID,
			&citation.DocumentChunkID,
			&citation.DocumentID,
			&citation.KnowledgeBaseID,
			&citation.RankNo,
			&citation.DocumentName,
			&sourcePage,
		); err != nil {
			return fmt.Errorf("scan citation: %w", err)
		}
		if sourcePage.Valid {
			page := int(sourcePage.Int32)
			citation.SourcePage = &page
		}
		idx, ok := messageIndex[messageID]
		if !ok {
			continue
		}
		messages[idx].Citations = append(messages[idx].Citations, citation)
	}
	if rows.Err() != nil {
		return fmt.Errorf("iterate citations: %w", rows.Err())
	}
	return nil
}

func scanSession(row pgx.Row) (*Session, error) {
	var (
		session         Session
		name            sql.NullString
		knowledgeBaseID uuid.NullUUID
	)

	if err := row.Scan(
		&session.ID,
		&session.UserID,
		&name,
		&session.Model,
		&knowledgeBaseID,
		&session.CreatedAt,
		&session.UpdatedAt,
	); err != nil {
		return nil, err
	}

	session.Name = nullStringPtr(name)
	session.KnowledgeBaseID = nullUUIDPtr(knowledgeBaseID)
	return &session, nil
}

func scanSessionWithKB(row pgx.Row) (*Session, error) {
	var (
		session           Session
		name              sql.NullString
		knowledgeBaseID   uuid.NullUUID
		knowledgeBaseName sql.NullString
	)

	if err := row.Scan(
		&session.ID,
		&session.UserID,
		&name,
		&session.Model,
		&knowledgeBaseID,
		&knowledgeBaseName,
		&session.CreatedAt,
		&session.UpdatedAt,
	); err != nil {
		return nil, err
	}

	session.Name = nullStringPtr(name)
	session.KnowledgeBaseID = nullUUIDPtr(knowledgeBaseID)
	session.KnowledgeBaseName = nullStringPtr(knowledgeBaseName)
	return &session, nil
}

func scanMessage(row pgx.Row) (*Message, error) {
	var (
		message          Message
		replyToMessageID uuid.NullUUID
		modelUsed        sql.NullString
	)

	if err := row.Scan(
		&message.ID,
		&message.SessionID,
		&message.Role,
		&replyToMessageID,
		&message.Content,
		&message.Status,
		&modelUsed,
		&message.Grounded,
		&message.PromptTokens,
		&message.CompletionTokens,
		&message.TotalTokens,
		&message.CreatedAt,
		&message.UpdatedAt,
	); err != nil {
		return nil, err
	}

	message.ReplyToMessageID = nullUUIDPtr(replyToMessageID)
	message.ModelUsed = nullStringPtr(modelUsed)
	return &message, nil
}

func nullableUUID(id *uuid.UUID) any {
	if id == nil {
		return nil
	}
	return *id
}

func nullUUIDPtr(value uuid.NullUUID) *uuid.UUID {
	if !value.Valid {
		return nil
	}
	v := value.UUID
	return &v
}

func nullStringPtr(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}
	v := value.String
	return &v
}

func normalizeOptionalString(value *string) any {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return trimmed
}
