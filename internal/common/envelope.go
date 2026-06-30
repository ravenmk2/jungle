// Package common provides transport-shared helpers: response envelope, request binding, and error mapping.
package common

import "github.com/ravenmk2/jungle/internal/apperrors"

// Envelope is the unified RPC-style response envelope.
type Envelope struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Error   *ErrorBody  `json:"error"`
}

// ErrorBody is the error part of the envelope.
type ErrorBody struct {
	Code    string             `json:"code"`
	Message string             `json:"message"`
	Details []apperrors.Detail `json:"details,omitempty"`
}

// OK builds a success envelope.
func OK(data interface{}) Envelope {
	return Envelope{Success: true, Data: data}
}

// Fail builds a failure envelope from a typed error.
func Fail(err *apperrors.Error) Envelope {
	return Envelope{Success: false, Error: &ErrorBody{
		Code:    string(err.Code),
		Message: err.Message,
		Details: err.Details,
	}}
}
