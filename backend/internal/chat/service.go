package chat

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"backend/internal/model"
	"backend/internal/platform/httpx"
	"backend/internal/rag"

	"github.com/google/uuid"
)

type Service struct {
	repo     chatRepository
	rag      ragRetriever
	provider Provider
	streams  *streamManager
	cfg      ServiceConfig
}

type chatRepository interface {
	Create(ctx context.Context, userID uuid.UUID, input CreateSessionInput) (*Session, error)
	List(ctx context.Context, userID uuid.UUID, page, size int, keyword string) (*ListSessionsResult, error)
	GetDetail(ctx context.Context, userID, sessionID uuid.UUID) (*SessionDetail, error)
	Delete(ctx context.Context, userID, sessionID uuid.UUID) error
	CreateMessage(ctx context.Context, userID, sessionID uuid.UUID, input CreateMessageInput) (*Message, error)
	UpdateMessage(ctx context.Context, userID, sessionID, messageID uuid.UUID, input UpdateMessageInput) (*Message, error)
}

type ragRetriever interface {
	Retrieve(ctx context.Context, knowledgeBaseID uuid.UUID, question string) (rag.RetrievalResult, error)
}

type ServiceConfig struct {
	DefaultModel       string
	SystemPrompt       string
	RequestTimeout     time.Duration
	MaxHistoryMessages int
	Temperature        float64
}

func NewService(repo chatRepository, provider Provider, cfg ServiceConfig) *Service {
	return NewServiceWithRAG(repo, nil, provider, cfg)
}

