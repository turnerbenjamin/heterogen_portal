package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/turnerbenjamin/heterogen_portal/internal/constants"
	"github.com/turnerbenjamin/heterogen_portal/internal/db"
	"github.com/turnerbenjamin/heterogen_portal/internal/services"
)

func TestParseJwtMiddleware_CallsNextWithUserNilWhenNoCookie(t *testing.T) {
	r := httptest.NewRequest("GET", "/", strings.NewReader(""))
	w := httptest.NewRecorder()
	c := &PipelineContext[UserState]{state: UserStateInit()}

	next := testHandler[UserState]{t: t}

	as := NewMockAuthService(t)
	mw := ParseJwtMiddleware[UserState](as)(next.handle)

	err := mw(w, r, c)

	assert.Nil(t, err)
	assert.Equal(t, 1, next.callCount)

	var wantUser *db.User = nil
	assert.Equal(t, wantUser, c.state.GetUser())
}

func TestParseJwtMiddleware_HandlesErrorsReturnedFromAuthService(t *testing.T) {
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

		as := NewMockAuthService(t)
		as.
			EXPECT().ParseUserJwtCookie(testCookieJwtToken).
			Return(&services.AppClaims{UserId: td.testUserId}, td.parseUserErr)
		as.
			EXPECT().RetrieveUserById(td.testUserId).
			Maybe().
			Return(&db.User{}, td.retrieveUserErr)

		mw := ParseJwtMiddleware[UserState](as)(next.handle)

		appErr := mw(w, r, c)
		c.logger.Log(context.Background(), slog.LevelInfo, "test")

		assert.Nil(t, appErr)
		assert.Equal(t, 1, next.callCount)

		var wantUser *db.User = nil
		assert.Equal(t, wantUser, c.state.GetUser())

		require.Equal(t, 1, len(w.Result().Cookies()))
		gotUnsetCookie := w.Result().Cookies()[0]
		gotUnsetCookie.Raw = ""
		gotUnsetCookie.RawExpires = ""

		wantUnsetCookie := buildExpectedUnsetAppJwtCookie()
		assert.EqualValues(t, wantUnsetCookie, gotUnsetCookie)
		assertLogsContain(
			t,
			logSink,
			map[string]any{
				constants.SlogKeyNonFatalErrParseAppJWT: td.wantErrorMessage,
			},
		)
	}
}

func TestParseJwtMiddleware_SetsUserOnContextOnSuccessfulParse(t *testing.T) {
	testCookieJwtToken := "some-jwt-token"
	testUserId := "test-user-id"
	testUser := &db.User{Id: testUserId}

	r := httptest.NewRequest("GET", "/", strings.NewReader(""))
	r.AddCookie(buildExpectedSetJwtCookie(testCookieJwtToken))

	w := httptest.NewRecorder()
	c := &PipelineContext[UserState]{state: UserStateInit()}
	next := testHandler[UserState]{t: t}

	as := NewMockAuthService(t)
	as.
		EXPECT().ParseUserJwtCookie(testCookieJwtToken).
		Return(&services.AppClaims{UserId: testUserId}, nil)

	as.
		EXPECT().RetrieveUserById(testUserId).
		Return(testUser, nil)

	mw := ParseJwtMiddleware[UserState](as)(next.handle)

	err := mw(w, r, c)

	assert.Nil(t, err)
	assert.Equal(t, 1, next.callCount)
	assert.Equal(t, testUser, c.state.GetUser())
}

func TestRequireSignInMiddleware_RedirectsUserToSignInWhenUserIsNil(t *testing.T) {
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
			Quoted:   true,
		}

		r := httptest.NewRequest("GET", "/", strings.NewReader(""))
		if td.isHtmx {
			r.Header.Set(constants.HxRequestHeaderRequest, "true")
		}

		w := httptest.NewRecorder()
		c := &PipelineContext[UserState]{state: UserStateInit()}

		as := NewMockAuthService(t)
		as.
			EXPECT().BuildSignInRedirectRequest(mock.Anything).
			Return(td.signInRedirectReq, nil)

		testHandler := &testHandler[UserState]{t: t}
		mw := RequireSignInMiddleware[UserState](as)(testHandler.handle)
		err := mw(w, r, c)

		assert.Nil(t, err)
		assert.Equal(t, wantCode, w.Code)
		assert.Equal(t, wantPath, w.Result().Header.Get(td.redirectKey))

		cookies := w.Result().Cookies()
		assert.Equal(t, 1, len(cookies))

		gotCookie := cookies[0]
		gotCookie.Raw = ""

		assert.EqualValues(t, wantCookie, gotCookie)
		assert.Equal(t, wantTestHandlerCallCount, testHandler.callCount)
	}
}

func TestRequireSignInMiddleware_HandlesAuthServicesErrors(t *testing.T) {
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

	as := NewMockAuthService(t)
	as.
		EXPECT().BuildSignInRedirectRequest(mock.Anything).
		Return(nil, wantInnerError)

	testHandler := &testHandler[UserState]{t: t}
	mw := RequireSignInMiddleware[UserState](as)(testHandler.handle)
	err := mw(w, r, c)

	assert.EqualValues(t, wantAppError, err)
	assert.Equal(t, wantTestHandlerCallCount, testHandler.callCount)
}

func TestRequireSignInMiddleware_CallsNextWhenUserIsNotNil(t *testing.T) {
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

	as := NewMockAuthService(t)

	testHandler := &testHandler[UserState]{t: t}
	mw := RequireSignInMiddleware[UserState](as)(testHandler.handle)
	err := mw(w, r, c)

	assert.Nil(t, err)
	assert.Equal(t, wantTestHandlerCallCount, testHandler.callCount)
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

func assertLogsContain(t testing.TB, logSink *bytes.Buffer, wantAttributes map[string]any) {
	t.Helper()

	gotLogAttributes := map[string]any{}
	err := json.Unmarshal(logSink.Bytes(), &gotLogAttributes)
	require.NoError(t, err)

	for k, v := range wantAttributes {
		gotV, ok := gotLogAttributes[k]
		require.Equal(t, true, ok)

		if v != mock.Anything {
			assert.Equal(t, v, gotV)
		}
	}
}
