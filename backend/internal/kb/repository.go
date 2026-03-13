package kb

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"backend/internal/task"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("knowledge base not found")

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, userID uuid.UUID, input CreateKnowledgeBaseInput) (*KnowledgeBase, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO knowledge_bases (
			user_id,
			name,
			description,
			embedding_model,
			prompt_template,
			retrieval_top_k,
			similarity_threshold
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, user_id, name, description, embedding_model, prompt_template, retrieval_top_k, similarity_threshold, last_indexed_at, created_at, updated_at
	`, userID, input.Name, input.Description, input.EmbeddingModel, input.PromptTemplate, input.RetrievalTopK, input.SimilarityThreshold)

	kb, err := scanKnowledgeBase(row)
	if err != nil {
		return nil, fmt.Errorf("create knowledge base: %w", err)
	}
	return kb, nil
}

func (r *Repository) List(ctx context.Context, userID uuid.UUID, page, size int, keyword string) (*ListKnowledgeBasesResult, error) {
	pattern := "%"
	if trimmed := strings.TrimSpace(keyword); trimmed != "" {
		pattern = "%" + trimmed + "%"
	}

	row := r.pool.QueryRow(ctx, `
		SELECT count(*)
		FROM knowledge_bases
		WHERE user_id = $1
		  AND deleted_at IS NULL
		  AND ($2 = '%' OR name ILIKE $2)
	`, userID, pattern)

	var total int
	if err := row.Scan(&total); err != nil {
		return nil, fmt.Errorf("count knowledge bases: %w", err)
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, name, description, embedding_model, prompt_template, retrieval_top_k, similarity_threshold, last_indexed_at, created_at, updated_at
		FROM knowledge_bases
		WHERE user_id = $1
		  AND deleted_at IS NULL
		  AND ($2 = '%' OR name ILIKE $2)
		ORDER BY updated_at DESC
		LIMIT $3 OFFSET $4
	`, userID, pattern, size, (page-1)*size)
	if err != nil {
		return nil, fmt.Errorf("list knowledge bases: %w", err)
	}
	defer rows.Close()

	items := make([]KnowledgeBase, 0, size)
	for rows.Next() {
		kb, err := scanKnowledgeBase(rows)
		if err != nil {
			return nil, fmt.Errorf("scan knowledge base: %w", err)
		}
		items = append(items, *kb)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("iterate knowledge bases: %w", rows.Err())
	}

	return &ListKnowledgeBasesResult{
		Items: items,
		Total: total,
		Page:  page,
		Size:  size,
	}, nil
}

func (r *Repository) Get(ctx context.Context, userID, knowledgeBaseID uuid.UUID) (*KnowledgeBase, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, user_id, name, description, embedding_model, prompt_template, retrieval_top_k, similarity_threshold, last_indexed_at, created_at, updated_at
		FROM knowledge_bases
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
	`, knowledgeBaseID, userID)

	kb, err := scanKnowledgeBase(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get knowledge base: %w", err)
	}
	return kb, nil
}

func (r *Repository) Update(ctx context.Context, userID, knowledgeBaseID uuid.UUID, input UpdateKnowledgeBaseInput) (*KnowledgeBase, error) {
	row := r.pool.QueryRow(ctx, `
		UPDATE knowledge_bases
		SET name = $3,
		    description = $4,
		    prompt_template = $5,
		    retrieval_top_k = $6,
		    similarity_threshold = $7
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
		RETURNING id, user_id, name, description, embedding_model, prompt_template, retrieval_top_k, similarity_threshold, last_indexed_at, created_at, updated_at
	`, knowledgeBaseID, userID, input.Name, input.Description, input.PromptTemplate, input.RetrievalTopK, input.SimilarityThreshold)

	kb, err := scanKnowledgeBase(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("update knowledge base: %w", err)
	}
	return kb, nil
}

func (r *Repository) Delete(ctx context.Context, userID, knowledgeBaseID uuid.UUID) error {
	result, err := r.pool.Exec(ctx, `
		UPDATE knowledge_bases
		SET deleted_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
	`, knowledgeBaseID, userID)
	if err != nil {
		return fmt.Errorf("delete knowledge base: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) ListDocuments(ctx context.Context, userID, knowledgeBaseID uuid.UUID, page, size int) (*ListDocumentsResult, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT count(*)
		FROM documents d
		INNER JOIN knowledge_bases kb ON kb.id = d.knowledge_base_id
		WHERE d.knowledge_base_id = $1
		  AND kb.user_id = $2
		  AND d.deleted_at IS NULL
		  AND kb.deleted_at IS NULL
	`, knowledgeBaseID, userID)

	var total int
	if err := row.Scan(&total); err != nil {
		return nil, fmt.Errorf("count documents: %w", err)
	}

	rows, err := r.pool.Query(ctx, `
		SELECT d.id, d.knowledge_base_id, d.file_id, d.title, f.original_filename, d.status, d.error_message, d.content_length, d.chunk_count, d.created_at, d.updated_at
		FROM documents d
		INNER JOIN knowledge_bases kb ON kb.id = d.knowledge_base_id
		LEFT JOIN files f ON f.id = d.file_id
		WHERE d.knowledge_base_id = $1
		  AND kb.user_id = $2
		  AND d.deleted_at IS NULL
		  AND kb.deleted_at IS NULL
		ORDER BY d.created_at DESC
		LIMIT $3 OFFSET $4
	`, knowledgeBaseID, userID, size, (page-1)*size)
	if err != nil {
		return nil, fmt.Errorf("list documents: %w", err)
	}
	defer rows.Close()

	items := make([]Document, 0, size)
	for rows.Next() {
		document, err := scanDocument(rows)
		if err != nil {
			return nil, fmt.Errorf("scan document: %w", err)
		}
		items = append(items, *document)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("iterate documents: %w", rows.Err())
	}

	return &ListDocumentsResult{
		Items: items,
		Total: total,
		Page:  page,
		Size:  size,
	}, nil
}

