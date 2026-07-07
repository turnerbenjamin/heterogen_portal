// Package services is responsible for the service layer of the application
//
// This file implements an oAuth 2.0 auth code + PKCE flow to authenticate the
// user
package services

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net/http"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/golang-jwt/jwt/v5"
	"github.com/turnerbenjamin/heterogen_portal/internal/constants"
	"github.com/turnerbenjamin/heterogen_portal/internal/db"
	"github.com/turnerbenjamin/heterogen_portal/internal/etc"
	"golang.org/x/oauth2"
)

// UserRepo performs database operations on the user table
type UserRepo interface {
	UpsertUser(
		oid string,
		givenName string,
		familyName string,
		userName string,
		emailAddress string,
	) (*db.User, error)
	RetrieveUserById(id string) (*db.User, error)
	RetrieveUserByOid(id string) (*db.User, error)
}

// JwtSigner can sign and parse JWT tokens
type JwtSigner interface {
	ParseWithClaims(
		tokenString string,
		claims jwt.Claims,
		keyFunc jwt.Keyfunc,
		parserOptions ...jwt.ParserOption,
	) (*jwt.Token, error)
	Sign(token *jwt.Token, privateKey []byte) (string, error)
}

// PayloadSigner can sign and verify payloads
type PayloadSigner interface {
	Sign(secret []byte, data []byte) (signedData string)
	Verify(secret []byte, value string) (data []byte, ok bool)
}

// OAuthClient can provide a Url to the auth provider and exchange
// authorisation codes received from the provider with an id token
type OAuthClient interface {
	AuthCodeURL(state string, opts ...oauth2.AuthCodeOption) string
	Exchange(
		ctx context.Context,
		code string,
		opts ...oauth2.AuthCodeOption,
	) (*oauth2.Token, error)
}

// RandReader reads random bytes into b
type RandReader func(b []byte) (n int, err error)

// IdToken exposes a method to read token claims
type IdToken interface {
	Claims(v any) error
}

// IdTokenVerifier can parse and verify a raw id token string
type IdTokenVerifier interface {
	Verify(ctx context.Context, rawIDToken string) (IdToken, error)
}

// OidcProvider can provide an endpoint to the auth provider and a token
// verifier for the provider
type OidcProvider interface {
	Endpoint() oauth2.Endpoint
	Verifier(config *oidc.Config) IdTokenVerifier
}

// Jsonserialiser can marshall and unmarshal json payloads
type JsonSerialiser interface {
	Marshal(v any) ([]byte, error)
	Unmarshal(data []byte, v any) error
}

// SignInRedirectRequest contains a redirect url to the auth provider and signed
// oidc state which must be persisted to handle the response from the auth
// provider
type SignInRedirectRequest struct {
	Url             string
	SignedOidcState string
}

// AuthenticateUserResponse includes a signed jwt app token and the original
// path requested by the user before the sign-in redirect
type AuthenticateUserResponse struct {
	AppToken      string
	RequestedPath string
}

// portalTokenClaims contains the claims expected in the id token returned from
// the auth provider
type portalTokenClaims struct {
	Nonce        string `json:"nonce"`
	Oid          string `json:"oid"`
	GivenName    string `json:"given_name"`
	FamilyName   string `json:"family_name"`
	UserName     string `json:"name"`
	EmailAddress string `json:"email"`
}

// oidcState contains state persisted during the oidc code + pkce auth flow
type oidcState struct {
	State         string
	Nonce         string
	CodeVerifier  string
	RequestedPath string
}

// AppClaims contains claims from the app jwt token
type AppClaims struct {
	IdToken   string
	UserId    string
	UserOid   string
	UserEmail string
}

// appClaims is struct for internal use within the package to parse app jwt
// tokens
type appClaims struct {
	jwt.RegisteredClaims
	IdToken   string `json:"id_token"`
	UserId    string `json:"user_id"`
	UserOid   string `json:"user_oid"`
	UserEmail string `json:"email"`
}

