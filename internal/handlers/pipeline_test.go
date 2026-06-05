package handlers

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/turnerbenjamin/heterogen_portal/internal/etc"
	"github.com/turnerbenjamin/heterogen_portal/testhelpers"
)

func TestPipelineBuilder_New_invokesHandler(t *testing.T) {
	wantBody := []byte("expected-body")
	wantStatusCode := 418

	middlewareStack := newTestMiddlewareStack(t, []testMiddleware[NoPipelineState]{{}})
	testHandler := &testAppHandler[NoPipelineState]{
		t: t,
		fn: func(
			r *http.Request,
			c *PipelineContext[NoPipelineState],
		) (request *http.Request, statusCode *int, response []byte, err *etc.AppError) {
			return r, &wantStatusCode, wantBody, nil
		},
	}

	b := NewPipelineBuilder[NoPipelineState](&testErrorHandler{}, &bytes.Buffer{})
	p := b.New(middlewareStack.stack, testHandler.handle)

	r := httptest.NewRequest("POST", "/test", strings.NewReader(""))
	w := httptest.NewRecorder()

	p.ServeHTTP(w, r)
	testhelpers.AssertBytesEqual(t, w.Body.Bytes(), wantBody)
	testhelpers.AssertIntEqual(t, w.Result().StatusCode, wantStatusCode)
}

func TestPipelineBuilder_New_defaultsStatusTo200_whenWriteWithoutHeader(t *testing.T) {
	middlewareStack := newTestMiddlewareStack(t, []testMiddleware[NoPipelineState]{{}})
	testHandler := &testAppHandler[NoPipelineState]{
		t: t,
		fn: func(
			r *http.Request,
			c *PipelineContext[NoPipelineState],
		) (request *http.Request, statusCode *int, response []byte, err *etc.AppError) {
			return r, nil, []byte("expected-body"), nil
		},
	}

	b := NewPipelineBuilder[NoPipelineState](&testErrorHandler{}, &bytes.Buffer{})
	p := b.New(middlewareStack.stack, testHandler.handle)

	r := httptest.NewRequest("POST", "/test", strings.NewReader(""))
	w := httptest.NewRecorder()

	p.ServeHTTP(w, r)

	wantStatusCode := 200
	gotStatusCode := w.Result().StatusCode
	testhelpers.AssertIntEqual(t, gotStatusCode, wantStatusCode)
}

func TestPipelineBuilder_New_appliesMiddlewaresCorrectly(t *testing.T) {
	middlewareCount := 5
	wantCalls := make([]string, 0, middlewareCount)
	gotCalls := make([]string, 0, middlewareCount)
	middlewares := make([]testMiddleware[NoPipelineState], 0, middlewareCount)

	for i := range middlewareCount {
		id := fmt.Sprintf("mw-id-%d", i)
		wantCalls = append(wantCalls, id)
		middlewares = append(middlewares, testMiddleware[NoPipelineState]{
			fn: func(
				r *http.Request,
				c *PipelineContext[NoPipelineState],
			) (request *http.Request, statusCode *int, response []byte, err *etc.AppError) {
				gotCalls = append(gotCalls, id)
				return r, nil, nil, nil
			},
		})
	}
	middlewareStack := newTestMiddlewareStack(t, middlewares)

	testHandler := &testAppHandler[NoPipelineState]{
		t: t,
		fn: func(
			r *http.Request,
			c *PipelineContext[NoPipelineState],
		) (request *http.Request, statusCode *int, response []byte, err *etc.AppError) {
			return r, nil, []byte("expected-body"), nil
		},
	}

	b := NewPipelineBuilder[NoPipelineState](&testErrorHandler{}, &bytes.Buffer{})
	p := b.New(middlewareStack.stack, testHandler.handle)

	r := httptest.NewRequest("POST", "/test", strings.NewReader(""))
	w := httptest.NewRecorder()

	p.ServeHTTP(w, r)

	testhelpers.AssertIntEqual(t, len(gotCalls), middlewareCount)
	for i, wantCall := range wantCalls {
		gotCall := gotCalls[i]
		testhelpers.AssertStringEqual(t, gotCall, wantCall)
	}
}

