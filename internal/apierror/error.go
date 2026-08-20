// Package apierror defines the structured JSON error envelope used by the HTTP API.
package apierror

import "net/http"

// Code is a stable machine-readable error identifier for API clients.
type Code string

const (
	CodeInvalidRequest   Code = "invalid_request"
	CodeInvalidName      Code = "invalid_name"
	CodeInvalidPath      Code = "invalid_path"
	CodeNotFound         Code = "not_found"
	CodeConflict         Code = "conflict"
	CodeUnsupportedType  Code = "unsupported_type"
	CodeTooLarge         Code = "too_large"
	CodeNotReady         Code = "not_ready"
	CodeInternal         Code = "internal"
	CodeMethodNotAllowed Code = "method_not_allowed"
)

// Body is the JSON object written on error responses.
type Body struct {
	Error Error `json:"error"`
}

// Error is a single structured API error.
type Error struct {
	Code      Code   `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id,omitempty"`
}

// New builds a Body with the given code, message, and request id.
func New(code Code, message, requestID string) Body {
	return Body{Error: Error{Code: code, Message: message, RequestID: requestID}}
}

// HTTPStatus maps an error code to an HTTP status.
func HTTPStatus(code Code) int {
	switch code {
	case CodeInvalidRequest, CodeInvalidName, CodeInvalidPath:
		return http.StatusBadRequest
	case CodeNotFound:
		return http.StatusNotFound
	case CodeConflict:
		return http.StatusConflict
	case CodeUnsupportedType:
		return http.StatusUnprocessableEntity
	case CodeTooLarge:
		return http.StatusRequestEntityTooLarge
	case CodeNotReady:
		return http.StatusServiceUnavailable
	case CodeMethodNotAllowed:
		return http.StatusMethodNotAllowed
	default:
		return http.StatusInternalServerError
	}
}
