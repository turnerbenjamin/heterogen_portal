package etc

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/turnerbenjamin/heterogen_portal/internal/constants"
)

type responseType int

const (
	ResponseTypeHtml responseType = iota
	ResponseTypeJson
)

// AppError holds data for errors returned from http handlers. It includes data
// intended for end users such as toast messages and page errors. It can also
// hold an inner error which may be logged.
type AppError struct {
	ResponseType     responseType `json:"-"`
	Code             int          `json:"-"`
	InnerError       error        `json:"-"`
	ErrorMessage     string       `json:"error_message"`
	SubErrorMessages []string     `json:"sub_error_messages,omitempty"`
}

func (e *AppError) WithJsonType() *AppError {
	return &AppError{
		ResponseType:     ResponseTypeJson,
		Code:             e.Code,
		ErrorMessage:     e.ErrorMessage,
		SubErrorMessages: e.SubErrorMessages,
		InnerError:       e.InnerError,
	}
}

// NewServerError is a helper for creating a server error which will display the
// generic server error message as both a toast and a server error. The
// underlying error is stored as an inner error for logging
func NewServerError(err error) *AppError {
	return &AppError{
		Code:             http.StatusInternalServerError,
		ErrorMessage:     constants.ErrMsgInternalServerError,
		SubErrorMessages: []string{constants.ErrMsgInternalServerError},
		InnerError:       err,
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

	if len(e.SubErrorMessages) > 0 {
		return fmt.Sprintf("[%s]", strings.Join(e.SubErrorMessages, ","))
	}

	return constants.EmptyAppErrorString
}
