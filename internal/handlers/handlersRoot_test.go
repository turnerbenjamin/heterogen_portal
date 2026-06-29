package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/turnerbenjamin/heterogen_portal/internal/constants"
	"github.com/turnerbenjamin/heterogen_portal/internal/db"
	"github.com/turnerbenjamin/heterogen_portal/internal/services"
	"github.com/turnerbenjamin/heterogen_portal/internal/templates"
	"github.com/turnerbenjamin/heterogen_portal/internal/testhelpers"
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

		ts := &mockTemplateStore{t: t, returns: nil}
		h := GetRootHandler(ts)

		err := h(w, r, c)

		AssertAppErrorNil(t, err)
		testhelpers.AssertIntEqual(t, w.Code, td.wantCode)
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

		ts := &mockTemplateStore{t: t, returns: nil}
		h := GetRootHandler(ts)

		err := h(w, r, c)

		AssertAppErrorNil(t, err)
		testhelpers.AssertIntEqual(t, w.Code, wantStatusCode)
		testhelpers.AssertIntEqual(t, len(ts.calls), 1)

		executeCall := ts.calls[0]
		testhelpers.AssertEqual(t, executeCall.templateId, wantTemplate)
		testhelpers.AssertEqual(t, executeCall.data.PageConfig.ContentOnly, td.wantContentOnlyValue)
		testhelpers.AssertEqual(t, executeCall.data.Data, wantState)
		testhelpers.AssertStringEqual(t, executeCall.data.PageConfig.Title, wantPageTitle)
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

	ts := &mockTemplateStore{t: t, returns: wantInnerError}
	h := GetRootHandler(ts)

	err := h(w, r, c)

	AssertAppErrorEqual(t, err, wantAppError)
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

		ts := &mockTemplateStore{t: t}
		as := &mockAuthService{t: t}
		h := GetSignInRedirectHandler(ts, as)

		err := h(w, r, c)

		AssertAppErrorEqual(t, err, wantAppError)
		if td.HasOidcStateCookie {
			wantCookie := buildExpectedUnsetOidcStateCookie()

			cookies := w.Result().Cookies()
			testhelpers.AssertIntEqual(t, len(cookies), 1)

			gotCookie := cookies[0]
			testhelpers.AssertCookieEqual(t, gotCookie, wantCookie)
		}
	}
}

func TestGetSignInRedirectHandler_HandlesAuthServiceErrors(t *testing.T) {
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

	ts := &mockTemplateStore{t: t}
	as := &mockAuthService{
		t:                          t,
		authenticateUserReturnsErr: wantInnerError,
	}
	h := GetSignInRedirectHandler(ts, as)

	err := h(w, r, c)

	AssertAppErrorEqual(t, err, wantAppError)
}

func TestGetSignInRedirectHandler_SetsAppCookieAndRedirectsUser(t *testing.T) {
	wantAppToken := "some-app-token"
	wantStatusCode := http.StatusSeeOther

	wantSetAppJwtCookie := buildExpectedSetJwtCookie(wantAppToken)
	wantClearOidcStateCookie := buildExpectedUnsetOidcStateCookie()

	testData := []struct {
		isHtmx               bool
		redirectKey          string
		wantRedirectPath     string
		codeValue            string
		stateValue           string
		signedOidcStateValue string
		appToken             string
	}{
		{
			isHtmx:               true,
			redirectKey:          "HX-Redirect",
			wantRedirectPath:     "/some-protected-endpoint",
			codeValue:            "some-code",
			stateValue:           "some-state-value",
			signedOidcStateValue: "some-oidc-state-value",
			appToken:             "an-app-token",
		},
		{
			isHtmx:               false,
			redirectKey:          "Location",
			wantRedirectPath:     "/some-other-endpoint",
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

		ts := &mockTemplateStore{t: t}
		as := &mockAuthService{
			t: t,
			authenticateUserReturnsResp: &services.AuthenticateUserResponse{
				AppToken:      wantAppToken,
				RequestedPath: td.wantRedirectPath,
			},
		}
		h := GetSignInRedirectHandler(ts, as)

		err := h(w, r, c)

		AssertAppErrorNil(t, err)
		testhelpers.AssertIntEqual(t, w.Code, wantStatusCode)
		testhelpers.AssertStringEqual(
			t,
			w.Result().Header.Get(td.redirectKey),
			td.wantRedirectPath,
		)

		testhelpers.AssertIntEqual(t, len(as.authenticateUserCallArgs), 1)
		authenticateUserCall := as.authenticateUserCallArgs[0]

		testhelpers.AssertEqual(t, authenticateUserCall.ctx, r.Context())
		testhelpers.AssertStringEqual(t, authenticateUserCall.authorisationCode, td.codeValue)
		testhelpers.AssertStringEqual(t, authenticateUserCall.returnedState, td.stateValue)
		testhelpers.AssertStringEqual(t, authenticateUserCall.signedOidcState, td.signedOidcStateValue)

		cookies := w.Result().Cookies()
		testhelpers.AssertIntEqual(t, len(cookies), 2)

		gotClearOidcStateCookie := cookies[0]
		gotSetAppJwtCookie := cookies[1]

		testhelpers.AssertCookieEqual(t, gotClearOidcStateCookie, wantClearOidcStateCookie)
		testhelpers.AssertCookieEqual(t, gotSetAppJwtCookie, wantSetAppJwtCookie)
	}
}

func TestGetSignOutHandler_UnsetsAppJwtCookieAndRedirectsUserToSignOut(t *testing.T) {
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

		as := &mockAuthService{
			t:                                     t,
			buildSignOutRedirectRequestReturnsReq: td.redirectValue,
		}

		h := GetSignOutHandler(as)
		err := h(w, r, c)

		AssertAppErrorNil(t, err)
		testhelpers.AssertIntEqual(t, w.Result().StatusCode, http.StatusSeeOther)
		testhelpers.AssertStringEqual(
			t,
			w.Result().Header.Get(td.redirectKey),
			td.redirectValue,
		)

		cookies := w.Result().Cookies()
		testhelpers.AssertIntEqual(t, len(cookies), 1)

		gotUnsetAppJwtCookie := cookies[0]
		wantUnsetAppJwtCookie := buildExpectedUnsetAppJwtCookie()
		testhelpers.AssertCookieEqual(t, gotUnsetAppJwtCookie, wantUnsetAppJwtCookie)
	}
}

