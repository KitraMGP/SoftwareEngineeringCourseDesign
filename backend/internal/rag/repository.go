package rag

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type KnowledgeBaseSettings struct {
	ID                  uuid.UUID
	Name                string
	EmbeddingModel      string
	PromptTemplate      *string
	RetrievalTopK       int
	SimilarityThreshold *float64
}

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) GetKnowledgeBaseSettings(ctx context.Context, knowledgeBaseID uuid.UUID) (*KnowledgeBaseSettings, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, name, embedding_model, prompt_template, retrieval_top_k, similarity_threshold
		FROM knowledge_bases
		WHERE id = $1
		  AND deleted_at IS NULL
	`, knowledgeBaseID)

	var (
		settings            KnowledgeBaseSettings
		promptTemplate      sql.NullString
		similarityThreshold sql.NullFloat64
	)
	if err := row.Scan(
		&settings.ID,
		&settings.Name,
		&settings.EmbeddingModel,
		&promptTemplate,
		&settings.RetrievalTopK,
		&similarityThreshold,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrKnowledgeBaseNotFound
		}
		return nil, fmt.Errorf("get knowledge base settings: %w", err)
	}

	settings.PromptTemplate = nullStringPtr(promptTemplate)
	settings.SimilarityThreshold = nullFloatPtr(similarityThreshold)
	return &settings, nil
}

func (r *Repository) SearchChunks(ctx context.Context, knowledgeBaseID uuid.UUID, queryVector string, topK int, similarityThreshold *float64) ([]RetrievedChunk, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT
			dc.id,
			dc.document_id,
			dc.knowledge_base_id,
			COALESCE(NULLIF(d.title, ''), f.original_filename, 'Untitled Document') AS document_name,
			dc.source_page,
			dc.content,
			1 - (dc.embedding <=> $2::vector) AS similarity
		FROM document_chunks dc
		INNER JOIN documents d ON d.id = dc.document_id
		INNER JOIN knowledge_bases kb ON kb.id = dc.knowledge_base_id
		LEFT JOIN files f ON f.id = d.file_id
		WHERE dc.knowledge_base_id = $1
		  AND d.deleted_at IS NULL
		  AND d.status = 'available'
		  AND kb.deleted_at IS NULL
		  AND ($3::double precision IS NULL OR (1 - (dc.embedding <=> $2::vector)) >= $3)
		ORDER BY dc.embedding <=> $2::vector ASC, dc.created_at ASC
		LIMIT $4
	`, knowledgeBaseID, queryVector, similarityThreshold, topK)
	if err != nil {
		return nil, fmt.Errorf("query chunks: %w", err)
	}
	defer rows.Close()

	chunks := make([]RetrievedChunk, 0, topK)
	for rows.Next() {
		var (
			chunk      RetrievedChunk
			sourcePage sql.NullInt32
		)
		if err := rows.Scan(
			&chunk.Citation.DocumentChunkID,
			&chunk.Citation.DocumentID,
			&chunk.Citation.KnowledgeBaseID,
			&chunk.Citation.DocumentName,
			&sourcePage,
			&chunk.Content,
			&chunk.Similarity,
		); err != nil {
			return nil, fmt.Errorf("scan chunk: %w", err)
		}
		chunk.Citation.Rank = len(chunks) + 1
		chunk.Citation.SourcePage = nullInt32Ptr(sourcePage)
		chunks = append(chunks, chunk)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("iterate chunks: %w", rows.Err())
	}

	return chunks, nil
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

func nullInt32Ptr(value sql.NullInt32) *int {
	if !value.Valid {
		return nil
	}
	v := int(value.Int32)
	return &v
}
