package model

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type OpenAICompatibleEmbeddingProvider struct {
	client  *http.Client
	baseURL string
	apiKey  string
}

type openAIEmbeddingRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type openAIEmbeddingResponse struct {
	Data  []openAIEmbeddingItem `json:"data"`
	Error *openAIErrorEnvelope  `json:"error,omitempty"`
}

type openAIEmbeddingItem struct {
	Index     int       `json:"index"`
	Embedding []float32 `json:"embedding"`
}

type openAIErrorEnvelope struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    any    `json:"code"`
}

func NewOpenAICompatibleEmbeddingProvider(baseURL, apiKey string, timeout time.Duration) *OpenAICompatibleEmbeddingProvider {
	if timeout <= 0 {
		timeout = 20 * time.Second
	}
	return &OpenAICompatibleEmbeddingProvider{
		client: &http.Client{
			Timeout: timeout,
		},
		baseURL: strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		apiKey:  strings.TrimSpace(apiKey),
	}
}

func (p *OpenAICompatibleEmbeddingProvider) Embed(ctx context.Context, req EmbeddingRequest) (EmbeddingResult, error) {
	if strings.TrimSpace(p.apiKey) == "" {
		return EmbeddingResult{}, &ProviderError{
			Kind:    ProviderErrorMisconfigured,
			Message: "embedding provider API key is not configured",
		}
	}
	if strings.TrimSpace(p.baseURL) == "" {
		return EmbeddingResult{}, &ProviderError{
			Kind:    ProviderErrorMisconfigured,
			Message: "embedding provider base URL is not configured",
		}
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

	payload, err := json.Marshal(openAIEmbeddingRequest{
		Model: req.Model,
		Input: req.Texts,
	})
	if err != nil {
		return EmbeddingResult{}, fmt.Errorf("marshal embedding request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/embeddings", bytes.NewReader(payload))
	if err != nil {
		return EmbeddingResult{}, fmt.Errorf("create embedding request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		if ctx.Err() != nil {
			return EmbeddingResult{}, ctx.Err()
		}
		return EmbeddingResult{}, &ProviderError{
			Kind:    ProviderErrorUnavailable,
			Message: "failed to call embedding provider",
			Err:     err,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return EmbeddingResult{}, mapOpenAICompatibleHTTPError(resp)
	}

	var parsed openAIEmbeddingResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, 8<<20)).Decode(&parsed); err != nil {
		return EmbeddingResult{}, &ProviderError{
			Kind:    ProviderErrorUnavailable,
			Message: "failed to decode embedding response",
			Err:     err,
		}
	}
	if parsed.Error != nil && strings.TrimSpace(parsed.Error.Message) != "" {
		return EmbeddingResult{}, &ProviderError{
			Kind:    ProviderErrorUnavailable,
			Message: strings.TrimSpace(parsed.Error.Message),
		}
	}
	if len(parsed.Data) != len(req.Texts) {
		return EmbeddingResult{}, &ProviderError{
			Kind:    ProviderErrorUnavailable,
			Message: "embedding response count mismatch",
		}
	}

	result := EmbeddingResult{
		Vectors: make([][]float32, len(req.Texts)),
	}
	for _, item := range parsed.Data {
		if item.Index < 0 || item.Index >= len(req.Texts) {
			return EmbeddingResult{}, &ProviderError{
				Kind:    ProviderErrorUnavailable,
				Message: "embedding response index out of range",
			}
		}
		if len(item.Embedding) != EmbeddingDimension {
			return EmbeddingResult{}, &ProviderError{
				Kind:    ProviderErrorBadRequest,
				Message: fmt.Sprintf("embedding dimension mismatch: got %d want %d", len(item.Embedding), EmbeddingDimension),
			}
		}
		result.Vectors[item.Index] = item.Embedding
	}

	for idx := range result.Vectors {
		if len(result.Vectors[idx]) == 0 {
			return EmbeddingResult{}, &ProviderError{
				Kind:    ProviderErrorUnavailable,
				Message: "embedding response is incomplete",
			}
		}
	}

	return result, nil
}

func mapOpenAICompatibleHTTPError(resp *http.Response) error {
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	message := strings.TrimSpace(string(body))

	var parsed struct {
		Error openAIErrorEnvelope `json:"error"`
	}
	if err := json.Unmarshal(body, &parsed); err == nil && strings.TrimSpace(parsed.Error.Message) != "" {
		message = strings.TrimSpace(parsed.Error.Message)
	}
	if message == "" {
		message = resp.Status
	}

	kind := ProviderErrorUnavailable
	switch resp.StatusCode {
	case http.StatusUnauthorized, http.StatusForbidden:
		kind = ProviderErrorAuthFailed
	case http.StatusTooManyRequests:
		kind = ProviderErrorRateLimited
	case http.StatusBadRequest:
		kind = ProviderErrorBadRequest
	}

	return &ProviderError{
		Kind:    kind,
		Message: message,
	}
}