func TestGetSignedOutHandler_ReturnsSignedOutPage(t *testing.T) {
	t.Parallel()

	wantStatusCode := http.StatusOK
	wantTemplate := templates.TmplPageUserSignedOut
	wantPageTitle := "HETEROGEN | SIGNED-OUT"
	wantContentOnlyValue := false

	r := httptest.NewRequest("GET", "/", strings.NewReader(""))

	w := httptest.NewRecorder()
	c := &PipelineContext[NoState]{}

	ts := &mockTemplateStore{t: t, returns: nil}
	h := GetSignedOutHandler(ts)

	err := h(w, r, c)

	AssertAppErrorNil(t, err)
	testhelpers.AssertIntEqual(t, w.Code, wantStatusCode)
	testhelpers.AssertIntEqual(t, len(ts.calls), 1)

	executeCall := ts.calls[0]
	testhelpers.AssertEqual(t, executeCall.templateId, wantTemplate)
	testhelpers.AssertEqual(t, executeCall.data.PageConfig.ContentOnly, wantContentOnlyValue)
	testhelpers.AssertStringEqual(t, executeCall.data.PageConfig.Title, wantPageTitle)
}

func TestGetSignedOutHandler_ReturnsErrIfHtmxRequest(t *testing.T) {
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

	ts := &mockTemplateStore{t: t, returns: nil}
	h := GetSignedOutHandler(ts)

	err := h(w, r, c)
	AssertAppErrorEqual(t, err, wantAppError)
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

	ts := &mockTemplateStore{t: t, returns: wantInnerError}
	h := GetSignedOutHandler(ts)

	err := h(w, r, c)
	AssertAppErrorEqual(t, err, wantAppError)
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

type mockAuthServiceAuthenticateUserArgs struct {
	ctx               context.Context
	authorisationCode string
	returnedState     string
	signedOidcState   string
}

type mockAuthService struct {
	t testing.TB

	parseUserJwtCookieCallArgs      []string
	parseUserJwtCookieReturnsClaims *services.AppClaims
	parseUserJwtCookieReturnsErr    error

	retrieveUserByIdCallArgs    []string
	retrieveUserByIdReturnsUser *db.User
	retrieveUserByIdReturnsErr  error

	buildSignInRedirectRequestCallArgs   []string
	buildSignInRedirectRequestReturnsReq *services.SignInRedirectRequest
	buildSignInRedirectRequestReturnsErr error

	buildSignOutRedirectRequestCallCount  int
	buildSignOutRedirectRequestReturnsReq string

	authenticateUserCallArgs    []mockAuthServiceAuthenticateUserArgs
	authenticateUserReturnsResp *services.AuthenticateUserResponse
	authenticateUserReturnsErr  error
}

func (s *mockAuthService) ParseUserJwtCookie(
	tokenString string,
) (*services.AppClaims, error) {
	s.t.Helper()

	s.parseUserJwtCookieCallArgs = append(s.parseUserJwtCookieCallArgs, tokenString)
	return s.parseUserJwtCookieReturnsClaims, s.parseUserJwtCookieReturnsErr
}

func (s *mockAuthService) RetrieveUserById(
	userId string,
) (*db.User, error) {
	s.t.Helper()

	s.retrieveUserByIdCallArgs = append(s.retrieveUserByIdCallArgs, userId)
	return s.retrieveUserByIdReturnsUser, s.retrieveUserByIdReturnsErr
}

func (s *mockAuthService) BuildSignInRedirectRequest(requestedPath string) (*services.SignInRedirectRequest, error) {
	s.t.Helper()

	s.buildSignInRedirectRequestCallArgs = append(
		s.buildSignInRedirectRequestCallArgs,
		requestedPath,
	)
	return s.buildSignInRedirectRequestReturnsReq, s.buildSignInRedirectRequestReturnsErr
}

func (s *mockAuthService) BuildSignOutRedirectRequest() string {
	s.t.Helper()

	s.buildSignOutRedirectRequestCallCount = s.buildSignOutRedirectRequestCallCount + 1
	return s.buildSignOutRedirectRequestReturnsReq
}

func (s *mockAuthService) AuthenticateUser(
	ctx context.Context,
	authorisationCode string,
	returnedState string,
	signedOidcState string,
) (resp *services.AuthenticateUserResponse, err error) {
	s.t.Helper()

	s.authenticateUserCallArgs = append(
		s.authenticateUserCallArgs,
		mockAuthServiceAuthenticateUserArgs{
			ctx:               ctx,
			authorisationCode: authorisationCode,
			returnedState:     returnedState,
			signedOidcState:   signedOidcState,
		},
	)
	return s.authenticateUserReturnsResp, s.authenticateUserReturnsErr
}
