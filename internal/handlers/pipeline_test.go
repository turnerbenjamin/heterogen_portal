package handlers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/turnerbenjamin/heterogen_portal/internal/etc"
	"github.com/turnerbenjamin/heterogen_portal/testhelpers"
)

func TestStatusSpyWriter_Write_CallsWriteOnUnderlyingWriter(t *testing.T) {
	w := httptest.NewRecorder()
	ssw := statusSpyWriter{ResponseWriter: w}

	wantStatusCode := 418
	wantBody := []byte("test_content")
	wantWritten := len(wantBody)

	ssw.WriteHeader(wantStatusCode)
	gotWritten, err := ssw.Write(wantBody)
	testhelpers.AssertErrorNil(t, err)

	res := w.Result()

	gotStatusCode := res.StatusCode
	gotBody, err := io.ReadAll(res.Body)
	defer func() {
		err := res.Body.Close()
		if err != nil {
			t.Fatal(err.Error())
		}
	}()

	testhelpers.AssertErrorNil(t, err)
	testhelpers.AssertIntEqual(t, gotStatusCode, wantStatusCode)
	testhelpers.AssertBytesEqual(t, gotBody, wantBody)
	testhelpers.AssertIntEqual(t, gotWritten, wantWritten)
}

func TestStatusSpyWriter_Write_DefaultsStatusCodeTo200(t *testing.T) {
	w := httptest.NewRecorder()
	ssw := statusSpyWriter{ResponseWriter: w}

	_, err := ssw.Write([]byte("body"))
	testhelpers.AssertErrorNil(t, err)

	wantStatusCode := 200
	gotStatusCode := w.Result().StatusCode

	testhelpers.AssertIntEqual(t, gotStatusCode, wantStatusCode)
}

func TestStatusSpyWriter_WriteHeader_TracksFinalStatusCode(t *testing.T) {
	w := httptest.NewRecorder()
	ssw := statusSpyWriter{ResponseWriter: w}

	ssw.WriteHeader(201)
	ssw.WriteHeader(500)
	ssw.WriteHeader(404)

	want := 401
	ssw.WriteHeader(want)

	_, err := ssw.Write([]byte("body"))
	testhelpers.AssertErrorNil(t, err)

	gotTracked := ssw.statusCode
	gotWritten := w.Result().StatusCode

	testhelpers.AssertIntEqual(t, gotTracked, want)
	testhelpers.AssertIntEqual(t, gotWritten, want)
}

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

	b := NewPipelineBuilder(&testErrorHandler{})
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

	b := NewPipelineBuilder(&testErrorHandler{})
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
	middlewares := make([]testMiddleware[NoPipelineState], middlewareCount)

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

	b := NewPipelineBuilder(&testErrorHandler{})
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

	b := NewPipelineBuilder(&testErrorHandler{})
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

	testHandler := &testAppHandler[NoPipelineState]{
		t: t,
		fn: func(
			r *http.Request,
			c *PipelineContext[NoPipelineState],
		) (request *http.Request, statusCode *int, response []byte, err *etc.AppError) {
			sc := 200
			return r, &sc, []byte("unexpected_body_value"), nil
		},
	}

	b := NewPipelineBuilder(&testErrorHandler{})
	p := b.New(middlewareStack.stack, testHandler.handle)

	r := httptest.NewRequest("POST", "/test", strings.NewReader(""))

	w := httptest.NewRecorder()

	p.ServeHTTP(w, r)

	gotStatusCode := w.Result().StatusCode
	testhelpers.AssertIntEqual(t, gotStatusCode, wantStatusCode)
	testhelpers.AssertBytesEqual(t, w.Body.Bytes(), wantBody)
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

	b := NewPipelineBuilder(&testErrorHandler{})
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

	b := NewPipelineBuilder(&testErrorHandler{})
	p := b.New(nil, testHandler.handle)

	r := httptest.NewRequest("POST", "/test", strings.NewReader(""))
	w := httptest.NewRecorder()

	p.ServeHTTP(w, r)
	testhelpers.AssertBytesEqual(t, w.Body.Bytes(), wantBody)
	testhelpers.AssertIntEqual(t, w.Result().StatusCode, wantStatusCode)
}

type errHandlerWriteArgs struct {
	responseWriter http.ResponseWriter
	error          *etc.AppError
}

type testErrorHandler struct {
	calls []errHandlerWriteArgs
}

func (eh *testErrorHandler) Write(r http.ResponseWriter, e *etc.AppError) error {
	if eh.calls == nil {
		eh.calls = []errHandlerWriteArgs{}
	}
	eh.calls = append(eh.calls, errHandlerWriteArgs{
		responseWriter: r,
		error:          e,
	})
	return nil
}

type testAppHandler[T any] struct {
	t  testing.TB
	fn func(r *http.Request, c *PipelineContext[T]) (request *http.Request, statusCode *int, response []byte, err *etc.AppError)
}

func (h *testAppHandler[T]) handle(w http.ResponseWriter, r *http.Request, c *PipelineContext[T]) *etc.AppError {
	h.t.Helper()

	req, statusCode, res, appErr := h.fn(r, c)
	if appErr != nil {
		return appErr
	}
	r = req

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
