package etc

import (
	"fmt"
	"strings"
)

type AppError interface {
	Code() int
	ToastError() string
	PageErrors() []string
}

var (
	ErrServer = ToastAndPageErrors(
		500,
		"An unexpected error has occurred. Please try again later",
		"An unexpected error has occurred. Please try again later",
	)
)

type appError struct {
	code       int
	toastError string
	pageErrors []string
}

func (e *appError) String() string {
	sb := strings.Builder{}
	if e.toastError != "" {
		sb.WriteString(": ")
		sb.WriteString(e.toastError)
	}

	if len(e.pageErrors) > 0 {
		sb.WriteString(": ")
		sb.WriteString(strings.Join(e.pageErrors, ", "))
	}

	errorMessage := sb.String()

	return fmt.Sprintf("%d%s", e.code, errorMessage)
}

func (e *appError) Code() int {
	return e.code
}

func (e *appError) ToastError() string {
	return e.toastError
}

func (e *appError) PageErrors() []string {
	return e.pageErrors
}

func ToastError(httpCode int, message string) AppError {
	return &appError{
		code:       httpCode,
		toastError: message,
		pageErrors: []string{},
	}
}

func ToastErrorf(httpCode int, format string, a ...any) AppError {
	return &appError{
		code:       httpCode,
		toastError: fmt.Sprintf(format, a...),
		pageErrors: []string{},
	}
}

func PageError(httpCode int, messages ...string) AppError {
	return &appError{
		code:       httpCode,
		toastError: "",
		pageErrors: messages,
	}
}

func ToastAndPageErrors(httpCode int, toastError string, pageErrors ...string) AppError {
	return &appError{
		code:       httpCode,
		toastError: toastError,
		pageErrors: pageErrors,
	}
}
