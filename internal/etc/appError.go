package etc

import (
	"fmt"
	"strings"

	"github.com/turnerbenjamin/heterogen_portal/internal/constants"
)

type AppError struct {
	Code       int
	ToastError string
	PageErrors []string
	InnerError error
}

func (e *AppError) String() string {
	if e.InnerError != nil {
		return e.InnerError.Error()
	}

	sb := strings.Builder{}
	if e.ToastError != "" {
		sb.WriteString(": ")
		sb.WriteString(e.ToastError)
	}

	if len(e.PageErrors) > 0 {
		sb.WriteString(": ")
		sb.WriteString(strings.Join(e.PageErrors, ", "))
	}

	return sb.String()
}

func NewServerError(err error) *AppError {
	appErr := ToastAndPageErrors(
		500,
		constants.ErrMsgInternalServerError,
		constants.ErrMsgInternalServerError,
	)
	appErr.InnerError = err
	return appErr
}

func ToastError(httpCode int, message string) *AppError {
	return &AppError{
		Code:       httpCode,
		ToastError: message,
		PageErrors: []string{},
	}
}

func ToastErrorf(httpCode int, format string, a ...any) *AppError {
	return &AppError{
		Code:       httpCode,
		ToastError: fmt.Sprintf(format, a...),
		PageErrors: []string{},
	}
}

func PageError(httpCode int, messages ...string) *AppError {
	return &AppError{
		Code:       httpCode,
		ToastError: "",
		PageErrors: messages,
	}
}

func ToastAndPageErrors(httpCode int, toastError string, pageErrors ...string) *AppError {
	return &AppError{
		Code:       httpCode,
		ToastError: toastError,
		PageErrors: pageErrors,
	}
}
