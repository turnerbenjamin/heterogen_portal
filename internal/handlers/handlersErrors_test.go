package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/turnerbenjamin/heterogen_portal/internal/constants"
	"github.com/turnerbenjamin/heterogen_portal/internal/templates"
)

func TestWrite_ReturnsError_ForUnsupportedContentType(t *testing.T) {
	t.Parallel()

	testData := []struct {
		responseType string
		wantError    error
	}{
		{
			responseType: ContentTypeHtml.String(),
			wantError:    nil,
		},
		{
			responseType: ContentTypeJson.String(),
			wantError:    nil,
		},
		{
			responseType: "multipart/form-data",
			wantError: fmt.Errorf(
				constants.ErrMsgPatternUnsupportedContentType,
				"multipart/form-data",
			),
		},
		{
			responseType: "text/javascript",
			wantError: fmt.Errorf(
				constants.ErrMsgPatternUnsupportedContentType,
				"text/javascript",
			),
		},
		{
			responseType: "",
			wantError:    errors.New("mime: no media type"),
		},
	}

	for _, td := range testData {
		testAppError := &AppError{Code: 500, ToastError: "Test error"}
		ts := NewMockTemplateStore(t)

		r := httptest.NewRequest("GET", "/", strings.NewReader(""))

		w := httptest.NewRecorder()
		w.Header().Set("Content-Type", td.responseType)

		ts.EXPECT().
			Execute(mock.Anything, mock.Anything, mock.Anything).
			Maybe().
			Return(nil)

		h := NewErrorHandler(ts)

		err := h.Write(w, r, testAppError)

		assert.Equal(t, td.wantError, err)
	}
}

func TestWrite_HandlesErrorResponseWhenContentTypeIsJson(t *testing.T) {
	t.Parallel()

	testAppError := &AppError{
		Code:       200,
		ToastError: "Test error",
		PageErrors: []string{"A page error", "and another"},
	}

	r := httptest.NewRequest("GET", "/", strings.NewReader(""))

	w := httptest.NewRecorder()
	w.Header().Set("Content-Type", ContentTypeJson.String())

	ts := NewMockTemplateStore(t)
	h := NewErrorHandler(ts)

	err := h.Write(w, r, testAppError)
	require.NoError(t, err)

	gotBody := w.Body.String()
	wantBodyBytes, err := json.Marshal(testAppError)
	require.NoError(t, err)

	assert.Equal(t, testAppError.Code, w.Code)
	assert.Equal(t, string(wantBodyBytes), gotBody)
}

func TestWrite_HandlesErrorResponseWhenContentTypeIsHtml(t *testing.T) {
	t.Parallel()

	testData := []struct {
		isHtmx               bool
		wantTemplate         templates.TemplateIdentifier
		wantContentOnlyValue bool
		wantStatusCode       int
		wantExecuteCallCount int
	}{
		{
			isHtmx:               true,
			wantTemplate:         templates.TmplComponentErrors,
			wantContentOnlyValue: true,
			wantStatusCode:       500,
			wantExecuteCallCount: 1,
		},
		{
			isHtmx:               false,
			wantTemplate:         templates.TmplPageOutOfAppErr,
			wantContentOnlyValue: false,
			wantStatusCode:       418,
			wantExecuteCallCount: 1,
		},
	}

	for _, td := range testData {
		testAppError := &AppError{
			Code:       td.wantStatusCode,
			ToastError: "Some toast error",
			PageErrors: []string{"A page error", "and another"},
		}

		ts := NewMockTemplateStore(t)

		r := httptest.NewRequest("GET", "/", strings.NewReader(""))
		if td.isHtmx {
			r.Header.Set(constants.HxRequestHeaderRequest, "true")
		}

		w := httptest.NewRecorder()
		w.Header().Set("Content-Type", ContentTypeHtml.String())

		var capturedTemplateArgs templates.TemplateArgs
		ts.EXPECT().
			Execute(td.wantTemplate, w, mock.Anything).
			Run(func(_ templates.TemplateIdentifier, _ io.Writer, data templates.TemplateArgs) {
				capturedTemplateArgs = data
			}).
			Once().
			Return(nil)

		h := NewErrorHandler(ts)

		err := h.Write(w, r, testAppError)

		assert.Nil(t, err)
		assert.Equal(t, td.wantStatusCode, w.Code)

		assert.Equal(t, td.wantContentOnlyValue, capturedTemplateArgs.PageConfig.ContentOnly)
		assert.Equal(t, testAppError, capturedTemplateArgs.Data)
	}
}

func TestWrite_ShouldReturnErrorsReturnedFromExecute(t *testing.T) {
	t.Parallel()

	wantError := errors.New("test error")

	w := httptest.NewRecorder()
	w.Header().Set("Content-Type", ContentTypeHtml.String())

	ts := NewMockTemplateStore(t)
	ts.EXPECT().
		Execute(mock.Anything, mock.Anything, mock.Anything).
		Return(wantError)

	h := NewErrorHandler(ts)

	r := httptest.NewRequest("GET", "/", strings.NewReader(""))
	gotErr := h.Write(w, r, &AppError{Code: 200})

	assert.EqualError(t, gotErr, wantError.Error())
}
