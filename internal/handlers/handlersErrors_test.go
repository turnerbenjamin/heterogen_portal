package handlers

import (
	"context"
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
	"github.com/turnerbenjamin/heterogen_portal/internal/etc"
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
		testAppError := &etc.AppError{Code: 500, ErrorMessage: "Test error"}

		r := httptest.NewRequest("GET", "/", strings.NewReader(""))

		w := httptest.NewRecorder()
		w.Header().Set("Content-Type", td.responseType)

		ts := NewMockTemplateStore(t)
		ts.EXPECT().
			ExecuteRenderAppError(mock.Anything, mock.Anything, mock.Anything).
			Maybe().
			Return(nil)

		ts.EXPECT().
			ExecutePageOutOfAppError(mock.Anything, mock.Anything, mock.Anything).
			Maybe().
			Return(nil)

		js := NewMockJsonSerialiser(t)
		js.EXPECT().Marshal(mock.Anything).Maybe().Return([]byte("test"), nil)

		h := NewErrorHandler(ts, js)

		err := h.Write(w, r, testAppError)
		assert.Equal(t, td.wantError, err)
	}
}

func TestWrite_HandlesErrorResponseWhenContentTypeIsJson(t *testing.T) {
	t.Parallel()

	testAppError := &etc.AppError{
		Code:         200,
		ErrorMessage: "Test error",
	}
	expectedResponse := "expected response"

	r := httptest.NewRequest("GET", "/", strings.NewReader(""))

	w := httptest.NewRecorder()
	w.Header().Set("Content-Type", ContentTypeJson.String())

	ts := NewMockTemplateStore(t)
	js := NewMockJsonSerialiser(t)
	js.EXPECT().Marshal(mock.Anything).Maybe().Return([]byte(expectedResponse), nil)

	h := NewErrorHandler(ts, js)

	err := h.Write(w, r, testAppError)
	require.NoError(t, err)

	gotBody := w.Body.String()

	assert.Equal(t, testAppError.Code, w.Code)
	assert.Equal(t, string(expectedResponse), gotBody)
}

func TestWrite_ReturnsErrorsFromJsonMarshal(t *testing.T) {
	t.Parallel()

	testAppError := &etc.AppError{
		Code:         500,
		ErrorMessage: "Test error",
	}
	expectedErr := errors.New("expected error")

	r := httptest.NewRequest("GET", "/", strings.NewReader(""))

	w := httptest.NewRecorder()
	w.Header().Set("Content-Type", ContentTypeJson.String())

	ts := NewMockTemplateStore(t)
	js := NewMockJsonSerialiser(t)
	js.EXPECT().Marshal(mock.Anything).Maybe().Return(nil, expectedErr)

	h := NewErrorHandler(ts, js)

	err := h.Write(w, r, testAppError)
	assert.EqualError(t, err, expectedErr.Error())
}

func TestWrite_SetsErrorCodeIfNotSet(t *testing.T) {
	t.Parallel()

	testAppError := &etc.AppError{
		Code:         0,
		ErrorMessage: "Test error",
	}

	r := httptest.NewRequest("GET", "/", strings.NewReader(""))

	w := httptest.NewRecorder()
	w.Header().Set("Content-Type", ContentTypeJson.String())

	ts := NewMockTemplateStore(t)
	js := NewMockJsonSerialiser(t)
	js.EXPECT().Marshal(mock.Anything).Maybe().Return([]byte("data"), nil)

	h := NewErrorHandler(ts, js)

	err := h.Write(w, r, testAppError)
	assert.NoError(t, err)
	assert.Equal(t, 500, w.Code)
}

