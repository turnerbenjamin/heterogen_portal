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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/turnerbenjamin/heterogen_portal/internal/constants"
)

func TestPipelineBuilder_New_invokesHandler(t *testing.T) {
	t.Parallel()

	wantBody := []byte("expected-body")
	wantStatusCode := 418

	middlewareStack := newTestMiddlewareStack(t, []testMiddleware[NoState]{{}})
	testHandler := &testAppHandler[NoState]{
		t: t,
		fn: func(
			r *http.Request,
			c *PipelineContext[NoState],
		) (request *http.Request, statusCode *int, response []byte, err *AppError) {
			return r, &wantStatusCode, wantBody, nil
		},
	}

	eh := NewMockErrorWriter(t)
	b := NewPipelineBuilder(eh, &bytes.Buffer{}, NoStateInit)
	p := b.New(middlewareStack.stack, testHandler.handle)

	r := httptest.NewRequest("POST", "/test", strings.NewReader(""))
	w := httptest.NewRecorder()

	p.ServeHTTP(w, r)
	assert.Equal(t, wantStatusCode, w.Result().StatusCode)
	assert.EqualValues(t, string(wantBody), w.Body.String())
}

func TestPipelineBuilder_New_defaultsStatusTo200_whenWriteWithoutHeader(t *testing.T) {
	t.Parallel()

	middlewareStack := newTestMiddlewareStack(t, []testMiddleware[NoState]{{}})
	testHandler := &testAppHandler[NoState]{
		t: t,
		fn: func(
			r *http.Request,
			c *PipelineContext[NoState],
		) (request *http.Request, statusCode *int, response []byte, err *AppError) {
			return r, nil, []byte("expected-body"), nil
		},
	}
	eh := NewMockErrorWriter(t)
	b := NewPipelineBuilder(eh, &bytes.Buffer{}, NoStateInit)
	p := b.New(middlewareStack.stack, testHandler.handle)

	r := httptest.NewRequest("POST", "/test", strings.NewReader(""))
	w := httptest.NewRecorder()

	p.ServeHTTP(w, r)

	wantStatusCode := 200
	gotStatusCode := w.Result().StatusCode
	assert.Equal(t, wantStatusCode, gotStatusCode)
}

func TestPipelineBuilder_New_appliesMiddlewaresCorrectly(t *testing.T) {
	t.Parallel()

	middlewareCount := 5
	wantCalls := make([]string, 0, middlewareCount)
	gotCalls := make([]string, 0, middlewareCount)
	middlewares := make([]testMiddleware[NoState], 0, middlewareCount)

	for i := range middlewareCount {
		id := fmt.Sprintf("mw-id-%d", i)
		wantCalls = append(wantCalls, id)
		middlewares = append(middlewares, testMiddleware[NoState]{
			fn: func(
				r *http.Request,
				c *PipelineContext[NoState],
			) (request *http.Request, statusCode *int, response []byte, err *AppError) {
				gotCalls = append(gotCalls, id)
				return r, nil, nil, nil
			},
		})
	}
	middlewareStack := newTestMiddlewareStack(t, middlewares)

	testHandler := &testAppHandler[NoState]{
		t: t,
		fn: func(
			r *http.Request,
			c *PipelineContext[NoState],
		) (request *http.Request, statusCode *int, response []byte, err *AppError) {
			return r, nil, []byte("expected-body"), nil
		},
	}

	b := NewPipelineBuilder(NewMockErrorWriter(t), &bytes.Buffer{}, NoStateInit)
	p := b.New(middlewareStack.stack, testHandler.handle)

	r := httptest.NewRequest("POST", "/test", strings.NewReader(""))
	w := httptest.NewRecorder()

	p.ServeHTTP(w, r)

	assert.Equal(t, middlewareCount, len(gotCalls))
	for i, wantCall := range wantCalls {
		gotCall := gotCalls[i]
		assert.Equal(t, wantCall, gotCall)
	}
}

