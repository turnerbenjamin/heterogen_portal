package etc

import (
	"net/http"

	"github.com/turnerbenjamin/heterogen_portal/internal/constants"
)

// AppError holds data for errors returned from http handlers. It includes data
// intended for end users such as toast messages and page errors. It can also
// hold an inner error which may be logged.
type AppError struct {
	Code         int
	ErrorMessage string
	InnerError   error
}

// NewServerError is a helper for creating a server error which will display the
// generic server error message as both a toast and a server error. The
// underlying error is stored as an inner error for logging
func NewServerError(err error) *AppError {
	return &AppError{
		Code:         http.StatusInternalServerError,
		ErrorMessage: constants.ErrMsgInternalServerError,
		InnerError:   err,
	}
}

// String returns a string representation of the error for logging purposes
func (e *AppError) String() string {
	if e.InnerError != nil {
		return e.InnerError.Error()
	}

	if e.ErrorMessage != "" {
		return e.ErrorMessage
	}

	return constants.EmptyAppErrorString
}
