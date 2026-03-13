package rag

import "context"

type Citation struct {
	DocumentID string
	ChunkID    string
	Rank       int
}

type RetrievalResult struct {
	Grounded  bool
	Context   []string
	Citations []Citation
}

type Service interface {
	Retrieve(ctx context.Context, knowledgeBaseID, question string) (RetrievalResult, error)
}
