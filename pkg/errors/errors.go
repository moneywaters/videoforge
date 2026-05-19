package errors

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

// ProblemDetails represents an RFC 7807 Problem Details response.
// This format is used to provide machine-readable details about errors
// in HTTP API responses.
type ProblemDetails struct {
	// Type is a URI reference that identifies the problem type.
	// Defaults to "about:blank" when not specified.
	Type string `json:"type,omitempty"`

	// Title is a short, human-readable summary of the problem type.
	// It should not change across occurrences of the problem.
	Title string `json:"title"`

	// Status is the HTTP status code.
	Status int `json:"status"`

	// Detail is a human-readable explanation specific to this occurrence.
	Detail string `json:"detail,omitempty"`

	// Instance is a URI reference that identifies the specific occurrence.
	Instance string `json:"instance,omitempty"`
}

// Error implements the error interface.
func (p *ProblemDetails) Error() string {
	return p.Detail
}

// JSON returns the JSON representation of the problem.
func (p *ProblemDetails) JSON() ([]byte, error) {
	return json.Marshal(p)
}

// ContentType returns the RFC 7807 content type.
func (p *ProblemDetails) ContentType() string {
	return "application/problem+json"
}

// StatusCode returns the HTTP status code.
func (p *ProblemDetails) StatusCode() int {
	return p.Status
}

// NewProblem creates a new ProblemDetails with the given parameters.
// The Instance field is automatically generated as a unique URI.
func NewProblem(status int, title, detail string) *ProblemDetails {
	return &ProblemDetails{
		Type:     "about:blank",
		Title:    title,
		Status:   status,
		Detail:   detail,
		Instance: "/errors/" + uuid.New().String(),
	}
}

// NewProblemWithType creates a new ProblemDetails with a custom type URI.
func NewProblemWithType(status int, title, detail, typeURI string) *ProblemDetails {
	return &ProblemDetails{
		Type:     typeURI,
		Title:    title,
		Status:   status,
		Detail:   detail,
		Instance: "/errors/" + uuid.New().String(),
	}
}

// BadRequest creates a ProblemDetails for a 400 Bad Request error.
func BadRequest(detail string) *ProblemDetails {
	return NewProblem(http.StatusBadRequest, "Bad Request", detail)
}

// NotFound creates a ProblemDetails for a 404 Not Found error.
func NotFound(detail string) *ProblemDetails {
	return NewProblem(http.StatusNotFound, "Not Found", detail)
}

// Unauthorized creates a ProblemDetails for a 401 Unauthorized error.
func Unauthorized(detail string) *ProblemDetails {
	return NewProblem(http.StatusUnauthorized, "Unauthorized", detail)
}

// Forbidden creates a ProblemDetails for a 403 Forbidden error.
func Forbidden(detail string) *ProblemDetails {
	return NewProblem(http.StatusForbidden, "Forbidden", detail)
}

// Internal creates a ProblemDetails for a 500 Internal Server Error.
// Note: In production, detailed internal errors should not be exposed to clients.
func Internal(detail string) *ProblemDetails {
	return NewProblem(http.StatusInternalServerError, "Internal Server Error", detail)
}

// InternalWithType creates a ProblemDetails for a 500 Internal Server Error
// with a custom type URI (useful for specific error types like validation failures).
func InternalWithType(detail string, typeURI string) *ProblemDetails {
	return NewProblemWithType(http.StatusInternalServerError, "Internal Server Error", detail, typeURI)
}

// Conflict creates a ProblemDetails for a 409 Conflict error.
func Conflict(detail string) *ProblemDetails {
	return NewProblem(http.StatusConflict, "Conflict", detail)
}

// UnprocessableEntity creates a ProblemDetails for a 422 Unprocessable Entity error.
func UnprocessableEntity(detail string) *ProblemDetails {
	return NewProblem(http.StatusUnprocessableEntity, "Unprocessable Entity", detail)
}

// TooManyRequests creates a ProblemDetails for a 429 Too Many Requests error.
func TooManyRequests(detail string) *ProblemDetails {
	return NewProblem(http.StatusTooManyRequests, "Too Many Requests", detail)
}

// ServiceUnavailable creates a ProblemDetails for a 503 Service Unavailable error.
func ServiceUnavailable(detail string) *ProblemDetails {
	return NewProblem(http.StatusServiceUnavailable, "Service Unavailable", detail)
}

// Write writes the ProblemDetails to the response writer.
// It sets the appropriate content type and status code.
func Write(ctx context.Context, w http.ResponseWriter, problem *ProblemDetails) {
	w.Header().Set("Content-Type", problem.ContentType())
	w.WriteHeader(problem.StatusCode())

	// Marshal with custom JSON encoder to handle special characters
	encoder := json.NewEncoder(w)
	if err := encoder.Encode(problem); err != nil {
		// If encoding fails, write a simple error response
		w.Write([]byte(`{"title":"Internal Server Error","status":500}`))
	}
}

// WriteError is a convenience function that writes an error to the response.
func WriteError(ctx context.Context, w http.ResponseWriter, err error) {
	if problem, ok := err.(*ProblemDetails); ok {
		Write(ctx, w, problem)
		return
	}

// For unknown errors, wrap as internal server error
	// In production, we should not expose internal error details
	internal := Internal("An internal error occurred")
	Write(ctx, w, internal)
}

// ToError converts a ProblemDetails to a standard error.
func (p *ProblemDetails) ToError() error {
	return p
}

// New creates a ProblemDetails with the given detail and status.
// This is a convenience function for use in handlers.
func New(detail string, status int) *ProblemDetails {
	return NewProblem(status, http.StatusText(status), detail)
}

// HTTPError wraps an error with HTTP status.
// This allows passing errors with specific HTTP status codes.
type HTTPError struct {
	Err    error
	Status int
	Title  string
}

// NewHTTPError creates a new HTTPError with the given parameters.
func NewHTTPError(status int, title, detail string) *HTTPError {
	return &HTTPError{
		Err:    NewProblem(status, title, detail),
		Status: status,
		Title:  title,
	}
}

func (e *HTTPError) Error() string {
	return e.Err.Error()
}