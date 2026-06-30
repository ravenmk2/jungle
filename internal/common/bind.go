package common

import (
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/ravenmk2/jungle/internal/apperrors"
)

var validate = validator.New()

// Bind decodes the JSON request body into req and validates it.
// Returns a typed ValidationError (with field-level details) on failure.
func Bind(c echo.Context, req interface{}) error {
	if err := c.Bind(req); err != nil {
		return apperrors.New(apperrors.ValidationError, "invalid request body")
	}
	if err := validate.Struct(req); err != nil {
		return toValidationError(err)
	}
	return nil
}

func toValidationError(err error) *apperrors.Error {
	details := []apperrors.Detail{}
	if ve, ok := err.(validator.ValidationErrors); ok {
		for _, fe := range ve {
			details = append(details, apperrors.Detail{
				Code:    fe.Tag(),
				Message: "field " + fe.Field() + " failed " + fe.Tag(),
				Target:  fe.Field(),
			})
		}
	}
	return apperrors.NewWithDetails(apperrors.ValidationError, "validation failed", details)
}