func (r *Repository) GetDocument(ctx context.Context, userID, knowledgeBaseID, documentID uuid.UUID) (*Document, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT d.id, d.knowledge_base_id, d.file_id, d.title, f.original_filename, d.status, d.error_message, d.content_length, d.chunk_count, d.created_at, d.updated_at
		FROM documents d
		INNER JOIN knowledge_bases kb ON kb.id = d.knowledge_base_id
		LEFT JOIN files f ON f.id = d.file_id
		WHERE d.id = $1
		  AND d.knowledge_base_id = $2
		  AND kb.user_id = $3
		  AND d.deleted_at IS NULL
		  AND kb.deleted_at IS NULL
	`, documentID, knowledgeBaseID, userID)

	document, err := scanDocument(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get document: %w", err)
	}
	return document, nil
}

func (r *Repository) DeleteDocument(ctx context.Context, userID, knowledgeBaseID, documentID uuid.UUID) error {
	result, err := r.pool.Exec(ctx, `
		UPDATE documents d
		SET deleted_at = CURRENT_TIMESTAMP,
		    status = 'deleting'
		FROM knowledge_bases kb
		WHERE d.id = $1
		  AND d.knowledge_base_id = $2
		  AND kb.id = d.knowledge_base_id
		  AND kb.user_id = $3
		  AND d.deleted_at IS NULL
		  AND kb.deleted_at IS NULL
	`, documentID, knowledgeBaseID, userID)
	if err != nil {
		return fmt.Errorf("delete document: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) HasDuplicateDocument(ctx context.Context, knowledgeBaseID uuid.UUID, sha256 string) (bool, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM documents d
			INNER JOIN files f ON f.id = d.file_id
			WHERE d.knowledge_base_id = $1
			  AND d.deleted_at IS NULL
			  AND f.deleted_at IS NULL
			  AND f.sha256 = $2
		)
	`, knowledgeBaseID, sha256)

	var exists bool
	if err := row.Scan(&exists); err != nil {
		return false, fmt.Errorf("check duplicate document: %w", err)
	}
	return exists, nil
}