func TestPipelineBuilder_New_middlewareCanModifyRequest_andHandlerSeesChange(t *testing.T) {
	t.Parallel()

	type mwKey string
	wantBodyKey := mwKey("want_body_key")
	wantBodyValue := []byte("want_body_value")

	middlewareStack := newTestMiddlewareStack(t, []testMiddleware[NoState]{{
		fn: func(
			r *http.Request,
			c *PipelineContext[NoState],
		) (request *http.Request, statusCode *int, response []byte, err *AppError) {
			updatedRequest := r.WithContext(
				context.WithValue(r.Context(), wantBodyKey, wantBodyValue),
			)
			return updatedRequest, nil, nil, nil
		},
	},
	})

	testHandler := &testAppHandler[NoState]{
		t: t,
		fn: func(
			r *http.Request,
			c *PipelineContext[NoState],
		) (request *http.Request, statusCode *int, response []byte, err *AppError) {
			gotWantBodyValue, ok := r.Context().Value(wantBodyKey).([]byte)
			if !ok {
				gotWantBodyValue = []byte("unexpected-body")
			}

			return r, nil, gotWantBodyValue, nil
		},
	}

	b := NewPipelineBuilder(NewMockErrorWriter(t), &bytes.Buffer{}, NoStateInit)
	p := b.New(middlewareStack.stack, testHandler.handle)

	r := httptest.NewRequest("POST", "/test", strings.NewReader(""))
	w := httptest.NewRecorder()

	p.ServeHTTP(w, r)

	assert.Equal(t, string(wantBodyValue), w.Body.String())
}

func TestPipelineBuilder_New_middlewareCanModifyResponseBeforeHandler(t *testing.T) {
	t.Parallel()

	wantBody := []byte("want_body")
	wantStatusCode := 400

	middlewareStack := newTestMiddlewareStack(t, []testMiddleware[NoState]{{
		fn: func(
			r *http.Request,
			c *PipelineContext[NoState],
		) (request *http.Request, statusCode *int, response []byte, err *AppError) {
			return r, &wantStatusCode, wantBody, nil
		},
	},
	})

	handlerCallCount := 0
	testHandler := &testAppHandler[NoState]{
		t: t,
		fn: func(
			r *http.Request,
			c *PipelineContext[NoState],
		) (request *http.Request, statusCode *int, response []byte, err *AppError) {
			handlerCallCount = handlerCallCount + 1
			sc := 200
			return r, &sc, []byte("unexpected_body_value"), nil
		},
	}

	b := NewPipelineBuilder(NewMockErrorWriter(t), &bytes.Buffer{}, NoStateInit)
	p := b.New(middlewareStack.stack, testHandler.handle)

	r := httptest.NewRequest("POST", "/test", strings.NewReader(""))

	w := httptest.NewRecorder()

	p.ServeHTTP(w, r)

	gotStatusCode := w.Result().StatusCode
	assert.Equal(t, wantStatusCode, gotStatusCode)
	assert.Equal(t, string(wantBody), w.Body.String())
	assert.Equal(t, 0, handlerCallCount)
}

func TestPipelineBuilder_New_invokesHandlerDirectly_WhenNoMiddlewares(t *testing.T) {
	t.Parallel()

	wantBody := []byte("expected-body")
	wantStatusCode := 418

	middlewareStack := newTestMiddlewareStack(t, []testMiddleware[NoState]{})
	testHandler := &testAppHandler[NoState]{
		t: t,
		fn: func(
			r *http.Request,
			c *PipelineContext[NoState],
		) (request *http.Request, statusCode *int, response []byte, err *AppError) {
			return r, &wantStatusCode, wantBody, nil
		},
	}

	b := NewPipelineBuilder(NewMockErrorWriter(t), &bytes.Buffer{}, NoStateInit)
	p := b.New(middlewareStack.stack, testHandler.handle)

	r := httptest.NewRequest("POST", "/test", strings.NewReader(""))
	w := httptest.NewRecorder()

	p.ServeHTTP(w, r)
	assert.Equal(t, string(wantBody), w.Body.String())
	assert.Equal(t, wantStatusCode, w.Result().StatusCode)
}

func TestPipelineBuilder_New_invokesHandlerDirectly_WhenMiddlewaresIsNil(t *testing.T) {
	t.Parallel()

	wantBody := []byte("expected-body")
	wantStatusCode := 418

	testHandler := &testAppHandler[NoState]{
		t: t,
		fn: func(
			r *http.Request,
			c *PipelineContext[NoState],
		) (request *http.Request, statusCode *int, response []byte, err *AppError) {
			return r, &wantStatusCode, wantBody, nil
		},
	}

	b := NewPipelineBuilder(NewMockErrorWriter(t), &bytes.Buffer{}, NoStateInit)
	p := b.New(nil, testHandler.handle)

	r := httptest.NewRequest("POST", "/test", strings.NewReader(""))
	w := httptest.NewRecorder()

	p.ServeHTTP(w, r)
	assert.Equal(t, string(wantBody), w.Body.String())
	assert.Equal(t, wantStatusCode, w.Result().StatusCode)
}

