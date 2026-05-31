package handlers

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"maps"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/turnerbenjamin/heterogen_portal/internal/etc"
	"github.com/turnerbenjamin/heterogen_portal/internal/logging"
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
	wantBody := []byte("body")
	middlewares := newTestMiddlewareStack(t, []testMiddlewareData{{id: "mw-1"}})
	testHandler := &testAppHandler{
		getBody: func(r *http.Request) []byte {
			return wantBody
		},
		statusCode: 418,
	}
	testLogger := newTestLogger()

	b := NewPipelineBuilder(testLogger.Get, &testErrorHandler{})
	p := b.New(middlewares.stack, testHandler.handle)

	r := httptest.NewRequest("POST", "/test", strings.NewReader(""))
	w := httptest.NewRecorder()

	p.ServeHTTP(w, r)
	testhelpers.AssertBytesEqual(t, w.Body.Bytes(), wantBody)
	testhelpers.AssertIntEqual(t, w.Result().StatusCode, testHandler.statusCode)
}

func TestPipelineBuilder_New_defaultsStatusTo200_whenWriteWithoutHeader(t *testing.T) {
	middlewares := []Middleware{}
	testHandler := &testAppHandler{
		getBody: func(r *http.Request) []byte {
			return []byte("body")
		},
	}
	testLogger := newTestLogger()

	b := NewPipelineBuilder(testLogger.Get, &testErrorHandler{})
	p := b.New(middlewares, testHandler.handle)

	r := httptest.NewRequest("POST", "/test", strings.NewReader(""))
	w := httptest.NewRecorder()

	p.ServeHTTP(w, r)

	wantStatusCode := 200
	gotStatusCode := w.Result().StatusCode
	testhelpers.AssertIntEqual(t, gotStatusCode, wantStatusCode)
}

func TestPipelineBuilder_New_appliesMiddlewaresCorrectly(t *testing.T) {
	middlewareData := []testMiddlewareData{}
	for i := range 5 {
		middlewareData = append(
			middlewareData,
			testMiddlewareData{t: t, id: fmt.Sprintf("mw-%d", i)},
		)
	}
	middlewares := newTestMiddlewareStack(t, middlewareData)
	testHandler := &testAppHandler{
		getBody: func(r *http.Request) []byte {
			return []byte("body")
		},
	}
	testLogger := newTestLogger()

	b := NewPipelineBuilder(testLogger.Get, &testErrorHandler{})
	p := b.New(middlewares.stack, testHandler.handle)

	r := httptest.NewRequest("POST", "/test", strings.NewReader(""))
	w := httptest.NewRecorder()

	p.ServeHTTP(w, r)

	testhelpers.AssertIntEqual(t, len(middlewares.calls), len(middlewareData))
	for i, d := range middlewareData {
		testhelpers.AssertStringEqual(t, middlewares.calls[i], d.id)
	}
}

func TestPipelineBuilder_New_middlewareCanModifyRequest_andHandlerSeesChange(t *testing.T) {
	type testContextKey string
	var requestKey testContextKey = "r-key"
	requestValue := "r-val"

	middlewares := newTestMiddlewareStack(t, []testMiddlewareData{
		{
			id: "mw-1",
			fn: func(r *http.Request) (*http.Request, etc.AppError) {
				return r.WithContext(context.WithValue(r.Context(), requestKey, requestValue)), nil
			},
		},
	})

	testHandler := &testAppHandler{
		getBody: func(r *http.Request) []byte {
			v := r.Context().Value(requestKey)
			body, ok := v.(string)
			if !ok {
				body = "FAIL"
			}
			return []byte(body)
		},
	}

	testLogger := newTestLogger()

	b := NewPipelineBuilder(testLogger.Get, &testErrorHandler{})
	p := b.New(middlewares.stack, testHandler.handle)

	r := httptest.NewRequest("POST", "/test", strings.NewReader(""))
	w := httptest.NewRecorder()

	p.ServeHTTP(w, r)

	testhelpers.AssertBytesEqual(t, w.Body.Bytes(), []byte(requestValue))
}

