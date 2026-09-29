package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/turnerbenjamin/heterogen_portal/internal/constants"
	"github.com/turnerbenjamin/querystack/queryerror"
	"github.com/turnerbenjamin/querystack/querymodel"
)

func TestProcessQuery_SendsTheCorrectQueryAndWritesTheResponse(t *testing.T) {
	t.Parallel()

	queryService := NewMockQueryService(t)
	jsonSerialiser := NewMockJsonSerialiser(t)

	resource := "test_resource"
	rawquery := "raw query"
	decodedQuery := "select=field1&expand=field2"

	decodeFunc := func(s string) (string, error) {
		assert.Equal(t, rawquery, s)
		return decodedQuery, nil
	}

	h := NewQueryApiHandler(queryService, decodeFunc, jsonSerialiser)
	r := httptest.NewRequest("GET", "/", strings.NewReader(""))

	r.SetPathValue(constants.UrlParamResource, resource)
	r.URL.RawQuery = rawquery

	w := httptest.NewRecorder()

	queryResponse := &querymodel.ExecuteResult{}
	queryService.
		EXPECT().
		ExecuteQuery(r.Context(), resource, decodedQuery).
		Return(queryResponse, nil)

	serialisedResponse := []byte("serialisedResponse")
	jsonSerialiser.
		EXPECT().
		Marshal(queryResponse).
		Return(serialisedResponse, nil)

	err := h.ProcessQuery(w, r, &PipelineContext[NoState]{})

	assert.Nil(t, err)
	assert.Equal(t, 200, w.Result().StatusCode)
	assert.Equal(t, string(serialisedResponse), w.Body.String())
}

func TestProcessQuery_HandlesErrorsFromDecodeUrl(t *testing.T) {
	t.Parallel()

	decodeUrlErr := errors.New("expected error")

	queryService := NewMockQueryService(t)
	jsonSerialiser := NewMockJsonSerialiser(t)

	decodeUrlFunc := newDecodeUrlFunc("", decodeUrlErr)

	h := NewQueryApiHandler(queryService, decodeUrlFunc, jsonSerialiser)
	r := httptest.NewRequest("GET", "/", strings.NewReader(""))
	w := httptest.NewRecorder()

	gotErr := h.ProcessQuery(w, r, &PipelineContext[NoState]{})
	wantErr := NewServerError(decodeUrlErr)

	assertAppErrorsEqual(t, wantErr, gotErr)
}

func TestProcessQuery_HandlesExecuteQueryErrors(t *testing.T) {
	t.Parallel()

	testData := []struct {
		executeQueryErr error
		isUserError     bool
	}{
		{
			executeQueryErr: queryerror.QueryError{
				Category: queryerror.QueryErrInternalErr,
				Err:      errors.New("an internal error"),
			},
			isUserError: false,
		},
		{
			executeQueryErr: queryerror.QueryError{
				Category: queryerror.QueryErrUnknown,
				Err:      errors.New("an unknown error type"),
			},
			isUserError: false,
		},
		{
			executeQueryErr: errors.New("not a query error"),
			isUserError:     false,
		},
		{
			executeQueryErr: queryerror.QueryError{
				Category: queryerror.QueryErrSyntaxErr,
				Err:      errors.New("a syntax error"),
			},
			isUserError: true,
		},
		{
			executeQueryErr: queryerror.QueryError{
				Category: queryerror.QueryErrBindingErr,
				Err:      errors.New("a binding error"),
			},
			isUserError: true,
		},
		{
			executeQueryErr: queryerror.QueryError{
				Category: queryerror.QueryErrAccessErr,
				Err:      errors.New("an access error"),
			},
			isUserError: true,
		},
		{
			executeQueryErr: queryerror.QueryError{
				Category: queryerror.QueryErrInvalidPagingTokenErr,
				Err:      errors.New("a paging token err"),
			},
			isUserError: true,
		},
	}

	for _, td := range testData {
		decodedUrl := "decoded_url"

		decodeUrlFunc := newDecodeUrlFunc(decodedUrl, nil)
		queryService := NewMockQueryService(t)
		jsonSerialiser := NewMockJsonSerialiser(t)

		h := NewQueryApiHandler(queryService, decodeUrlFunc, jsonSerialiser)
		r := httptest.NewRequest("GET", "/", strings.NewReader(""))
		w := httptest.NewRecorder()

		queryService.
			EXPECT().
			ExecuteQuery(mock.Anything, mock.Anything, mock.Anything).
			Return(nil, td.executeQueryErr)

		gotErr := h.ProcessQuery(w, r, &PipelineContext[NoState]{})

		var wantErr *AppError
		if td.isUserError {
			wantErr = &AppError{
				Code:       400,
				ToastError: td.executeQueryErr.Error(),
				innerError: td.executeQueryErr,
			}
		} else {
			wantErr = NewServerError(td.executeQueryErr)
		}

		assertAppErrorsEqual(t, wantErr, gotErr)
	}
}

