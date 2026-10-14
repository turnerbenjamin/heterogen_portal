package etc

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turnerbenjamin/heterogen_portal/internal/constants"
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
			Code:         http.StatusInternalServerError,
			ErrorMessage: constants.ErrMsgInternalServerError,
			InnerError:   innerError,
		}

		gotErr := NewServerError(innerError)

		assert.EqualValues(t, gotErr, wantErr)
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
				Code:         418,
				ErrorMessage: "test_toast_error",
				InnerError:   errors.New("inner_error_present"),
			},
			want: "inner_error_present",
		},
		{
			appError: &AppError{
				Code:         418,
				ErrorMessage: "test_toast_error",
				InnerError:   nil,
			},
			want: "test_toast_error",
		},
		{
			appError: &AppError{
				Code:         418,
				ErrorMessage: "",
				InnerError:   nil,
			},
			want: constants.EmptyAppErrorString,
		},
	}

	for _, td := range testData {
		got := td.appError.String()
		assert.Equal(t, td.want, got)
	}
}
