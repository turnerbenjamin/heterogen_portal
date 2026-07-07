package handlers

import (
	"errors"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/turnerbenjamin/heterogen_portal/internal/constants"
	"github.com/turnerbenjamin/heterogen_portal/internal/templates"
)

func TestWrite_HandlesErrorResponse(t *testing.T) {
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
	ts := NewMockTemplateStore(t)
	ts.EXPECT().
		Execute(mock.Anything, mock.Anything, mock.Anything).
		Return(wantError)

	h := NewErrorHandler(ts)

	r := httptest.NewRequest("GET", "/", strings.NewReader(""))
	gotErr := h.Write(w, r, &AppError{Code: 200})

	assert.EqualError(t, gotErr, wantError.Error())
}