func TestPipelineBuilder_New_middlewareCanModifyRequest_andHandlerSeesChange(t *testing.T) {
	type mwKey string
	wantBodyKey := mwKey("want_body_key")
	wantBodyValue := []byte("want_body_value")

	middlewareStack := newTestMiddlewareStack(t, []testMiddleware[NoPipelineState]{{
		fn: func(
			r *http.Request,
			c *PipelineContext[NoPipelineState],
		) (request *http.Request, statusCode *int, response []byte, err *etc.AppError) {
			updatedRequest := r.WithContext(
				context.WithValue(r.Context(), wantBodyKey, wantBodyValue),
			)
			return updatedRequest, nil, nil, nil
		},
	},
	})

	testHandler := &testAppHandler[NoPipelineState]{
		t: t,
		fn: func(
			r *http.Request,
			c *PipelineContext[NoPipelineState],
		) (request *http.Request, statusCode *int, response []byte, err *etc.AppError) {
			gotWantBodyValue, ok := r.Context().Value(wantBodyKey).([]byte)
			if !ok {
				gotWantBodyValue = []byte("unexpected-body")
			}

			return r, nil, gotWantBodyValue, nil
		},
	}

	b := NewPipelineBuilder[NoPipelineState](&testErrorHandler{}, &bytes.Buffer{})
	p := b.New(middlewareStack.stack, testHandler.handle)

	r := httptest.NewRequest("POST", "/test", strings.NewReader(""))

	w := httptest.NewRecorder()

	p.ServeHTTP(w, r)

	testhelpers.AssertBytesEqual(t, w.Body.Bytes(), wantBodyValue)
}

func TestPipelineBuilder_New_middlewareCanModifyResponseBeforeHandler(t *testing.T) {
	wantBody := []byte("want_body")
	wantStatusCode := 400

	middlewareStack := newTestMiddlewareStack(t, []testMiddleware[NoPipelineState]{{
		fn: func(
			r *http.Request,
			c *PipelineContext[NoPipelineState],
		) (request *http.Request, statusCode *int, response []byte, err *etc.AppError) {
			return r, &wantStatusCode, wantBody, nil
		},
	},
	})

	handlerCallCount := 0
	testHandler := &testAppHandler[NoPipelineState]{
		t: t,
		fn: func(
			r *http.Request,
			c *PipelineContext[NoPipelineState],
		) (request *http.Request, statusCode *int, response []byte, err *etc.AppError) {
			handlerCallCount = handlerCallCount + 1
			sc := 200
			return r, &sc, []byte("unexpected_body_value"), nil
		},
	}

	b := NewPipelineBuilder[NoPipelineState](&testErrorHandler{}, &bytes.Buffer{})
	p := b.New(middlewareStack.stack, testHandler.handle)

	r := httptest.NewRequest("POST", "/test", strings.NewReader(""))

	w := httptest.NewRecorder()

	p.ServeHTTP(w, r)

	gotStatusCode := w.Result().StatusCode
	testhelpers.AssertIntEqual(t, gotStatusCode, wantStatusCode)
	testhelpers.AssertBytesEqual(t, w.Body.Bytes(), wantBody)
	testhelpers.AssertIntEqual(t, handlerCallCount, 0)
}

func TestPipelineBuilder_New_invokesHandlerDirectly_WhenNoMiddlewares(t *testing.T) {
	wantBody := []byte("expected-body")
	wantStatusCode := 418

	middlewareStack := newTestMiddlewareStack(t, []testMiddleware[NoPipelineState]{})
	testHandler := &testAppHandler[NoPipelineState]{
		t: t,
		fn: func(
			r *http.Request,
			c *PipelineContext[NoPipelineState],
		) (request *http.Request, statusCode *int, response []byte, err *etc.AppError) {
			return r, &wantStatusCode, wantBody, nil
		},
	}

	b := NewPipelineBuilder[NoPipelineState](&testErrorHandler{}, &bytes.Buffer{})
	p := b.New(middlewareStack.stack, testHandler.handle)

	r := httptest.NewRequest("POST", "/test", strings.NewReader(""))
	w := httptest.NewRecorder()

	p.ServeHTTP(w, r)
	testhelpers.AssertBytesEqual(t, w.Body.Bytes(), wantBody)
	testhelpers.AssertIntEqual(t, w.Result().StatusCode, wantStatusCode)
}