func TestPipelineBuilder_New_middlewareCanModifyResponseBeforeHandler(t *testing.T) {
	middlewareResponse := []byte("middleware-response")
	handlerResponse := []byte("handler-response")

	middlewares := newTestMiddlewareStack(t, []testMiddlewareData{
		{
			id:       "mw-1",
			response: middlewareResponse,
		},
	})

	testHandler := &testAppHandler{
		getBody: func(r *http.Request) []byte {
			return handlerResponse
		},
	}

	testLogger := newTestLogger()

	b := NewPipelineBuilder(testLogger.Get, &testErrorHandler{})
	p := b.New(middlewares.stack, testHandler.handle)

	r := httptest.NewRequest("POST", "/test", strings.NewReader(""))
	w := httptest.NewRecorder()

	p.ServeHTTP(w, r)

	testhelpers.AssertBytesEqual(t, w.Body.Bytes(), middlewareResponse)
}

func TestPipelineBuilder_New_invokesHandlerDirectly_WhenNoMiddlewares(t *testing.T) {
	wantBody := []byte("body")
	middlewares := []Middleware{}
	testHandler := &testAppHandler{
		getBody: func(r *http.Request) []byte {
			return wantBody
		},
		statusCode: 418,
	}
	testLogger := newTestLogger()

	b := NewPipelineBuilder(testLogger.Get, &testErrorHandler{})
	p := b.New(middlewares, testHandler.handle)

	r := httptest.NewRequest("POST", "/test", strings.NewReader(""))
	w := httptest.NewRecorder()

	p.ServeHTTP(w, r)
	testhelpers.AssertBytesEqual(t, w.Body.Bytes(), wantBody)
	testhelpers.AssertIntEqual(t, w.Result().StatusCode, testHandler.statusCode)
}

func TestPipelineBuilder_New_invokesHandlerDirectly_WhenMiddlewaresIsNil(t *testing.T) {
	wantBody := []byte("body")
	testHandler := &testAppHandler{
		getBody: func(r *http.Request) []byte {
			return wantBody
		},
		statusCode: 418,
	}
	testLogger := newTestLogger()

	b := NewPipelineBuilder(testLogger.Get, &testErrorHandler{})
	p := b.New(nil, testHandler.handle)

	r := httptest.NewRequest("POST", "/test", strings.NewReader(""))
	w := httptest.NewRecorder()

	p.ServeHTTP(w, r)
	testhelpers.AssertBytesEqual(t, w.Body.Bytes(), wantBody)
	testhelpers.AssertIntEqual(t, w.Result().StatusCode, testHandler.statusCode)
}

func TestPipelineBuilder_New_passesLogger_fromNewLogger(t *testing.T) {
	handlerId := "handler-1"
	middlewareData := []testMiddlewareData{}
	for i := range 5 {
		middlewareData = append(
			middlewareData,
			testMiddlewareData{t: t, id: fmt.Sprintf("mw-%d", i)},
		)
	}
	middlewares := newTestMiddlewareStack(t, middlewareData)
	testHandler := &testAppHandler{
		id:      handlerId,
		getBody: func(r *http.Request) []byte { return []byte("body") },
	}
	testLogger := newTestLogger()

	b := NewPipelineBuilder(testLogger.Get, &testErrorHandler{})
	p := b.New(middlewares.stack, testHandler.handle)

	r := httptest.NewRequest("POST", "/test", strings.NewReader(""))
	w := httptest.NewRecorder()

	p.ServeHTTP(w, r)

	for _, d := range middlewareData {
		_, ok := testLogger.properties[d.id]
		if !ok {
			t.Fatalf("got nil, but want middleware %s to have logged call", d.id)
		}
	}
	_, ok := testLogger.properties[handlerId]
	if !ok {
		t.Fatal("got nil, but want handler to have logged call")
	}
}

