package handlers

import (
	"bytes"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/turnerbenjamin/heterogen_portal/internal/db"
	"github.com/turnerbenjamin/heterogen_portal/internal/etc"
	"github.com/turnerbenjamin/heterogen_portal/internal/templates"
	"github.com/turnerbenjamin/heterogen_portal/testhelpers"
)

func TestGetRoot_Returns404_WhenPathIsNotRoot(t *testing.T) {
	t.Parallel()

	wantStatusCode := http.StatusNotFound

	templateStore := mockTemplateStore{t: t, returns: nil}
	handler := GetRootHandler(&templateStore)

	r := httptest.NewRequest("GET", "/test", strings.NewReader(""))
	w := httptest.NewRecorder()

	userState := &UserState{}
	logSink := &bytes.Buffer{}
	c := makeUserStatePipelineContext(t, userState, logSink)

	err := handler(w, r, c)

	testhelpers.AssertAppErrorNil(t, err)
	testhelpers.AssertIntEqual(t, w.Code, wantStatusCode)
}

func TestGetRoot_RedirectsToSignIn_WhenUserIsNil(t *testing.T) {
	t.Parallel()

	wantStatusCode := http.StatusSeeOther
	wantRedirectPath := "/sign-in"

	testData := []struct {
		isHtmxRequest bool
		redirectKey   string
	}{
		{isHtmxRequest: true, redirectKey: "HX-Redirect"},
		{isHtmxRequest: false, redirectKey: "Location"},
	}

	templateStore := mockTemplateStore{t: t, returns: nil}
	handler := GetRootHandler(&templateStore)

	for _, td := range testData {
		r := httptest.NewRequest("GET", "/", strings.NewReader(""))
		if td.isHtmxRequest {
			r.Header.Set("HX-Request", "true")
		}

		w := httptest.NewRecorder()
		c := makeUserStatePipelineContext(t, &UserState{}, &bytes.Buffer{})

		err := handler(w, r, c)

		testhelpers.AssertAppErrorNil(t, err)
		testhelpers.AssertIntEqual(t, w.Code, wantStatusCode)
		testhelpers.AssertStringEqual(t, w.Result().Header.Get(td.redirectKey), wantRedirectPath)
	}
}

func TestGetRoot_ReturnsAppPage_WhenUserIsNotNil(t *testing.T) {
	t.Parallel()

	wantStatusCode := http.StatusOK
	wantTemplate := templates.TmplPageApp
	wantPageTitle := "HETEROGEN"
	wantState := &UserState{
		User: &db.User{},
	}

	testData := []struct {
		isHtmxRequest        bool
		wantContentOnlyValue bool
	}{
		{isHtmxRequest: true, wantContentOnlyValue: true},
		{isHtmxRequest: false, wantContentOnlyValue: false},
	}

	for _, td := range testData {
		templateStore := mockTemplateStore{t: t, returns: nil}
		handler := GetRootHandler(&templateStore)

		r := httptest.NewRequest("GET", "/", strings.NewReader(""))
		if td.isHtmxRequest {
			r.Header.Set("HX-Request", "true")
		}

		w := httptest.NewRecorder()
		c := makeUserStatePipelineContext(t, wantState, &bytes.Buffer{})

		err := handler(w, r, c)

		testhelpers.AssertAppErrorNil(t, err)
		testhelpers.AssertIntEqual(t, w.Code, wantStatusCode)
		testhelpers.AssertIntEqual(t, len(templateStore.calls), 1)

		executeCall := templateStore.calls[0]
		testhelpers.AssertEqual(t, executeCall.templateId, wantTemplate)
		testhelpers.AssertEqual(t, executeCall.data.PageConfig.ContentOnly, td.wantContentOnlyValue)
		testhelpers.AssertEqual(t, executeCall.data.Data, wantState)
		testhelpers.AssertStringEqual(t, executeCall.data.PageConfig.Title, wantPageTitle)
	}
}

func TestGetRoot_ReturnsServerError_WhenExecuteReturnsAnError(t *testing.T) {
	t.Parallel()

	wantInnerError := errors.New("expected_test_error")
	wantAppError := &etc.AppError{
		Code:       http.StatusInternalServerError,
		ToastError: etc.ErrMessageInternalServerError,
		PageErrors: []string{etc.ErrMessageInternalServerError},
		InnerError: wantInnerError,
	}

	templateStore := mockTemplateStore{t: t, returns: wantInnerError}
	handler := GetRootHandler(&templateStore)

	r := httptest.NewRequest("GET", "/", strings.NewReader(""))
	s := &UserState{
		User: &db.User{},
	}
	w := httptest.NewRecorder()
	c := makeUserStatePipelineContext(t, s, &bytes.Buffer{})

	gotErr := handler(w, r, c)
	testhelpers.AssertAppErrorEqual(t, gotErr, wantAppError)
}

