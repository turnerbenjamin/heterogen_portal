package handlers

import (
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/turnerbenjamin/heterogen_portal/internal/etc"
	"github.com/turnerbenjamin/heterogen_portal/internal/templates"
	"github.com/turnerbenjamin/heterogen_portal/internal/testhelpers"
)

func TestWrite_HandlesErrorResponse(t *testing.T) {
	t.Parallel()

	wantExecuteCallCount := 1
	wantTemplate := templates.TmplComponentErrors
	wantStatusCode := 418
	wantContentOnlyValue := true

	testAppError := &etc.AppError{
		Code:       wantStatusCode,
		ToastError: "Some toast error",
		PageErrors: []string{"A page error", "and another"},
	}
	ts := &mockTemplateStore{t: t}

	w := httptest.NewRecorder()
	h := NewErrorHandler(ts)

	err := h.Write(w, testAppError)

	testhelpers.AssertErrorNil(t, err)
	testhelpers.AssertIntEqual(t, w.Code, wantStatusCode)
	testhelpers.AssertIntEqual(t, len(ts.calls), wantExecuteCallCount)

	gotExecuteCall := ts.calls[0]
	testhelpers.AssertEqual(t, gotExecuteCall.templateId, wantTemplate)
	testhelpers.AssertEqual(t, gotExecuteCall.data.PageConfig.ContentOnly, wantContentOnlyValue)
	testhelpers.AssertEqual(t, gotExecuteCall.data.Data, testAppError)
}

func TestWrite_ShouldReturnErrorsReturnedFromExecute(t *testing.T) {
	t.Parallel()

	wantError := errors.New("test error")
	ts := &mockTemplateStore{t: t, returns: wantError}

	w := httptest.NewRecorder()
	h := NewErrorHandler(ts)

	gotErr := h.Write(w, &etc.AppError{Code: 200})

	testhelpers.AssertErrorNotNil(t, gotErr, wantError)
	testhelpers.AssertErrorEqual(t, gotErr, wantError)
}
