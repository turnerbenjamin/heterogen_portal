package handlers

import (
	"errors"
	"net/http"
	"testing"

	"github.com/turnerbenjamin/heterogen_portal/internal/constants"
	"github.com/turnerbenjamin/heterogen_portal/internal/testhelpers"
)

func TestNewServerError_ReturnsAppErrorWithCorrectProperties(t *testing.T) {
	t.Parallel()

	innerErrors := []error{
		errors.New("test_error_1"),
		errors.New("test_error_2"),
		nil,
	}

	for _, innerError := range innerErrors {
		wantErr := &AppError{
			Code:       http.StatusInternalServerError,
			ToastError: constants.ErrMsgInternalServerError,
			PageErrors: []string{constants.ErrMsgInternalServerError},
			innerError: innerError,
		}

		gotErr := NewServerError(innerError)

		AssertAppErrorEqual(t, gotErr, wantErr)
	}
}

func TestAppError_String_ReturnsCorrectString(t *testing.T) {
	t.Parallel()

	testData := []struct {
		appError *AppError
		want     string
	}{
		{
			appError: &AppError{
				Code:       418,
				ToastError: "test_toast_error",
				PageErrors: []string{"test_page_error"},
				innerError: errors.New("inner_error_present"),
			},
			want: "inner_error_present",
		},
		{
			appError: &AppError{
				Code:       418,
				ToastError: "test_toast_error",
				PageErrors: []string{"test_page_error"},
				innerError: nil,
			},
			want: "test_toast_error",
		},
		{
			appError: &AppError{
				Code:       418,
				ToastError: "",
				PageErrors: []string{"test_page_error"},
				innerError: nil,
			},
			want: "[test_page_error]",
		},
		{
			appError: &AppError{
				Code:       418,
				ToastError: "",
				PageErrors: []string{"test_page_error1", "test_page_error2"},
				innerError: nil,
			},
			want: "[test_page_error1,test_page_error2]",
		},
		{
			appError: &AppError{
				Code:       418,
				ToastError: "",
				PageErrors: []string{},
				innerError: nil,
			},
			want: constants.EmptyAppErrorString,
		},
	}

	for _, td := range testData {
		got := td.appError.String()
		testhelpers.AssertStringEqual(t, got, td.want)
	}
}
