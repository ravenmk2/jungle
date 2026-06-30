// Package apperrors provides typed errors and stable error codes.
package apperrors

// Code is a stable, UPPER_SNAKE error code.
type Code string

const (
	ValidationError   Code = "VALIDATION_ERROR"
	NotFound          Code = "NOT_FOUND"
	MethodNotAllowed  Code = "METHOD_NOT_ALLOWED"
	InternalError     Code = "INTERNAL_ERROR"
	BuildFailed       Code = "BUILD_FAILED"
	RunFailed         Code = "RUN_FAILED"
	ServiceNotRunning Code = "SERVICE_NOT_RUNNING"
	DBQueryFailed     Code = "DB_QUERY_FAILED"
	DBResetFailed     Code = "DB_RESET_FAILED"
	SearchFailed      Code = "SEARCH_FAILED"
	AmbiguousTarget   Code = "AMBIGUOUS_TARGET"
)

// Detail is a field-level error.
type Detail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Target  string `json:"target"`
}

// Error is a typed error carrying a stable code and optional details.
type Error struct {
	Code    Code
	Message string
	Details []Detail
}

func (e *Error) Error() string { return e.Message }

// New creates a typed error.
func New(code Code, msg string) *Error {
	return &Error{Code: code, Message: msg}
}

// NewWithDetails creates a typed error with field-level details.
func NewWithDetails(code Code, msg string, details []Detail) *Error {
	return &Error{Code: code, Message: msg, Details: details}
}