func NewServiceWithRAG(repo chatRepository, ragService ragRetriever, provider Provider, cfg ServiceConfig) *Service {
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
		rag:      ragService,
		provider: provider,
		streams:  newStreamManager(),
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

	handle, err := s.streams.Acquire(ctx, sessionID)
	if err != nil {
		if errors.Is(err, ErrGenerationInProgress) {
			return httpx.Conflict("a message is already being generated for this session")
		}
		return httpx.Internal("failed to initialize stream generation").WithErr(err)
	}
	defer handle.Release()

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

	retrievalResult, shouldContinue := s.retrieveKnowledgeContext(handle.Context(), detail.Session, trimmedContent, stream)
	if !shouldContinue {
		return nil
	}

	grounded := retrievalResult.Grounded
	assistantMessageID := uuid.New()
	if err := stream.SendMeta(StreamMeta{
		MessageID: assistantMessageID,
		Grounded:  grounded,
		Model:     model,
	}); err != nil {
		return nil
	}

	requestCtx, cancel := context.WithTimeout(handle.Context(), s.cfg.RequestTimeout)
	defer cancel()

	var assistantContent strings.Builder
	result, err := s.provider.StreamChat(requestCtx, ProviderRequest{
		Model:       model,
		Messages:    s.buildProviderMessages(detail.Session, detail.Messages, *userMessage, retrievalResult),
		Temperature: s.cfg.Temperature,
	}, func(delta string) error {
		if delta == "" {
			return nil
		}
		assistantContent.WriteString(delta)
		return stream.SendDelta(delta)
	})
	if err != nil {
		if handle.StopRequested() {
			_ = stream.SendDone("cancelled")
			return nil
		}
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
		Grounded:         grounded,
		Citations:        buildCitationInputs(retrievalResult.Chunks),
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

func (s *Service) RegenerateMessageStream(ctx context.Context, userID, sessionID, messageID uuid.UUID, stream StreamWriter) error {
	detail, err := s.repo.GetDetail(ctx, userID, sessionID)
	if err != nil {
		if err == ErrNotFound {
			return httpx.NotFound("session not found")
		}
		return httpx.Internal("failed to load session detail").WithErr(err)
	}

	targetMessage, userMessage, historyBeforeQuestion, err := findRegenerationContext(detail.Messages, messageID)
	if err != nil {
		return err
	}

	handle, err := s.streams.Acquire(ctx, sessionID)
	if err != nil {
		if errors.Is(err, ErrGenerationInProgress) {
			return httpx.Conflict("a message is already being generated for this session")
		}
		return httpx.Internal("failed to initialize stream generation").WithErr(err)
	}
	defer handle.Release()

	if err := stream.Start(); err != nil {
		return httpx.Internal("failed to start streaming response").WithErr(err)
	}
	defer stream.Close()

	model := strings.TrimSpace(detail.Session.Model)
	if model == "" {
		model = strings.TrimSpace(s.cfg.DefaultModel)
	}

	retrievalResult, shouldContinue := s.retrieveKnowledgeContext(handle.Context(), detail.Session, userMessage.Content, stream)
	if !shouldContinue {
		return nil
	}

	grounded := retrievalResult.Grounded
	if err := stream.SendMeta(StreamMeta{
		MessageID: targetMessage.ID,
		Grounded:  grounded,
		Model:     model,
	}); err != nil {
		return nil
	}

	requestCtx, cancel := context.WithTimeout(handle.Context(), s.cfg.RequestTimeout)
	defer cancel()

	var assistantContent strings.Builder
	result, err := s.provider.StreamChat(requestCtx, ProviderRequest{
		Model:       model,
		Messages:    s.buildProviderMessages(detail.Session, historyBeforeQuestion, *userMessage, retrievalResult),
		Temperature: s.cfg.Temperature,
	}, func(delta string) error {
		if delta == "" {
			return nil
		}
		assistantContent.WriteString(delta)
		return stream.SendDelta(delta)
	})
	if err != nil {
		if handle.StopRequested() {
			_ = stream.SendDone("cancelled")
			return nil
		}
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

	updateInput := UpdateMessageInput{
		Content:   fullAssistantContent,
		Status:    "completed",
		ModelUsed: &modelUsed,
		Grounded:  grounded,
		Citations: buildCitationInputs(retrievalResult.Chunks),
	}
	if result != nil {
		updateInput.PromptTokens = result.Usage.PromptTokens
		updateInput.CompletionTokens = result.Usage.CompletionTokens
		updateInput.TotalTokens = result.Usage.TotalTokens
	}

	if _, err := s.repo.UpdateMessage(ctx, userID, sessionID, messageID, updateInput); err != nil {
		if err == ErrNotFound {
			_ = stream.SendError(httpx.CodeResourceNotFound, "message not found")
			return nil
		}
		_ = stream.SendError(httpx.CodeInternal, "failed to persist regenerated assistant message")
		return nil
	}

	finishReason := "stop"
	if result != nil && strings.TrimSpace(result.FinishReason) != "" {
		finishReason = strings.TrimSpace(result.FinishReason)
	}
	_ = stream.SendDone(finishReason)
	return nil
}

func (s *Service) StopStream(ctx context.Context, userID, sessionID uuid.UUID) (bool, error) {
	if _, err := s.repo.GetDetail(ctx, userID, sessionID); err != nil {
		if err == ErrNotFound {
			return false, httpx.NotFound("session not found")
		}
		return false, httpx.Internal("failed to load session detail").WithErr(err)
	}
	return s.streams.Stop(sessionID), nil
}

func (s *Service) retrieveKnowledgeContext(ctx context.Context, session Session, content string, stream StreamWriter) (rag.RetrievalResult, bool) {
	if session.KnowledgeBaseID == nil {
		return rag.RetrievalResult{}, true
	}
	if s.rag == nil {
		_ = stream.SendError(httpx.CodeInternal, "knowledge-base retrieval is not configured")
		return rag.RetrievalResult{}, false
	}

	result, err := s.rag.Retrieve(ctx, *session.KnowledgeBaseID, content)
	if err == nil {
		return result, true
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) || ctx.Err() != nil {
		_ = stream.SendDone("cancelled")
		return rag.RetrievalResult{}, false
	}
	if errors.Is(err, rag.ErrKnowledgeBaseNotFound) {
		_ = stream.SendError(httpx.CodeResourceNotFound, "knowledge base not found")
		return rag.RetrievalResult{}, false
	}
	_ = stream.SendError(mapEmbeddingErrorCode(err), mapEmbeddingErrorMessage(err))
	return rag.RetrievalResult{}, false
}

func (s *Service) buildProviderMessages(session Session, history []Message, userMessage Message, retrieval rag.RetrievalResult) []ProviderMessage {
	items := make([]ProviderMessage, 0, len(history)+4)
	if systemPrompt := strings.TrimSpace(s.cfg.SystemPrompt); systemPrompt != "" {
		items = append(items, ProviderMessage{
			Role:    "system",
			Content: systemPrompt,
		})
	}
	if prompt := buildKnowledgeBaseSystemPrompt(session, retrieval); prompt != "" {
		items = append(items, ProviderMessage{
			Role:    "system",
			Content: prompt,
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

func buildKnowledgeBaseSystemPrompt(session Session, retrieval rag.RetrievalResult) string {
	if session.KnowledgeBaseID == nil {
		return ""
	}

	var builder strings.Builder
	if retrieval.Grounded && len(retrieval.Chunks) > 0 {
		builder.WriteString("This session is linked to a knowledge base. Use the retrieved context below as the primary source of truth when it is relevant.\n")
		builder.WriteString("If the context is insufficient, say that briefly and then continue with the best general answer without inventing citations.\n")
		if session.KnowledgeBaseName != nil && strings.TrimSpace(*session.KnowledgeBaseName) != "" {
			builder.WriteString("Knowledge base: ")
			builder.WriteString(strings.TrimSpace(*session.KnowledgeBaseName))
			builder.WriteString("\n")
		}
		if retrieval.PromptTemplate != nil && strings.TrimSpace(*retrieval.PromptTemplate) != "" {
			builder.WriteString("\nKnowledge-base instructions:\n")
			builder.WriteString(strings.TrimSpace(*retrieval.PromptTemplate))
			builder.WriteString("\n")
		}
		builder.WriteString("\nRetrieved context:\n")
		for _, chunk := range retrieval.Chunks {
			builder.WriteString(formatRetrievedChunk(chunk))
			builder.WriteString("\n")
		}
		return strings.TrimSpace(builder.String())
	}

	builder.WriteString("This session is linked to a knowledge base, but retrieval did not return relevant context for the latest user message.\n")
	builder.WriteString("Answer as a general assistant. Do not claim that the answer is grounded in the knowledge base and do not invent citations.")
	if session.KnowledgeBaseName != nil && strings.TrimSpace(*session.KnowledgeBaseName) != "" {
		builder.WriteString("\nKnowledge base: ")
		builder.WriteString(strings.TrimSpace(*session.KnowledgeBaseName))
	}
	return builder.String()
}

func formatRetrievedChunk(chunk rag.RetrievedChunk) string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("[%d] %s", chunk.Citation.Rank, chunk.Citation.DocumentName))
	if chunk.Citation.SourcePage != nil {
		builder.WriteString(fmt.Sprintf(" (page %d)", *chunk.Citation.SourcePage))
	}
	builder.WriteString("\n")
	builder.WriteString(strings.TrimSpace(chunk.Content))
	return builder.String()
}

func buildCitationInputs(chunks []rag.RetrievedChunk) []CreateCitationInput {
	if len(chunks) == 0 {
		return nil
	}

	citations := make([]CreateCitationInput, 0, len(chunks))
	for _, chunk := range chunks {
		citations = append(citations, CreateCitationInput{
			DocumentChunkID: chunk.Citation.DocumentChunkID,
			DocumentID:      chunk.Citation.DocumentID,
			KnowledgeBaseID: chunk.Citation.KnowledgeBaseID,
			RankNo:          chunk.Citation.Rank,
		})
	}
	return citations
}

func findRegenerationContext(messages []Message, targetMessageID uuid.UUID) (*Message, *Message, []Message, error) {
	messageIndex := make(map[uuid.UUID]int, len(messages))
	for idx := range messages {
		messageIndex[messages[idx].ID] = idx
	}

	targetIdx, ok := messageIndex[targetMessageID]
	if !ok {
		return nil, nil, nil, httpx.NotFound("message not found")
	}

	target := messages[targetIdx]
	if target.Role != "assistant" {
		return nil, nil, nil, httpx.ValidationFailed(httpx.FieldError{Field: "message_id", Message: "message must reference an assistant message"})
	}
	if target.ReplyToMessageID == nil {
		return nil, nil, nil, httpx.ValidationFailed(httpx.FieldError{Field: "message_id", Message: "assistant message is missing its source question"})
	}

	userIdx, ok := messageIndex[*target.ReplyToMessageID]
	if !ok {
		return nil, nil, nil, httpx.NotFound("source user message not found")
	}
	userMessage := messages[userIdx]
	if userMessage.Role != "user" {
		return nil, nil, nil, httpx.ValidationFailed(httpx.FieldError{Field: "message_id", Message: "assistant message is linked to an invalid source message"})
	}

	history := make([]Message, 0, userIdx)
	for idx := 0; idx < userIdx; idx++ {
		history = append(history, messages[idx])
	}
	return &target, &userMessage, history, nil
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

func mapEmbeddingErrorCode(err error) int {
	providerErr, ok := model.AsProviderError(err)
	if !ok {
		return httpx.CodeInternal
	}

	switch providerErr.Kind {
	case model.ProviderErrorAuthFailed:
		return httpx.CodeProviderAuthFailed
	case model.ProviderErrorRateLimited:
		return httpx.CodeProviderRateLimited
	case model.ProviderErrorUnavailable, model.ProviderErrorMisconfigured:
		return httpx.CodeProviderUnavailable
	default:
		return httpx.CodeInternal
	}
}

func mapEmbeddingErrorMessage(err error) string {
	providerErr, ok := model.AsProviderError(err)
	if !ok {
		return "knowledge-base retrieval failed"
	}

	switch providerErr.Kind {
	case model.ProviderErrorAuthFailed:
		return "embedding provider authentication failed"
	case model.ProviderErrorRateLimited:
		return "embedding provider rate limit exceeded"
	case model.ProviderErrorMisconfigured:
		return providerErr.Message
	case model.ProviderErrorBadRequest:
		if strings.TrimSpace(providerErr.Message) != "" {
			return providerErr.Message
		}
		return "embedding request is invalid"
	case model.ProviderErrorUnavailable:
		if strings.TrimSpace(providerErr.Message) != "" {
			return providerErr.Message
		}
		return "embedding provider is currently unavailable"
	default:
		return "knowledge-base retrieval failed"
	}
}
