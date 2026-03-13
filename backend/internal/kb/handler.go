package kb

import (
	"errors"
	"io"
	"net/http"
	"strings"

	"backend/internal/platform/auth"
	"backend/internal/platform/httpx"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Handler struct {
	service *Service
}

type createKnowledgeBaseRequest struct {
	Name                string   `json:"name"`
	Description         *string  `json:"description"`
	EmbeddingModel      string   `json:"embedding_model"`
	PromptTemplate      *string  `json:"prompt_template"`
	RetrievalTopK       *int     `json:"retrieval_top_k"`
	SimilarityThreshold *float64 `json:"similarity_threshold"`
}

type updateKnowledgeBaseRequest struct {
	Name                string   `json:"name"`
	Description         *string  `json:"description"`
	PromptTemplate      *string  `json:"prompt_template"`
	RetrievalTopK       *int     `json:"retrieval_top_k"`
	SimilarityThreshold *float64 `json:"similarity_threshold"`
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/knowledge-bases", func(r chi.Router) {
		r.Get("/", httpx.Adapt(h.ListKnowledgeBases))
		r.Post("/", httpx.Adapt(h.CreateKnowledgeBase))
		r.Get("/{kbId}", httpx.Adapt(h.GetKnowledgeBase))
		r.Put("/{kbId}", httpx.Adapt(h.UpdateKnowledgeBase))
		r.Delete("/{kbId}", httpx.Adapt(h.DeleteKnowledgeBase))
		r.Post("/{kbId}/reindex", httpx.Adapt(h.ReindexKnowledgeBase))
		r.Get("/{kbId}/documents", httpx.Adapt(h.ListDocuments))
		r.Post("/{kbId}/documents", httpx.Adapt(h.UploadDocument))
		r.Get("/{kbId}/documents/{docId}", httpx.Adapt(h.GetDocument))
		r.Delete("/{kbId}/documents/{docId}", httpx.Adapt(h.DeleteDocument))
	})
}

func (h *Handler) CreateKnowledgeBase(w http.ResponseWriter, r *http.Request) error {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		return httpx.Unauthorized("missing auth context")
	}

	var req createKnowledgeBaseRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		return err
	}

	retrievalTopK := 5
	if req.RetrievalTopK != nil {
		retrievalTopK = *req.RetrievalTopK
	}

	kb, err := h.service.CreateKnowledgeBase(r.Context(), principal.UserID, CreateKnowledgeBaseInput{
		Name:                strings.TrimSpace(req.Name),
		Description:         normalizeOptionalString(req.Description),
		EmbeddingModel:      strings.TrimSpace(req.EmbeddingModel),
		PromptTemplate:      normalizeOptionalString(req.PromptTemplate),
		RetrievalTopK:       retrievalTopK,
		SimilarityThreshold: req.SimilarityThreshold,
	})
	if err != nil {
		return err
	}

	httpx.Success(w, http.StatusOK, map[string]any{"knowledge_base_id": kb.ID})
	return nil
}

func (h *Handler) ListKnowledgeBases(w http.ResponseWriter, r *http.Request) error {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		return httpx.Unauthorized("missing auth context")
	}

	page, size := httpx.ParsePageSize(r, 20, 100)
	result, err := h.service.ListKnowledgeBases(r.Context(), principal.UserID, page, size, r.URL.Query().Get("keyword"))
	if err != nil {
		return httpx.Internal("failed to list knowledge bases").WithErr(err)
	}

	httpx.Success(w, http.StatusOK, result)
	return nil
}

func (h *Handler) GetKnowledgeBase(w http.ResponseWriter, r *http.Request) error {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		return httpx.Unauthorized("missing auth context")
	}

	kbID, err := uuid.Parse(chi.URLParam(r, "kbId"))
	if err != nil {
		return httpx.BadRequest("invalid knowledge base id")
	}

	kb, err := h.service.GetKnowledgeBase(r.Context(), principal.UserID, kbID)
	if err != nil {
		return err
	}

	httpx.Success(w, http.StatusOK, kb)
	return nil
}

func (h *Handler) UpdateKnowledgeBase(w http.ResponseWriter, r *http.Request) error {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		return httpx.Unauthorized("missing auth context")
	}

	kbID, err := uuid.Parse(chi.URLParam(r, "kbId"))
	if err != nil {
		return httpx.BadRequest("invalid knowledge base id")
	}

	var req updateKnowledgeBaseRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		return err
	}

	retrievalTopK := 5
	if req.RetrievalTopK != nil {
		retrievalTopK = *req.RetrievalTopK
	}

	kb, err := h.service.UpdateKnowledgeBase(r.Context(), principal.UserID, kbID, UpdateKnowledgeBaseInput{
		Name:                strings.TrimSpace(req.Name),
		Description:         normalizeOptionalString(req.Description),
		PromptTemplate:      normalizeOptionalString(req.PromptTemplate),
		RetrievalTopK:       retrievalTopK,
		SimilarityThreshold: req.SimilarityThreshold,
	})
	if err != nil {
		return err
	}

	httpx.Success(w, http.StatusOK, kb)
	return nil
}

