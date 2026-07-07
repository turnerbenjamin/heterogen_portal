package handlers

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/turnerbenjamin/heterogen_portal/internal/constants"
	"github.com/turnerbenjamin/heterogen_portal/internal/db"
	"github.com/turnerbenjamin/heterogen_portal/internal/services"
	"github.com/turnerbenjamin/heterogen_portal/internal/templates"
)

func TestGetRootHandler_Returns404WhenPathIsNotRoot(t *testing.T) {
	t.Parallel()

	testData := []struct {
		path     string
		wantCode int
	}{
		{path: "/a", wantCode: http.StatusNotFound},
		{path: "/sign-in", wantCode: http.StatusNotFound},
		{path: "//", wantCode: http.StatusNotFound},
		{path: "/", wantCode: http.StatusOK},
	}

	for _, td := range testData {
		r := httptest.NewRequest("GET", td.path, strings.NewReader(""))
		w := httptest.NewRecorder()
		c := &PipelineContext[UserState]{state: UserStateInit()}
		c.state.SetUser(&db.User{})

		ts := NewMockTemplateStore(t)
		ts.EXPECT().
			Execute(mock.Anything, mock.Anything, mock.Anything).
			Maybe().
			Return(nil)

		as := NewMockAuthService(t)

		h := NewAuthHandler(ts, as)

		err := h.GetRoot(w, r, c)

		assert.Nil(t, err)
		assert.Equal(t, td.wantCode, w.Code)
	}
}

func TestGetRootHandler_ReturnsMainAppTemplate(t *testing.T) {
	t.Parallel()

	wantStatusCode := http.StatusOK
	wantTemplate := templates.TmplPageApp
	wantPageTitle := "HETEROGEN"

	wantState := UserStateInit()
	wantState.SetUser(&db.User{})

	testData := []struct {
		isHtmxRequest        bool
		wantContentOnlyValue bool
	}{
		{isHtmxRequest: true, wantContentOnlyValue: true},
		{isHtmxRequest: false, wantContentOnlyValue: false},
	}

	for _, td := range testData {
		r := httptest.NewRequest("GET", "/", strings.NewReader(""))
		if td.isHtmxRequest {
			r.Header.Set("HX-Request", "true")
		}

		w := httptest.NewRecorder()
		c := &PipelineContext[UserState]{state: wantState}

		ts := NewMockTemplateStore(t)
		var capturedTemplateArgs templates.TemplateArgs
		ts.EXPECT().
			Execute(wantTemplate, w, mock.Anything).
			Run(func(_ templates.TemplateIdentifier, _ io.Writer, data templates.TemplateArgs) {
				capturedTemplateArgs = data
			}).
			Return(nil)

		as := NewMockAuthService(t)

		h := NewAuthHandler(ts, as)

		err := h.GetRoot(w, r, c)

		assert.Nil(t, err)
		assert.Equal(t, wantStatusCode, w.Code)
		assert.Equal(t, td.wantContentOnlyValue, capturedTemplateArgs.PageConfig.ContentOnly)
		assert.Equal(t, wantState, capturedTemplateArgs.Data)
		assert.Equal(t, wantPageTitle, capturedTemplateArgs.PageConfig.Title)
	}
}

func TestGetRootHandler_HandlesExecuteTemplateErrors(t *testing.T) {
	t.Parallel()

	wantInnerError := errors.New("test auth services err")
	wantAppError := &AppError{
		Code:       http.StatusInternalServerError,
		ToastError: constants.ErrMsgInternalServerError,
		PageErrors: []string{constants.ErrMsgInternalServerError},
		innerError: wantInnerError,
	}

	r := httptest.NewRequest("GET", "/", strings.NewReader(""))
	w := httptest.NewRecorder()
	c := &PipelineContext[UserState]{state: UserStateInit()}
	c.state.SetUser(&db.User{})

	ts := NewMockTemplateStore(t)
	ts.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything).Return(wantInnerError)

	as := NewMockAuthService(t)
	h := NewAuthHandler(ts, as)

	err := h.GetRoot(w, r, c)

	assert.EqualValues(t, wantAppError, err)
}