func TestPipelineBuilder_New_invokesHandlerDirectly_WhenMiddlewaresIsNil(t *testing.T) {
	wantBody := []byte("expected-body")
	wantStatusCode := 418

	testHandler := &testAppHandler[NoPipelineState]{
		t: t,
		fn: func(
			r *http.Request,
			c *PipelineContext[NoPipelineState],
		) (request *http.Request, statusCode *int, response []byte, err *etc.AppError) {
			return r, &wantStatusCode, wantBody, nil
		},
	}

	b := NewPipelineBuilder[NoPipelineState](&testErrorHandler{}, &bytes.Buffer{})
	p := b.New(nil, testHandler.handle)

	r := httptest.NewRequest("POST", "/test", strings.NewReader(""))
	w := httptest.NewRecorder()

	p.ServeHTTP(w, r)
	testhelpers.AssertBytesEqual(t, w.Body.Bytes(), wantBody)
	testhelpers.AssertIntEqual(t, w.Result().StatusCode, wantStatusCode)
}

func TestPipelineBuilder_New_initializesPipelineContext_withCorrectState(t *testing.T) {
	wantPipelineStateValue := 0
	middlewareStack := newTestMiddlewareStack(t, []testMiddleware[testPipelineState]{{}})

	gotPipelineStateValue := wantPipelineStateValue - 1
	testHandler := &testAppHandler[testPipelineState]{
		t: t,
		fn: func(
			r *http.Request,
			c *PipelineContext[testPipelineState],
		) (request *http.Request, statusCode *int, response []byte, err *etc.AppError) {
			gotPipelineStateValue = c.state.value
			return r, nil, nil, nil
		},
	}

	b := NewPipelineBuilder[testPipelineState](&testErrorHandler{}, &bytes.Buffer{})
	p := b.New(middlewareStack.stack, testHandler.handle)

	r := httptest.NewRequest("POST", "/test", strings.NewReader(""))
	w := httptest.NewRecorder()

	p.ServeHTTP(w, r)
	testhelpers.AssertIntEqual(t, gotPipelineStateValue, wantPipelineStateValue)
}

func TestPipelineBuilder_New_carriesPipelineContextState_throughMiddlewareChain(t *testing.T) {
	middlewareCount := 5
	middlewares := make([]testMiddleware[testPipelineState], 0, middlewareCount)

	wantStateValue := 0
	for i := range middlewareCount {
		wantStateValue = wantStateValue + (i + 1)
		middlewares = append(middlewares, testMiddleware[testPipelineState]{
			fn: func(
				r *http.Request,
				c *PipelineContext[testPipelineState],
			) (request *http.Request, statusCode *int, response []byte, err *etc.AppError) {
				c.state.value = c.state.value + (i + 1)
				return r, nil, nil, nil
			},
		})
	}
	middlewareStack := newTestMiddlewareStack(t, middlewares)

	gotStateValue := 0
	testHandler := &testAppHandler[testPipelineState]{
		t: t,
		fn: func(
			r *http.Request,
			c *PipelineContext[testPipelineState],
		) (request *http.Request, statusCode *int, response []byte, err *etc.AppError) {
			gotStateValue = c.state.value
			return r, nil, []byte("expected-body"), nil
		},
	}

	b := NewPipelineBuilder[testPipelineState](&testErrorHandler{}, &bytes.Buffer{})
	p := b.New(middlewareStack.stack, testHandler.handle)

	r := httptest.NewRequest("POST", "/test", strings.NewReader(""))
	w := httptest.NewRecorder()

	p.ServeHTTP(w, r)

	testhelpers.AssertIntEqual(t, gotStateValue, wantStateValue)
}

func TestPipelineBuilder_New_invokesErrorHandler_whenHandlerReturnsError(t *testing.T) {
	testAppError := &etc.AppError{
		Code:       403,
		ToastError: "expected_error",
	}

	middlewareStack := newTestMiddlewareStack(t, []testMiddleware[NoPipelineState]{{}})
	testHandler := &testAppHandler[NoPipelineState]{
		t: t,
		fn: func(
			r *http.Request,
			c *PipelineContext[NoPipelineState],
		) (request *http.Request, statusCode *int, response []byte, err *etc.AppError) {
			return r, nil, []byte("expected_body"), testAppError
		},
	}

	errorHandler := &testErrorHandler{}
	b := NewPipelineBuilder[NoPipelineState](errorHandler, &bytes.Buffer{})
	p := b.New(middlewareStack.stack, testHandler.handle)

	r := httptest.NewRequest("POST", "/test", strings.NewReader(""))
	w := httptest.NewRecorder()

	p.ServeHTTP(w, r)
	testhelpers.AssertBytesEqual(t, w.Body.Bytes(), []byte(testAppError.ToastError))
	testhelpers.AssertIntEqual(t, w.Result().StatusCode, testAppError.Code)
}

