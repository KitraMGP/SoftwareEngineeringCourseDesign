package rag

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"backend/internal/model"

	"github.com/google/uuid"
)

var ErrKnowledgeBaseNotFound = errors.New("knowledge base not found")

type Citation struct {
	DocumentChunkID uuid.UUID
	DocumentID      uuid.UUID
	KnowledgeBaseID uuid.UUID
	DocumentName    string
	Rank            int
	SourcePage      *int
}

type RetrievedChunk struct {
	Citation   Citation
	Content    string
	Similarity float64
}

type RetrievalResult struct {
	Grounded       bool
	PromptTemplate *string
	Chunks         []RetrievedChunk
}

type ServiceConfig struct {
	MaxContextChunks int
}

type repository interface {
	GetKnowledgeBaseSettings(ctx context.Context, knowledgeBaseID uuid.UUID) (*KnowledgeBaseSettings, error)
	SearchChunks(ctx context.Context, knowledgeBaseID uuid.UUID, queryVector string, topK int, similarityThreshold *float64) ([]RetrievedChunk, error)
}

type Service struct {
	repo     repository
	embedder model.EmbeddingProvider
	cfg      ServiceConfig
}

func NewService(repo repository, embedder model.EmbeddingProvider, cfg ServiceConfig) *Service {
	if cfg.MaxContextChunks <= 0 {
		cfg.MaxContextChunks = 5
	}
	if embedder == nil {
		embedder = model.NewDisabledEmbeddingProvider("embedding provider is not configured")
	}
	return &Service{
		repo:     repo,
		embedder: embedder,
		cfg:      cfg,
	}
}

func (s *Service) Retrieve(ctx context.Context, knowledgeBaseID uuid.UUID, question string) (RetrievalResult, error) {
	question = strings.TrimSpace(question)
	if question == "" {
		return RetrievalResult{}, nil
	}

	settings, err := s.repo.GetKnowledgeBaseSettings(ctx, knowledgeBaseID)
	if err != nil {
		if errors.Is(err, ErrKnowledgeBaseNotFound) {
			return RetrievalResult{}, err
		}
		return RetrievalResult{}, fmt.Errorf("load knowledge base settings: %w", err)
	}

	embeddingResult, err := s.embedder.Embed(ctx, model.EmbeddingRequest{
		Model: settings.EmbeddingModel,
		Texts: []string{question},
	})
	if err != nil {
		return RetrievalResult{}, err
	}
	if len(embeddingResult.Vectors) != 1 || len(embeddingResult.Vectors[0]) != model.EmbeddingDimension {
		return RetrievalResult{}, &model.ProviderError{
			Kind:    model.ProviderErrorUnavailable,
			Message: "embedding provider returned an invalid query vector",
		}
	}

	topK := settings.RetrievalTopK
	if topK <= 0 {
		topK = s.cfg.MaxContextChunks
	}
	if topK > s.cfg.MaxContextChunks {
		topK = s.cfg.MaxContextChunks
	}

	chunks, err := s.repo.SearchChunks(ctx, knowledgeBaseID, model.FormatVector(embeddingResult.Vectors[0]), topK, settings.SimilarityThreshold)
	if err != nil {
		return RetrievalResult{}, fmt.Errorf("search knowledge base chunks: %w", err)
	}

	result := RetrievalResult{
		Grounded:       len(chunks) > 0,
		PromptTemplate: settings.PromptTemplate,
		Chunks:         chunks,
	}
	return result, nil
}