func (r *Repository) CreateUploadedDocument(ctx context.Context, params CreateUploadedDocumentParams) (*UploadDocumentResult, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin upload transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var fileID uuid.UUID
	if err := tx.QueryRow(ctx, `
		INSERT INTO files (
			user_id,
			storage_provider,
			bucket_name,
			object_key,
			original_filename,
			mime_type,
			size_bytes,
			sha256
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`, params.UserID, params.StorageProvider, params.BucketName, params.ObjectKey, params.OriginalFilename, params.MIMEType, params.SizeBytes, params.SHA256).Scan(&fileID); err != nil {
		return nil, fmt.Errorf("insert file: %w", err)
	}

	var documentID uuid.UUID
	if err := tx.QueryRow(ctx, `
		INSERT INTO documents (
			knowledge_base_id,
			file_id,
			title,
			status
		)
		VALUES ($1, $2, $3, 'pending')
		RETURNING id
	`, params.KnowledgeBaseID, fileID, params.Title).Scan(&documentID); err != nil {
		return nil, fmt.Errorf("insert document: %w", err)
	}

	payload, err := json.Marshal(map[string]any{
		"knowledge_base_id": params.KnowledgeBaseID,
		"document_id":       documentID,
		"file_id":           fileID,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal ingest task payload: %w", err)
	}

	var taskID uuid.UUID
	var status string
	if err := tx.QueryRow(ctx, `
		INSERT INTO tasks (
			task_type,
			resource_type,
			resource_id,
			user_id,
			status,
			payload,
			max_attempts
		)
		VALUES ($1, $2, $3, $4, 'pending', $5::jsonb, 3)
		RETURNING id, status
	`, task.TaskTypeDocumentIngest, task.ResourceTypeDocument, documentID, params.UserID, string(payload)).Scan(&taskID, &status); err != nil {
		return nil, fmt.Errorf("insert document ingest task: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit upload transaction: %w", err)
	}

	return &UploadDocumentResult{
		DocumentID: documentID,
		TaskID:     taskID,
		Status:     status,
	}, nil
}

func (r *Repository) GetDocumentIngestSource(ctx context.Context, documentID uuid.UUID) (*DocumentIngestSource, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT d.id, d.knowledge_base_id, f.id, f.bucket_name, f.object_key, f.original_filename, f.mime_type
		FROM documents d
		INNER JOIN knowledge_bases kb ON kb.id = d.knowledge_base_id
		INNER JOIN files f ON f.id = d.file_id
		WHERE d.id = $1
		  AND d.deleted_at IS NULL
		  AND kb.deleted_at IS NULL
	`, documentID)

	var source DocumentIngestSource
	if err := row.Scan(
		&source.DocumentID,
		&source.KnowledgeBaseID,
		&source.FileID,
		&source.BucketName,
		&source.ObjectKey,
		&source.OriginalFilename,
		&source.MIMEType,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get document ingest source: %w", err)
	}
	return &source, nil
}

func (r *Repository) MarkDocumentProcessing(ctx context.Context, documentID uuid.UUID) error {
	return r.setDocumentState(ctx, documentID, "processing", nil)
}

func (r *Repository) MarkDocumentPending(ctx context.Context, documentID uuid.UUID, errorMessage string) error {
	return r.setDocumentState(ctx, documentID, "pending", &errorMessage)
}

func (r *Repository) MarkDocumentFailed(ctx context.Context, documentID uuid.UUID, errorMessage string) error {
	return r.setDocumentState(ctx, documentID, "failed", &errorMessage)
}

func (r *Repository) ReplaceDocumentContent(ctx context.Context, documentID uuid.UUID, content string, chunks []DocumentChunkInput) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin document ingest transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var knowledgeBaseID uuid.UUID
	if err := tx.QueryRow(ctx, `
		SELECT knowledge_base_id
		FROM documents
		WHERE id = $1 AND deleted_at IS NULL
	`, documentID).Scan(&knowledgeBaseID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return fmt.Errorf("load document knowledge base: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		DELETE FROM document_chunks
		WHERE document_id = $1
	`, documentID); err != nil {
		return fmt.Errorf("delete existing document chunks: %w", err)
	}

	if len(chunks) > 0 {
		var batch pgx.Batch
		for _, chunk := range chunks {
			batch.Queue(`
				INSERT INTO document_chunks (
					knowledge_base_id,
					document_id,
					chunk_index,
					heading_path,
					content,
					token_count,
					source_page,
					embedding
				)
				VALUES ($1, $2, $3, NULL, $4, $5, NULL, $6::vector)
			`, knowledgeBaseID, documentID, chunk.ChunkIndex, chunk.Content, chunk.TokenCount, chunk.Embedding)
		}

		results := tx.SendBatch(ctx, &batch)
		for range chunks {
			if _, err := results.Exec(); err != nil {
				results.Close()
				return fmt.Errorf("insert document chunk: %w", err)
			}
		}
		if err := results.Close(); err != nil {
			return fmt.Errorf("close document chunk batch: %w", err)
		}
	}

	contentLength := len([]rune(content))
	if _, err := tx.Exec(ctx, `
		UPDATE documents
		SET status = 'available',
		    error_message = NULL,
		    content_text = $2,
		    content_length = $3,
		    chunk_count = $4
		WHERE id = $1
	`, documentID, content, contentLength, len(chunks)); err != nil {
		return fmt.Errorf("update document content: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		UPDATE knowledge_bases
		SET last_indexed_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`, knowledgeBaseID); err != nil {
		return fmt.Errorf("update knowledge base last indexed time: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit document ingest transaction: %w", err)
	}
	return nil
}

func (r *Repository) ListDocumentIDsForKnowledgeBase(ctx context.Context, knowledgeBaseID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id
		FROM documents
		WHERE knowledge_base_id = $1
		  AND deleted_at IS NULL
		  AND file_id IS NOT NULL
		ORDER BY created_at ASC
	`, knowledgeBaseID)
	if err != nil {
		return nil, fmt.Errorf("list knowledge base documents: %w", err)
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var documentID uuid.UUID
		if err := rows.Scan(&documentID); err != nil {
			return nil, fmt.Errorf("scan knowledge base document: %w", err)
		}
		ids = append(ids, documentID)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("iterate knowledge base documents: %w", rows.Err())
	}
	return ids, nil
}

func (r *Repository) CleanupDocumentResource(ctx context.Context, documentID uuid.UUID) ([]FileObjectRef, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin document cleanup transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `
		DELETE FROM document_chunks
		WHERE document_id = $1
	`, documentID); err != nil {
		return nil, fmt.Errorf("delete document chunks: %w", err)
	}

	row := tx.QueryRow(ctx, `
		SELECT f.id, f.bucket_name, f.object_key, f.deleted_at IS NOT NULL
		FROM documents d
		INNER JOIN files f ON f.id = d.file_id
		WHERE d.id = $1
	`, documentID)

	fileRef, found, err := scanCleanupFileRef(ctx, tx, row)
	if err != nil {
		return nil, err
	}
	if !found {
		if err := tx.Commit(ctx); err != nil {
			return nil, fmt.Errorf("commit empty document cleanup transaction: %w", err)
		}
		return nil, nil
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit document cleanup transaction: %w", err)
	}
	return []FileObjectRef{*fileRef}, nil
}