func TestGetSignInRedirectHandler_ReturnsErrorWhenCodeStateOrOidcStateMissing(t *testing.T) {
	t.Parallel()

	testData := []struct {
		codeParam          string
		stateParam         string
		HasOidcStateCookie bool
		WantInnerErr       error
	}{
		{
			codeParam:          "",
			stateParam:         "state-val",
			HasOidcStateCookie: true,
			WantInnerErr:       errors.New(constants.ErrMissingOIDCCodeParam),
		},
		{
			codeParam:          "code-val",
			stateParam:         "",
			HasOidcStateCookie: true,
			WantInnerErr:       errors.New(constants.ErrMissingOIDCStateParam),
		},
		{
			codeParam:          "code-val",
			stateParam:         "state-val",
			HasOidcStateCookie: false,
			WantInnerErr:       errors.New(constants.ErrMissingOIDCStateCookie),
		},
	}

	for _, td := range testData {
		wantAppError := &AppError{
			Code:       http.StatusInternalServerError,
			ToastError: constants.ErrMsgInternalServerError,
			PageErrors: []string{constants.ErrMsgInternalServerError},
			innerError: td.WantInnerErr,
		}

		path := buildRedirectPathWithParams("/sign-in-redirect", td.codeParam, td.stateParam)
		r := httptest.NewRequest("GET", path, strings.NewReader(""))
		if td.HasOidcStateCookie {
			r.AddCookie(&http.Cookie{
				Name:  constants.IdentifierOidcStateCookie,
				Value: "signed-oidc-state",
			})
		}

		w := httptest.NewRecorder()
		c := &PipelineContext[NoState]{}

		ts := NewMockTemplateStore(t)
		as := NewMockAuthService(t)

		h := NewAuthHandler(ts, as)

		err := h.GetSignInRedirect(w, r, c)

		assert.EqualValues(t, wantAppError, err)
		if td.HasOidcStateCookie {
			wantCookie := buildExpectedUnsetOidcStateCookie()

			cookies := w.Result().Cookies()
			assert.Equal(t, 1, len(cookies))

			gotCookie := cookies[0]
			gotCookie.Raw = ""
			assert.EqualValues(t, wantCookie, gotCookie)
		}
	}
}

func TestGetSignInRedirectHandler_HandlesAuthServiceErrors(t *testing.T) {
	t.Parallel()

	wantInnerError := errors.New("test inner error")
	wantAppError := &AppError{
		Code:       http.StatusInternalServerError,
		ToastError: constants.ErrMsgInternalServerError,
		PageErrors: []string{constants.ErrMsgInternalServerError},
		innerError: wantInnerError,
	}

	testCodeValue := "test-code-value"
	testStateValue := "test-state-value"
	testSignedOidcStateValue := "test-oidc-state-value"

	path := buildRedirectPathWithParams("/sign-in-redirect", testCodeValue, testStateValue)
	r := httptest.NewRequest("GET", path, strings.NewReader(""))
	r.AddCookie(&http.Cookie{
		Name:  constants.IdentifierOidcStateCookie,
		Value: testSignedOidcStateValue,
	})

	w := httptest.NewRecorder()
	c := &PipelineContext[NoState]{}

	ts := NewMockTemplateStore(t)
	as := NewMockAuthService(t)
	as.EXPECT().AuthenticateUser(
		mock.Anything,
		testCodeValue,
		testStateValue,
		testSignedOidcStateValue,
	).Return(nil, wantInnerError)

	h := NewAuthHandler(ts, as)

	err := h.GetSignInRedirect(w, r, c)

	assert.EqualValues(t, wantAppError, err)
}

