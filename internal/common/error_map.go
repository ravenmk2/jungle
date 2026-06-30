package common

import (
	"errors"

	"github.com/ravenmk2/jungle/internal/apperrors"
)

// MapError maps any error to an envelope. Typed errors keep their code;
// unknown errors become INTERNAL_ERROR.
func MapError(err error) Envelope {
	var te *apperrors.Error
	if errors.As(err, &te) {
		return Fail(te)
	}
	return Fail(apperrors.New(apperrors.InternalError, "internal error"))
}
