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
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/turnerbenjamin/heterogen_portal/internal/constants"
	"github.com/turnerbenjamin/heterogen_portal/internal/db"
	"github.com/turnerbenjamin/heterogen_portal/internal/testhelpers"
)

var testUser = db.User{
	Id:           "anid",
	Oid:          "oid",
	GivenName:    "given_name",
	FamilyName:   "family_name",
	UserName:     "user_name",
	EmailAddress: "email_address",
}

var testPortalClaims = PortalTokenClaims{
	Oid:          testUser.Oid,
	GivenName:    testUser.GivenName,
	FamilyName:   testUser.FamilyName,
	UserName:     testUser.UserName,
	EmailAddress: testUser.EmailAddress,
}

func TestPostSignIn_CallsValidatePortalTokenCorrectly(t *testing.T) {
	t.Parallel()

	tokenValidator := &mockTokenValidator{t: t, returnsClaims: &testPortalClaims}
	userRepo := &mockUserRepo{
		t: t,
		upsertUserResponse: &userRepoGetUserResponse{
			user: &testUser,
		},
	}
	tokenSigner := &mockTokenSignerAndParser{t: t}
	templateStore := &mockTemplateStore{t: t}

	h := PostSignInHandler(tokenValidator, tokenSigner, templateStore, userRepo)

	r := httptest.NewRequest("POST", "/sign-in", strings.NewReader(""))
	testBearerToken := "bearer my_jwt_token"
	r.Header.Set("Authorization", testBearerToken)

	w := httptest.NewRecorder()

	err := h(w, r, &PipelineContext[NoState]{})

	AssertAppErrorNil(t, err)
	testhelpers.AssertIntEqual(t, len(tokenValidator.calls), 1)

	validatorCall := tokenValidator.calls[0]
	testhelpers.AssertEqual(t, validatorCall.ctx, r.Context())
	testhelpers.AssertStringEqual(t, validatorCall.tokenString, testBearerToken)
}

func TestPostSignIn_ReturnsUnauthorisedWhenValidatePortalTokenReturnsAnError(t *testing.T) {
	t.Parallel()

	validatorError := errors.New("some validator error")
	wantAppErr := &AppError{
		Code:       http.StatusUnauthorized,
		ToastError: constants.ErrMsgUnauthorised,
		PageErrors: []string{constants.ErrMsgUnauthorised},
		innerError: validatorError,
	}

	tokenValidator := &mockTokenValidator{t: t, returnsError: validatorError}
	userRepo := &mockUserRepo{t: t}
	tokenSigner := &mockTokenSignerAndParser{t: t}
	templateStore := &mockTemplateStore{t: t}

	h := PostSignInHandler(tokenValidator, tokenSigner, templateStore, userRepo)

	r := httptest.NewRequest("POST", "/sign-in", strings.NewReader(""))
	r.Header.Set("Authorization", "bearer my_jwt_token")

	w := httptest.NewRecorder()

	gotAppErr := h(w, r, &PipelineContext[NoState]{})

	AssertAppErrorEqual(t, gotAppErr, wantAppErr)
}

func TestPostSignIn_ShouldCallUpsertWithUserData(t *testing.T) {
	t.Parallel()

	tokenValidator := &mockTokenValidator{t: t, returnsClaims: &testPortalClaims}
	userRepo := &mockUserRepo{
		t: t,
		upsertUserResponse: &userRepoGetUserResponse{
			user: &testUser,
		},
	}
	tokenSigner := &mockTokenSignerAndParser{t: t}
	templateStore := &mockTemplateStore{t: t}

	h := PostSignInHandler(tokenValidator, tokenSigner, templateStore, userRepo)

	r := httptest.NewRequest("POST", "/sign-in", strings.NewReader(""))
	r.Header.Set("Authorization", "bearer my_jwt_token")

	w := httptest.NewRecorder()

	err := h(w, r, &PipelineContext[NoState]{})

	AssertAppErrorNil(t, err)
	testhelpers.AssertIntEqual(t, len(userRepo.upsertUserCalls), 1)

	gotUpsertCall := userRepo.upsertUserCalls[0]
	testhelpers.AssertEqual(t, gotUpsertCall.ctx, r.Context())
	testhelpers.AssertStringEqual(t, gotUpsertCall.oid, testUser.Oid)
	testhelpers.AssertStringEqual(t, gotUpsertCall.givenName, testUser.GivenName)
	testhelpers.AssertStringEqual(t, gotUpsertCall.familyName, testUser.FamilyName)
	testhelpers.AssertStringEqual(t, gotUpsertCall.userName, testUser.UserName)
	testhelpers.AssertStringEqual(t, gotUpsertCall.emailAddress, testUser.EmailAddress)
}

