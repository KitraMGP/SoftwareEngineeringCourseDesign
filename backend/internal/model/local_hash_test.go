package model

import (
	"context"
	"testing"
)

func TestLocalHashEmbeddingProviderDeterministic(t *testing.T) {
	t.Parallel()

	provider := NewLocalHashEmbeddingProvider(EmbeddingDimension)

	first, err := provider.Embed(context.Background(), EmbeddingRequest{
		Model: "local-hash-1536",
		Texts: []string{"Go backend knowledge base"},
	})
	if err != nil {
		t.Fatalf("Embed() error = %v", err)
	}

	second, err := provider.Embed(context.Background(), EmbeddingRequest{
		Model: "local-hash-1536",
		Texts: []string{"Go backend knowledge base"},
	})
	if err != nil {
		t.Fatalf("Embed() second call error = %v", err)
	}

	if len(first.Vectors) != 1 || len(first.Vectors[0]) != EmbeddingDimension {
		t.Fatalf("unexpected first embedding shape: %+v", first)
	}
	if len(second.Vectors) != 1 || len(second.Vectors[0]) != EmbeddingDimension {
		t.Fatalf("unexpected second embedding shape: %+v", second)
	}

	for idx := range first.Vectors[0] {
		if first.Vectors[0][idx] != second.Vectors[0][idx] {
			t.Fatalf("vector mismatch at index %d", idx)
		}
	}
}

func TestFormatVector(t *testing.T) {
	t.Parallel()

	got := FormatVector([]float32{1.25, -0.5, 0})
	want := "[1.25,-0.5,0]"
	if got != want {
		t.Fatalf("FormatVector() = %q, want %q", got, want)
	}
}