func TestWrite_HandlesErrorResponseWhenContentTypeIsHtmlForHTMXRequest(t *testing.T) {
	t.Parallel()

	wantStatusCode := 500

	testAppError := &etc.AppError{
		Code:         wantStatusCode,
		ErrorMessage: "Some toast error",
	}

	ts := NewMockTemplateStore(t)

	r := httptest.NewRequest("GET", "/", strings.NewReader(""))
	r.Header.Set(constants.HxRequestHeaderRequest, "true")

	w := httptest.NewRecorder()
	w.Header().Set("Content-Type", ContentTypeHtml.String())

	var capturedErrorData *etc.AppError
	ts.EXPECT().
		ExecuteRenderAppError(r.Context(), w, mock.Anything).
		Run(func(_ context.Context, _ io.Writer, ed *etc.AppError) {
			capturedErrorData = ed
		}).
		Once().
		Return(nil)

	js := NewMockJsonSerialiser(t)
	js.EXPECT().Marshal(mock.Anything).Maybe().Return([]byte("test"), nil)

	h := NewErrorHandler(ts, js)

	err := h.Write(w, r, testAppError)

	assert.Nil(t, err)
	assert.Equal(t, wantStatusCode, w.Code)

	assert.Equal(t, testAppError.Code, capturedErrorData.Code)
	assert.Equal(t, testAppError.ErrorMessage, capturedErrorData.ErrorMessage)
	assert.Equal(t, testAppError.InnerError, capturedErrorData.InnerError)
}

func TestWrite_HandlesErrorResponseWhenContentTypeIsHtmlForNonHTMXRequest(t *testing.T) {
	t.Parallel()

	wantStatusCode := 418

	testAppError := &etc.AppError{
		Code:         wantStatusCode,
		ErrorMessage: "Some toast error",
	}

	ts := NewMockTemplateStore(t)

	r := httptest.NewRequest("GET", "/", strings.NewReader(""))

	w := httptest.NewRecorder()
	w.Header().Set("Content-Type", ContentTypeHtml.String())

	var capturedPageConfig templates.PageConfig
	ts.EXPECT().
		ExecutePageOutOfAppError(r.Context(), w, mock.Anything).
		Run(func(_ context.Context, _ io.Writer, c templates.PageConfig) {
			capturedPageConfig = c
		}).
		Once().
		Return(nil)

	js := NewMockJsonSerialiser(t)
	js.EXPECT().Marshal(mock.Anything).Maybe().Return([]byte("test"), nil)

	h := NewErrorHandler(ts, js)

	err := h.Write(w, r, testAppError)

	assert.Nil(t, err)
	assert.Equal(t, wantStatusCode, w.Code)

	assert.False(t, capturedPageConfig.ContentOnly)
}

func TestWrite_ShouldReturnErrorsReturnedFromExecuteForNonHtmxRequests(t *testing.T) {
	t.Parallel()

	wantError := errors.New("test error")

	w := httptest.NewRecorder()
	w.Header().Set("Content-Type", ContentTypeHtml.String())

	ts := NewMockTemplateStore(t)
	ts.EXPECT().
		ExecutePageOutOfAppError(mock.Anything, mock.Anything, mock.Anything).
		Return(wantError)

	js := NewMockJsonSerialiser(t)
	js.EXPECT().Marshal(mock.Anything).Maybe().Return([]byte("test"), nil)

	h := NewErrorHandler(ts, js)

	r := httptest.NewRequest("GET", "/", strings.NewReader(""))
	gotErr := h.Write(w, r, &etc.AppError{Code: 200})

	assert.EqualError(t, gotErr, wantError.Error())
}

func TestWrite_ShouldReturnErrorsReturnedFromExecuteForHtmxRequests(t *testing.T) {
	t.Parallel()

	wantError := errors.New("test error")

	w := httptest.NewRecorder()
	w.Header().Set("Content-Type", ContentTypeHtml.String())

	ts := NewMockTemplateStore(t)
	ts.EXPECT().
		ExecuteRenderAppError(mock.Anything, mock.Anything, mock.Anything).
		Return(wantError)

	js := NewMockJsonSerialiser(t)
	js.EXPECT().Marshal(mock.Anything).Maybe().Return([]byte("test"), nil)

	h := NewErrorHandler(ts, js)

	r := httptest.NewRequest("GET", "/", strings.NewReader(""))
	r.Header.Set(constants.HxRequestHeaderRequest, "true")

	gotErr := h.Write(w, r, &etc.AppError{Code: 200})

	assert.EqualError(t, gotErr, wantError.Error())
}
