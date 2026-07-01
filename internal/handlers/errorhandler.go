package handlers

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/ravenmk2/jungle/internal/apperrors"
	"github.com/ravenmk2/jungle/internal/common"
)

// ErrorHandler is Echo's centralized HTTP error handler. Every error — business
// apperrors, Echo transport errors (404/405/415/500, including sentinel errors
// like ErrNotFound and *echo.HTTPError), panics recovered by middleware.Recover,
// and unknown errors — is rendered as a response envelope.
//
// Per spec 5.1: business errors return HTTP 200 with success=false; transport
// and system errors return the matching 4xx/5xx status. Messages are fixed
// strings so internal details are never leaked.
func ErrorHandler(c *echo.Context, err error) {
	// Business error -> HTTP 200 + envelope.
	var ne *apperrors.Error
	if errors.As(err, &ne) {
		_ = c.JSON(http.StatusOK, common.Fail(ne))
		return
	}
	// Echo transport/system error (implements echo.HTTPStatusCoder) -> its status + envelope.
	if code := echo.StatusCode(err); code != 0 {
		_ = c.JSON(code, common.Fail(mapStatusToAppError(code)))
		return
	}
	// Unknown error (e.g. recovered panic not wrapped as HTTPStatusCoder) -> 500 + envelope.
	_ = c.JSON(http.StatusInternalServerError, common.Fail(apperrors.New(apperrors.InternalError, "internal error")))
}

// mapStatusToAppError converts an HTTP status code into a typed apperror for the
// envelope. Uses fixed messages to avoid leaking internals.
func mapStatusToAppError(code int) *apperrors.Error {
	switch code {
	case http.StatusNotFound:
		return apperrors.New(apperrors.NotFound, "not found")
	case http.StatusMethodNotAllowed:
		return apperrors.New(apperrors.MethodNotAllowed, "method not allowed")
	default:
		return apperrors.New(apperrors.InternalError, "internal error")
	}
}