func TestPipelineBuilder_New_initializesPipelineContext_withCorrectState(t *testing.T) {
	t.Parallel()

	wantPipelineStateValue := 0
	middlewareStack := newTestMiddlewareStack(t, []testMiddleware[testPipelineStateInterface]{{}})

	gotPipelineStateValue := wantPipelineStateValue - 1
	testHandler := &testAppHandler[testPipelineStateInterface]{
		t: t,
		fn: func(
			r *http.Request,
			c *PipelineContext[testPipelineStateInterface],
		) (request *http.Request, statusCode *int, response []byte, err *AppError) {
			gotPipelineStateValue = c.state.getValue()
			return r, nil, nil, nil
		},
	}

	b := NewPipelineBuilder(NewMockErrorWriter(t), &bytes.Buffer{}, newTestPipelineState)
	p := b.New(middlewareStack.stack, testHandler.handle)

	r := httptest.NewRequest("POST", "/test", strings.NewReader(""))
	w := httptest.NewRecorder()

	p.ServeHTTP(w, r)
	assert.Equal(t, wantPipelineStateValue, gotPipelineStateValue)
}

func TestPipelineBuilder_New_carriesPipelineContextState_throughMiddlewareChain(t *testing.T) {
	t.Parallel()

	middlewareCount := 5
	middlewares := make([]testMiddleware[testPipelineStateInterface], 0, middlewareCount)

	wantStateValue := 0
	for i := range middlewareCount {
		wantStateValue = wantStateValue + (i + 1)
		middlewares = append(middlewares, testMiddleware[testPipelineStateInterface]{
			fn: func(
				r *http.Request,
				c *PipelineContext[testPipelineStateInterface],
			) (request *http.Request, statusCode *int, response []byte, err *AppError) {
				c.state.setValue(c.state.getValue() + (i + 1))
				return r, nil, nil, nil
			},
		})
	}
	middlewareStack := newTestMiddlewareStack(t, middlewares)

	gotStateValue := 0
	testHandler := &testAppHandler[testPipelineStateInterface]{
		t: t,
		fn: func(
			r *http.Request,
			c *PipelineContext[testPipelineStateInterface],
		) (request *http.Request, statusCode *int, response []byte, err *AppError) {
			gotStateValue = c.state.getValue()
			return r, nil, []byte("expected-body"), nil
		},
	}

	b := NewPipelineBuilder(NewMockErrorWriter(t), &bytes.Buffer{}, newTestPipelineState)
	p := b.New(middlewareStack.stack, testHandler.handle)

	r := httptest.NewRequest("POST", "/test", strings.NewReader(""))
	w := httptest.NewRecorder()

	p.ServeHTTP(w, r)

	assert.Equal(t, wantStateValue, gotStateValue)
}

func TestPipelineBuilder_New_invokesErrorHandler_whenHandlerReturnsError(t *testing.T) {
	t.Parallel()

	testAppError := &AppError{
		Code:       403,
		ToastError: "expected_error",
	}

	middlewareStack := newTestMiddlewareStack(t, []testMiddleware[NoState]{{}})
	testHandler := &testAppHandler[NoState]{
		t: t,
		fn: func(
			r *http.Request,
			c *PipelineContext[NoState],
		) (request *http.Request, statusCode *int, response []byte, err *AppError) {
			return r, nil, []byte("expected_body"), testAppError
		},
	}

	errorHandler := NewMockErrorWriter(t)

	b := NewPipelineBuilder(errorHandler, &bytes.Buffer{}, NoStateInit)
	p := b.New(middlewareStack.stack, testHandler.handle)

	r := httptest.NewRequest("POST", "/test", strings.NewReader(""))
	w := httptest.NewRecorder()

	errorHandler.EXPECT().Write(mock.Anything, r, testAppError).Return(nil)
	p.ServeHTTP(w, r)
}

