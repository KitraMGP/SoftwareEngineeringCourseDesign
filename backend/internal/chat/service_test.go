package chat

import (
	"context"
	"sync"
	"testing"
	"time"

	"backend/internal/platform/httpx"
	"backend/internal/rag"

	"github.com/google/uuid"
)

type stubChatRepository struct {
	detail                *SessionDetail
	createdMessages       []CreateMessageInput
	createdAssistantReply *CreateMessageInput
	updatedMessageInput   *UpdateMessageInput
}

func (r *stubChatRepository) Create(context.Context, uuid.UUID, CreateSessionInput) (*Session, error) {
	return nil, nil
}

func (r *stubChatRepository) List(context.Context, uuid.UUID, int, int, string) (*ListSessionsResult, error) {
	return nil, nil
}

func (r *stubChatRepository) GetDetail(context.Context, uuid.UUID, uuid.UUID) (*SessionDetail, error) {
	return r.detail, nil
}

func (r *stubChatRepository) Delete(context.Context, uuid.UUID, uuid.UUID) error {
	return nil
}

func (r *stubChatRepository) CreateMessage(_ context.Context, _ uuid.UUID, sessionID uuid.UUID, input CreateMessageInput) (*Message, error) {
	r.createdMessages = append(r.createdMessages, input)
	if input.Role == "assistant" {
		copied := input
		r.createdAssistantReply = &copied
	}

	messageID := input.ID
	if messageID == uuid.Nil {
		messageID = uuid.New()
	}

	return &Message{
		ID:               messageID,
		SessionID:        sessionID,
		Role:             input.Role,
		ReplyToMessageID: input.ReplyToMessageID,
		Content:          input.Content,
		Status:           input.Status,
		ModelUsed:        input.ModelUsed,
		Grounded:         input.Grounded,
		PromptTokens:     input.PromptTokens,
		CompletionTokens: input.CompletionTokens,
		TotalTokens:      input.TotalTokens,
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}, nil
}

func (r *stubChatRepository) UpdateMessage(_ context.Context, _ uuid.UUID, sessionID, messageID uuid.UUID, input UpdateMessageInput) (*Message, error) {
	copied := input
	r.updatedMessageInput = &copied

	return &Message{
		ID:               messageID,
		SessionID:        sessionID,
		Role:             "assistant",
		Content:          input.Content,
		Status:           input.Status,
		ModelUsed:        input.ModelUsed,
		Grounded:         input.Grounded,
		PromptTokens:     input.PromptTokens,
		CompletionTokens: input.CompletionTokens,
		TotalTokens:      input.TotalTokens,
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}, nil
}

type stubRAGRetriever struct {
	result rag.RetrievalResult
	err    error
}

func (r *stubRAGRetriever) Retrieve(context.Context, uuid.UUID, string) (rag.RetrievalResult, error) {
	return r.result, r.err
}

type stubProvider struct {
	result        *CompletionResult
	deltas        []string
	receivedInput ProviderRequest
	err           error
	waitForCtx    bool
	mu            sync.Mutex
}

func (p *stubProvider) StreamChat(ctx context.Context, req ProviderRequest, onDelta func(string) error) (*CompletionResult, error) {
	p.mu.Lock()
	p.receivedInput = req
	p.mu.Unlock()
	if p.waitForCtx {
		<-ctx.Done()
		return nil, ctx.Err()
	}
	for _, delta := range p.deltas {
		if err := onDelta(delta); err != nil {
			return nil, err
		}
	}
	if p.err != nil {
		return nil, p.err
	}
	return p.result, nil
}

type recordingStream struct {
	started bool
	meta    StreamMeta
	deltas  []string
	done    string
	errors  []struct {
		code    int
		message string
	}
}

func (s *recordingStream) Start() error {
	s.started = true
	return nil
}

func (s *recordingStream) SendMeta(meta StreamMeta) error {
	s.meta = meta
	return nil
}

func (s *recordingStream) SendDelta(content string) error {
	s.deltas = append(s.deltas, content)
	return nil
}

func (s *recordingStream) SendError(code int, message string) error {
	s.errors = append(s.errors, struct {
		code    int
		message string
	}{code: code, message: message})
	return nil
}

func (s *recordingStream) SendDone(finishReason string) error {
	s.done = finishReason
	return nil
}

func (s *recordingStream) Close() {}

