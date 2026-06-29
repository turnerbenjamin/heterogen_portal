package handlers

import (
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/turnerbenjamin/heterogen_portal/internal/constants"
	"github.com/turnerbenjamin/heterogen_portal/internal/templates"
	"github.com/turnerbenjamin/heterogen_portal/internal/testhelpers"
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
		ts := &mockTemplateStore{t: t}

		r := httptest.NewRequest("GET", "/", strings.NewReader(""))
		if td.isHtmx {
			r.Header.Set(constants.HxRequestHeaderRequest, "true")
		}

		w := httptest.NewRecorder()
		h := NewErrorHandler(ts)

		err := h.Write(w, r, testAppError)

		testhelpers.AssertErrorNil(t, err)
		testhelpers.AssertIntEqual(t, w.Code, td.wantStatusCode)
		testhelpers.AssertIntEqual(t, len(ts.calls), td.wantExecuteCallCount)

		gotExecuteCall := ts.calls[0]
		testhelpers.AssertEqual(t, gotExecuteCall.templateId, td.wantTemplate)
		testhelpers.AssertEqual(t, gotExecuteCall.data.PageConfig.ContentOnly, td.wantContentOnlyValue)
		testhelpers.AssertEqual(t, gotExecuteCall.data.Data, testAppError)
	}
}

func TestWrite_ShouldReturnErrorsReturnedFromExecute(t *testing.T) {
	t.Parallel()

	wantError := errors.New("test error")
	ts := &mockTemplateStore{t: t, returns: wantError}

	w := httptest.NewRecorder()
	h := NewErrorHandler(ts)

	r := httptest.NewRequest("GET", "/", strings.NewReader(""))
	gotErr := h.Write(w, r, &AppError{Code: 200})

	testhelpers.AssertErrorNotNil(t, gotErr, wantError)
	testhelpers.AssertErrorEqual(t, gotErr, wantError)
}