func TestProcessQuery_HandlesErrorsFromJsonMarshal(t *testing.T) {
	t.Parallel()

	queryService := NewMockQueryService(t)
	jsonSerialiser := NewMockJsonSerialiser(t)

	decodeUrlFunc := newDecodeUrlFunc("query", nil)

	h := NewQueryApiHandler(queryService, decodeUrlFunc, jsonSerialiser)
	r := httptest.NewRequest("GET", "/", strings.NewReader(""))
	w := httptest.NewRecorder()

	queryService.
		EXPECT().
		ExecuteQuery(mock.Anything, mock.Anything, mock.Anything).
		Return(&querymodel.ExecuteResult{}, nil)

	expectedErr := errors.New("json marshal err")
	jsonSerialiser.
		EXPECT().
		Marshal(mock.Anything).
		Return(nil, expectedErr)

	gotErr := h.ProcessQuery(w, r, &PipelineContext[NoState]{})
	wantErr := NewServerError(expectedErr)

	assertAppErrorsEqual(t, wantErr, gotErr)
}

func TestProcessQuery_HandlesErrorsFromWrite(t *testing.T) {
	t.Parallel()

	queryService := NewMockQueryService(t)
	jsonSerialiser := NewMockJsonSerialiser(t)

	decodeUrlFunc := newDecodeUrlFunc("query", nil)

	h := NewQueryApiHandler(queryService, decodeUrlFunc, jsonSerialiser)
	r := httptest.NewRequest("GET", "/", strings.NewReader(""))

	expectedErr := errors.New("writer error")
	w := newMockWriter(expectedErr)

	queryService.
		EXPECT().
		ExecuteQuery(mock.Anything, mock.Anything, mock.Anything).
		Return(&querymodel.ExecuteResult{}, nil)

	jsonSerialiser.
		EXPECT().
		Marshal(mock.Anything).
		Return([]byte("response json"), nil)

	gotErr := h.ProcessQuery(w, r, &PipelineContext[NoState]{})
	wantErr := NewServerError(expectedErr)

	assertAppErrorsEqual(t, wantErr, gotErr)
}

func newDecodeUrlFunc(s string, err error) func(string) (string, error) {
	return func(_ string) (string, error) {
		return s, err
	}
}

func assertAppErrorsEqual(t testing.TB, expected *AppError, got *AppError) {
	t.Helper()

	if expected == nil {
		assert.Nil(t, got)
	} else {
		assert.NotNil(t, got)
	}

	assert.Equal(t, expected.Code, got.Code)
	assert.Equal(t, expected.ToastError, got.ToastError)
	assert.Equal(t, expected.innerError, got.innerError)
	assert.Equal(t, expected.PageErrors, got.PageErrors)
}

type mockResponseWriter struct {
	StatusCode int
	Body       []byte
	header     http.Header

	ErrOnWrite error
}

func newMockWriter(errorOnWrite error) http.ResponseWriter {
	return &mockResponseWriter{
		StatusCode: -1,
		Body:       make([]byte, 0),
		header:     make(http.Header),
		ErrOnWrite: errorOnWrite,
	}
}

func (w *mockResponseWriter) Header() http.Header {
	return w.header
}

func (w *mockResponseWriter) Write(b []byte) (int, error) {
	w.Body = b
	if w.ErrOnWrite != nil {
		return 0, w.ErrOnWrite
	}
	return len(b), nil
}

func (w *mockResponseWriter) WriteHeader(s int) {
	w.StatusCode = s
}
