package etc

import (
	"fmt"
	"strings"
)

var (
	ErrServer = ToastAndPageErrors(
		500,
		"An unexpected error has occurred. Please try again later",
		"An unexpected error has occurred. Please try again later",
	)
)

type AppError struct {
	Code       int
	ToastError string
	PageErrors []string
}

func (e *AppError) String() string {
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

func (e *AppError) Error() string {
	return e.String()
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