func TestPostSignIn_ReturnsServerErrorWhenUpsertUserReturnsAnError(t *testing.T) {
	t.Parallel()

	upsertError := errors.New("some upsert error")
	wantAppErr := &AppError{
		Code:       http.StatusInternalServerError,
		ToastError: constants.ErrMsgInternalServerError,
		PageErrors: []string{constants.ErrMsgInternalServerError},
		innerError: upsertError,
	}

	tokenValidator := &mockTokenValidator{t: t, returnsClaims: &testPortalClaims}
	userRepo := &mockUserRepo{
		t: t,
		upsertUserResponse: &userRepoGetUserResponse{
			user: nil,
			err:  upsertError,
		},
	}
	tokenSigner := &mockTokenSignerAndParser{t: t}
	templateStore := &mockTemplateStore{t: t}

	h := PostSignInHandler(tokenValidator, tokenSigner, templateStore, userRepo)

	r := httptest.NewRequest("POST", "/sign-in", strings.NewReader(""))
	r.Header.Set("Authorization", "bearer my_jwt_token")

	w := httptest.NewRecorder()

	gotAppErr := h(w, r, &PipelineContext[NoState]{})

	AssertAppErrorEqual(t, gotAppErr, wantAppErr)
}

func TestPostSignIn_ReturnsServerErrorWhenTokenSignerReturnsAnError(t *testing.T) {
	t.Parallel()

	signError := errors.New("some sign error")
	wantAppErr := &AppError{
		Code:       http.StatusInternalServerError,
		ToastError: constants.ErrMsgInternalServerError,
		PageErrors: []string{constants.ErrMsgInternalServerError},
		innerError: signError,
	}

	tokenValidator := &mockTokenValidator{t: t, returnsClaims: &testPortalClaims}
	userRepo := &mockUserRepo{
		t: t,
		upsertUserResponse: &userRepoGetUserResponse{
			user: &testUser,
			err:  nil,
		},
	}
	tokenSigner := &mockTokenSignerAndParser{t: t, signReturnsError: signError}
	templateStore := &mockTemplateStore{t: t}

	h := PostSignInHandler(tokenValidator, tokenSigner, templateStore, userRepo)

	r := httptest.NewRequest("POST", "/sign-in", strings.NewReader(""))
	r.Header.Set("Authorization", "bearer my_jwt_token")

	w := httptest.NewRecorder()

	gotAppErr := h(w, r, &PipelineContext[NoState]{})

	AssertAppErrorEqual(t, gotAppErr, wantAppErr)
}

func TestPostSignIn_SetsCookieWithJwtToken(t *testing.T) {
	t.Parallel()

	testTokenString := "test_jwt_token_string"
	wantCookie := &http.Cookie{
		Name:        constants.IdentifierJwtCookie,
		Value:       testTokenString,
		SameSite:    http.SameSiteStrictMode,
		MaxAge:      int((time.Hour * 24) / time.Second),
		Secure:      true,
		Partitioned: true,
		HttpOnly:    true,
	}

	tokenValidator := &mockTokenValidator{t: t, returnsClaims: &testPortalClaims}
	userRepo := &mockUserRepo{
		t: t,
		upsertUserResponse: &userRepoGetUserResponse{
			user: &testUser,
		},
	}
	tokenSigner := &mockTokenSignerAndParser{t: t, signReturnsString: testTokenString}
	templateStore := &mockTemplateStore{t: t}

	h := PostSignInHandler(tokenValidator, tokenSigner, templateStore, userRepo)

	r := httptest.NewRequest("POST", "/sign-in", strings.NewReader(""))
	r.Header.Set("Authorization", "bearer my_jwt_token")

	w := httptest.NewRecorder()

	err := h(w, r, &PipelineContext[NoState]{})

	AssertAppErrorNil(t, err)

	var gotCookie *http.Cookie = nil
	for _, c := range w.Result().Cookies() {
		if c.Name == constants.IdentifierJwtCookie {
			gotCookie = c
			break
		}
	}

	testhelpers.AssertNotNil(t, gotCookie, wantCookie)
	testhelpers.AssertStringEqual(t, gotCookie.Value, wantCookie.Value)
	testhelpers.AssertEqual(t, gotCookie.SameSite, wantCookie.SameSite)
	testhelpers.AssertIntEqual(t, gotCookie.MaxAge, wantCookie.MaxAge)
	testhelpers.AssertEqual(t, gotCookie.Secure, wantCookie.Secure)
	testhelpers.AssertEqual(t, gotCookie.Partitioned, wantCookie.Partitioned)
	testhelpers.AssertEqual(t, gotCookie.HttpOnly, wantCookie.HttpOnly)
}