func TestPipelineBuilder_New_invokesErrorHandler_whenMiddlewareReturnsError(t *testing.T) {
	testAppError := &etc.AppError{
		Code:       405,
		ToastError: "expected_error",
	}

	middlewareCount := 10
	middlewares := make([]testMiddleware[NoPipelineState], 0, middlewareCount)

	failingMiddlewareIndex := 3
	gotMiddlewareCalls := 0
	for i := range middlewareCount {
		middlewares = append(middlewares, testMiddleware[NoPipelineState]{
			fn: func(
				r *http.Request,
				c *PipelineContext[NoPipelineState],
			) (request *http.Request, statusCode *int, response []byte, err *etc.AppError) {
				gotMiddlewareCalls = gotMiddlewareCalls + 1
				if i == failingMiddlewareIndex {
					return r, nil, nil, testAppError
				}
				return r, nil, nil, nil
			},
		})
	}
	middlewareStack := newTestMiddlewareStack(t, middlewares)

	gotHandlerCalls := 0
	testHandler := &testAppHandler[NoPipelineState]{
		t: t,
		fn: func(
			r *http.Request,
			c *PipelineContext[NoPipelineState],
		) (request *http.Request, statusCode *int, response []byte, err *etc.AppError) {
			gotHandlerCalls = gotHandlerCalls + 1
			return r, nil, []byte("expected_body"), nil
		},
	}

	errorHandler := &testErrorHandler{}
	b := NewPipelineBuilder[NoPipelineState](errorHandler, &bytes.Buffer{})
	p := b.New(middlewareStack.stack, testHandler.handle)

	r := httptest.NewRequest("POST", "/test", strings.NewReader(""))
	w := httptest.NewRecorder()

	p.ServeHTTP(w, r)
	testhelpers.AssertIntEqual(t, gotMiddlewareCalls, failingMiddlewareIndex+1)
	testhelpers.AssertIntEqual(t, gotHandlerCalls, 0)
	testhelpers.AssertIntEqual(t, w.Result().StatusCode, testAppError.Code)
	testhelpers.AssertBytesEqual(t, w.Body.Bytes(), []byte(testAppError.ToastError))
}

func TestPipelineBuilder_New_handlesErrorWriterReturningError(t *testing.T) {
	testAppError := &etc.AppError{
		Code:       402,
		ToastError: "some_error",
	}

	middlewareStack := newTestMiddlewareStack(t, []testMiddleware[NoPipelineState]{{}})
	testHandler := &testAppHandler[NoPipelineState]{
		t: t,
		fn: func(
			r *http.Request,
			c *PipelineContext[NoPipelineState],
		) (request *http.Request, statusCode *int, response []byte, err *etc.AppError) {
			return r, nil, []byte("expected_body"), testAppError
		},
	}

	writeError := errors.New("this is a test error message")
	errorHandler := &testErrorHandler{Err: writeError}
	logSink := &bytes.Buffer{}

	b := NewPipelineBuilder[NoPipelineState](errorHandler, logSink)
	p := b.New(middlewareStack.stack, testHandler.handle)

	r := httptest.NewRequest("POST", "/test", strings.NewReader(""))
	w := httptest.NewRecorder()

	p.ServeHTTP(w, r)

	testhelpers.AssertSlogsContain(
		t,
		logSink.Bytes(),
		map[string]any{
			"writer_error":   writeError.Error(),
			"response_error": testAppError.Error(),
		},
	)
	testhelpers.AssertIntEqual(t, w.Result().StatusCode, 500)
}

