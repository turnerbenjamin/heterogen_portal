package handlers

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/turnerbenjamin/heterogen_portal/internal/constants"
	"github.com/turnerbenjamin/heterogen_portal/internal/db"
	"github.com/turnerbenjamin/heterogen_portal/internal/services"
	"github.com/turnerbenjamin/heterogen_portal/internal/testhelpers"
)

func TestNewParseJwtMiddleware_CallsNextWithUserNilWhenNoCookie(t *testing.T) {
	r := httptest.NewRequest("GET", "/", strings.NewReader(""))
	w := httptest.NewRecorder()
	c := &PipelineContext[UserState]{state: UserStateInit()}

	next := testHandler[UserState]{t: t}

	as := &mockAuthService{t: t}
	mw := NewParseJwtMiddleware(as)(next.handle)

	err := mw(w, r, c)

	AssertAppErrorNil(t, err)
	testhelpers.AssertIntEqual(t, next.callCount, 1)

	var wantUser *db.User = nil
	testhelpers.AssertEqual(t, c.state.GetUser(), wantUser)
}

func TestNewParseJwtMiddleware_HandlesErrorsReturnedFromAuthService(t *testing.T) {
	testData := []struct {
		testUserId       string
		parseUserErr     error
		retrieveUserErr  error
		wantErrorMessage string
	}{
		{
			testUserId:       "a-test-user-id",
			parseUserErr:     errors.New("test-parse-user-error"),
			wantErrorMessage: "test-parse-user-error",
		},
		{
			testUserId:       "another-test-user-id",
			retrieveUserErr:  errors.New("test-retrieve-user-error"),
			wantErrorMessage: "test-retrieve-user-error",
		},
	}

	for _, td := range testData {
		testCookieJwtToken := "some-jwt-token"

		r := httptest.NewRequest("GET", "/", strings.NewReader(""))
		r.AddCookie(buildExpectedSetJwtCookie(testCookieJwtToken))

		w := httptest.NewRecorder()

		logSink := &bytes.Buffer{}
		c := &PipelineContext[UserState]{
			state:  UserStateInit(),
			logger: slog.New(slog.NewJSONHandler(logSink, &slog.HandlerOptions{})),
		}

		next := testHandler[UserState]{t: t}

		as := &mockAuthService{
			t:                               t,
			parseUserJwtCookieReturnsClaims: &services.AppClaims{UserId: td.testUserId},
			parseUserJwtCookieReturnsErr:    td.parseUserErr,
			retrieveUserByIdReturnsUser:     &db.User{},
			retrieveUserByIdReturnsErr:      td.retrieveUserErr,
		}
		mw := NewParseJwtMiddleware(as)(next.handle)

		err := mw(w, r, c)
		c.logger.Log(context.Background(), slog.LevelInfo, "test")

		AssertAppErrorNil(t, err)
		testhelpers.AssertIntEqual(t, next.callCount, 1)

		var wantUser *db.User = nil
		testhelpers.AssertEqual(t, c.state.GetUser(), wantUser)

		testhelpers.AssertIntEqual(t, len(w.Result().Cookies()), 1)
		gotUnsetCookie := w.Result().Cookies()[0]
		wantUnsetCookie := buildExpectedUnsetAppJwtCookie()

		testhelpers.AssertCookieEqual(t, gotUnsetCookie, wantUnsetCookie)

		testhelpers.AssertSlogsContain(
			t,
			logSink.Bytes(),
			map[string]any{
				constants.SlogKeyNonFatalErrParseAppJWT: td.wantErrorMessage,
			},
		)
	}
}

func TestNewParseJwtMiddleware_SetsUserOnContextOnSuccessfulParse(t *testing.T) {
	testCookieJwtToken := "some-jwt-token"
	testUserId := "test-user-id"
	testUser := &db.User{Id: testUserId}

	r := httptest.NewRequest("GET", "/", strings.NewReader(""))
	r.AddCookie(buildExpectedSetJwtCookie(testCookieJwtToken))

	w := httptest.NewRecorder()
	c := &PipelineContext[UserState]{state: UserStateInit()}
	next := testHandler[UserState]{t: t}

	as := &mockAuthService{
		t:                               t,
		parseUserJwtCookieReturnsClaims: &services.AppClaims{UserId: testUserId},
		retrieveUserByIdReturnsUser:     testUser,
	}

	mw := NewParseJwtMiddleware(as)(next.handle)

	err := mw(w, r, c)

	AssertAppErrorNil(t, err)
	testhelpers.AssertIntEqual(t, next.callCount, 1)
	testhelpers.AssertEqual(t, c.state.GetUser(), testUser)
}

