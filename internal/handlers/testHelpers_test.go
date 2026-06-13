package handlers

import (
	"testing"

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

	testhelpers.AssertErrorEqual(t, got.InnerError, want.InnerError)

	if got.ToastError != want.ToastError {
		t.Fatalf("got toastError %s, but want toastError %s", got.ToastError, want.ToastError)
	}

	testhelpers.AssertStringSliceEqual(t, got.PageErrors, want.PageErrors)
}