func (h *Handler) DeleteKnowledgeBase(w http.ResponseWriter, r *http.Request) error {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		return httpx.Unauthorized("missing auth context")
	}

	kbID, err := uuid.Parse(chi.URLParam(r, "kbId"))
	if err != nil {
		return httpx.BadRequest("invalid knowledge base id")
	}

	if err := h.service.DeleteKnowledgeBase(r.Context(), principal.UserID, kbID); err != nil {
		return err
	}

	httpx.Success(w, http.StatusOK, nil)
	return nil
}

func (h *Handler) ReindexKnowledgeBase(w http.ResponseWriter, r *http.Request) error {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		return httpx.Unauthorized("missing auth context")
	}

	kbID, err := uuid.Parse(chi.URLParam(r, "kbId"))
	if err != nil {
		return httpx.BadRequest("invalid knowledge base id")
	}

	task, err := h.service.ReindexKnowledgeBase(r.Context(), principal.UserID, kbID)
	if err != nil {
		return err
	}

	httpx.Success(w, http.StatusOK, map[string]any{"task_id": task.ID})
	return nil
}

func (h *Handler) ListDocuments(w http.ResponseWriter, r *http.Request) error {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		return httpx.Unauthorized("missing auth context")
	}

	kbID, err := uuid.Parse(chi.URLParam(r, "kbId"))
	if err != nil {
		return httpx.BadRequest("invalid knowledge base id")
	}

	page, size := httpx.ParsePageSize(r, 20, 100)
	result, err := h.service.ListDocuments(r.Context(), principal.UserID, kbID, page, size)
	if err != nil {
		return httpx.Internal("failed to list documents").WithErr(err)
	}

	httpx.Success(w, http.StatusOK, result)
	return nil
}

func (h *Handler) UploadDocument(w http.ResponseWriter, r *http.Request) error {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		return httpx.Unauthorized("missing auth context")
	}

	kbID, err := uuid.Parse(chi.URLParam(r, "kbId"))
	if err != nil {
		return httpx.BadRequest("invalid knowledge base id")
	}

	maxBytes := h.service.MaxUploadBytes()
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes+1024)
	if err := r.ParseMultipartForm(maxBytes); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			return httpx.FileTooLarge()
		}
		return httpx.BadRequest("invalid multipart form").WithErr(err)
	}
	if r.MultipartForm != nil {
		defer r.MultipartForm.RemoveAll()
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		return httpx.ValidationFailed(httpx.FieldError{Field: "file", Message: "file is required"})
	}
	defer file.Close()

	content, err := io.ReadAll(io.LimitReader(file, maxBytes+1))
	if err != nil {
		return httpx.BadRequest("failed to read uploaded file").WithErr(err)
	}
	if int64(len(content)) > maxBytes {
		return httpx.FileTooLarge()
	}

	result, err := h.service.UploadDocument(r.Context(), principal.UserID, kbID, UploadDocumentInput{
		Filename:    fileHeader.Filename,
		ContentType: fileHeader.Header.Get("Content-Type"),
		Content:     content,
	})
	if err != nil {
		return err
	}

	httpx.Success(w, http.StatusOK, result)
	return nil
}

func (h *Handler) GetDocument(w http.ResponseWriter, r *http.Request) error {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		return httpx.Unauthorized("missing auth context")
	}

	kbID, err := uuid.Parse(chi.URLParam(r, "kbId"))
	if err != nil {
		return httpx.BadRequest("invalid knowledge base id")
	}

	docID, err := uuid.Parse(chi.URLParam(r, "docId"))
	if err != nil {
		return httpx.BadRequest("invalid document id")
	}

	document, err := h.service.GetDocument(r.Context(), principal.UserID, kbID, docID)
	if err != nil {
		return err
	}

	httpx.Success(w, http.StatusOK, document)
	return nil
}

func (h *Handler) DeleteDocument(w http.ResponseWriter, r *http.Request) error {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		return httpx.Unauthorized("missing auth context")
	}

	kbID, err := uuid.Parse(chi.URLParam(r, "kbId"))
	if err != nil {
		return httpx.BadRequest("invalid knowledge base id")
	}

	docID, err := uuid.Parse(chi.URLParam(r, "docId"))
	if err != nil {
		return httpx.BadRequest("invalid document id")
	}

	if err := h.service.DeleteDocument(r.Context(), principal.UserID, kbID, docID); err != nil {
		return err
	}

	httpx.Success(w, http.StatusOK, nil)
	return nil
}

func normalizeOptionalString(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