func TestNewRequireSignIn_RedirectsUserToSignInWhenUserIsNil(t *testing.T) {
	t.Parallel()

	testData := []struct {
		isHtmx            bool
		redirectKey       string
		signInRedirectReq *services.SignInRedirectRequest
	}{
		{
			isHtmx:      true,
			redirectKey: "HX-Redirect",
			signInRedirectReq: &services.SignInRedirectRequest{
				Url:             "https://test-redirect-link-1/sign-in",
				SignedOidcState: "some state",
			},
		},
		{
			isHtmx:      false,
			redirectKey: "Location",
			signInRedirectReq: &services.SignInRedirectRequest{
				Url:             "https://test-redirect-link-2/sign-in",
				SignedOidcState: "more state",
			},
		},
	}

	for _, td := range testData {
		wantCode := 303
		wantPath := td.signInRedirectReq.Url
		wantTestHandlerCallCount := 0
		wantCookie := &http.Cookie{
			Name:     constants.IdentifierOidcStateCookie,
			Value:    td.signInRedirectReq.SignedOidcState,
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
			Path:     "/sign-in-redirect",
			MaxAge:   600,
		}

		r := httptest.NewRequest("GET", "/", strings.NewReader(""))
		if td.isHtmx {
			r.Header.Set(constants.HxRequestHeaderRequest, "true")
		}

		w := httptest.NewRecorder()
		c := &PipelineContext[UserState]{state: UserStateInit()}

		as := &mockAuthService{
			t:                                    t,
			buildSignInRedirectRequestReturnsReq: td.signInRedirectReq,
			buildSignInRedirectRequestReturnsErr: nil,
		}

		testHandler := &testHandler[UserState]{t: t}
		mw := NewRequireSignInMiddleware(as)(testHandler.handle)
		err := mw(w, r, c)

		AssertAppErrorNil(t, err)
		testhelpers.AssertIntEqual(t, w.Code, wantCode)
		testhelpers.AssertStringEqual(
			t,
			w.Result().Header.Get(td.redirectKey),
			wantPath,
		)
		cookies := w.Result().Cookies()

		testhelpers.AssertIntEqual(t, len(cookies), 1)
		gotCookie := cookies[0]
		testhelpers.AssertCookieEqual(t, gotCookie, wantCookie)
		testhelpers.AssertIntEqual(t, testHandler.callCount, wantTestHandlerCallCount)
	}
}

func TestNewRequireSignInMiddleware_HandlesAuthServicesErrors(t *testing.T) {
	t.Parallel()

	wantInnerError := errors.New("test auth services err")
	wantAppError := &AppError{
		Code:       http.StatusInternalServerError,
		ToastError: constants.ErrMsgInternalServerError,
		PageErrors: []string{constants.ErrMsgInternalServerError},
		innerError: wantInnerError,
	}
	wantTestHandlerCallCount := 0

	r := httptest.NewRequest("GET", "/", strings.NewReader(""))
	w := httptest.NewRecorder()
	c := &PipelineContext[UserState]{state: UserStateInit()}

	as := &mockAuthService{
		t:                                    t,
		buildSignInRedirectRequestReturnsErr: wantInnerError,
	}

	testHandler := &testHandler[UserState]{t: t}
	mw := NewRequireSignInMiddleware(as)(testHandler.handle)
	err := mw(w, r, c)

	AssertAppErrorEqual(t, err, wantAppError)
	testhelpers.AssertIntEqual(t, testHandler.callCount, wantTestHandlerCallCount)
}

func TestNewRequireSignInMiddleware_CallsNextWhenUserIsNotNil(t *testing.T) {
	t.Parallel()
	wantTestHandlerCallCount := 1

	r := httptest.NewRequest("GET", "/", strings.NewReader(""))
	w := httptest.NewRecorder()
	c := &PipelineContext[UserState]{state: UserStateInit()}
	c.state.SetUser(&db.User{
		Id:           "123",
		GivenName:    "Test",
		FamilyName:   "User",
		EmailAddress: "test@user.com",
	})

	as := &mockAuthService{t: t}

	testHandler := &testHandler[UserState]{t: t}
	mw := NewRequireSignInMiddleware(as)(testHandler.handle)
	err := mw(w, r, c)

	AssertAppErrorNil(t, err)
	testhelpers.AssertIntEqual(t, testHandler.callCount, wantTestHandlerCallCount)
}

type testHandler[T any] struct {
	t          *testing.T
	returnsErr *AppError
	callCount  int
}

func (m *testHandler[T]) handle(_ http.ResponseWriter, _ *http.Request, _ *PipelineContext[T]) *AppError {
	m.t.Helper()

	m.callCount = m.callCount + 1
	return m.returnsErr
}