func TestPipelineBuilder_New_stopsExecutionChain_whenMiddlewareReturnsError(t *testing.T) {
	testAppError := &etc.AppError{
		Code:       405,
		ToastError: "expected_error",
	}

	middlewareStack := newTestMiddlewareStack(t, []testMiddleware[NoPipelineState]{{
		fn: func(
			r *http.Request,
			c *PipelineContext[NoPipelineState],
		) (request *http.Request, statusCode *int, response []byte, err *etc.AppError) {
			return r, nil, nil, testAppError
		},
	}})

	testHandler := &testAppHandler[NoPipelineState]{
		t: t,
		fn: func(
			r *http.Request,
			c *PipelineContext[NoPipelineState],
		) (request *http.Request, statusCode *int, response []byte, err *etc.AppError) {
			return r, nil, []byte("expected_body"), nil
		},
	}

	errorHandler := &testErrorHandler{}
	b := NewPipelineBuilder[NoPipelineState](errorHandler, &bytes.Buffer{})
	p := b.New(middlewareStack.stack, testHandler.handle)

	r := httptest.NewRequest("POST", "/test", strings.NewReader(""))
	w := httptest.NewRecorder()

	p.ServeHTTP(w, r)
	testhelpers.AssertBytesEqual(t, w.Body.Bytes(), []byte(testAppError.ToastError))
	testhelpers.AssertIntEqual(t, w.Result().StatusCode, testAppError.Code)
}

func TestPipelineBuilder_New_logsResponseErrorWhenErrorWriterSucceeds(t *testing.T) {
	testAppError := &etc.AppError{
		Code:       402,
		ToastError: "some_error",
	}

	middlewareStack := newTestMiddlewareStack(t, []testMiddleware[NoPipelineState]{{}})
	testHandler := &testAppHandler[NoPipelineState]{
		t: t,
		fn: func(
			r *http.Request,
			c *PipelineContext[NoPipelineState],
		) (request *http.Request, statusCode *int, response []byte, err *etc.AppError) {
			return r, nil, []byte("expected_body"), testAppError
		},
	}

	logSink := &bytes.Buffer{}
	b := NewPipelineBuilder[NoPipelineState](&testErrorHandler{}, logSink)
	p := b.New(middlewareStack.stack, testHandler.handle)

	r := httptest.NewRequest("POST", "/test", strings.NewReader(""))
	w := httptest.NewRecorder()

	p.ServeHTTP(w, r)

	testhelpers.AssertSlogsContain(
		t,
		logSink.Bytes(),
		map[string]any{
			"response_error": testAppError.Error(),
		},
	)
}

func TestPipelineBuilder_New_includesRequestDataInLogs(t *testing.T) {
	wantStatusCode := 201
	middlewareStack := newTestMiddlewareStack(t, []testMiddleware[NoPipelineState]{{}})
	testHandler := &testAppHandler[NoPipelineState]{
		t: t,
		fn: func(
			r *http.Request,
			c *PipelineContext[NoPipelineState],
		) (request *http.Request, statusCode *int, response []byte, err *etc.AppError) {
			return r, &wantStatusCode, []byte("expected_body"), nil
		},
	}

	logSink := &bytes.Buffer{}
	b := NewPipelineBuilder[NoPipelineState](&testErrorHandler{}, logSink)
	p := b.New(middlewareStack.stack, testHandler.handle)

	r := httptest.NewRequest("POST", "/test", strings.NewReader(""))
	w := httptest.NewRecorder()

	p.ServeHTTP(w, r)

	testhelpers.AssertSlogsContain(
		t,
		logSink.Bytes(),
		map[string]any{
			"request_method":       r.Method,
			"request_path":         r.URL.Path,
			"request_time":         testhelpers.AnySlogValue,
			"request_duration_ms":  testhelpers.AnySlogValue,
			"response_status_code": wantStatusCode,
		},
	)
}

func TestPipelineBuilder_New_RecoversFromHandlerPanic(t *testing.T) {
	middlewareStack := newTestMiddlewareStack(t, []testMiddleware[NoPipelineState]{{}})
	testHandler := &testAppHandler[NoPipelineState]{
		t: t,
		fn: func(
			r *http.Request,
			c *PipelineContext[NoPipelineState],
		) (request *http.Request, statusCode *int, response []byte, err *etc.AppError) {
			panic("handler panic")
		},
	}

	b := NewPipelineBuilder[NoPipelineState](&testErrorHandler{}, &bytes.Buffer{})
	p := b.New(middlewareStack.stack, testHandler.handle)

	r := httptest.NewRequest("POST", "/test", strings.NewReader(""))
	w := httptest.NewRecorder()

	p.ServeHTTP(w, r)

	wantStatusCode := 500
	gotStatusCode := w.Result().StatusCode
	testhelpers.AssertIntEqual(t, gotStatusCode, wantStatusCode)
}