// AuthService is responsible for the auth flow in the application
type AuthService struct {
	appSettings    *etc.AppSettings
	userRepo       UserRepo
	jsonSerialiser JsonSerialiser
	jwtSigner      JwtSigner
	payloadSigner  PayloadSigner
	randRead       RandReader
	oauthClient    OAuthClient
	verifier       IdTokenVerifier
}

// NewAuthService builds a new authService
func NewAuthService(
	ctx context.Context,
	appSettings *etc.AppSettings,
	httpClient *http.Client,
	jsonSerialiser JsonSerialiser,
	randReader RandReader,
	jwtSigner JwtSigner,
	payloadSigner PayloadSigner,
	newOidcProvider func(ctx context.Context, issuer string) (OidcProvider, error),
	userRepo UserRepo,
) (*AuthService, error) {
	oidcCtx := oidc.ClientContext(ctx, httpClient)

	provider, err := newOidcProvider(
		oidcCtx,
		appSettings.UserPortalIssuerUrl+"/v2.0",
	)
	if err != nil {
		return nil, err
	}

	oauthClient := buildOAuthConfig(
		appSettings,
		provider,
	)
	verifier := provider.Verifier(&oidc.Config{
		ClientID: appSettings.UserPortalClientId,
	})

	return &AuthService{
		appSettings:    appSettings,
		userRepo:       userRepo,
		jsonSerialiser: jsonSerialiser,
		jwtSigner:      jwtSigner,
		payloadSigner:  payloadSigner,
		randRead:       randReader,
		oauthClient:    oauthClient,
		verifier:       verifier,
	}, nil
}

