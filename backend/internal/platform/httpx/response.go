package httpx

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

type successEnvelope struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

type errorEnvelope struct {
	Code      int          `json:"code"`
	Message   string       `json:"message"`
	Data      any          `json:"data"`
	RequestID string       `json:"request_id,omitempty"`
	Details   []FieldError `json:"details,omitempty"`
}

func Success(w http.ResponseWriter, status int, data any) {
	if data == nil {
		data = struct{}{}
	}
	writeJSON(w, status, successEnvelope{
		Code:    0,
		Message: "ok",
		Data:    data,
	})
}

func Error(w http.ResponseWriter, r *http.Request, err error) {
	appErr, ok := AsAppError(err)
	if !ok {
		appErr = Internal("internal server error").WithErr(err)
	}

	writeJSON(w, appErr.Status, errorEnvelope{
		Code:      appErr.Code,
		Message:   appErr.Message,
		Data:      nil,
		RequestID: RequestIDFromContext(r.Context()),
		Details:   appErr.Details,
	})
}

func DecodeJSON(r *http.Request, dst any) error {
	defer r.Body.Close()

	decoder := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		if errors.Is(err, io.EOF) {
			return BadRequest("request body is required")
		}
		return BadRequest("invalid request body").WithErr(err)
	}

	if decoder.More() {
		return BadRequest("request body must contain a single JSON object")
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