func TestPipelineBuilder_New_WritesLogWithFinalStatusCodeAfterHandlerCalled(t *testing.T) {
	handlerId := "handler-id"
	middlewares := []Middleware{}
	testHandler := &testAppHandler{
		id: handlerId,
		getBody: func(r *http.Request) []byte {
			return []byte("body")
		},
		statusCode: 418,
	}
	testLogger := newTestLogger()

	b := NewPipelineBuilder(testLogger.Get, &testErrorHandler{})
	p := b.New(middlewares, testHandler.handle)

	r := httptest.NewRequest("POST", "/test", strings.NewReader(""))
	w := httptest.NewRecorder()

	p.ServeHTTP(w, r)

	expectedLogProperties := map[string]string{
		handlerId:     ".",
		"status_code": "^418$",
	}

	for wk, wv := range expectedLogProperties {
		gv, ok := testLogger.loggedProperties[wk]
		if !ok {
			t.Fatalf("got nil, but wanted key %s", wk)
		}
		testhelpers.AssertStringMatches(t, gv, wv)
	}
}

// ERRORS

type testLogger struct {
	properties       map[string]string
	loggedProperties map[string]string
	request          *http.Request
	writer           bytes.Buffer
}

func newTestLogger() *testLogger {
	return &testLogger{
		properties: map[string]string{},
		request:    nil,
		writer:     bytes.Buffer{},
	}
}

func (l *testLogger) Get(r *http.Request) logging.Logger {
	l.request = r
	return l
}

func (l *testLogger) AddKV(k, v string) {
	l.properties[k] = v
}

func (l *testLogger) WriteLog() error {
	l.loggedProperties = map[string]string{}
	maps.Copy(l.loggedProperties, l.properties)
	return nil
}

type errHandlerWriteArgs struct {
	responseWriter http.ResponseWriter
	logger         logging.Logger
	error          etc.AppError
}

type testErrorHandler struct {
	calls []errHandlerWriteArgs
}

func (eh *testErrorHandler) Write(r http.ResponseWriter, l logging.Logger, e etc.AppError) {
	if eh.calls == nil {
		eh.calls = []errHandlerWriteArgs{}
	}
	eh.calls = append(eh.calls, errHandlerWriteArgs{
		responseWriter: r,
		logger:         l,
		error:          e,
	})
}

type testAppHandler struct {
	id         string
	getBody    func(r *http.Request) []byte
	statusCode int
	throws     etc.AppError
}

func (h *testAppHandler) handle(w http.ResponseWriter, r *http.Request, l logging.Logger) etc.AppError {
	if h.statusCode != 0 {
		w.WriteHeader(h.statusCode)
	}

	l.AddKV(h.id, "CALLED")

	body := h.getBody(r)
	_, err := w.Write(body)
	if err != nil {
		return newAppError(500, err.Error())
	}
	return h.throws
}

type testAppError struct {
	code    int
	message string
}

func newAppError(code int, message string) etc.AppError {
	return &testAppError{
		message: message,
	}
}

func (e *testAppError) Code() int {
	return e.code
}

func (e *testAppError) ToastError() string {
	return e.message
}

func (e *testAppError) PageErrors() []string {
	return []string{e.message}
}

type testMiddlewareData struct {
	t          testing.TB
	id         string
	recordCall func(id string)
	fn         func(r *http.Request) (*http.Request, etc.AppError)
	response   []byte
}

func (m *testMiddlewareData) handle(h AppHandler) AppHandler {
	return func(w http.ResponseWriter, r *http.Request, l logging.Logger) etc.AppError {
		m.t.Helper()

		m.recordCall(m.id)
		l.AddKV(m.id, "CALLED")

		if m.response != nil {
			_, err := w.Write(m.response)
			if err != nil {
				m.t.Fatal(err.Error())
			}
			return nil
		}

		var err etc.AppError = nil
		if m.fn != nil {
			r, err = m.fn(r)
		}

		if err != nil {
			return err
		}
		return h(w, r, l)
	}
}

type testMiddlewareStack struct {
	stack  []Middleware
	calls  []string
	errors []etc.AppError
	count  int
}

func newTestMiddlewareStack(t testing.TB, data []testMiddlewareData) *testMiddlewareStack {
	t.Helper()

	c := len(data)
	s := &testMiddlewareStack{
		stack:  make([]Middleware, c),
		calls:  make([]string, 0),
		errors: make([]etc.AppError, c),
		count:  c,
	}

	for i, mw := range data {
		mw.t = t
		mw.recordCall = func(callData string) {
			s.calls = append(s.calls, callData)
		}
		s.stack[i] = mw.handle
	}
	return s
}
