package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/turnerbenjamin/heterogen_portal/internal/constants"
)

// AppError holds errors returned from http handlers. It includes data intended
// for users such as toast messages and page errors. It can also hold an
// inner error which may be logged.
type AppError struct {
	Code       int
	ToastError string
	PageErrors []string
	InnerError error
}

// NewServerError is a helper for creating a server error which will display the
// generic server error message as both a toast and a server error. The
// underlying error is stored as an inner error for logging
func NewServerError(err error) *AppError {
	return &AppError{
		Code:       http.StatusInternalServerError,
		ToastError: constants.ErrMsgInternalServerError,
		PageErrors: []string{constants.ErrMsgInternalServerError},
		InnerError: err,
	}
}

// String returns a string representation of the error for logging purposes
func (e *AppError) String() string {
	if e.InnerError != nil {
		return e.InnerError.Error()
	}

	if e.ToastError != "" {
		return e.ToastError
	}

	if len(e.PageErrors) > 0 {
		return fmt.Sprintf("[%ss]", strings.Join(e.PageErrors, ","))
	}

	return constants.DefaultAppErrorString
}
