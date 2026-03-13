package model

import "context"

type ChatMessage struct {
	Role    string
	Content string
}

type ChatRequest struct {
	Model    string
	Messages []ChatMessage
}

type ChatDelta struct {
	Content string
}

type StreamResult interface {
	Recv() (ChatDelta, error)
	Close() error
}

type EmbeddingRequest struct {
	Model string
	Texts []string
}

type EmbeddingResult struct {
	Vectors [][]float32
}

type ChatProvider interface {
	StreamChat(ctx context.Context, req ChatRequest) (StreamResult, error)
}

type EmbeddingProvider interface {
	Embed(ctx context.Context, req EmbeddingRequest) (EmbeddingResult, error)
}