func TestGetSignInRedirectHandler_SetsAppCookieAndRedirectsUser(t *testing.T) {
	t.Parallel()

	wantAppToken := "some-app-token"
	wantStatusCode := http.StatusSeeOther

	wantSetAppJwtCookie := buildExpectedSetJwtCookie(wantAppToken)
	wantClearOidcStateCookie := buildExpectedUnsetOidcStateCookie()

	testData := []struct {
		isHtmx               bool
		redirectKey          string
		wantRequestedPath    string
		codeValue            string
		stateValue           string
		signedOidcStateValue string
		appToken             string
	}{
		{
			isHtmx:               true,
			redirectKey:          "HX-Redirect",
			wantRequestedPath:    "/some-protected-endpoint",
			codeValue:            "some-code",
			stateValue:           "some-state-value",
			signedOidcStateValue: "some-oidc-state-value",
			appToken:             "an-app-token",
		},
		{
			isHtmx:               false,
			redirectKey:          "Location",
			wantRequestedPath:    "/some-other-endpoint",
			codeValue:            "some-other-code",
			stateValue:           "some-other-state-value",
			signedOidcStateValue: "some-other-oidc-state-value",
			appToken:             "an-app-token-also",
		},
	}

	for _, td := range testData {
		path := buildRedirectPathWithParams("/sign-in-redirect", td.codeValue, td.stateValue)

		r := httptest.NewRequest("GET", path, strings.NewReader(""))
		if td.isHtmx {
			r.Header.Add(constants.HxRequestHeaderRequest, "true")
		}

		r.AddCookie(&http.Cookie{
			Name:  constants.IdentifierOidcStateCookie,
			Value: td.signedOidcStateValue,
		})

		w := httptest.NewRecorder()
		c := &PipelineContext[NoState]{}

		ts := NewMockTemplateStore(t)
		as := NewMockAuthService(t)
		as.EXPECT().AuthenticateUser(
			mock.Anything,
			td.codeValue,
			td.stateValue,
			td.signedOidcStateValue,
		).Return(
			&services.AuthenticateUserResponse{
				AppToken:      wantAppToken,
				RequestedPath: td.wantRequestedPath,
			},
			nil,
		)

		h := NewAuthHandler(ts, as)

		err := h.GetSignInRedirect(w, r, c)

		assert.Nil(t, err)
		assert.Equal(t, wantStatusCode, w.Code)
		assert.Equal(t, td.wantRequestedPath, w.Result().Header.Get(td.redirectKey))

		cookies := w.Result().Cookies()
		assert.Equal(t, 2, len(cookies))

		gotClearOidcStateCookie := cookies[0]
		gotClearOidcStateCookie.Raw = ""

		gotSetAppJwtCookie := cookies[1]
		gotSetAppJwtCookie.Raw = ""

		assert.EqualValues(t, wantClearOidcStateCookie, gotClearOidcStateCookie)
		assert.EqualValues(t, wantSetAppJwtCookie, gotSetAppJwtCookie)
	}
}

func TestSignOutHandler_UnsetsAppJwtCookieAndRedirectsUserToSignOut(t *testing.T) {
	t.Parallel()

	testData := []struct {
		isHtmx        bool
		redirectKey   string
		redirectValue string
	}{
		{
			isHtmx:        true,
			redirectKey:   "HX-Redirect",
			redirectValue: "/some-redirect-url",
		},
		{
			isHtmx:        false,
			redirectKey:   "Location",
			redirectValue: "/some-other-redirect-url",
		},
	}

	for _, td := range testData {

		r := httptest.NewRequest("GET", "/sign-out", strings.NewReader(""))
		if td.isHtmx {
			r.Header.Add(constants.HxRequestHeaderRequest, "true")
		}

		w := httptest.NewRecorder()
		c := &PipelineContext[NoState]{}

		ts := NewMockTemplateStore(t)
		as := NewMockAuthService(t)
		as.EXPECT().BuildSignOutRedirectRequest().Return(td.redirectValue)

		h := NewAuthHandler(ts, as)

		err := h.GetSignOut(w, r, c)

		assert.Nil(t, err)
		assert.Equal(t, http.StatusSeeOther, w.Result().StatusCode)
		assert.Equal(t, td.redirectValue, w.Result().Header.Get(td.redirectKey))

		cookies := w.Result().Cookies()
		assert.Equal(t, 1, len(cookies))

		gotUnsetAppJwtCookie := cookies[0]
		gotUnsetAppJwtCookie.Raw = ""
		gotUnsetAppJwtCookie.RawExpires = ""

		wantUnsetAppJwtCookie := buildExpectedUnsetAppJwtCookie()
		assert.EqualValues(t, wantUnsetAppJwtCookie, gotUnsetAppJwtCookie)
	}
}

