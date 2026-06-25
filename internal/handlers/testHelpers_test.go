package handlers

import (
	"io"
	"testing"

	"github.com/turnerbenjamin/heterogen_portal/internal/templates"
	"github.com/turnerbenjamin/heterogen_portal/internal/testhelpers"
)

func AssertAppErrorNil(t testing.TB, err *AppError) {
	t.Helper()

	if err != nil {
		t.Fatalf("got %s, but want nil", err.String())
	}
}

func AssertAppErrorEqual(t testing.TB, got, want *AppError) {
	t.Helper()

	if got == nil {
		t.Fatalf("got nil, but want %s", want.String())
		return
	}

	if got.Code != want.Code {
		t.Fatalf("got status code %d, but want %d", got.Code, want.Code)
	}

	if want.innerError != nil {
		testhelpers.AssertErrorEqual(t, got.innerError, want.innerError)
	}

	if got.ToastError != want.ToastError {
		t.Fatalf("got toastError %s, but want toastError %s", got.ToastError, want.ToastError)
	}

	testhelpers.AssertStringSliceEqual(t, got.PageErrors, want.PageErrors)
}

type mockTemplateStoreExecuteCallArgs struct {
	templateId templates.TemplateIdentifier
	writer     io.Writer
	data       templates.TemplateArgs
}

type mockTemplateStore struct {
	t       testing.TB
	returns error
	calls   []mockTemplateStoreExecuteCallArgs
}

func (is *mockTemplateStore) Execute(
	id templates.TemplateIdentifier,
	w io.Writer,
	data templates.TemplateArgs,
) error {
	is.t.Helper()
	if is.calls == nil {
		is.calls = []mockTemplateStoreExecuteCallArgs{}
	}

	is.calls = append(is.calls, mockTemplateStoreExecuteCallArgs{
		templateId: id,
		writer:     w,
		data:       data,
	})

	return is.returns
}