func (r *Repository) CleanupKnowledgeBaseResources(ctx context.Context, knowledgeBaseID uuid.UUID) ([]FileObjectRef, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin knowledge base cleanup transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `
		UPDATE documents
		SET deleted_at = COALESCE(deleted_at, CURRENT_TIMESTAMP),
		    status = 'deleting'
		WHERE knowledge_base_id = $1
	`, knowledgeBaseID); err != nil {
		return nil, fmt.Errorf("mark knowledge base documents deleting: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		DELETE FROM document_chunks dc
		USING documents d
		WHERE d.id = dc.document_id
		  AND d.knowledge_base_id = $1
	`, knowledgeBaseID); err != nil {
		return nil, fmt.Errorf("delete knowledge base chunks: %w", err)
	}

	rows, err := tx.Query(ctx, `
		SELECT DISTINCT f.id, f.bucket_name, f.object_key, f.deleted_at IS NOT NULL
		FROM documents d
		INNER JOIN files f ON f.id = d.file_id
		WHERE d.knowledge_base_id = $1
	`, knowledgeBaseID)
	if err != nil {
		return nil, fmt.Errorf("list knowledge base files for cleanup: %w", err)
	}
	type cleanupCandidate struct {
		fileID      uuid.UUID
		bucketName  string
		objectKey   string
		softDeleted bool
	}

	var candidates []cleanupCandidate
	for rows.Next() {
		var candidate cleanupCandidate
		if err := rows.Scan(&candidate.fileID, &candidate.bucketName, &candidate.objectKey, &candidate.softDeleted); err != nil {
			return nil, fmt.Errorf("scan knowledge base cleanup file: %w", err)
		}
		candidates = append(candidates, candidate)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("iterate knowledge base files for cleanup: %w", rows.Err())
	}
	rows.Close()

	var refs []FileObjectRef
	for _, candidate := range candidates {
		fileRef, found, err := prepareCleanupFileRef(ctx, tx, candidate.fileID, candidate.bucketName, candidate.objectKey, candidate.softDeleted)
		if err != nil {
			return nil, err
		}
		if found {
			refs = append(refs, *fileRef)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit knowledge base cleanup transaction: %w", err)
	}
	return refs, nil
}

func (r *Repository) setDocumentState(ctx context.Context, documentID uuid.UUID, status string, errorMessage *string) error {
	result, err := r.pool.Exec(ctx, `
		UPDATE documents
		SET status = $2,
		    error_message = $3
		WHERE id = $1
	`, documentID, status, errorMessage)
	if err != nil {
		return fmt.Errorf("update document state: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func scanKnowledgeBase(row pgx.Row) (*KnowledgeBase, error) {
	var (
		kb                  KnowledgeBase
		description         sql.NullString
		promptTemplate      sql.NullString
		similarityThreshold sql.NullFloat64
		lastIndexedAt       sql.NullTime
	)

	if err := row.Scan(
		&kb.ID,
		&kb.UserID,
		&kb.Name,
		&description,
		&kb.EmbeddingModel,
		&promptTemplate,
		&kb.RetrievalTopK,
		&similarityThreshold,
		&lastIndexedAt,
		&kb.CreatedAt,
		&kb.UpdatedAt,
	); err != nil {
		return nil, err
	}

	kb.Description = nullStringPtr(description)
	kb.PromptTemplate = nullStringPtr(promptTemplate)
	kb.SimilarityThreshold = nullFloatPtr(similarityThreshold)
	kb.LastIndexedAt = nullTimePtr(lastIndexedAt)
	return &kb, nil
}

func scanDocument(row pgx.Row) (*Document, error) {
	var (
		document         Document
		fileID           uuid.NullUUID
		title            sql.NullString
		originalFilename sql.NullString
		errorMessage     sql.NullString
	)

	if err := row.Scan(
		&document.ID,
		&document.KnowledgeBaseID,
		&fileID,
		&title,
		&originalFilename,
		&document.Status,
		&errorMessage,
		&document.ContentLength,
		&document.ChunkCount,
		&document.CreatedAt,
		&document.UpdatedAt,
	); err != nil {
		return nil, err
	}

	document.FileID = nullUUIDPtr(fileID)
	document.Title = nullStringPtr(title)
	document.OriginalFilename = nullStringPtr(originalFilename)
	document.ErrorMessage = nullStringPtr(errorMessage)
	return &document, nil
}

func scanCleanupFileRef(ctx context.Context, tx pgx.Tx, row pgx.Row) (*FileObjectRef, bool, error) {
	var (
		fileID      uuid.UUID
		bucketName  string
		objectKey   string
		softDeleted bool
	)

	if err := row.Scan(&fileID, &bucketName, &objectKey, &softDeleted); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("scan cleanup file: %w", err)
	}

	return prepareCleanupFileRef(ctx, tx, fileID, bucketName, objectKey, softDeleted)
}

func prepareCleanupFileRef(ctx context.Context, tx pgx.Tx, fileID uuid.UUID, bucketName, objectKey string, softDeleted bool) (*FileObjectRef, bool, error) {
	var activeRefs int
	if err := tx.QueryRow(ctx, `
		SELECT count(*)
		FROM documents
		WHERE file_id = $1
		  AND deleted_at IS NULL
	`, fileID).Scan(&activeRefs); err != nil {
		return nil, false, fmt.Errorf("count active file references: %w", err)
	}
	if activeRefs > 0 {
		return nil, false, nil
	}

	if _, err := tx.Exec(ctx, `
		UPDATE files
		SET deleted_at = COALESCE(deleted_at, CURRENT_TIMESTAMP)
		WHERE id = $1
	`, fileID); err != nil {
		return nil, false, fmt.Errorf("soft delete file: %w", err)
	}

	return &FileObjectRef{
		FileID:         fileID,
		BucketName:     bucketName,
		ObjectKey:      objectKey,
		WasSoftDeleted: softDeleted,
	}, true, nil
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

func nullFloatPtr(value sql.NullFloat64) *float64 {
	if !value.Valid {
		return nil
	}
	v := value.Float64
	return &v
}

func nullTimePtr(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}
	v := value.Time
	return &v
}
