package services

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/golang-jwt/jwt/v5"
	"github.com/turnerbenjamin/heterogen_portal/internal/db"
	"github.com/turnerbenjamin/heterogen_portal/internal/etc"
	"golang.org/x/oauth2"
)

// UserRepo performs database operations on the user table
type UserRepo interface {
	UpsertUser(
		ctx context.Context,
		oid, givenName, familyName, userName, emailAddress string,
	) (*db.User, error)
	RetrieveUserById(id string) (*db.User, error)
	RetrieveUserByOid(id string) (*db.User, error)
}

// JWTSignerAndParser can sign and parse JWT tokens
type JWTSignerAndParser interface {
	ParseWithClaims(
		tokenString string,
		claims jwt.Claims,
		keyFunc jwt.Keyfunc,
		parserOptions ...jwt.ParserOption,
	) (*jwt.Token, error)
	Sign(token *jwt.Token, privateKey []byte) (string, error)
}

type SignInRedirectRequest struct {
	Url             string
	SignedOidcState string
}

type AuthenticateUserResponse struct {
	AppToken string
	RequestedPath  string
}

type portalTokenClaims struct {
	Nonce        string `json:"nonce"`
	Oid          string `json:"oid"`
	GivenName    string `json:"given_name"`
	FamilyName   string `json:"family_name"`
	UserName     string `json:"name"`
	EmailAddress string `json:"email"`
}

type oidcState struct {
	State        string
	Nonce        string
	CodeVerifier string
	RequestedPath      string
}

type AppClaims struct {
	IdToken   string
	UserId    string
	UserOid   string
	UserEmail string
}

type appClaims struct {
	jwt.RegisteredClaims
	IdToken   string `json:"id_token"`
	UserId    string `json:"user_id"`
	UserOid   string `json:"user_oid"`
	UserEmail string `json:"email"`
}

type authService struct {
	appSettings        *etc.AppSettings
	userRepo           UserRepo
	jwtSignerAndParser JWTSignerAndParser

	oauthConfig *oauth2.Config
	verifier    *oidc.IDTokenVerifier
}

func NewAuthService(
	ctx context.Context,
	appSettings *etc.AppSettings,
	httpClient *http.Client,
	userRepo UserRepo,
	jwtSignerAndParser JWTSignerAndParser,
) (*authService, error) {
	oidcCtx := oidc.ClientContext(ctx, httpClient)

	provider, err := oidc.NewProvider(
		oidcCtx,
		appSettings.UserPortalIssuerUrl+"/v2.0",
	)
	if err != nil {
		return nil, err
	}
	provider.Endpoint()
	verifier := provider.Verifier(&oidc.Config{
		ClientID: appSettings.UserPortalClientId,
	})

	oauthConfig := &oauth2.Config{
		ClientID:     appSettings.UserPortalClientId,
		ClientSecret: appSettings.UserPortalClientSecret,
		RedirectURL:  appSettings.AppUrlBase + "/sign-in-redirect",

		Endpoint: provider.Endpoint(),

		Scopes: []string{
			oidc.ScopeOpenID,
			"profile",
			"email",
			"offline_access",
		},
	}

	return &authService{
		appSettings:        appSettings,
		userRepo:           userRepo,
		jwtSignerAndParser: jwtSignerAndParser,
		oauthConfig:        oauthConfig,
		verifier:           verifier,
	}, nil
}

// BuildSignInRedirectRequest generates OIDC state, used to validate any Id
// token provided in a response. It returns a struct containing a redirect Url
// and the OIDC state as a signed string
func (s *authService) BuildSignInRedirectRequest(
	requestedPath string,
) (*SignInRedirectRequest, error) {
	r := &SignInRedirectRequest{}

	// Generate OIDC security values
	state := generateRandomString(32)
	nonce := generateRandomString(32)
	codeVerifier := generateRandomString(64)
	codeChallenge := sha256Base64URL(codeVerifier)

	// Store OIDC security values in a hmac signed string. The handler will
	// persist this in a http-only cookie
	oidcState := oidcState{
		State:        state,
		Nonce:        nonce,
		CodeVerifier: codeVerifier,
		RequestedPath:      requestedPath,
	}

	oidcStateString, err := json.Marshal(oidcState)
	if err != nil {
		return nil, err
	}
	r.SignedOidcState = signPayload(s.appSettings.OidcStateSecret, oidcStateString)

	// Construct url to redirect user to sign-in
	r.Url = s.oauthConfig.AuthCodeURL(
		state,
		oidc.Nonce(nonce),
		oauth2.SetAuthURLParam(
			"code_challenge",
			codeChallenge,
		),
		oauth2.SetAuthURLParam(
			"code_challenge_method",
			"S256",
		),
		oauth2.SetAuthURLParam("prompt", "select_account"),
	)

	// Return the request containing the redirect url and the signed oidc state
	// used to verify responses to that request and meet the code challenge when
	// exchanging the returned code for a token
	return r, nil
}

func (s *authService) BuildSignOutRedirectRequest() string {
	logoutURL := s.appSettings.UserPortalIssuerUrl + "/oauth2/v2.0/logout"
	postLogoutRedirectURI := s.appSettings.AppUrlBase + "/signed-out"

	u, err := url.Parse(logoutURL)
	if err != nil {
		return logoutURL
	}

	q := u.Query()
	q.Set("post_logout_redirect_uri", postLogoutRedirectURI)

	u.RawQuery = q.Encode()

	return u.String()
}