func TestPostSignIn_RedirectsUserToRoot(t *testing.T) {
	t.Parallel()

	tokenValidator := &mockTokenValidator{t: t, returnsClaims: &testPortalClaims}
	userRepo := &mockUserRepo{
		t: t,
		upsertUserResponse: &userRepoGetUserResponse{
			user: &testUser,
		},
	}
	tokenSigner := &mockTokenSignerAndParser{t: t, signReturnsString: "a_token"}
	templateStore := &mockTemplateStore{t: t}

	h := PostSignInHandler(tokenValidator, tokenSigner, templateStore, userRepo)

	wantStatusCode := http.StatusSeeOther
	wantRedirectPath := "/"
	testData := []struct {
		isHtmxRequest bool
		redirectKey   string
	}{
		{isHtmxRequest: true, redirectKey: "HX-Redirect"},
		{isHtmxRequest: false, redirectKey: "Location"},
	}

	for _, td := range testData {

		r := httptest.NewRequest("POST", "/sign-in", strings.NewReader(""))
		r.Header.Set("Authorization", "bearer my_jwt_token")
		if td.isHtmxRequest {
			r.Header.Set("HX-Request", "true")
		}

		w := httptest.NewRecorder()

		err := h(w, r, &PipelineContext[NoState]{})

		AssertAppErrorNil(t, err)
		testhelpers.AssertIntEqual(t, w.Code, wantStatusCode)
		testhelpers.AssertStringEqual(t, w.Result().Header.Get(td.redirectKey), wantRedirectPath)
	}
}

func TestParseJWTMiddleware_ShouldRetrieveUserAndUpdatePipelineContext(t *testing.T) {
	t.Parallel()

	nextHandlerCallCount := 0
	var userReceivedInHandler *db.User = nil

	nextHandler := &testAppHandler[UserState]{
		t: t,
		fn: func(
			r *http.Request,
			c *PipelineContext[UserState],
		) (request *http.Request, statusCode *int, response []byte, err *AppError) {
			nextHandlerCallCount = nextHandlerCallCount + 1
			userReceivedInHandler = c.state.User
			return r, nil, []byte("body"), nil
		},
	}

	tokenParser := &mockTokenSignerAndParser{
		t: t,
		parseReturnsToken: &jwt.Token{
			Claims: &mockClaims{
				t:              t,
				returnsSubject: "jwt_subject",
			},
		}}
	userRepo := &mockUserRepo{
		t: t,
		retrieveUserByIdResponse: &userRepoGetUserResponse{
			user: &testUser,
			err:  nil,
		},
	}

	r := httptest.NewRequest("POST", "/test", strings.NewReader(""))
	r.AddCookie(&http.Cookie{Name: constants.IdentifierJwtCookie})

	w := httptest.NewRecorder()
	c := &PipelineContext[UserState]{
		logger: slog.New(slog.NewJSONHandler(&bytes.Buffer{}, &slog.HandlerOptions{})),
		state:  &UserState{},
	}

	mw := NewParseJwtMiddleware(tokenParser, userRepo)
	h := mw(nextHandler.handle)
	err := h(w, r, c)

	AssertAppErrorNil(t, err)
	testhelpers.AssertIntEqual(t, nextHandlerCallCount, 1)
	testhelpers.AssertEqual(t, userReceivedInHandler, &testUser)
}

