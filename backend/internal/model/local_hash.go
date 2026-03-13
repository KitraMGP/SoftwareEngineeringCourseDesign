package model

import (
	"context"
	"hash/fnv"
	"math"
	"strings"
	"unicode"
)

type LocalHashEmbeddingProvider struct {
	dimension int
}

func NewLocalHashEmbeddingProvider(dimension int) *LocalHashEmbeddingProvider {
	if dimension <= 0 {
		dimension = EmbeddingDimension
	}
	return &LocalHashEmbeddingProvider{dimension: dimension}
}

func (p *LocalHashEmbeddingProvider) Embed(ctx context.Context, req EmbeddingRequest) (EmbeddingResult, error) {
	if err := ctx.Err(); err != nil {
		return EmbeddingResult{}, err
	}
	if strings.TrimSpace(req.Model) == "" {
		return EmbeddingResult{}, &ProviderError{
			Kind:    ProviderErrorBadRequest,
			Message: "embedding model is required",
		}
	}
	if len(req.Texts) == 0 {
		return EmbeddingResult{}, &ProviderError{
			Kind:    ProviderErrorBadRequest,
			Message: "embedding texts are required",
		}
	}

	result := EmbeddingResult{
		Vectors: make([][]float32, 0, len(req.Texts)),
	}
	for _, text := range req.Texts {
		result.Vectors = append(result.Vectors, p.embedText(req.Model, text))
	}
	return result, nil
}

func (p *LocalHashEmbeddingProvider) embedText(modelName, text string) []float32 {
	vector := make([]float64, p.dimension)
	tokens := tokenizeForEmbedding(text)
	if len(tokens) == 0 {
		tokens = []string{"__empty__"}
	}

	for _, token := range tokens {
		hashA := hashToken(modelName + "\x00" + token)
		hashB := hashToken(token + "\x00" + modelName)

		indexA := int(hashA % uint64(p.dimension))
		indexB := int(hashB % uint64(p.dimension))

		vector[indexA] += signedWeight(hashA, 1.0)
		vector[indexB] += signedWeight(hashB, 0.5)
	}

	var norm float64
	for _, value := range vector {
		norm += value * value
	}
	if norm == 0 {
		vector[0] = 1
		norm = 1
	}

	norm = math.Sqrt(norm)
	normalized := make([]float32, p.dimension)
	for idx, value := range vector {
		normalized[idx] = float32(value / norm)
	}
	return normalized
}

func tokenizeForEmbedding(text string) []string {
	normalized := strings.TrimSpace(strings.ToLower(text))
	if normalized == "" {
		return nil
	}

	tokens := make([]string, 0, len(normalized)/4)
	var current strings.Builder
	var hanRunes []rune

	flushWord := func() {
		if current.Len() == 0 {
			return
		}
		tokens = append(tokens, current.String())
		current.Reset()
	}

	for _, r := range normalized {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			if unicode.Is(unicode.Han, r) {
				flushWord()
				hanRunes = append(hanRunes, r)
				tokens = append(tokens, string(r))
				continue
			}
			current.WriteRune(r)
		default:
			flushWord()
		}
	}
	flushWord()

	for idx := 0; idx+1 < len(hanRunes); idx++ {
		tokens = append(tokens, string([]rune{hanRunes[idx], hanRunes[idx+1]}))
	}

	if len(tokens) > 0 {
		return tokens
	}

	runes := []rune(normalized)
	for idx := 0; idx+1 < len(runes); idx++ {
		pair := strings.TrimSpace(string([]rune{runes[idx], runes[idx+1]}))
		if pair != "" {
			tokens = append(tokens, pair)
		}
	}
	return tokens
}

func hashToken(value string) uint64 {
	hasher := fnv.New64a()
	_, _ = hasher.Write([]byte(value))
	return hasher.Sum64()
}

func signedWeight(hash uint64, weight float64) float64 {
	if hash&1 == 0 {
		return weight
	}
	return -weight
}