// BuildSignInRedirectRequest generates OIDC state, used to validate any Id
// token provided in a response. It returns a struct containing a redirect Url
// and the OIDC state as a signed string
func (s *AuthService) BuildSignInRedirectRequest(
	requestedPath string,
) (*SignInRedirectRequest, error) {
	r := &SignInRedirectRequest{}

	// Generate OIDC security values
	state, err := s.generateRandomString(32)
	if err != nil {
		return nil, errors.New(constants.ErrMsgFailedToGenerateOidcStateValue)
	}

	nonce, err := s.generateRandomString(32)
	if err != nil {
		return nil, errors.New(constants.ErrMsgFailedToGenerateOidcNonceValue)
	}

	codeVerifier, err := s.generateRandomString(64)
	if err != nil {
		return nil, errors.New(constants.ErrMsgFailedToGenerateOidcCodeVerifierValue)
	}
	codeChallenge := sha256Base64Url(codeVerifier)

	// Store OIDC security values in a hmac signed string. The handler will
	// persist this in a http-only cookie
	oidcState := oidcState{
		State:         state,
		Nonce:         nonce,
		CodeVerifier:  codeVerifier,
		RequestedPath: requestedPath,
	}

	oidcStateBytes, err := s.jsonSerialiser.Marshal(oidcState)
	if err != nil {
		return nil, err
	}
	r.SignedOidcState = s.payloadSigner.Sign(
		s.appSettings.OidcStateSecret,
		oidcStateBytes,
	)

	// Construct url to redirect user to sign-in
	r.Url = s.oauthClient.AuthCodeURL(
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

// BuildSignOutRedirectRequest returns a url to redirect to so that the user can
// lock out from the auth provider
func (s *AuthService) BuildSignOutRedirectRequest() string {
	postLogoutRedirectUri := s.appSettings.AppUrlBase + "/signed-out"

	logoutUrl := s.appSettings.UserPortalIssuerUrl +
		"/oauth2/v2.0/logout" +
		"?post_logout_redirect_uri=" +
		postLogoutRedirectUri

	return logoutUrl
}

// AuthenticateUser uses an authorisation code and oidc state to retrieve user
// information from an id token. Once authenticated, the user is upserted into
// the database and a signed token is returned to allow direct authentication
// with the app for subsequent requests
func (s *AuthService) AuthenticateUser(
	ctx context.Context,
	authorisationCode string,
	returnedStateValue string,
	signedOidcState string,
) (resp *AuthenticateUserResponse, err error) {
	// validate and parse OIDC state obtained from cookie
	raw, ok := s.payloadSigner.Verify(
		s.appSettings.OidcStateSecret,
		signedOidcState,
	)
	if !ok {
		return nil, errors.New(constants.ErrMsgInvalidOidcState)
	}

	var oidcState oidcState
	if err := s.jsonSerialiser.Unmarshal(raw, &oidcState); err != nil {
		return nil, err
	}

	// Validate that the state value matches that of the original redirect
	// request
	if oidcState.State != returnedStateValue {
		return nil, errors.New(constants.ErrMsgOidcStateValueMismatch)
	}

	// Exchange the auth code for an id token using the oidc code verifier to
	// meet the code challenge sent in the original redirect request
	oauthToken, err := s.oauthClient.Exchange(
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
		return nil, errors.New(constants.ErrMsgUnableToAccessIdToken)
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
		return nil, errors.New(constants.ErrMsgOidcNonceValueMismatch)
	}

	// Upsert the user with the data from the claims
	user, err := s.userRepo.UpsertUser(
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
		AppToken:      appToken,
		RequestedPath: oidcState.RequestedPath,
	}, nil
}

// ParseUserJwtCookie is a helper method for parsing a jwt token stored in a
// jwt cookie and returning the user id, contained in the token's subject claim.
// It will handle any errors with logging and return false if there is any issue
// parsing the cookie.
func (s *AuthService) ParseUserJwtCookie(tokenString string) (*AppClaims, error) {
	claims := &appClaims{}
	_, err := s.jwtSigner.ParseWithClaims(
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

	if claims.IdToken == "" {
		return nil, errors.New(constants.ErrMsgParseAppJwtErrUnableToReadIdToken)
	}

	if claims.UserId == "" {
		return nil, errors.New(constants.ErrMsgParseAppJwtErrUnableToReadUserId)
	}

	if claims.UserOid == "" {
		return nil, errors.New(constants.ErrMsgParseAppJwtErrUnableToReadUserOid)
	}

	if claims.UserEmail == "" {
		return nil, errors.New(constants.ErrMsgParseAppJwtErrUnableToReadUserEmail)
	}

	return &AppClaims{
		IdToken:   claims.IdToken,
		UserId:    claims.UserId,
		UserOid:   claims.UserOid,
		UserEmail: claims.UserEmail,
	}, nil
}

// RetrieveUserById returns the user with the matching id or an error if the
// user cannot be accessed
func (s *AuthService) RetrieveUserById(userId string) (*db.User, error) {
	return s.userRepo.RetrieveUserById(userId)
}

// buildAppToken creates an app token containing user details and the original
// idToken used to authenticate the user with the auth provider
func (s *AuthService) buildAppToken(user *db.User, idToken string) (string, error) {
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

	return s.jwtSigner.Sign(appToken, s.appSettings.AppJwtSecret)
}

// generateRandomString returns a random string with raw url base64 encoding
func (s *AuthService) generateRandomString(nBytes int) (string, error) {
	b := make([]byte, nBytes)
	_, err := s.randRead(b)
	if err != nil {
		return "", errors.New(constants.ErrMsgFailedToGenerateRandomString)
	}

	return base64.RawURLEncoding.EncodeToString(b), nil
}

// sha256Base64Url returns the checksum of the input with raw url base64
// encoding
func sha256Base64Url(input string) string {
	sum := sha256.Sum256([]byte(input))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

// buildOAuthConfig uses the app settings and oidc provider to configure an
// oauth client
func buildOAuthConfig(
	appSettings *etc.AppSettings,
	provider OidcProvider,
) *oauth2.Config {
	return &oauth2.Config{
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
}