func TestParseJWTMiddleware_IsIndestuctable(t *testing.T) {
	t.Parallel()

	testData := []struct {
		cookieIdentifier         string
		parserError              error
		getClaimsErr             error
		subject                  string
		retrieveUserReturnsUser  *db.User
		retrieveUserReturnsError error
	}{
		// no cookie set
		{
			cookieIdentifier: "",
		},
		// incorrect cookie identifier
		{
			cookieIdentifier: constants.IdentifierJwtCookie + "x",
		},
		// token parser error
		{
			cookieIdentifier:        constants.IdentifierJwtCookie,
			parserError:             errors.New("parser-error"),
			subject:                 "some_subject",
			retrieveUserReturnsUser: &testUser,
		},
		// get claims from token error
		{
			cookieIdentifier:        constants.IdentifierJwtCookie,
			getClaimsErr:            errors.New("get-claims-error"),
			subject:                 "some_subject",
			retrieveUserReturnsUser: &testUser,
		},
		// no errors but empty subject
		{
			cookieIdentifier:        constants.IdentifierJwtCookie,
			parserError:             nil,
			subject:                 "",
			retrieveUserReturnsUser: &testUser,
		},
		// retrieve user by id error
		{
			cookieIdentifier:         constants.IdentifierJwtCookie,
			subject:                  "some_subject",
			retrieveUserReturnsUser:  &testUser,
			retrieveUserReturnsError: errors.New("retrieve-user-error"),
		},
		// retrieve user by id returns nil user
		{
			cookieIdentifier:        constants.IdentifierJwtCookie,
			subject:                 "some_subject",
			retrieveUserReturnsUser: nil,
		},
	}

	for _, td := range testData {
		wantLogs := map[string]any{}
		if td.parserError != nil {
			wantLogs[constants.SlogKeyNonFatalErrParseWithClaims] =
				td.parserError.Error()
		}

		if td.getClaimsErr != nil {
			wantLogs[constants.SlogKeyNonFatalErrClaimsGetSubject] =
				td.getClaimsErr.Error()
		}

		if td.retrieveUserReturnsError != nil {
			wantLogs[constants.SlogKeyNonFatalErrRetrieveUserById] =
				td.retrieveUserReturnsError.Error()
		}

		nextHandlerCallCount := 0
		nextHandler := &testAppHandler[UserState]{
			t: t,
			fn: func(
				r *http.Request,
				c *PipelineContext[UserState],
			) (request *http.Request, statusCode *int, response []byte, err *AppError) {
				nextHandlerCallCount = nextHandlerCallCount + 1
				return r, nil, []byte("body"), nil
			},
		}
		userRepo := &mockUserRepo{
			t: t,
			retrieveUserByIdResponse: &userRepoGetUserResponse{
				user: td.retrieveUserReturnsUser,
				err:  td.retrieveUserReturnsError,
			},
		}
		tokenParser := &mockTokenSignerAndParser{
			t:                 t,
			parseReturnsError: td.parserError,
			parseReturnsToken: &jwt.Token{
				Claims: &mockClaims{
					t:              t,
					returnsSubject: td.subject,
					returnsError:   td.getClaimsErr,
				},
			}}

		r := httptest.NewRequest("POST", "/test", strings.NewReader(""))
		if td.cookieIdentifier != "" {
			r.AddCookie(&http.Cookie{Name: td.cookieIdentifier})
		}

		logSink := &bytes.Buffer{}
		w := httptest.NewRecorder()
		c := &PipelineContext[UserState]{
			logger: slog.New(slog.NewJSONHandler(logSink, &slog.HandlerOptions{})),
			state:  &UserState{},
		}

		mw := NewParseJwtMiddleware(tokenParser, userRepo)
		h := mw(nextHandler.handle)
		err := h(w, r, c)

		// Should always continue to handler without returning an error
		AssertAppErrorNil(t, err)
		testhelpers.AssertIntEqual(t, nextHandlerCallCount, 1)

		// pipeline state user should always be nil
		if c.state.User != nil {
			t.Fatalf("got %v, but want nil", c.state.User)
		}

		// should log errors encountered
		if len(wantLogs) > 0 {
			c.logger.Info("")
			testhelpers.AssertSlogsContain(t, logSink.Bytes(), wantLogs)
		}

		// should unset cookie if set
		if td.cookieIdentifier == constants.IdentifierJwtCookie {
			assertJWTCookieUnset(t, w)
		}
	}
}

type validatePortalTokenCallArgs struct {
	ctx         context.Context
	tokenString string
}

type mockTokenValidator struct {
	t             testing.TB
	returnsClaims *PortalTokenClaims
	returnsError  error
	calls         []validatePortalTokenCallArgs
}

func (v *mockTokenValidator) ValidatePortalToken(
	ctx context.Context,
	tokenString string,
) (*PortalTokenClaims, error) {
	if v.calls == nil {
		v.calls = []validatePortalTokenCallArgs{}
	}
	v.calls = append(v.calls, validatePortalTokenCallArgs{
		ctx:         ctx,
		tokenString: tokenString,
	})

	return v.returnsClaims, v.returnsError
}

