package chat

import (
	"context"
	"errors"
	"strings"
	"time"

	"backend/internal/platform/httpx"

	"github.com/google/uuid"
)

type Service struct {
	repo     chatRepository
	provider Provider
	cfg      ServiceConfig
}

type chatRepository interface {
	Create(ctx context.Context, userID uuid.UUID, input CreateSessionInput) (*Session, error)
	List(ctx context.Context, userID uuid.UUID, page, size int, keyword string) (*ListSessionsResult, error)
	GetDetail(ctx context.Context, userID, sessionID uuid.UUID) (*SessionDetail, error)
	Delete(ctx context.Context, userID, sessionID uuid.UUID) error
	CreateMessage(ctx context.Context, userID, sessionID uuid.UUID, input CreateMessageInput) (*Message, error)
}

type ServiceConfig struct {
	DefaultModel       string
	SystemPrompt       string
	RequestTimeout     time.Duration
	MaxHistoryMessages int
	Temperature        float64
}

func NewService(repo chatRepository, provider Provider, cfg ServiceConfig) *Service {
	if cfg.RequestTimeout <= 0 {
		cfg.RequestTimeout = 60 * time.Second
	}
	if cfg.MaxHistoryMessages <= 0 {
		cfg.MaxHistoryMessages = 12
	}
	if provider == nil {
		provider = &DisabledProvider{reason: "chat provider is not configured"}
	}
	return &Service{
		repo:     repo,
		provider: provider,
		cfg:      cfg,
	}
}

func (s *Service) CreateSession(ctx context.Context, userID uuid.UUID, input CreateSessionInput) (*Session, error) {
	model := strings.TrimSpace(input.Model)
	if model == "" {
		model = strings.TrimSpace(s.cfg.DefaultModel)
	}
	if model == "" {
		return nil, httpx.ValidationFailed(httpx.FieldError{Field: "model", Message: "model is required"})
	}

	session, err := s.repo.Create(ctx, userID, CreateSessionInput{
		Name:            input.Name,
		Model:           model,
		KnowledgeBaseID: input.KnowledgeBaseID,
	})
	if err != nil {
		if err == ErrNotFound && input.KnowledgeBaseID != nil {
			return nil, httpx.NotFound("knowledge base not found")
		}
		return nil, httpx.Internal("failed to create session").WithErr(err)
	}
	return session, nil
}

func (s *Service) ListSessions(ctx context.Context, userID uuid.UUID, page, size int, keyword string) (*ListSessionsResult, error) {
	return s.repo.List(ctx, userID, page, size, keyword)
}

func (s *Service) GetSessionDetail(ctx context.Context, userID, sessionID uuid.UUID) (*SessionDetail, error) {
	detail, err := s.repo.GetDetail(ctx, userID, sessionID)
	if err != nil {
		if err == ErrNotFound {
			return nil, httpx.NotFound("session not found")
		}
		return nil, httpx.Internal("failed to load session detail").WithErr(err)
	}
	return detail, nil
}

func (s *Service) DeleteSession(ctx context.Context, userID, sessionID uuid.UUID) error {
	if err := s.repo.Delete(ctx, userID, sessionID); err != nil {
		if err == ErrNotFound {
			return httpx.NotFound("session not found")
		}
		return httpx.Internal("failed to delete session").WithErr(err)
	}
	return nil
}