func TestSignedOutHandler_ReturnsSignedOutPage(t *testing.T) {
	t.Parallel()

	wantStatusCode := http.StatusOK
	wantTemplate := templates.TmplPageUserSignedOut
	wantPageTitle := "HETEROGEN | SIGNED-OUT"
	wantContentOnlyValue := false

	r := httptest.NewRequest("GET", "/", strings.NewReader(""))

	w := httptest.NewRecorder()
	c := &PipelineContext[NoState]{}

	var capturedTemplateArgs templates.TemplateArgs
	ts := NewMockTemplateStore(t)
	ts.
		EXPECT().Execute(wantTemplate, w, mock.Anything).
		Run(func(_ templates.TemplateIdentifier, _ io.Writer, data templates.TemplateArgs) {
			capturedTemplateArgs = data
		}).
		Return(nil)

	as := NewMockAuthService(t)
	h := NewAuthHandler(ts, as)

	err := h.GetSignedOut(w, r, c)

	assert.Nil(t, err)
	assert.Equal(t, wantStatusCode, w.Code)
	assert.Equal(t, wantContentOnlyValue, capturedTemplateArgs.PageConfig.ContentOnly)
	assert.Equal(t, wantPageTitle, capturedTemplateArgs.PageConfig.Title)
}

func TestSignedOutHandler_ReturnsErrIfHtmxRequest(t *testing.T) {
	t.Parallel()

	wantInnerError := errors.New(constants.ErrMsgHtmxNotSupported)
	wantAppError := &AppError{
		Code:       http.StatusInternalServerError,
		ToastError: constants.ErrMsgInternalServerError,
		PageErrors: []string{constants.ErrMsgInternalServerError},
		innerError: wantInnerError,
	}

	r := httptest.NewRequest("GET", "/", strings.NewReader(""))
	r.Header.Add(constants.HxRequestHeaderRequest, "true")

	w := httptest.NewRecorder()
	c := &PipelineContext[NoState]{}

	ts := NewMockTemplateStore(t)
	as := NewMockAuthService(t)

	h := NewAuthHandler(ts, as)

	err := h.GetSignedOut(w, r, c)
	assert.EqualValues(t, wantAppError, err)
}

func TestGetSignedOutHandler_HandlesExecuteErr(t *testing.T) {
	t.Parallel()

	wantInnerError := errors.New("some template error")
	wantAppError := &AppError{
		Code:       http.StatusInternalServerError,
		ToastError: constants.ErrMsgInternalServerError,
		PageErrors: []string{constants.ErrMsgInternalServerError},
		innerError: wantInnerError,
	}

	r := httptest.NewRequest("GET", "/", strings.NewReader(""))

	w := httptest.NewRecorder()
	c := &PipelineContext[NoState]{}

	ts := NewMockTemplateStore(t)
	ts.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything).Return(wantInnerError)
	as := NewMockAuthService(t)

	h := NewAuthHandler(ts, as)

	err := h.GetSignedOut(w, r, c)
	assert.EqualValues(t, wantAppError, err)
}

func buildRedirectPathWithParams(path, code, state string) string {
	params := []string{}
	if code != "" {
		params = append(params, "code="+code)
	}
	if state != "" {
		params = append(params, "state="+state)
	}

	if len(params) == 0 {
		return path
	}

	return path + "?" + strings.Join(params, "&")
}

func buildExpectedSetJwtCookie(tokenString string) *http.Cookie {
	return &http.Cookie{
		Name:        constants.IdentifierJwtCookie,
		Value:       tokenString,
		HttpOnly:    true,
		Secure:      true,
		Partitioned: true,
		SameSite:    http.SameSiteLaxMode,
		Path:        "/",
		MaxAge:      int(time.Second) * 60 * 60,
	}
}

func buildExpectedUnsetOidcStateCookie() *http.Cookie {
	return &http.Cookie{
		Name:     constants.IdentifierOidcStateCookie,
		Value:    "",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/sign-in-redirect",
		MaxAge:   -1,
	}
}

func buildExpectedUnsetAppJwtCookie() *http.Cookie {
	return &http.Cookie{
		Name:        constants.IdentifierJwtCookie,
		SameSite:    http.SameSiteLaxMode,
		Path:        "/",
		MaxAge:      -1,
		Expires:     time.Unix(0, 0).UTC(),
		Secure:      true,
		Partitioned: true,
		HttpOnly:    true,
	}
}