func TestPipelineBuilder_New_RecoversFromMiddlewarePanic(t *testing.T) {
	middlewareStack := newTestMiddlewareStack(t, []testMiddleware[NoPipelineState]{{
		fn: func(
			r *http.Request,
			c *PipelineContext[NoPipelineState],
		) (request *http.Request, statusCode *int, response []byte, err *etc.AppError) {
			panic("middleware panic")
		},
	}})

	testHandler := &testAppHandler[NoPipelineState]{
		t: t,
		fn: func(
			r *http.Request,
			c *PipelineContext[NoPipelineState],
		) (request *http.Request, statusCode *int, response []byte, err *etc.AppError) {
			return r, nil, nil, nil
		},
	}

	b := NewPipelineBuilder[NoPipelineState](&testErrorHandler{}, &bytes.Buffer{})
	p := b.New(middlewareStack.stack, testHandler.handle)

	r := httptest.NewRequest("POST", "/test", strings.NewReader(""))
	w := httptest.NewRecorder()

	p.ServeHTTP(w, r)

	wantStatusCode := 500
	gotStatusCode := w.Result().StatusCode
	testhelpers.AssertIntEqual(t, gotStatusCode, wantStatusCode)
}

type testPipelineState struct {
	value int
}

type testErrorHandler struct {
	Err error
}

func (eh *testErrorHandler) Write(w http.ResponseWriter, e *etc.AppError) error {
	if eh.Err != nil {
		return eh.Err
	}
	w.WriteHeader(e.Code)
	_, err := w.Write([]byte(e.ToastError))
	return err
}

type testAppHandler[T any] struct {
	t  testing.TB
	fn func(r *http.Request, c *PipelineContext[T]) (request *http.Request, statusCode *int, response []byte, err *etc.AppError)
}

func (h *testAppHandler[T]) handle(w http.ResponseWriter, r *http.Request, c *PipelineContext[T]) *etc.AppError {
	h.t.Helper()

	_, statusCode, res, appErr := h.fn(r, c)
	if appErr != nil {
		return appErr
	}

	if statusCode != nil {
		w.WriteHeader(*statusCode)
	}

	if res == nil {
		res = []byte("default_body")
	}
	_, err := w.Write(res)
	if err != nil {
		h.t.Fatalf("failed to write response: %s", err.Error())
	}
	return nil
}

type testMiddleware[T any] struct {
	t  testing.TB
	fn func(r *http.Request, c *PipelineContext[T]) (request *http.Request, statusCode *int, response []byte, err *etc.AppError)
}

func (m *testMiddleware[T]) handle(h AppHandler[T]) AppHandler[T] {
	return func(w http.ResponseWriter, r *http.Request, c *PipelineContext[T]) *etc.AppError {
		m.t.Helper()

		if m.fn == nil {
			return h(w, r, c)
		}

		request, statusCode, response, err := m.fn(r, c)

		if err != nil {
			return err
		}
		r = request

		if statusCode != nil {
			w.WriteHeader(*statusCode)
		}

		if response != nil {
			_, err := w.Write(response)
			if err != nil {
				m.t.Fatal(err.Error())
			}
			return nil
		}

		return h(w, r, c)
	}
}

type testMiddlewareStack[T any] struct {
	stack  []Middleware[T]
	calls  []string
	errors []etc.AppError
	count  int
}

func newTestMiddlewareStack[T any](t testing.TB, data []testMiddleware[T]) *testMiddlewareStack[T] {
	t.Helper()

	c := len(data)
	s := &testMiddlewareStack[T]{
		stack:  make([]Middleware[T], c),
		calls:  make([]string, 0),
		errors: make([]etc.AppError, c),
		count:  c,
	}

	for i, mw := range data {
		mw.t = t
		s.stack[i] = mw.handle
	}
	return s
}