func TestPipelineBuilder_New_invokesErrorHandler_whenMiddlewareReturnsError(t *testing.T) {
	t.Parallel()

	testAppError := &AppError{
		Code:       405,
		ToastError: "expected_error",
	}

	middlewareCount := 10
	middlewares := make([]testMiddleware[NoState], 0, middlewareCount)

	failingMiddlewareIndex := 3
	gotMiddlewareCalls := 0
	for i := range middlewareCount {
		middlewares = append(middlewares, testMiddleware[NoState]{
			fn: func(
				r *http.Request,
				c *PipelineContext[NoState],
			) (request *http.Request, statusCode *int, response []byte, err *AppError) {
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
	testHandler := &testAppHandler[NoState]{
		t: t,
		fn: func(
			r *http.Request,
			c *PipelineContext[NoState],
		) (request *http.Request, statusCode *int, response []byte, err *AppError) {
			gotHandlerCalls = gotHandlerCalls + 1
			return r, nil, []byte("expected_body"), nil
		},
	}

	errorHandler := NewMockErrorWriter(t)

	b := NewPipelineBuilder(errorHandler, &bytes.Buffer{}, NoStateInit)
	p := b.New(middlewareStack.stack, testHandler.handle)

	r := httptest.NewRequest("POST", "/test", strings.NewReader(""))
	w := httptest.NewRecorder()

	errorHandler.EXPECT().Write(mock.Anything, r, testAppError).Return(nil)

	p.ServeHTTP(w, r)

	assert.Equal(t, failingMiddlewareIndex+1, gotMiddlewareCalls)
	assert.Equal(t, 0, gotHandlerCalls)
}

func TestPipelineBuilder_New_handlesErrorWriterReturningError(t *testing.T) {
	t.Parallel()

	testAppError := &AppError{
		Code:       402,
		ToastError: "some_error",
	}

	middlewareStack := newTestMiddlewareStack(t, []testMiddleware[NoState]{{}})
	testHandler := &testAppHandler[NoState]{
		t: t,
		fn: func(
			r *http.Request,
			c *PipelineContext[NoState],
		) (request *http.Request, statusCode *int, response []byte, err *AppError) {
			return r, nil, []byte("expected_body"), testAppError
		},
	}

	writeError := errors.New("this is a test error message")
	errorHandler := NewMockErrorWriter(t)

	logSink := &bytes.Buffer{}

	b := NewPipelineBuilder(errorHandler, logSink, NoStateInit)
	p := b.New(middlewareStack.stack, testHandler.handle)

	r := httptest.NewRequest("POST", "/test", strings.NewReader(""))
	w := httptest.NewRecorder()

	errorHandler.EXPECT().Write(mock.Anything, r, testAppError).Return(writeError)

	p.ServeHTTP(w, r)

	assertLogsContain(
		t,
		logSink,
		map[string]any{
			constants.SlogKeyResponseWriterErr: writeError.Error(),
			constants.SlogKeyRequestErr:        testAppError.String(),
		},
	)
	assert.Equal(t, http.StatusInternalServerError, w.Result().StatusCode)
}

func TestPipelineBuilder_New_logsResponseErrorWhenErrorWriterSucceeds(t *testing.T) {
	t.Parallel()

	testAppError := &AppError{
		Code:       402,
		ToastError: "some_error",
	}

	middlewareStack := newTestMiddlewareStack(t, []testMiddleware[NoState]{{}})
	testHandler := &testAppHandler[NoState]{
		t: t,
		fn: func(
			r *http.Request,
			c *PipelineContext[NoState],
		) (request *http.Request, statusCode *int, response []byte, err *AppError) {
			return r, nil, []byte("expected_body"), testAppError
		},
	}

	logSink := &bytes.Buffer{}
	eh := NewMockErrorWriter(t)

	b := NewPipelineBuilder(eh, logSink, NoStateInit)
	p := b.New(middlewareStack.stack, testHandler.handle)

	r := httptest.NewRequest("POST", "/test", strings.NewReader(""))
	w := httptest.NewRecorder()

	eh.EXPECT().Write(mock.Anything, r, testAppError).Return(nil)

	p.ServeHTTP(w, r)

	assertLogsContain(
		t,
		logSink,
		map[string]any{
			constants.SlogKeyRequestErr: testAppError.String(),
		},
	)
}

func TestPipelineBuilder_New_includesRequestDataInLogs(t *testing.T) {
	t.Parallel()

	wantStatusCode := 201
	middlewareStack := newTestMiddlewareStack(t, []testMiddleware[NoState]{{}})
	testHandler := &testAppHandler[NoState]{
		t: t,
		fn: func(
			r *http.Request,
			c *PipelineContext[NoState],
		) (request *http.Request, statusCode *int, response []byte, err *AppError) {
			return r, &wantStatusCode, []byte("expected_body"), nil
		},
	}

	logSink := &bytes.Buffer{}
	b := NewPipelineBuilder(NewMockErrorWriter(t), logSink, NoStateInit)
	p := b.New(middlewareStack.stack, testHandler.handle)

	r := httptest.NewRequest("POST", "/test", strings.NewReader(""))
	w := httptest.NewRecorder()

	p.ServeHTTP(w, r)

	assertLogsContain(
		t,
		logSink,
		map[string]any{
			constants.SlogKeyRequestMethod:      r.Method,
			constants.SlogKeyRequestPath:        r.URL.Path,
			constants.SlogKeyRequestTime:        mock.Anything,
			constants.SlogKeyRequestDurationMs:  mock.Anything,
			constants.SlogKeyResponseStatusCode: float64(wantStatusCode),
		},
	)
}

func TestPipelineBuilder_New_RecoversFromHandlerPanic(t *testing.T) {
	t.Parallel()

	middlewareStack := newTestMiddlewareStack(t, []testMiddleware[NoState]{{}})
	testHandler := &testAppHandler[NoState]{
		t: t,
		fn: func(
			r *http.Request,
			c *PipelineContext[NoState],
		) (request *http.Request, statusCode *int, response []byte, err *AppError) {
			panic("handler panic")
		},
	}

	b := NewPipelineBuilder(NewMockErrorWriter(t), &bytes.Buffer{}, NoStateInit)
	p := b.New(middlewareStack.stack, testHandler.handle)

	r := httptest.NewRequest("POST", "/test", strings.NewReader(""))
	w := httptest.NewRecorder()

	p.ServeHTTP(w, r)

	wantStatusCode := 500
	gotStatusCode := w.Result().StatusCode
	assert.Equal(t, wantStatusCode, gotStatusCode)
}

func TestPipelineBuilder_New_RecoversFromMiddlewarePanic(t *testing.T) {
	t.Parallel()

	middlewareStack := newTestMiddlewareStack(t, []testMiddleware[NoState]{{
		fn: func(
			r *http.Request,
			c *PipelineContext[NoState],
		) (request *http.Request, statusCode *int, response []byte, err *AppError) {
			panic("middleware panic")
		},
	}})

	testHandler := &testAppHandler[NoState]{
		t: t,
		fn: func(
			r *http.Request,
			c *PipelineContext[NoState],
		) (request *http.Request, statusCode *int, response []byte, err *AppError) {
			return r, nil, nil, nil
		},
	}

	b := NewPipelineBuilder(NewMockErrorWriter(t), &bytes.Buffer{}, NoStateInit)
	p := b.New(middlewareStack.stack, testHandler.handle)

	r := httptest.NewRequest("POST", "/test", strings.NewReader(""))
	w := httptest.NewRecorder()

	p.ServeHTTP(w, r)

	wantStatusCode := 500
	gotStatusCode := w.Result().StatusCode
	assert.Equal(t, wantStatusCode, gotStatusCode)
}

type testPipelineStateInterface interface {
	getValue() int
	setValue(int)
}

type testPipelineState struct {
	value int
}

func newTestPipelineState() testPipelineStateInterface {
	return &testPipelineState{}
}

func (s *testPipelineState) getValue() int {
	return s.value
}

func (s *testPipelineState) setValue(v int) {
	s.value = v
}

type testAppHandler[T any] struct {
	t  testing.TB
	fn func(r *http.Request, c *PipelineContext[T]) (request *http.Request, statusCode *int, response []byte, err *AppError)
}

func (h *testAppHandler[T]) handle(w http.ResponseWriter, r *http.Request, c *PipelineContext[T]) *AppError {
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
	fn func(r *http.Request, c *PipelineContext[T]) (request *http.Request, statusCode *int, response []byte, err *AppError)
}

func (m *testMiddleware[T]) handle(h AppHandler[T]) AppHandler[T] {
	return func(w http.ResponseWriter, r *http.Request, c *PipelineContext[T]) *AppError {
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
	errors []AppError
	count  int
}

func newTestMiddlewareStack[T any](t testing.TB, data []testMiddleware[T]) *testMiddlewareStack[T] {
	t.Helper()

	c := len(data)
	s := &testMiddlewareStack[T]{
		stack:  make([]Middleware[T], c),
		calls:  make([]string, 0),
		errors: make([]AppError, c),
		count:  c,
	}

	for i, mw := range data {
		mw.t = t
		s.stack[i] = mw.handle
	}
	return s
}