type upsertUserCallArgs struct {
	ctx          context.Context
	oid          string
	givenName    string
	familyName   string
	userName     string
	emailAddress string
}

type userRepoGetUserResponse struct {
	user *db.User
	err  error
}

type userRepoGetUserByIdOrOidCallArgs struct {
	id string
}

type mockUserRepo struct {
	t                         testing.TB
	upsertUserResponse        *userRepoGetUserResponse
	upsertUserCalls           []upsertUserCallArgs
	retrieveUserByIdResponse  *userRepoGetUserResponse
	retrieveUserByIdCalls     []userRepoGetUserByIdOrOidCallArgs
	retrieveUserByOidResponse *userRepoGetUserResponse
	retrieveUserByOidCalls    []userRepoGetUserByIdOrOidCallArgs
}

func (r *mockUserRepo) UpsertUser(
	ctx context.Context,
	oid, givenName, familyName, userName, emailAddress string,
) (*db.User, error) {
	r.t.Helper()
	if r.upsertUserResponse == nil {
		r.t.Fatal("upsert user called before set-up")
		return nil, nil
	}

	if r.upsertUserCalls == nil {
		r.upsertUserCalls = []upsertUserCallArgs{}
	}
	r.upsertUserCalls = append(r.upsertUserCalls, upsertUserCallArgs{
		ctx:          ctx,
		oid:          oid,
		givenName:    givenName,
		familyName:   familyName,
		userName:     userName,
		emailAddress: emailAddress,
	})

	return r.upsertUserResponse.user, r.upsertUserResponse.err
}

func (r *mockUserRepo) RetrieveUserById(id string) (*db.User, error) {
	r.t.Helper()
	if r.retrieveUserByIdResponse == nil {
		r.t.Fatal("retrieve user by id called before set-up")
		return nil, nil
	}

	if r.retrieveUserByIdCalls == nil {
		r.retrieveUserByIdCalls = []userRepoGetUserByIdOrOidCallArgs{}
	}
	r.retrieveUserByIdCalls = append(r.retrieveUserByIdCalls, userRepoGetUserByIdOrOidCallArgs{
		id: id,
	})

	return r.retrieveUserByIdResponse.user, r.retrieveUserByIdResponse.err
}

func (r *mockUserRepo) RetrieveUserByOid(id string) (*db.User, error) {
	r.t.Helper()
	if r.retrieveUserByOidResponse == nil {
		r.t.Fatal("retrieve user by oid called before set-up")
		return nil, nil
	}

	if r.retrieveUserByOidCalls == nil {
		r.retrieveUserByOidCalls = []userRepoGetUserByIdOrOidCallArgs{}
	}
	r.retrieveUserByOidCalls = append(r.retrieveUserByOidCalls, userRepoGetUserByIdOrOidCallArgs{
		id: id,
	})

	return r.retrieveUserByOidResponse.user, r.retrieveUserByOidResponse.err
}

type mocktokenSignerAndParserParseCallArgs struct {
	jwtString string
	claims    *jwt.RegisteredClaims
}

type mockTokenSignerAndParser struct {
	t                 *testing.T
	signCalls         []*jwt.Token
	signReturnsString string
	signReturnsError  error
	parseCalls        []mocktokenSignerAndParserParseCallArgs
	parseReturnsToken *jwt.Token
	parseReturnsError error
}

func (sp *mockTokenSignerAndParser) Sign(token *jwt.Token) (string, error) {
	if sp.signCalls == nil {
		sp.signCalls = []*jwt.Token{}
	}
	sp.signCalls = append(sp.signCalls, token)
	return sp.signReturnsString, sp.signReturnsError
}

func (sp *mockTokenSignerAndParser) ParseWithClaims(
	jwtString string,
	claims *jwt.RegisteredClaims,
) (*jwt.Token, error) {
	if sp.parseCalls == nil {
		sp.parseCalls = []mocktokenSignerAndParserParseCallArgs{}
	}
	sp.parseCalls = append(sp.parseCalls, mocktokenSignerAndParserParseCallArgs{
		jwtString: jwtString,
		claims:    claims,
	})
	return sp.parseReturnsToken, sp.parseReturnsError
}

type mockClaims struct {
	jwt.RegisteredClaims
	t              *testing.T
	returnsSubject string
	returnsError   error
}

func (c *mockClaims) GetSubject() (string, error) {
	c.t.Helper()

	return c.returnsSubject, c.returnsError
}
