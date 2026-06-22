package services

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/turnerbenjamin/heterogen_portal/internal/db"
	"github.com/turnerbenjamin/heterogen_portal/internal/etc"
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

type portalTokenClaims struct {
	jwt.RegisteredClaims
	Nonce        string `json:"nonce"`
	Oid          string `json:"oid"`
	GivenName    string `json:"given_name"`
	FamilyName   string `json:"family_name"`
	UserName     string `json:"name"`
	EmailAddress string `json:"email"`
}

type oidcConfiguration struct {
	Issuer  string `json:"issuer"`
	JwksURI string `json:"jwks_uri"`
}

type oidcState struct {
	State        string
	Nonce        string
	CodeVerifier string
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	IDToken     string `json:"id_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

type SignInRedirectRequest struct {
	Url             string
	SignedOIDCState string
}

type authService struct {
	appSettings        *etc.AppSettings
	userRepo           UserRepo
	jwtSignerAndParser JWTSignerAndParser
	tokenIssuer        string
	jwks               keyfunc.Keyfunc
}

func NewAuthService(
	context context.Context,
	appSettings *etc.AppSettings,
	httpClient *http.Client,
	userRepo UserRepo,
	jwtSignerAndParser JWTSignerAndParser,
) (*authService, error) {
	jwks, issuer, err := initialiseKeyFunc(
		context,
		httpClient,
		appSettings.UserPortalGetOidcConfigUrl,
	)
	if err != nil {
		return nil, err
	}

	return &authService{
		appSettings:        appSettings,
		userRepo:           userRepo,
		jwtSignerAndParser: jwtSignerAndParser,
		jwks:               jwks,
		tokenIssuer:        issuer,
	}, nil
}

// BuildSignInRedirectRequest generates OIDC state, used to validate any Id
// token provided in a response. It returns a struct containing a redirect Url
// and the OIDC state as a signed string
func (s *authService) BuildSignInRedirectRequest() (*SignInRedirectRequest, error) {
	r := &SignInRedirectRequest{}

	// Generate OIDC security values
	state := generateRandomString(32)
	nonce := generateRandomString(32)
	codeVerifier := generateRandomString(64)

	codeChallenge := sha256Base64URL(codeVerifier)

	// Store OIDC security values in a signed cookie
	oidcState := oidcState{
		State:        state,
		Nonce:        nonce,
		CodeVerifier: codeVerifier,
	}

	oidcStateString, err := json.Marshal(oidcState)
	if err != nil {
		return nil, err
	}
	r.SignedOIDCState = signPayload(s.appSettings.OIDCStateSecret, oidcStateString)

	// Build redirect URL
	r.Url = s.appSettings.UserPortalOAuthUrl + "/authorize" +
		"?client_id=" + s.appSettings.UserPortalClientId +
		"&response_type=code" +
		"&response_mode=query" +
		"&redirect_uri=" + s.appSettings.AppUrlBase + "/sign-in-redirect" +
		"&scope=" + "openid profile email offline_access" +
		"&state=" + state +
		"&nonce=" + nonce +
		"&code_challenge=" + codeChallenge +
		"&code_challenge_method=S256"

	return r, nil
}

func (s *authService) AuthenticateUser(
	ctx context.Context,
	authorisationCode string,
	returnedState string,
	signedOidcState string,
) (*db.User, error) {
	// validate and parse OIDC state obtained from cookie
	raw, ok := verifySignedCookie(s.appSettings.OIDCStateSecret, signedOidcState)
	if !ok {
		return nil, errors.New("invalid oidc state signature")
	}

	var oidcState oidcState
	if err := json.Unmarshal(raw, &oidcState); err != nil {
		return nil, err
	}

	// Validate returned state against state stored in the cookie
	if oidcState.State != returnedState {
		return nil, errors.New("state mismatch")
	}

	redirectUri := s.appSettings.AppUrlBase + "/sign-in-redirect"
	tokenUrl := s.appSettings.UserPortalOAuthUrl + "/token"

	// Exchange code for a token
	tokens, err := exchangeCodeForTokens(
		s.appSettings,
		authorisationCode,
		oidcState.CodeVerifier,
		redirectUri,
		tokenUrl,
	)
	if err != nil {
		return nil, err
	}

	// Parse and validate claims from the Id token
	claims, err := parseIdToken(
		s.jwtSignerAndParser,
		s.jwks,
		tokens.IDToken,
		s.appSettings.UserPortalClientId,
		s.tokenIssuer,
	)
	if err != nil {
		return nil, err
	}

	// Validate nonce against oidc state
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

	return user, nil
}

// ParseUserJwtCookie is a helper method for parsing a jwt token stored in a
// jwt cookie and returning the user id, contained in the token's subject claim.
// It will handle any errors with logging and return false if there is any issue
// parsing the cookie.
func (s *authService) ParseUserJwtCookie(tokenString string) (*db.User, error) {
	token, err := s.jwtSignerAndParser.ParseWithClaims(
		tokenString,
		&jwt.RegisteredClaims{},
		func(token *jwt.Token) (any, error) {
			return s.appSettings.AppJWTSecret, nil
		},
		jwt.WithExpirationRequired(),
	)

	if err != nil {
		return nil, err
	}

	userId, err := token.Claims.GetSubject()
	if err != nil {
		return nil, err
	}

	if userId == "" {
		return nil, errors.New("unable to access user id from token claims")
	}

	return s.userRepo.RetrieveUserById(userId)
}

func (s *authService) SignJWT(tokenToSign *jwt.Token, secret []byte) (string, error) {
	return s.jwtSignerAndParser.Sign(tokenToSign, secret)
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

func exchangeCodeForTokens(
	appSettings *etc.AppSettings,
	code string,
	codeVerifier string,
	redirectURI string,
	tokenURL string,
) (*tokenResponse, error) {
	assertion, assertionType, err := buildClientAssertion(
		tokenURL,
		appSettings,
	)
	if err != nil {
		return nil, err
	}

	// Build token request
	data := url.Values{}
	data.Set("client_id", appSettings.UserPortalClientId)
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)
	data.Set("code_verifier", codeVerifier)
	data.Set("client_assertion_type", assertionType)
	data.Set("client_assertion", assertion)

	req, err := http.NewRequestWithContext(
		context.Background(),
		"POST",
		tokenURL,
		strings.NewReader(data.Encode()),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Execute token request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf(
			"token endpoint returned %d: %s",
			resp.StatusCode,
			string(body),
		)
	}

	// parse response
	var tr tokenResponse
	if err := json.Unmarshal(body, &tr); err != nil {
		return nil, err
	}

	return &tr, nil
}

func buildClientAssertion(
	tokenUrl string,
	appSettings *etc.AppSettings,
) (string, string, error) {
	assertionType := "urn:ietf:params:oauth:client-assertion-type:jwt-bearer"

	claims := jwt.MapClaims{
		"aud": tokenUrl,
		"iss": appSettings.UserPortalClientId,
		"sub": appSettings.UserPortalClientId,
		"jti": uuid.NewString(),
		"nbf": time.Now().Unix(),
		"exp": time.Now().Add(5 * time.Minute).Unix(),
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	h := sha1.Sum(appSettings.Cert.Raw)
	x5t := base64.RawURLEncoding.EncodeToString(h[:])
	jwtToken.Header["x5t"] = x5t

	jwtToken.Header["typ"] = "JWT"

	assertion, err := jwtToken.SignedString(appSettings.RsaKey)
	if err != nil {
		return "", assertionType, fmt.Errorf("failed to sign client assertion: %w", err)
	}
	return assertion, assertionType, nil
}

// ValidateToken validates and returns claims; uses cached JWKS.
func parseIdToken(
	jwtSignerAndParser JWTSignerAndParser,
	jwks keyfunc.Keyfunc,
	idTokenString string,
	clientId string,
	issuer string,
) (*portalTokenClaims, error) {

	// parse the token string as a jwt token
	token, err := jwtSignerAndParser.ParseWithClaims(
		idTokenString,
		&portalTokenClaims{},
		func(token *jwt.Token) (any, error) {
			if token.Method != jwt.SigningMethodRS256 {
				return nil, errors.New("unexpected signing algorithm")
			}

			return jwks.Keyfunc(token)
		},
		jwt.WithAudience(clientId),
		jwt.WithIssuer(issuer),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return nil, err
	}

	if token == nil {
		return nil, errors.New("token is nil")
	}

	if !token.Valid {
		return nil, errors.New("token validation failed")
	}

	claims, ok := token.Claims.(*portalTokenClaims)
	if !ok {
		return nil, errors.New("unexpected claims type")
	}

	if claims.Nonce == "" {
		return nil, errors.New("missing nonce claim")
	}

	return claims, nil
}

func initialiseKeyFunc(
	appContext context.Context,
	httpClient *http.Client,
	getOidcConfigUrl string,
) (keyfunc.Keyfunc, string, error) {
	// Request OIDC config
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, getOidcConfigUrl, nil)

	if err != nil {
		return nil, "", err
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, "", err
	}

	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			fmt.Fprintf(os.Stderr, "error closing response body: %s\n", cerr.Error())
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf(
			"unexpected status code %d from OIDC configuration endpoint",
			resp.StatusCode,
		)
	}

	// Parse and validate the oidc config
	config := &oidcConfiguration{}
	if err := json.NewDecoder(resp.Body).Decode(config); err != nil {
		return nil, "", err
	}

	if config.JwksURI == "" {
		return nil, "", errors.New("jwks_uri missing from openid config")
	}

	if config.Issuer == "" {
		return nil, "", errors.New("issuer missing from openid config")
	}

	// create a keyfunc using the jwks URI from the key func
	jwks, err := keyfunc.NewDefaultCtx(appContext, []string{config.JwksURI})
	if err != nil {
		return nil, "", err
	}

	return jwks, config.Issuer, nil
}