func (s *Service) SendMessageStream(ctx context.Context, userID, sessionID uuid.UUID, content string, stream StreamWriter) error {
	trimmedContent := strings.TrimSpace(content)
	if trimmedContent == "" {
		return httpx.ValidationFailed(httpx.FieldError{Field: "content", Message: "content is required"})
	}

	detail, err := s.repo.GetDetail(ctx, userID, sessionID)
	if err != nil {
		if err == ErrNotFound {
			return httpx.NotFound("session not found")
		}
		return httpx.Internal("failed to load session detail").WithErr(err)
	}
	if detail.Session.KnowledgeBaseID != nil {
		return httpx.FeatureNotReady("knowledge-base grounded chat is not available in the current phase")
	}

	userMessage, err := s.repo.CreateMessage(ctx, userID, sessionID, CreateMessageInput{
		Role:    "user",
		Content: trimmedContent,
		Status:  "completed",
	})
	if err != nil {
		if err == ErrNotFound {
			return httpx.NotFound("session not found")
		}
		return httpx.Internal("failed to create user message").WithErr(err)
	}

	if err := stream.Start(); err != nil {
		return httpx.Internal("failed to start streaming response").WithErr(err)
	}
	defer stream.Close()

	model := strings.TrimSpace(detail.Session.Model)
	if model == "" {
		model = strings.TrimSpace(s.cfg.DefaultModel)
	}

	assistantMessageID := uuid.New()
	if err := stream.SendMeta(StreamMeta{
		MessageID: assistantMessageID,
		Grounded:  false,
		Model:     model,
	}); err != nil {
		return nil
	}

	requestCtx, cancel := context.WithTimeout(ctx, s.cfg.RequestTimeout)
	defer cancel()

	var assistantContent strings.Builder
	result, err := s.provider.StreamChat(requestCtx, ProviderRequest{
		Model:       model,
		Messages:    s.buildProviderMessages(detail.Messages, *userMessage),
		Temperature: s.cfg.Temperature,
	}, func(delta string) error {
		if delta == "" {
			return nil
		}
		assistantContent.WriteString(delta)
		return stream.SendDelta(delta)
	})
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) || ctx.Err() != nil {
			return nil
		}
		_ = stream.SendError(mapProviderErrorCode(err), mapProviderErrorMessage(err))
		return nil
	}

	fullAssistantContent := assistantContent.String()
	if strings.TrimSpace(fullAssistantContent) == "" {
		_ = stream.SendError(httpx.CodeInternal, "empty response from chat provider")
		return nil
	}

	modelUsed := model
	if result != nil && strings.TrimSpace(result.Model) != "" {
		modelUsed = strings.TrimSpace(result.Model)
	}

	createAssistantInput := CreateMessageInput{
		ID:               assistantMessageID,
		Role:             "assistant",
		ReplyToMessageID: &userMessage.ID,
		Content:          fullAssistantContent,
		Status:           "completed",
		ModelUsed:        &modelUsed,
		Grounded:         false,
	}
	if result != nil {
		createAssistantInput.PromptTokens = result.Usage.PromptTokens
		createAssistantInput.CompletionTokens = result.Usage.CompletionTokens
		createAssistantInput.TotalTokens = result.Usage.TotalTokens
	}

	if _, err := s.repo.CreateMessage(ctx, userID, sessionID, createAssistantInput); err != nil {
		_ = stream.SendError(httpx.CodeInternal, "failed to persist assistant message")
		return nil
	}

	finishReason := "stop"
	if result != nil && strings.TrimSpace(result.FinishReason) != "" {
		finishReason = strings.TrimSpace(result.FinishReason)
	}
	_ = stream.SendDone(finishReason)
	return nil
}

func (s *Service) buildProviderMessages(history []Message, userMessage Message) []ProviderMessage {
	items := make([]ProviderMessage, 0, len(history)+2)
	if systemPrompt := strings.TrimSpace(s.cfg.SystemPrompt); systemPrompt != "" {
		items = append(items, ProviderMessage{
			Role:    "system",
			Content: systemPrompt,
		})
	}

	historyMessages := make([]ProviderMessage, 0, len(history)+1)
	for _, message := range history {
		if message.Status != "completed" {
			continue
		}
		if message.Role != "user" && message.Role != "assistant" {
			continue
		}
		if strings.TrimSpace(message.Content) == "" {
			continue
		}
		historyMessages = append(historyMessages, ProviderMessage{
			Role:    message.Role,
			Content: message.Content,
		})
	}
	historyMessages = append(historyMessages, ProviderMessage{
		Role:    userMessage.Role,
		Content: userMessage.Content,
	})

	if len(historyMessages) > s.cfg.MaxHistoryMessages {
		historyMessages = historyMessages[len(historyMessages)-s.cfg.MaxHistoryMessages:]
	}

	items = append(items, historyMessages...)
	return items
}

func mapProviderErrorCode(err error) int {
	providerErr, ok := AsProviderError(err)
	if !ok {
		return httpx.CodeInternal
	}

	switch providerErr.Kind {
	case ProviderErrorAuthFailed:
		return httpx.CodeProviderAuthFailed
	case ProviderErrorRateLimited:
		return httpx.CodeProviderRateLimited
	case ProviderErrorPromptTooLong:
		return httpx.CodePromptTooLarge
	case ProviderErrorUnavailable, ProviderErrorMisconfigured:
		return httpx.CodeProviderUnavailable
	default:
		return httpx.CodeInternal
	}
}

func mapProviderErrorMessage(err error) string {
	providerErr, ok := AsProviderError(err)
	if !ok {
		return "chat provider request failed"
	}

	switch providerErr.Kind {
	case ProviderErrorAuthFailed:
		return "DeepSeek API authentication failed"
	case ProviderErrorRateLimited:
		return "DeepSeek API rate limit exceeded"
	case ProviderErrorPromptTooLong:
		return "chat prompt is too large"
	case ProviderErrorMisconfigured:
		return providerErr.Message
	case ProviderErrorBadRequest:
		if strings.TrimSpace(providerErr.Message) != "" {
			return providerErr.Message
		}
		return "chat request is invalid"
	case ProviderErrorUnavailable:
		if strings.TrimSpace(providerErr.Message) != "" {
			return providerErr.Message
		}
		return "DeepSeek API is currently unavailable"
	default:
		return "chat provider request failed"
	}
}
