package chat

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestDeepSeekProviderStreamChat(t *testing.T) {
	t.Parallel()

	provider := &DeepSeekProvider{
		client: &http.Client{
			Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
				if r.Method != http.MethodPost {
					t.Fatalf("unexpected method: got %s want POST", r.Method)
				}
				if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
					t.Fatalf("unexpected authorization header: got %q", got)
				}
				if got := r.Header.Get("Accept"); got != "text/event-stream" {
					t.Fatalf("unexpected accept header: got %q", got)
				}

				var req deepSeekChatRequest
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					t.Fatalf("decode request: %v", err)
				}
				if !req.Stream {
					t.Fatal("expected stream=true")
				}
				if req.StreamOptions == nil || !req.StreamOptions.IncludeUsage {
					t.Fatal("expected include_usage=true")
				}
				if req.Model != "deepseek-chat" {
					t.Fatalf("unexpected model: got %q", req.Model)
				}

				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     make(http.Header),
					Body: io.NopCloser(strings.NewReader(
						"data: {\"model\":\"deepseek-chat\",\"choices\":[{\"delta\":{\"content\":\"Hello\"},\"finish_reason\":null}]}\n\n" +
							"data: {\"model\":\"deepseek-chat\",\"choices\":[{\"delta\":{\"content\":\" world\"},\"finish_reason\":\"stop\"}]}\n\n" +
							"data: {\"model\":\"deepseek-chat\",\"choices\":[],\"usage\":{\"prompt_tokens\":8,\"completion_tokens\":2,\"total_tokens\":10}}\n\n" +
							"data: [DONE]\n\n",
					)),
				}, nil
			}),
		},
		baseURL: "https://deepseek.invalid",
		apiKey:  "test-key",
	}
	var chunks []string

	result, err := provider.StreamChat(context.Background(), ProviderRequest{
		Model: "deepseek-chat",
		Messages: []ProviderMessage{
			{Role: "system", Content: "You are a helpful assistant."},
			{Role: "user", Content: "Say hello."},
		},
		Temperature: 0.7,
	}, func(delta string) error {
		chunks = append(chunks, delta)
		return nil
	})
	if err != nil {
		t.Fatalf("StreamChat() error = %v", err)
	}

	if got := strings.Join(chunks, ""); got != "Hello world" {
		t.Fatalf("unexpected streamed content: got %q want %q", got, "Hello world")
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Model != "deepseek-chat" {
		t.Fatalf("unexpected result model: got %q", result.Model)
	}
	if result.FinishReason != "stop" {
		t.Fatalf("unexpected finish reason: got %q", result.FinishReason)
	}
	if result.Usage.TotalTokens != 10 || result.Usage.PromptTokens != 8 || result.Usage.CompletionTokens != 2 {
		t.Fatalf("unexpected usage: %+v", result.Usage)
	}
}

func TestDeepSeekProviderMapsHTTPError(t *testing.T) {
	t.Parallel()

	provider := &DeepSeekProvider{
		client: &http.Client{
			Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusUnauthorized,
					Header:     make(http.Header),
					Body:       io.NopCloser(strings.NewReader(`{"error":{"message":"invalid api key"}}`)),
				}, nil
			}),
		},
		baseURL: "https://deepseek.invalid",
		apiKey:  "bad-key",
	}
	_, err := provider.StreamChat(context.Background(), ProviderRequest{
		Model: "deepseek-chat",
		Messages: []ProviderMessage{
			{Role: "user", Content: "Hello"},
		},
	}, func(string) error {
		return nil
	})
	if err == nil {
		t.Fatal("expected error")
	}

	providerErr, ok := AsProviderError(err)
	if !ok {
		t.Fatalf("expected provider error, got %T", err)
	}
	if providerErr.Kind != ProviderErrorAuthFailed {
		t.Fatalf("unexpected provider error kind: got %s want %s", providerErr.Kind, ProviderErrorAuthFailed)
	}
	if providerErr.Message != "invalid api key" {
		t.Fatalf("unexpected provider error message: got %q", providerErr.Message)
	}
}

func TestReadSSEEventData(t *testing.T) {
	t.Parallel()

	reader := strings.NewReader("event: delta\ndata: {\"content\":\"hello\"}\ndata: {\"content\":\" world\"}\n\n")
	data, err := readSSEEventData(bufio.NewReader(reader))
	if err != nil {
		t.Fatalf("readSSEEventData() error = %v", err)
	}
	if data != "{\"content\":\"hello\"}\n{\"content\":\" world\"}" {
		t.Fatalf("unexpected event data: got %q", data)
	}
}

func TestDisabledProvider(t *testing.T) {
	t.Parallel()

	provider := &DisabledProvider{reason: "missing api key"}
	_, err := provider.StreamChat(context.Background(), ProviderRequest{}, func(string) error { return nil })
	if err == nil {
		t.Fatal("expected error")
	}

	var providerErr *ProviderError
	if !errors.As(err, &providerErr) {
		t.Fatalf("expected ProviderError, got %T", err)
	}
	if providerErr.Kind != ProviderErrorMisconfigured {
		t.Fatalf("unexpected provider error kind: got %s want %s", providerErr.Kind, ProviderErrorMisconfigured)
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (fn roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return fn(r)
}
