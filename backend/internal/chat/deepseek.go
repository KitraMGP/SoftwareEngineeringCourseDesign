package chat

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type DeepSeekProvider struct {
	client  *http.Client
	baseURL string
	apiKey  string
}

type deepSeekChatRequest struct {
	Model         string                `json:"model"`
	Messages      []ProviderMessage     `json:"messages"`
	Temperature   float64               `json:"temperature,omitempty"`
	Stream        bool                  `json:"stream"`
	StreamOptions *deepSeekStreamOption `json:"stream_options,omitempty"`
}

type deepSeekStreamOption struct {
	IncludeUsage bool `json:"include_usage"`
}

type deepSeekStreamChunk struct {
	Model   string                    `json:"model"`
	Choices []deepSeekStreamChoice    `json:"choices"`
	Usage   *deepSeekUsage            `json:"usage"`
	Error   *deepSeekAPIErrorEnvelope `json:"error,omitempty"`
}

type deepSeekStreamChoice struct {
	Delta        deepSeekDelta `json:"delta"`
	FinishReason *string       `json:"finish_reason"`
}

type deepSeekDelta struct {
	Content string `json:"content"`
}

type deepSeekUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type deepSeekAPIErrorEnvelope struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    any    `json:"code"`
}

type deepSeekErrorResponse struct {
	Error deepSeekAPIErrorEnvelope `json:"error"`
}

func NewDeepSeekProvider(baseURL, apiKey string) *DeepSeekProvider {
	return &DeepSeekProvider{
		client:  &http.Client{},
		baseURL: strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		apiKey:  strings.TrimSpace(apiKey),
	}
}

func (p *DeepSeekProvider) StreamChat(ctx context.Context, req ProviderRequest, onDelta func(string) error) (*CompletionResult, error) {
	if strings.TrimSpace(p.apiKey) == "" {
		return nil, &ProviderError{
			Kind:    ProviderErrorMisconfigured,
			Message: "DeepSeek API key is not configured",
		}
	}
	if strings.TrimSpace(req.Model) == "" {
		return nil, &ProviderError{
			Kind:    ProviderErrorBadRequest,
			Message: "chat model is required",
		}
	}
	if len(req.Messages) == 0 {
		return nil, &ProviderError{
			Kind:    ProviderErrorBadRequest,
			Message: "chat messages are required",
		}
	}

	payload, err := json.Marshal(deepSeekChatRequest{
		Model:       req.Model,
		Messages:    req.Messages,
		Temperature: req.Temperature,
		Stream:      true,
		StreamOptions: &deepSeekStreamOption{
			IncludeUsage: true,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("marshal deepseek chat request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/chat/completions", bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("create deepseek request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, &ProviderError{
			Kind:    ProviderErrorUnavailable,
			Message: "failed to call DeepSeek API",
			Err:     err,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, p.mapHTTPError(resp)
	}

	result, err := parseDeepSeekStream(resp.Body, req.Model, onDelta)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, err
	}
	return result, nil
}

func (p *DeepSeekProvider) mapHTTPError(resp *http.Response) error {
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	message := strings.TrimSpace(string(body))

	var parsed deepSeekErrorResponse
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
		lowerMessage := strings.ToLower(message)
		if strings.Contains(lowerMessage, "context length") || strings.Contains(lowerMessage, "maximum context length") || strings.Contains(lowerMessage, "too long") {
			kind = ProviderErrorPromptTooLong
		} else {
			kind = ProviderErrorBadRequest
		}
	}

	return &ProviderError{
		Kind:    kind,
		Message: message,
	}
}

func parseDeepSeekStream(body io.Reader, fallbackModel string, onDelta func(string) error) (*CompletionResult, error) {
	reader := bufio.NewReader(body)
	result := &CompletionResult{
		Model:        fallbackModel,
		FinishReason: "stop",
	}

	for {
		eventData, err := readSSEEventData(reader)
		if err != nil {
			if err == io.EOF {
				return result, nil
			}
			return nil, &ProviderError{
				Kind:    ProviderErrorUnavailable,
				Message: "failed to read DeepSeek stream",
				Err:     err,
			}
		}

		if eventData == "" {
			continue
		}
		if eventData == "[DONE]" {
			return result, nil
		}

		var chunk deepSeekStreamChunk
		if err := json.Unmarshal([]byte(eventData), &chunk); err != nil {
			return nil, &ProviderError{
				Kind:    ProviderErrorUnavailable,
				Message: "failed to decode DeepSeek stream response",
				Err:     err,
			}
		}

		if chunk.Error != nil && strings.TrimSpace(chunk.Error.Message) != "" {
			return nil, &ProviderError{
				Kind:    ProviderErrorUnavailable,
				Message: strings.TrimSpace(chunk.Error.Message),
			}
		}
		if strings.TrimSpace(chunk.Model) != "" {
			result.Model = strings.TrimSpace(chunk.Model)
		}
		if chunk.Usage != nil {
			result.Usage = Usage{
				PromptTokens:     chunk.Usage.PromptTokens,
				CompletionTokens: chunk.Usage.CompletionTokens,
				TotalTokens:      chunk.Usage.TotalTokens,
			}
		}

		for _, choice := range chunk.Choices {
			if choice.FinishReason != nil && strings.TrimSpace(*choice.FinishReason) != "" {
				result.FinishReason = strings.TrimSpace(*choice.FinishReason)
			}
			if choice.Delta.Content == "" {
				continue
			}
			if err := onDelta(choice.Delta.Content); err != nil {
				return nil, err
			}
		}
	}
}

func readSSEEventData(reader *bufio.Reader) (string, error) {
	var dataLines []string

	for {
		line, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			return "", err
		}

		line = strings.TrimRight(line, "\r\n")

		if line == "" {
			if len(dataLines) > 0 {
				return strings.Join(dataLines, "\n"), nil
			}
			if err == io.EOF {
				return "", io.EOF
			}
			continue
		}

		if strings.HasPrefix(line, "data:") {
			dataLines = append(dataLines, strings.TrimSpace(strings.TrimPrefix(line, "data:")))
		}

		if err == io.EOF {
			if len(dataLines) == 0 {
				return "", io.EOF
			}
			return strings.Join(dataLines, "\n"), nil
		}
	}
}