func TestGetSignInHandler_ReturnsSignInPage(t *testing.T) {
	t.Parallel()

	wantStatusCode := http.StatusOK
	wantTemplate := templates.TmplPageUserSignIn
	wantPageTitle := "HETEROGEN | SIGN-IN"

	testData := []struct {
		isHtmxRequest        bool
		wantContentOnlyValue bool
	}{
		{isHtmxRequest: true, wantContentOnlyValue: false},
		{isHtmxRequest: false, wantContentOnlyValue: false},
	}

	for _, td := range testData {
		templateStore := mockTemplateStore{t: t, returns: nil}
		handler := GetSignInHandler(&templateStore)

		r := httptest.NewRequest("GET", "/sign-in", strings.NewReader(""))
		if td.isHtmxRequest {
			r.Header.Set("HX-Request", "true")
		}

		w := httptest.NewRecorder()
		c := &PipelineContext[NoState]{}

		err := handler(w, r, c)

		testhelpers.AssertAppErrorNil(t, err)
		testhelpers.AssertIntEqual(t, w.Code, wantStatusCode)
		testhelpers.AssertIntEqual(t, len(templateStore.calls), 1)

		executeCall := templateStore.calls[0]
		testhelpers.AssertEqual(t, executeCall.templateId, wantTemplate)
		testhelpers.AssertEqual(t, executeCall.data.PageConfig.ContentOnly, td.wantContentOnlyValue)
		testhelpers.AssertStringEqual(t, executeCall.data.PageConfig.Title, wantPageTitle)
	}
}

func TestGetSignInHandler_ReturnsServerError_WhenExecuteReturnsAnError(t *testing.T) {
	t.Parallel()

	wantInnerError := errors.New("expected_test_error")
	wantAppError := &etc.AppError{
		Code:       http.StatusInternalServerError,
		ToastError: etc.ErrMessageInternalServerError,
		PageErrors: []string{etc.ErrMessageInternalServerError},
		InnerError: wantInnerError,
	}

	templateStore := mockTemplateStore{t: t, returns: wantInnerError}
	handler := GetSignInHandler(&templateStore)

	r := httptest.NewRequest("GET", "/sign-in", strings.NewReader(""))
	w := httptest.NewRecorder()
	c := &PipelineContext[NoState]{}

	gotErr := handler(w, r, c)
	testhelpers.AssertAppErrorEqual(t, gotErr, wantAppError)
}

func TestMsalFlowHandlers_ReturnServerError_WhenHTMXRequest(t *testing.T) {
	t.Parallel()

	wantAppError := errHtmxNotSupported

	templateStore := mockTemplateStore{t: t}

	msalFlowHandlerBuilders := []func(TemplateStore) AppHandler[NoState]{
		GetSignInRedirectHandler,
		GetSignOutHandler,
		GetSignedOutHandler,
	}

	for _, hb := range msalFlowHandlerBuilders {
		handler := hb(&templateStore)

		r := httptest.NewRequest("GET", "/sign-in", strings.NewReader(""))
		r.Header.Add("HX-Request", "true")

		w := httptest.NewRecorder()
		c := &PipelineContext[NoState]{}

		gotErr := handler(w, r, c)
		testhelpers.AssertAppErrorEqual(t, gotErr, wantAppError)
	}
}

func TestMsalFlowHandlers_ReturnServerError_WhenExecuteReturnsAnError(t *testing.T) {
	t.Parallel()

	wantInnerError := errors.New("expected_test_error")
	wantAppError := &etc.AppError{
		Code:       http.StatusInternalServerError,
		ToastError: etc.ErrMessageInternalServerError,
		PageErrors: []string{etc.ErrMessageInternalServerError},
		InnerError: wantInnerError,
	}

	msalFlowHandlerBuilders := []func(TemplateStore) AppHandler[NoState]{
		GetSignInRedirectHandler,
		GetSignOutHandler,
		GetSignedOutHandler,
	}

	for _, hb := range msalFlowHandlerBuilders {
		templateStore := mockTemplateStore{t: t, returns: wantInnerError}
		handler := hb(&templateStore)

		r := httptest.NewRequest("GET", "/", strings.NewReader(""))

		w := httptest.NewRecorder()
		c := &PipelineContext[NoState]{}

		gotErr := handler(w, r, c)
		testhelpers.AssertAppErrorEqual(t, gotErr, wantAppError)
	}
}

func TestGetSignInRedirectHandler_ReturnsSignInRedirectPage(t *testing.T) {
	t.Parallel()

	wantStatusCode := http.StatusOK
	wantTemplate := templates.TmpPageUserSignInRedirect
	wantPageTitle := "HETEROGEN | SIGN-IN"
	wantContentOnlyValue := false

	templateStore := mockTemplateStore{t: t, returns: nil}
	handler := GetSignInRedirectHandler(&templateStore)

	r := httptest.NewRequest("GET", "/sign-in-redirect", strings.NewReader(""))

	w := httptest.NewRecorder()
	c := &PipelineContext[NoState]{}

	err := handler(w, r, c)

	testhelpers.AssertAppErrorNil(t, err)
	testhelpers.AssertIntEqual(t, w.Code, wantStatusCode)
	testhelpers.AssertIntEqual(t, len(templateStore.calls), 1)

	executeCall := templateStore.calls[0]
	testhelpers.AssertEqual(t, executeCall.templateId, wantTemplate)
	testhelpers.AssertEqual(t, executeCall.data.PageConfig.ContentOnly, wantContentOnlyValue)
	testhelpers.AssertStringEqual(t, executeCall.data.PageConfig.Title, wantPageTitle)
}