func TestSendMessageStreamWithKnowledgeBaseRetrieval(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	sessionID := uuid.New()
	kbID := uuid.New()
	docID := uuid.New()
	chunkID := uuid.New()
	kbName := "Course Design Docs"
	promptTemplate := "Answer in concise Chinese."

	repo := &stubChatRepository{
		detail: &SessionDetail{
			Session: Session{
				ID:                sessionID,
				UserID:            userID,
				Model:             "deepseek-chat",
				KnowledgeBaseID:   &kbID,
				KnowledgeBaseName: &kbName,
			},
		},
	}
	retriever := &stubRAGRetriever{
		result: rag.RetrievalResult{
			Grounded:       true,
			PromptTemplate: &promptTemplate,
			Chunks: []rag.RetrievedChunk{
				{
					Citation: rag.Citation{
						DocumentChunkID: chunkID,
						DocumentID:      docID,
						KnowledgeBaseID: kbID,
						DocumentName:    "system-architecture.md",
						Rank:            1,
					},
					Content: "The system uses a modular monolith API plus a dedicated worker.",
				},
			},
		},
	}
	provider := &stubProvider{
		deltas: []string{"系统采用模块化单体架构。"},
		result: &CompletionResult{
			Model:        "deepseek-chat",
			FinishReason: "stop",
			Usage: Usage{
				PromptTokens:     20,
				CompletionTokens: 10,
				TotalTokens:      30,
			},
		},
	}
	stream := &recordingStream{}

	service := NewServiceWithRAG(repo, retriever, provider, ServiceConfig{
		DefaultModel:       "deepseek-chat",
		SystemPrompt:       "You are a helpful assistant.",
		RequestTimeout:     10 * time.Second,
		MaxHistoryMessages: 8,
		Temperature:        0.3,
	})

	if err := service.SendMessageStream(context.Background(), userID, sessionID, "请总结系统架构。", stream); err != nil {
		t.Fatalf("SendMessageStream() error = %v", err)
	}

	if !stream.started {
		t.Fatal("expected stream to start")
	}
	if !stream.meta.Grounded {
		t.Fatal("expected grounded meta=true")
	}
	if stream.done != "stop" {
		t.Fatalf("unexpected finish reason: got %q", stream.done)
	}
	if len(stream.errors) != 0 {
		t.Fatalf("expected no stream errors, got %+v", stream.errors)
	}
	if len(repo.createdMessages) != 2 {
		t.Fatalf("expected 2 created messages, got %d", len(repo.createdMessages))
	}
	if repo.createdAssistantReply == nil {
		t.Fatal("expected assistant message to be persisted")
	}
	if !repo.createdAssistantReply.Grounded {
		t.Fatal("expected assistant message grounded=true")
	}
	if len(repo.createdAssistantReply.Citations) != 1 {
		t.Fatalf("expected 1 citation, got %d", len(repo.createdAssistantReply.Citations))
	}
	if repo.createdAssistantReply.Citations[0].DocumentChunkID != chunkID {
		t.Fatalf("unexpected citation chunk id: got %s want %s", repo.createdAssistantReply.Citations[0].DocumentChunkID, chunkID)
	}
	if len(provider.receivedInput.Messages) < 2 {
		t.Fatalf("expected provider to receive system + kb prompt messages, got %d", len(provider.receivedInput.Messages))
	}
	if provider.receivedInput.Messages[1].Role != "system" {
		t.Fatalf("expected second provider message to be system prompt, got %q", provider.receivedInput.Messages[1].Role)
	}
	if provider.receivedInput.Messages[1].Content == "" {
		t.Fatal("expected knowledge base system prompt to be populated")
	}
}

func TestSendMessageStreamFallsBackWhenKnowledgeBaseHasNoHits(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	sessionID := uuid.New()
	kbID := uuid.New()
	kbName := "Empty KB"

	repo := &stubChatRepository{
		detail: &SessionDetail{
			Session: Session{
				ID:                sessionID,
				UserID:            userID,
				Model:             "deepseek-chat",
				KnowledgeBaseID:   &kbID,
				KnowledgeBaseName: &kbName,
			},
		},
	}
	retriever := &stubRAGRetriever{
		result: rag.RetrievalResult{},
	}
	provider := &stubProvider{
		deltas: []string{"这是通用回答。"},
		result: &CompletionResult{
			Model:        "deepseek-chat",
			FinishReason: "stop",
		},
	}
	stream := &recordingStream{}

	service := NewServiceWithRAG(repo, retriever, provider, ServiceConfig{
		DefaultModel:       "deepseek-chat",
		SystemPrompt:       "You are a helpful assistant.",
		RequestTimeout:     10 * time.Second,
		MaxHistoryMessages: 8,
	})

	if err := service.SendMessageStream(context.Background(), userID, sessionID, "今天天气怎么样？", stream); err != nil {
		t.Fatalf("SendMessageStream() error = %v", err)
	}

	if stream.meta.Grounded {
		t.Fatal("expected grounded meta=false")
	}
	if repo.createdAssistantReply == nil {
		t.Fatal("expected assistant reply to persist")
	}
	if repo.createdAssistantReply.Grounded {
		t.Fatal("expected assistant reply grounded=false")
	}
	if len(repo.createdAssistantReply.Citations) != 0 {
		t.Fatalf("expected no citations, got %d", len(repo.createdAssistantReply.Citations))
	}
}

