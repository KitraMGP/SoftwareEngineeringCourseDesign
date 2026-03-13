package httpx

import (
	"errors"
	"net/http"
)

const (
	CodeValidationFailed    = 40000
	CodePromptTooLarge      = 40002
	CodeInvalidCredentials  = 40001
	CodeUnauthorized        = 40100
	CodeForbidden           = 40300
	CodeResourceNotFound    = 40400
	CodeConflict            = 40900
	CodeDuplicateDocument   = 40903
	CodeUnsupportedType     = 42201
	CodeFileTooLarge        = 42202
	CodeFeatureNotReady     = 50100
	CodeInternal            = 50000
	CodeStorageError        = 50003
	CodeTaskDispatchFailed  = 50004
	CodeProviderUnavailable = 50301
	CodeProviderAuthFailed  = 50302
	CodeProviderRateLimited = 50303
)

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type AppError struct {
	Status  int
	Code    int
	Message string
	Details []FieldError
	Err     error
}

func (e *AppError) Error() string {
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func (e *AppError) WithErr(err error) *AppError {
	e.Err = err
	return e
}

func (e *AppError) WithDetails(details ...FieldError) *AppError {
	e.Details = append(e.Details, details...)
	return e
}

func AsAppError(err error) (*AppError, bool) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}

func ValidationFailed(details ...FieldError) *AppError {
	return &AppError{
		Status:  http.StatusBadRequest,
		Code:    CodeValidationFailed,
		Message: "validation failed",
		Details: details,
	}
}

func BadRequest(message string) *AppError {
	return &AppError{Status: http.StatusBadRequest, Code: CodeValidationFailed, Message: message}
}

func Unauthorized(message string) *AppError {
	return &AppError{Status: http.StatusUnauthorized, Code: CodeUnauthorized, Message: message}
}

func InvalidCredentials() *AppError {
	return &AppError{Status: http.StatusUnauthorized, Code: CodeInvalidCredentials, Message: "invalid credentials"}
}

func Forbidden(message string) *AppError {
	return &AppError{Status: http.StatusForbidden, Code: CodeForbidden, Message: message}
}

func NotFound(message string) *AppError {
	return &AppError{Status: http.StatusNotFound, Code: CodeResourceNotFound, Message: message}
}

func Conflict(message string) *AppError {
	return &AppError{Status: http.StatusConflict, Code: CodeConflict, Message: message}
}

func DuplicateDocument() *AppError {
	return &AppError{Status: http.StatusConflict, Code: CodeDuplicateDocument, Message: "duplicate document"}
}

func UnsupportedFileType() *AppError {
	return &AppError{Status: http.StatusUnprocessableEntity, Code: CodeUnsupportedType, Message: "unsupported file type"}
}

func FileTooLarge() *AppError {
	return &AppError{Status: http.StatusUnprocessableEntity, Code: CodeFileTooLarge, Message: "file too large"}
}

func FeatureNotReady(message string) *AppError {
	return &AppError{Status: http.StatusNotImplemented, Code: CodeFeatureNotReady, Message: message}
}

func Internal(message string) *AppError {
	return &AppError{Status: http.StatusInternalServerError, Code: CodeInternal, Message: message}
}

func StorageError(message string) *AppError {
	return &AppError{Status: http.StatusInternalServerError, Code: CodeStorageError, Message: message}
}

func TaskDispatchFailed(message string) *AppError {
	return &AppError{Status: http.StatusInternalServerError, Code: CodeTaskDispatchFailed, Message: message}
}