func (s *authService) AuthenticateUser(
	ctx context.Context,
	authorisationCode string,
	returnedState string,
	signedOidcState string,
) (resp *AuthenticateUserResponse, err error) {
	// validate and parse OIDC state obtained from cookie
	raw, ok := verifySignedCookie(s.appSettings.OidcStateSecret, signedOidcState)
	if !ok {
		return nil, errors.New("invalid oidc state signature")
	}

	var oidcState oidcState
	if err := json.Unmarshal(raw, &oidcState); err != nil {
		return nil, err
	}

	// Validate that the state value matches that of the original redirect
	// request
	if oidcState.State != returnedState {
		return nil, errors.New("state mismatch")
	}

	// Exchange the auth code for an id token using the oidc code verifier to
	// meet the code challenge sent in the original redirect request
	oauthToken, err := s.oauthConfig.Exchange(
		ctx,
		authorisationCode,
		oauth2.SetAuthURLParam(
			"code_verifier",
			oidcState.CodeVerifier,
		),
	)
	if err != nil {
		return nil, err
	}

	// Validate the id token
	rawIDToken, ok := oauthToken.Extra("id_token").(string)
	if !ok {
		return nil, errors.New("missing id_token")
	}

	idToken, err := s.verifier.Verify(
		ctx,
		rawIDToken,
	)
	if err != nil {
		return nil, err
	}

	// Extract claims from the id token and validate that the nonce value
	// matches that of the original redirect request
	var claims portalTokenClaims
	if err := idToken.Claims(&claims); err != nil {
		return nil, err
	}

	if claims.Nonce != oidcState.Nonce {
		return nil, errors.New("nonce mismatch")
	}

	// Upsert the user with the data from the claims
	user, err := s.userRepo.UpsertUser(
		ctx,
		claims.Oid,
		claims.GivenName,
		claims.FamilyName,
		claims.UserName,
		claims.EmailAddress,
	)
	if err != nil {
		return nil, err
	}

	appToken, err := s.buildAppToken(user, rawIDToken)
	if err != nil {
		return nil, err
	}

	return &AuthenticateUserResponse{
		AppToken: appToken,
		RequestedPath:  oidcState.RequestedPath,
	}, nil
}

// ParseUserJwtCookie is a helper method for parsing a jwt token stored in a
// jwt cookie and returning the user id, contained in the token's subject claim.
// It will handle any errors with logging and return false if there is any issue
// parsing the cookie.
func (s *authService) ParseUserJwtCookie(tokenString string) (*AppClaims, error) {
	claims := &appClaims{}
	token, err := s.jwtSignerAndParser.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (any, error) {
			return s.appSettings.AppJwtSecret, nil
		},
		jwt.WithExpirationRequired(),
	)

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	if claims.UserId == "" {
		return nil, errors.New("unable to read user id from claims")
	}

	if claims.UserOid == "" {
		return nil, errors.New("unable to read user oid from claims")
	}

	if claims.UserEmail == "" {
		return nil, errors.New("unable to read user email from claims")
	}

	return &AppClaims{
		IdToken:   claims.IdToken,
		UserId:    claims.UserId,
		UserOid:   claims.UserOid,
		UserEmail: claims.UserEmail,
	}, nil
}

func (s *authService) RetrieveUserById(userId string) (*db.User, error) {
	return s.userRepo.RetrieveUserById(userId)
}

func (s *authService) buildAppToken(user *db.User, idToken string) (string, error) {
	appToken := jwt.NewWithClaims(jwt.SigningMethodHS256, appClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.Id,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
		IdToken:   idToken,
		UserId:    user.Id,
		UserOid:   user.Oid,
		UserEmail: user.EmailAddress,
	})

	return s.jwtSignerAndParser.Sign(appToken, s.appSettings.AppJwtSecret)
}

func generateRandomString(nBytes int) string {
	b := make([]byte, nBytes)
	_, err := rand.Read(b)
	if err != nil {
		panic("failed to generate secure random string")
	}

	return base64.RawURLEncoding.EncodeToString(b)
}

func sha256Base64URL(input string) string {
	sum := sha256.Sum256([]byte(input))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func signPayload(secret []byte, data []byte) string {
	mac := hmac.New(sha256.New, secret)
	mac.Write(data)

	sig := mac.Sum(nil)

	payload := base64.RawURLEncoding.EncodeToString(data)
	signature := base64.RawURLEncoding.EncodeToString(sig)

	return payload + "." + signature
}

func verifySignedCookie(cookieSecret []byte, value string) ([]byte, bool) {
	parts := strings.Split(value, ".")
	if len(parts) != 2 {
		return nil, false
	}

	payloadB64 := parts[0]
	sigB64 := parts[1]

	payload, err := base64.RawURLEncoding.DecodeString(payloadB64)
	if err != nil {
		return nil, false
	}

	sig, err := base64.RawURLEncoding.DecodeString(sigB64)
	if err != nil {
		return nil, false
	}

	mac := hmac.New(sha256.New, cookieSecret)
	mac.Write(payload)
	expected := mac.Sum(nil)

	if !hmac.Equal(sig, expected) {
		return nil, false
	}

	return payload, true
}