func TestRegenerateMessageStreamOverwritesAssistantMessage(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	sessionID := uuid.New()
	kbID := uuid.New()
	userMessageID := uuid.New()
	assistantMessageID := uuid.New()
	chunkID := uuid.New()
	docID := uuid.New()
	kbName := "RAG KB"

	repo := &stubChatRepository{
		detail: &SessionDetail{
			Session: Session{
				ID:                sessionID,
				UserID:            userID,
				Model:             "deepseek-chat",
				KnowledgeBaseID:   &kbID,
				KnowledgeBaseName: &kbName,
			},
			Messages: []Message{
				{
					ID:        userMessageID,
					SessionID: sessionID,
					Role:      "user",
					Content:   "请总结系统架构。",
					Status:    "completed",
				},
				{
					ID:               assistantMessageID,
					SessionID:        sessionID,
					Role:             "assistant",
					ReplyToMessageID: &userMessageID,
					Content:          "旧答案",
					Status:           "completed",
				},
			},
		},
	}
	retriever := &stubRAGRetriever{
		result: rag.RetrievalResult{
			Grounded: true,
			Chunks: []rag.RetrievedChunk{
				{
					Citation: rag.Citation{
						DocumentChunkID: chunkID,
						DocumentID:      docID,
						KnowledgeBaseID: kbID,
						DocumentName:    "arch.md",
						Rank:            1,
					},
					Content: "API + worker",
				},
			},
		},
	}
	provider := &stubProvider{
		deltas: []string{"新答案"},
		result: &CompletionResult{
			Model:        "deepseek-chat",
			FinishReason: "stop",
			Usage: Usage{
				PromptTokens:     9,
				CompletionTokens: 4,
				TotalTokens:      13,
			},
		},
	}
	stream := &recordingStream{}

	service := NewServiceWithRAG(repo, retriever, provider, ServiceConfig{
		DefaultModel:       "deepseek-chat",
		SystemPrompt:       "You are a helpful assistant.",
		RequestTimeout:     10 * time.Second,
		MaxHistoryMessages: 8,
	})

	if err := service.RegenerateMessageStream(context.Background(), userID, sessionID, assistantMessageID, stream); err != nil {
		t.Fatalf("RegenerateMessageStream() error = %v", err)
	}

	if stream.meta.MessageID != assistantMessageID {
		t.Fatalf("unexpected regenerated message id: got %s want %s", stream.meta.MessageID, assistantMessageID)
	}
	if repo.updatedMessageInput == nil {
		t.Fatal("expected assistant message to be overwritten")
	}
	if repo.updatedMessageInput.Content != "新答案" {
		t.Fatalf("unexpected regenerated content: got %q", repo.updatedMessageInput.Content)
	}
	if !repo.updatedMessageInput.Grounded {
		t.Fatal("expected regenerated answer grounded=true")
	}
	if len(repo.updatedMessageInput.Citations) != 1 {
		t.Fatalf("expected 1 regenerated citation, got %d", len(repo.updatedMessageInput.Citations))
	}
}

func TestStopStreamCancelsActiveGeneration(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	sessionID := uuid.New()

	repo := &stubChatRepository{
		detail: &SessionDetail{
			Session: Session{
				ID:     sessionID,
				UserID: userID,
				Model:  "deepseek-chat",
			},
		},
	}
	provider := &stubProvider{
		waitForCtx: true,
	}
	stream := &recordingStream{}
	service := NewServiceWithRAG(repo, nil, provider, ServiceConfig{
		DefaultModel:       "deepseek-chat",
		SystemPrompt:       "You are a helpful assistant.",
		RequestTimeout:     10 * time.Second,
		MaxHistoryMessages: 8,
	})

	doneCh := make(chan error, 1)
	go func() {
		doneCh <- service.SendMessageStream(context.Background(), userID, sessionID, "hello", stream)
	}()

	deadline := time.Now().Add(2 * time.Second)
	for {
		if time.Now().After(deadline) {
			t.Fatal("timed out waiting for stream registration")
		}
		if stopped, err := service.StopStream(context.Background(), userID, sessionID); err == nil && stopped {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	select {
	case err := <-doneCh:
		if err != nil {
			t.Fatalf("SendMessageStream() returned error after stop: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for stopped stream")
	}

	if stream.done != "cancelled" {
		t.Fatalf("unexpected cancelled finish reason: got %q", stream.done)
	}
	if repo.createdAssistantReply != nil {
		t.Fatal("expected no assistant reply to be persisted after stop")
	}
}

func TestFindRegenerationContextRejectsUserMessage(t *testing.T) {
	t.Parallel()

	userMessageID := uuid.New()
	_, _, _, err := findRegenerationContext([]Message{
		{ID: userMessageID, Role: "user", Content: "question", Status: "completed"},
	}, userMessageID)
	if err == nil {
		t.Fatal("expected error")
	}

	appErr, ok := httpx.AsAppError(err)
	if !ok {
		t.Fatalf("expected AppError, got %T", err)
	}
	if appErr.Code != httpx.CodeValidationFailed {
		t.Fatalf("unexpected app error code: got %d", appErr.Code)
	}
}