func TestGetSignOutHandler_ReturnsSignOutPage(t *testing.T) {
	t.Parallel()

	wantStatusCode := http.StatusOK
	wantTemplate := templates.TmplPageUserSignOut
	wantPageTitle := "HETEROGEN | SIGN-OUT"
	wantContentOnlyValue := false

	templateStore := mockTemplateStore{t: t, returns: nil}
	handler := GetSignOutHandler(&templateStore)

	r := httptest.NewRequest("GET", "/sign-out", strings.NewReader(""))

	w := httptest.NewRecorder()
	c := &PipelineContext[NoState]{}

	err := handler(w, r, c)

	testhelpers.AssertAppErrorNil(t, err)
	testhelpers.AssertIntEqual(t, w.Code, wantStatusCode)
	testhelpers.AssertIntEqual(t, len(templateStore.calls), 1)

	executeCall := templateStore.calls[0]
	testhelpers.AssertEqual(t, executeCall.templateId, wantTemplate)
	testhelpers.AssertEqual(t, executeCall.data.PageConfig.ContentOnly, wantContentOnlyValue)
	testhelpers.AssertStringEqual(t, executeCall.data.PageConfig.Title, wantPageTitle)
}

func TestGetSignOutHandler_SetsHeadersToUnsetJWT(t *testing.T) {
	t.Parallel()

	wantCookie := &http.Cookie{
		Name:        jwtCookieIdentifier,
		SameSite:    http.SameSiteStrictMode,
		MaxAge:      -1,
		Expires:     time.Unix(0, 0).UTC(),
		Secure:      true,
		Partitioned: true,
		HttpOnly:    true,
	}

	templateStore := mockTemplateStore{t: t, returns: nil}
	handler := GetSignOutHandler(&templateStore)

	r := httptest.NewRequest("GET", "/sign-out", strings.NewReader(""))

	w := httptest.NewRecorder()
	c := &PipelineContext[NoState]{}

	err := handler(w, r, c)

	testhelpers.AssertAppErrorNil(t, err)
	var gotCookie *http.Cookie
	for _, c := range w.Result().Cookies() {
		if c.Name == wantCookie.Name {
			gotCookie = c
			break
		}
	}

	testhelpers.AssertNotNil(t, gotCookie, wantCookie)
	testhelpers.AssertEqual(t, gotCookie.SameSite, wantCookie.SameSite)
	testhelpers.AssertIntEqual(t, gotCookie.MaxAge, wantCookie.MaxAge)
	testhelpers.AssertEqual(t, gotCookie.Expires, wantCookie.Expires)
	testhelpers.AssertEqual(t, gotCookie.Secure, wantCookie.Secure)
	testhelpers.AssertEqual(t, gotCookie.Partitioned, wantCookie.Partitioned)
	testhelpers.AssertEqual(t, gotCookie.HttpOnly, wantCookie.HttpOnly)
}

func TestGetSignedOutHandler_ReturnsSignedOutPage(t *testing.T) {
	t.Parallel()

	wantStatusCode := http.StatusOK
	wantTemplate := templates.TmplPageUserSignedOut
	wantPageTitle := "HETEROGEN | SIGNED-OUT"
	wantContentOnlyValue := false

	templateStore := mockTemplateStore{t: t, returns: nil}
	handler := GetSignedOutHandler(&templateStore)

	r := httptest.NewRequest("GET", "/signed-out", strings.NewReader(""))
	w := httptest.NewRecorder()
	c := &PipelineContext[NoState]{}

	err := handler(w, r, c)

	testhelpers.AssertAppErrorNil(t, err)
	testhelpers.AssertIntEqual(t, w.Code, wantStatusCode)
	testhelpers.AssertIntEqual(t, len(templateStore.calls), 1)

	executeCall := templateStore.calls[0]
	testhelpers.AssertEqual(t, executeCall.templateId, wantTemplate)
	testhelpers.AssertEqual(t, executeCall.data.PageConfig.ContentOnly, wantContentOnlyValue)
	testhelpers.AssertStringEqual(t, executeCall.data.PageConfig.Title, wantPageTitle)
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

func makeUserStatePipelineContext(
	t testing.TB,
	state *UserState,
	logSink io.Writer,
) *PipelineContext[UserState] {
	t.Helper()
	return &PipelineContext[UserState]{
		logger: slog.New(slog.NewJSONHandler(logSink, &slog.HandlerOptions{})),
		state:  state,
	}
}
