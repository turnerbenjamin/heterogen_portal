package services

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net/http"
	"net/url"
	"testing"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/turnerbenjamin/heterogen_portal/internal/constants"
	"github.com/turnerbenjamin/heterogen_portal/internal/db"
	"github.com/turnerbenjamin/heterogen_portal/internal/etc"

	"golang.org/x/oauth2"
)

func TestNewAuthService(t *testing.T) {
	t.Parallel()
	t.Run("returns configured auth service", func(t *testing.T) {

		ctx := context.Background()
		appSettings := &etc.AppSettings{
			UserPortalIssuerUrl:    "https://login.authprovider.com",
			UserPortalClientId:     "client-id",
			UserPortalClientSecret: "client-secret",
			AppUrlBase:             "https://portal.test",
		}
		wantIssuerUrl := appSettings.UserPortalIssuerUrl + "/v2.0"

		httpClient := &http.Client{}
		randReader := &deterministicRandReader{t: t}
		jwtSigner := NewMockJwtSigner(t)
		payloadSigner := NewMockPayloadSigner(t)
		userRepo := NewMockUserRepo(t)
		jsonSerialiser := NewMockJsonSerialiser(t)

		wantEndpoint := oauth2.Endpoint{
			AuthURL: appSettings.UserPortalIssuerUrl,
		}
		wantVerifier := NewMockIdTokenVerifier(t)
		oidcProvider := NewMockOidcProvider(t)
		oidcProvider.EXPECT().Endpoint().Return(wantEndpoint)
		oidcProvider.EXPECT().Verifier(mock.MatchedBy(func(c *oidc.Config) bool {
			assert.Equal(t, c.ClientID, appSettings.UserPortalClientId)
			return true
		})).Return(wantVerifier)

		oidcProviderFactory := oidcProviderFactory{
			t:            t,
			buildReturns: buildOidcProviderResp{provider: oidcProvider},
		}

		service, err := NewAuthService(
			ctx,
			appSettings,
			httpClient,
			jsonSerialiser,
			randReader.ReadNext,
			jwtSigner,
			payloadSigner,
			oidcProviderFactory.Build,
			userRepo,
		)

		assert.Nil(t, err)
		assert.Equal(t, len(oidcProviderFactory.buildCalls), 1)
		gotBuildCall := oidcProviderFactory.buildCalls[0]
		assert.Equal(t, gotBuildCall.issuer, wantIssuerUrl)

		assert.Equal(t, service.appSettings, appSettings)
		assert.Equal(t, service.userRepo, userRepo)
		assert.Equal(t, service.jsonSerialiser, jsonSerialiser)
		assert.Equal(t, service.jwtSigner, jwtSigner)
		assert.Equal(t, service.payloadSigner, payloadSigner)
		assert.Equal(t, service.verifier, wantVerifier)

		gotConfig, ok := service.oauthClient.(*oauth2.Config)
		assert.Equal(t, ok, true)
		assert.Equal(t, gotConfig.ClientID, appSettings.UserPortalClientId)
		assert.Equal(t, gotConfig.ClientSecret, appSettings.UserPortalClientSecret)
		assert.Equal(t, gotConfig.RedirectURL, appSettings.AppUrlBase+"/sign-in-redirect")
		assert.Equal(t, gotConfig.Endpoint, wantEndpoint)
		wantScopes := []string{oidc.ScopeOpenID, "profile", "email", "offline_access"}
		assert.EqualValues(t, gotConfig.Scopes, wantScopes)
	})

	t.Run("handles errors returned by newOidcProvider func", func(t *testing.T) {
		ctx := context.Background()
		appSettings := &etc.AppSettings{
			UserPortalIssuerUrl:    "https://login.authprovider.com",
			UserPortalClientId:     "client-id",
			UserPortalClientSecret: "client-secret",
			AppUrlBase:             "https://portal.test",
		}

		wantError := errors.New("test build provider error")
		randReader := &deterministicRandReader{t: t}
		userRepo := NewMockUserRepo(t)
		oidcProviderFactory := oidcProviderFactory{
			t:            t,
			buildReturns: buildOidcProviderResp{err: wantError},
		}

		_, err := NewAuthService(
			ctx,
			appSettings,
			&http.Client{},
			NewMockJsonSerialiser(t),
			randReader.ReadNext,
			NewMockJwtSigner(t),
			NewMockPayloadSigner(t),
			oidcProviderFactory.Build,
			userRepo,
		)

		assert.EqualValues(t, err, wantError)
	})
}

func TestBuildSignInRedirectRequest(t *testing.T) {
	t.Parallel()
	t.Run("returns redirect request with security values correctly set", func(t *testing.T) {
		appSettings := &etc.AppSettings{
			OidcStateSecret:        []byte("very-secret-key"),
			UserPortalClientId:     "client-id",
			UserPortalClientSecret: "client-secret",
			AppUrlBase:             "https://my-test-app.com",
		}

		wantRequestedPath := "/some-endpoint"
		wantHost := "an-auth-portal.com"
		oAuthClient := &oauth2.Config{
			ClientID:     appSettings.UserPortalClientId,
			ClientSecret: appSettings.UserPortalClientSecret,
			RedirectURL:  appSettings.AppUrlBase + "/sign-in-redirect",
			Endpoint: oauth2.Endpoint{
				AuthURL: "https://" + wantHost,
			},
		}

		wantSignedState := "signed-state"

		capturedState := oidcState{}
		stringifiedState := []byte("state-string")
		jsonSerialiser := NewMockJsonSerialiser(t)
		jsonSerialiser.
			EXPECT().Marshal(mock.Anything).
			Run(func(v any) {
				if s, ok := v.(oidcState); ok {
					capturedState = s
				}
			}).
			Return(stringifiedState, nil)

		payloadSigner := NewMockPayloadSigner(t)
		payloadSigner.
			EXPECT().Sign(appSettings.OidcStateSecret, stringifiedState).
			Return(wantSignedState)

		randReader := &deterministicRandReader{
			t:        t,
			patterns: [][]byte{},
		}

		stateBytes := make([]byte, 32)
		nonceBytes := make([]byte, 32)
		codeVerifierBytes := make([]byte, 64)

		_, err := randReader.ReadPattern(stateBytes, []byte("state-value"))
		assert.Nil(t, err)

		_, err = randReader.ReadPattern(nonceBytes, []byte("nonce-value"))
		assert.Nil(t, err)

		_, err = randReader.ReadPattern(codeVerifierBytes, []byte("code-verifier-value"))
		assert.Nil(t, err)

		randReader.patterns = [][]byte{stateBytes, nonceBytes, codeVerifierBytes}

		wantState := base64.RawStdEncoding.EncodeToString(stateBytes)
		wantNonce := base64.RawStdEncoding.EncodeToString(nonceBytes)
		wantCodeVerifier := base64.RawStdEncoding.EncodeToString(codeVerifierBytes)

		challengeSum := sha256.Sum256([]byte(wantCodeVerifier))
		wantChallenge := base64.RawURLEncoding.EncodeToString(challengeSum[:])

		wantQueryParams := map[string]string{
			"code_challenge_method": "S256",
			"code_challenge":        wantChallenge,
			"state":                 wantState,
			"nonce":                 wantNonce,
			"response_type":         "code",
			"prompt":                "select_account",
		}

		service := &AuthService{
			appSettings:    appSettings,
			oauthClient:    oAuthClient,
			randRead:       randReader.ReadNext,
			payloadSigner:  payloadSigner,
			jsonSerialiser: jsonSerialiser,
		}

		gotRedirectRes, err := service.BuildSignInRedirectRequest(wantRequestedPath)
		assert.Nil(t, err)

		// Test url host and params
		gotUrl, err := url.Parse(gotRedirectRes.Url)
		assert.Nil(t, err)

		assert.Equal(t, gotUrl.Host, wantHost)
		gotQueryParams := gotUrl.Query()
		for k, v := range wantQueryParams {
			gotValue := gotQueryParams.Get(k)
			assert.Equal(t, gotValue, v)
		}

		assert.Equal(t, gotRedirectRes.SignedOidcState, wantSignedState)

		payloadSigner.AssertExpectations(t)

		assert.Equal(t, capturedState.State, wantState)
		assert.Equal(t, capturedState.Nonce, wantNonce)
		assert.Equal(t, capturedState.CodeVerifier, wantCodeVerifier)
		assert.Equal(t, capturedState.RequestedPath, wantRequestedPath)
	})

	t.Run("handles errors returned from generate random string", func(t *testing.T) {
		testData := []struct {
			wantErrMsg        string
			randReaderReturns []error
		}{
			{
				wantErrMsg: constants.ErrMsgFailedToGenerateOidcStateValue,
				randReaderReturns: []error{
					errors.New(constants.ErrMsgFailedToGenerateRandomString),
					nil,
					nil,
				},
			},
			{
				wantErrMsg: constants.ErrMsgFailedToGenerateOidcNonceValue,
				randReaderReturns: []error{
					nil,
					errors.New(constants.ErrMsgFailedToGenerateRandomString),
					nil,
				},
			},
			{
				wantErrMsg: constants.ErrMsgFailedToGenerateOidcCodeVerifierValue,
				randReaderReturns: []error{
					nil,
					nil,
					errors.New(constants.ErrMsgFailedToGenerateRandomString),
				},
			},
		}

		for _, td := range testData {

			randReader := &deterministicRandReader{
				t:       t,
				returns: td.randReaderReturns,
			}

			service := &AuthService{
				appSettings:   &etc.AppSettings{},
				oauthClient:   NewMockOAuthClient(t),
				randRead:      randReader.ReadNext,
				payloadSigner: NewMockPayloadSigner(t),
			}

			_, err := service.BuildSignInRedirectRequest("/")
			assert.Equal(t, err, errors.New(td.wantErrMsg))
		}
	})

	t.Run("handles json serialiser errors", func(t *testing.T) {
		wantErr := errors.New("test error")

		jsonSerialiser := NewMockJsonSerialiser(t)
		jsonSerialiser.EXPECT().Marshal(mock.Anything).Return(nil, wantErr)

		randReader := &deterministicRandReader{t: t}
		service := AuthService{
			appSettings:    &etc.AppSettings{},
			randRead:       randReader.ReadNext,
			jsonSerialiser: jsonSerialiser,
			payloadSigner:  NewMockPayloadSigner(t),
			oauthClient:    NewMockOAuthClient(t),
		}

		_, err := service.BuildSignInRedirectRequest("/")
		assert.EqualError(t, err, wantErr.Error())
	})
}

func TestBuildSignOutRedirectRequest(t *testing.T) {
	t.Parallel()

	wantHost := "user-portal-issuer"
	appSettings := &etc.AppSettings{
		UserPortalIssuerUrl: "https://" + wantHost,
		AppUrlBase:          "https://my-test-app",
	}

	wantPath := "/oauth2/v2.0/logout"
	wantRedirectUri := appSettings.AppUrlBase + "/signed-out"
	wantQueryParams := map[string]string{
		"post_logout_redirect_uri": wantRedirectUri,
	}

	service := &AuthService{appSettings: appSettings}
	rawUrl := service.BuildSignOutRedirectRequest()
	gotUrl, err := url.Parse(rawUrl)

	assert.Nil(t, err)
	assert.Equal(t, gotUrl.Host, wantHost)
	assert.Equal(t, gotUrl.Path, wantPath)
	gotQueryParams := gotUrl.Query()
	for k, v := range wantQueryParams {
		gotValue := gotQueryParams.Get(k)
		assert.Equal(t, gotValue, v)
	}
}

func TestAuthenticateUser(t *testing.T) {
	t.Parallel()
	t.Run("handles invalid oidc state", func(t *testing.T) {
		wantErrMsg := constants.ErrMsgInvalidOidcState

		ctx := context.Background()
		authCode := "an-auth-code"
		returnedStateValue := "original-state-value"
		signedOidcState := "signed-oidc-state"

		appSettings := &etc.AppSettings{
			OidcStateSecret: []byte("super-secret-secret"),
		}

		payloadSigner := NewMockPayloadSigner(t)
		payloadSigner.EXPECT().Verify(
			appSettings.OidcStateSecret,
			signedOidcState,
		).Return(nil, false)

		service := &AuthService{
			appSettings:   appSettings,
			payloadSigner: payloadSigner,
		}

		gotResp, err := service.AuthenticateUser(
			ctx,
			authCode,
			returnedStateValue,
			signedOidcState,
		)
		assert.Nil(t, gotResp)
		assert.EqualError(t, err, wantErrMsg)
	})

	t.Run("handles serialiser errors", func(t *testing.T) {
		wantErr := errors.New("test unmarshal error")

		ctx := context.Background()
		authCode := "an-auth-code"
		returnedStateValue := "original-state-value"
		signedOidcState := "signed-oidc-state"

		appSettings := &etc.AppSettings{
			OidcStateSecret: []byte("super-secret-secret"),
		}

		payload := []byte("oidc-state-payload")
		payloadSigner := NewMockPayloadSigner(t)
		payloadSigner.EXPECT().Verify(
			appSettings.OidcStateSecret,
			signedOidcState,
		).Return(payload, true)

		jsonSerialiser := NewMockJsonSerialiser(t)
		jsonSerialiser.EXPECT().Unmarshal(payload, mock.Anything).Return(wantErr)

		service := &AuthService{
			appSettings:    appSettings,
			payloadSigner:  payloadSigner,
			jsonSerialiser: jsonSerialiser,
		}

		gotResp, err := service.AuthenticateUser(
			ctx,
			authCode,
			returnedStateValue,
			signedOidcState,
		)
		assert.Nil(t, gotResp)
		assert.EqualError(t, err, wantErr.Error())
	})

	t.Run("validates persisted oidc state against state returned by the auth provider", func(t *testing.T) {
		wantErrMsg := constants.ErrMsgOidcStateValueMismatch

		ctx := context.Background()
		authCode := "an-auth-code"
		signedOidcState := "signed-oidc-state"

		wantOidcState := &oidcState{State: "want state"}
		returnedState := "incorrect state"

		appSettings := &etc.AppSettings{
			OidcStateSecret: []byte("super-secret-secret"),
		}

		payload := []byte("oidc-state-payload")
		payloadSigner := NewMockPayloadSigner(t)
		payloadSigner.EXPECT().Verify(
			appSettings.OidcStateSecret,
			signedOidcState,
		).Return(payload, true)

		jsonSerialiser := NewMockJsonSerialiser(t)
		jsonSerialiser.EXPECT().Unmarshal(payload, mock.Anything).
			Run(func(_ []byte, v any) {
				if persistedState, ok := v.(*oidcState); ok {
					persistedState.State = wantOidcState.State
				}
			}).
			Return(nil)

		service := &AuthService{
			appSettings:    appSettings,
			payloadSigner:  payloadSigner,
			jsonSerialiser: jsonSerialiser,
		}

		gotResp, err := service.AuthenticateUser(
			ctx,
			authCode,
			returnedState,
			signedOidcState,
		)
		assert.Nil(t, gotResp)
		assert.EqualError(t, err, wantErrMsg)
	})

	t.Run("handles errors returned when exchanging auth code for an id token", func(t *testing.T) {
		wantErr := errors.New("unable to exchange code for token")

		ctx := context.Background()
		authCode := "an-auth-code"
		signedOidcState := "signed-oidc-state"

		wantOidcState := &oidcState{State: "want state"}
		returnedState := wantOidcState.State

		appSettings := &etc.AppSettings{
			OidcStateSecret: []byte("super-secret-secret"),
		}

		payload := []byte("oidc-state-payload")
		payloadSigner := NewMockPayloadSigner(t)
		payloadSigner.EXPECT().Verify(
			appSettings.OidcStateSecret,
			signedOidcState,
		).Return(payload, true)

		jsonSerialiser := NewMockJsonSerialiser(t)
		jsonSerialiser.EXPECT().Unmarshal(payload, mock.Anything).
			Run(func(_ []byte, v any) {
				if persistedState, ok := v.(*oidcState); ok {
					persistedState.State = wantOidcState.State
				}
			}).
			Return(nil)

		oauthClient := NewMockOAuthClient(t)
		oauthClient.
			EXPECT().Exchange(ctx, authCode, mock.Anything).
			Return(nil, wantErr)

		service := &AuthService{
			appSettings:    appSettings,
			payloadSigner:  payloadSigner,
			jsonSerialiser: jsonSerialiser,
			oauthClient:    oauthClient,
		}

		gotResp, err := service.AuthenticateUser(
			ctx,
			authCode,
			returnedState,
			signedOidcState,
		)
		assert.Nil(t, gotResp)
		assert.EqualError(t, err, wantErr.Error())
	})

	t.Run("handles id token not being returned by the code exchange", func(t *testing.T) {
		wantErrMsg := constants.ErrMsgUnableToAccessIdToken

		ctx := context.Background()
		authCode := "an-auth-code"
		signedOidcState := "signed-oidc-state"

		wantOidcState := &oidcState{State: "want state"}
		returnedState := wantOidcState.State

		appSettings := &etc.AppSettings{
			OidcStateSecret: []byte("super-secret-secret"),
		}

		payload := []byte("oidc-state-payload")
		payloadSigner := NewMockPayloadSigner(t)
		payloadSigner.EXPECT().Verify(
			appSettings.OidcStateSecret,
			signedOidcState,
		).Return(payload, true)

		jsonSerialiser := NewMockJsonSerialiser(t)
		jsonSerialiser.EXPECT().Unmarshal(payload, mock.Anything).
			Run(func(_ []byte, v any) {
				if persistedState, ok := v.(*oidcState); ok {
					persistedState.State = wantOidcState.State
				}
			}).
			Return(nil)

		oauthToken := &oauth2.Token{}
		oauthClient := NewMockOAuthClient(t)
		oauthClient.
			EXPECT().Exchange(ctx, authCode, mock.Anything).
			Return(oauthToken, nil)

		service := &AuthService{
			appSettings:    appSettings,
			payloadSigner:  payloadSigner,
			jsonSerialiser: jsonSerialiser,
			oauthClient:    oauthClient,
		}

		gotResp, err := service.AuthenticateUser(
			ctx,
			authCode,
			returnedState,
			signedOidcState,
		)
		assert.Nil(t, gotResp)
		assert.EqualError(t, err, wantErrMsg)
	})

	t.Run("handles invalid id tokens", func(t *testing.T) {
		wantErr := errors.New("unable to verify id token")

		ctx := context.Background()
		authCode := "an-auth-code"
		signedOidcState := "signed-oidc-state"

		wantOidcState := &oidcState{State: "want state"}
		returnedState := wantOidcState.State

		appSettings := &etc.AppSettings{
			OidcStateSecret: []byte("super-secret-secret"),
		}

		payload := []byte("oidc-state-payload")
		payloadSigner := NewMockPayloadSigner(t)
		payloadSigner.EXPECT().Verify(
			appSettings.OidcStateSecret,
			signedOidcState,
		).Return(payload, true)

		jsonSerialiser := NewMockJsonSerialiser(t)
		jsonSerialiser.EXPECT().Unmarshal(payload, mock.Anything).
			Run(func(_ []byte, v any) {
				if persistedState, ok := v.(*oidcState); ok {
					persistedState.State = wantOidcState.State
				}
			}).
			Return(nil)

		idToken := "mock-id-token-string"
		oauthToken := &oauth2.Token{}
		oauthToken = oauthToken.WithExtra(map[string]any{
			"id_token": idToken,
		})

		oauthClient := NewMockOAuthClient(t)
		oauthClient.
			EXPECT().Exchange(ctx, authCode, mock.Anything).
			Return(oauthToken, nil)

		verifier := NewMockIdTokenVerifier(t)
		verifier.
			EXPECT().Verify(ctx, idToken).
			Return(nil, wantErr)

		service := &AuthService{
			appSettings:    appSettings,
			payloadSigner:  payloadSigner,
			jsonSerialiser: jsonSerialiser,
			oauthClient:    oauthClient,
			verifier:       verifier,
		}

		gotResp, err := service.AuthenticateUser(
			ctx,
			authCode,
			returnedState,
			signedOidcState,
		)
		assert.Nil(t, gotResp)
		assert.EqualError(t, err, wantErr.Error())
	})

	t.Run("handles errors returned when parsing claims", func(t *testing.T) {
		wantErr := errors.New("unable to parse claims")

		ctx := context.Background()
		authCode := "an-auth-code"
		signedOidcState := "signed-oidc-state"

		wantOidcState := &oidcState{State: "want state"}
		returnedState := wantOidcState.State

		appSettings := &etc.AppSettings{
			OidcStateSecret: []byte("super-secret-secret"),
		}

		payload := []byte("oidc-state-payload")
		payloadSigner := NewMockPayloadSigner(t)
		payloadSigner.EXPECT().Verify(
			appSettings.OidcStateSecret,
			signedOidcState,
		).Return(payload, true)

		jsonSerialiser := NewMockJsonSerialiser(t)
		jsonSerialiser.EXPECT().Unmarshal(payload, mock.Anything).
			Run(func(_ []byte, v any) {
				if persistedState, ok := v.(*oidcState); ok {
					persistedState.State = wantOidcState.State
				}
			}).
			Return(nil)

		idTokenString := "mock-id-token-string"
		oauthToken := &oauth2.Token{}
		oauthToken = oauthToken.WithExtra(map[string]any{
			"id_token": idTokenString,
		})

		oauthClient := NewMockOAuthClient(t)
		oauthClient.
			EXPECT().Exchange(ctx, authCode, mock.Anything).
			Return(oauthToken, nil)

		idToken := NewMockIdToken(t)
		verifier := NewMockIdTokenVerifier(t)
		verifier.
			EXPECT().Verify(ctx, idTokenString).
			Return(idToken, nil)

		idToken.
			EXPECT().Claims(mock.Anything).
			Return(wantErr)

		service := &AuthService{
			appSettings:    appSettings,
			payloadSigner:  payloadSigner,
			jsonSerialiser: jsonSerialiser,
			oauthClient:    oauthClient,
			verifier:       verifier,
		}

		gotResp, err := service.AuthenticateUser(
			ctx,
			authCode,
			returnedState,
			signedOidcState,
		)
		assert.Nil(t, gotResp)
		assert.EqualError(t, err, wantErr.Error())
	})

	t.Run("validates id token nonce against persisted oidc state", func(t *testing.T) {
		wantErr := errors.New(constants.ErrMsgOidcNonceValueMismatch)

		ctx := context.Background()
		authCode := "an-auth-code"
		signedOidcState := "signed-oidc-state"

		wantOidcState := &oidcState{State: "want state", Nonce: "want nonce"}
		returnedState := wantOidcState.State
		returnedClaims := portalTokenClaims{
			Nonce: "incorrect nonce",
		}

		appSettings := &etc.AppSettings{
			OidcStateSecret: []byte("super-secret-secret"),
		}

		payload := []byte("oidc-state-payload")
		payloadSigner := NewMockPayloadSigner(t)
		payloadSigner.EXPECT().Verify(
			appSettings.OidcStateSecret,
			signedOidcState,
		).Return(payload, true)

		jsonSerialiser := NewMockJsonSerialiser(t)
		jsonSerialiser.EXPECT().Unmarshal(payload, mock.Anything).
			Run(func(_ []byte, v any) {
				if persistedState, ok := v.(*oidcState); ok {
					persistedState.State = wantOidcState.State
					persistedState.Nonce = wantOidcState.Nonce
				}
			}).
			Return(nil)

		idTokenString := "mock-id-token-string"
		oauthToken := &oauth2.Token{}
		oauthToken = oauthToken.WithExtra(map[string]any{
			"id_token": idTokenString,
		})

		oauthClient := NewMockOAuthClient(t)
		oauthClient.
			EXPECT().Exchange(ctx, authCode, mock.Anything).
			Return(oauthToken, nil)

		idToken := NewMockIdToken(t)
		verifier := NewMockIdTokenVerifier(t)
		verifier.
			EXPECT().Verify(ctx, idTokenString).
			Return(idToken, nil)

		idToken.
			EXPECT().Claims(mock.Anything).
			Run(func(v any) {
				if claims, ok := v.(*portalTokenClaims); ok {
					claims.Nonce = returnedClaims.Nonce
				}
			}).
			Return(nil)

		service := &AuthService{
			appSettings:    appSettings,
			payloadSigner:  payloadSigner,
			jsonSerialiser: jsonSerialiser,
			oauthClient:    oauthClient,
			verifier:       verifier,
		}

		gotResp, err := service.AuthenticateUser(
			ctx,
			authCode,
			returnedState,
			signedOidcState,
		)
		assert.Nil(t, gotResp)
		assert.EqualError(t, err, wantErr.Error())
	})

	t.Run("handles upsert errors", func(t *testing.T) {
		wantErr := errors.New("some upsert error")

		ctx := context.Background()
		authCode := "an-auth-code"
		signedOidcState := "signed-oidc-state"

		wantOidcState := &oidcState{State: "want state", Nonce: "want nonce"}
		returnedState := wantOidcState.State
		returnedClaims := portalTokenClaims{
			Nonce:        wantOidcState.Nonce,
			Oid:          "want-oid",
			GivenName:    "want given name",
			FamilyName:   "want family name",
			UserName:     "want username",
			EmailAddress: "want email address",
		}

		appSettings := &etc.AppSettings{
			OidcStateSecret: []byte("super-secret-secret"),
		}

		payload := []byte("oidc-state-payload")
		payloadSigner := NewMockPayloadSigner(t)
		payloadSigner.EXPECT().Verify(
			appSettings.OidcStateSecret,
			signedOidcState,
		).Return(payload, true)

		jsonSerialiser := NewMockJsonSerialiser(t)
		jsonSerialiser.EXPECT().Unmarshal(payload, mock.Anything).
			Run(func(_ []byte, v any) {
				if persistedState, ok := v.(*oidcState); ok {
					persistedState.State = wantOidcState.State
					persistedState.Nonce = wantOidcState.Nonce
				}
			}).
			Return(nil)

		idTokenString := "mock-id-token-string"
		oauthToken := &oauth2.Token{}
		oauthToken = oauthToken.WithExtra(map[string]any{
			"id_token": idTokenString,
		})

		oauthClient := NewMockOAuthClient(t)
		oauthClient.
			EXPECT().Exchange(ctx, authCode, mock.Anything).
			Return(oauthToken, nil)

		idToken := NewMockIdToken(t)
		verifier := NewMockIdTokenVerifier(t)
		verifier.
			EXPECT().Verify(ctx, idTokenString).
			Return(idToken, nil)

		idToken.
			EXPECT().Claims(mock.Anything).
			Run(func(v any) {
				if claims, ok := v.(*portalTokenClaims); ok {
					claims.Nonce = returnedClaims.Nonce
					claims.Oid = returnedClaims.Oid
					claims.GivenName = returnedClaims.GivenName
					claims.FamilyName = returnedClaims.FamilyName
					claims.UserName = returnedClaims.UserName
					claims.EmailAddress = returnedClaims.EmailAddress
				}
			}).
			Return(nil)

		userRepo := NewMockUserRepo(t)
		userRepo.
			EXPECT().UpsertUser(
			returnedClaims.Oid,
			returnedClaims.GivenName,
			returnedClaims.FamilyName,
			returnedClaims.UserName,
			returnedClaims.EmailAddress,
		).
			Return(nil, wantErr)

		service := &AuthService{
			appSettings:    appSettings,
			payloadSigner:  payloadSigner,
			jsonSerialiser: jsonSerialiser,
			oauthClient:    oauthClient,
			verifier:       verifier,
			userRepo:       userRepo,
		}

		gotResp, err := service.AuthenticateUser(
			ctx,
			authCode,
			returnedState,
			signedOidcState,
		)
		assert.Nil(t, gotResp)
		assert.EqualError(t, err, wantErr.Error())
	})

	t.Run("handles errors returned when signing the app token", func(t *testing.T) {
		wantErr := errors.New("some sign error")

		ctx := context.Background()
		authCode := "an-auth-code"
		signedOidcState := "signed-oidc-state"

		wantOidcState := &oidcState{State: "want state", Nonce: "want nonce"}
		returnedState := wantOidcState.State
		returnedClaims := portalTokenClaims{
			Nonce:        wantOidcState.Nonce,
			Oid:          "want-oid",
			GivenName:    "want given name",
			FamilyName:   "want family name",
			UserName:     "want username",
			EmailAddress: "want email address",
		}
		upsertedUser := &db.User{Id: "upserted-user-id"}

		appSettings := &etc.AppSettings{
			OidcStateSecret: []byte("super-secret-secret"),
			AppJwtSecret:    []byte("app jwt secret"),
		}

		payload := []byte("oidc-state-payload")
		payloadSigner := NewMockPayloadSigner(t)
		payloadSigner.EXPECT().Verify(
			appSettings.OidcStateSecret,
			signedOidcState,
		).Return(payload, true)

		jsonSerialiser := NewMockJsonSerialiser(t)
		jsonSerialiser.EXPECT().Unmarshal(payload, mock.Anything).
			Run(func(_ []byte, v any) {
				if persistedState, ok := v.(*oidcState); ok {
					persistedState.State = wantOidcState.State
					persistedState.Nonce = wantOidcState.Nonce
				}
			}).
			Return(nil)

		idTokenString := "mock-id-token-string"
		oauthToken := &oauth2.Token{}
		oauthToken = oauthToken.WithExtra(map[string]any{
			"id_token": idTokenString,
		})

		oauthClient := NewMockOAuthClient(t)
		oauthClient.
			EXPECT().Exchange(ctx, authCode, mock.Anything).
			Return(oauthToken, nil)

		idToken := NewMockIdToken(t)
		verifier := NewMockIdTokenVerifier(t)
		verifier.
			EXPECT().Verify(ctx, idTokenString).
			Return(idToken, nil)

		idToken.
			EXPECT().Claims(mock.Anything).
			Run(func(v any) {
				if claims, ok := v.(*portalTokenClaims); ok {
					claims.Nonce = returnedClaims.Nonce
					claims.Oid = returnedClaims.Oid
					claims.GivenName = returnedClaims.GivenName
					claims.FamilyName = returnedClaims.FamilyName
					claims.UserName = returnedClaims.UserName
					claims.EmailAddress = returnedClaims.EmailAddress
				}
			}).
			Return(nil)

		userRepo := NewMockUserRepo(t)
		userRepo.
			EXPECT().UpsertUser(
			returnedClaims.Oid,
			returnedClaims.GivenName,
			returnedClaims.FamilyName,
			returnedClaims.UserName,
			returnedClaims.EmailAddress,
		).
			Return(upsertedUser, nil)

		jwtSigner := NewMockJwtSigner(t)
		jwtSigner.
			EXPECT().Sign(mock.Anything, appSettings.AppJwtSecret).
			Return("", wantErr)

		service := &AuthService{
			appSettings:    appSettings,
			payloadSigner:  payloadSigner,
			jsonSerialiser: jsonSerialiser,
			oauthClient:    oauthClient,
			verifier:       verifier,
			userRepo:       userRepo,
			jwtSigner:      jwtSigner,
		}

		gotResp, err := service.AuthenticateUser(
			ctx,
			authCode,
			returnedState,
			signedOidcState,
		)
		assert.Nil(t, gotResp)
		assert.EqualError(t, err, wantErr.Error())
	})

	t.Run("returns valid response", func(t *testing.T) {
		ctx := context.Background()
		authCode := "an-auth-code"
		signedOidcState := "signed-oidc-state"
		idTokenString := "mock-id-token-string"

		wantOidcState := &oidcState{
			State:         "want state",
			Nonce:         "want nonce",
			RequestedPath: "want requested path",
		}
		returnedState := wantOidcState.State
		returnedClaims := portalTokenClaims{
			Nonce:        wantOidcState.Nonce,
			Oid:          "want-oid",
			GivenName:    "want given name",
			FamilyName:   "want family name",
			UserName:     "want username",
			EmailAddress: "want email address",
		}
		upsertedUser := &db.User{
			Id:           "upserted-user-id",
			Oid:          "upserted-user-oid",
			EmailAddress: "upserted-user-email-address",
		}
		wantSignedToken := "want-signed-token"

		wantAppTokenClaims := &appClaims{
			IdToken:   idTokenString,
			UserId:    upsertedUser.Id,
			UserOid:   upsertedUser.Oid,
			UserEmail: upsertedUser.EmailAddress,
		}
		var capturedToken *jwt.Token = nil

		appSettings := &etc.AppSettings{
			OidcStateSecret: []byte("super-secret-secret"),
			AppJwtSecret:    []byte("app jwt secret"),
		}

		payload := []byte("oidc-state-payload")
		payloadSigner := NewMockPayloadSigner(t)
		payloadSigner.EXPECT().Verify(
			appSettings.OidcStateSecret,
			signedOidcState,
		).Return(payload, true)

		jsonSerialiser := NewMockJsonSerialiser(t)
		jsonSerialiser.EXPECT().Unmarshal(payload, mock.Anything).
			Run(func(_ []byte, v any) {
				if persistedState, ok := v.(*oidcState); ok {
					persistedState.State = wantOidcState.State
					persistedState.Nonce = wantOidcState.Nonce
					persistedState.RequestedPath = wantOidcState.RequestedPath
				}
			}).
			Return(nil)

		oauthToken := &oauth2.Token{}
		oauthToken = oauthToken.WithExtra(map[string]any{
			"id_token": idTokenString,
		})

		oauthClient := NewMockOAuthClient(t)
		oauthClient.
			EXPECT().Exchange(ctx, authCode, mock.Anything).
			Return(oauthToken, nil)

		idToken := NewMockIdToken(t)
		verifier := NewMockIdTokenVerifier(t)
		verifier.
			EXPECT().Verify(ctx, idTokenString).
			Return(idToken, nil)

		idToken.
			EXPECT().Claims(mock.Anything).
			Run(func(v any) {
				if claims, ok := v.(*portalTokenClaims); ok {
					claims.Nonce = returnedClaims.Nonce
					claims.Oid = returnedClaims.Oid
					claims.GivenName = returnedClaims.GivenName
					claims.FamilyName = returnedClaims.FamilyName
					claims.UserName = returnedClaims.UserName
					claims.EmailAddress = returnedClaims.EmailAddress
				}
			}).
			Return(nil)

		userRepo := NewMockUserRepo(t)
		userRepo.
			EXPECT().UpsertUser(
			returnedClaims.Oid,
			returnedClaims.GivenName,
			returnedClaims.FamilyName,
			returnedClaims.UserName,
			returnedClaims.EmailAddress,
		).
			Return(upsertedUser, nil)

		jwtSigner := NewMockJwtSigner(t)
		jwtSigner.
			EXPECT().Sign(mock.Anything, appSettings.AppJwtSecret).
			Run(func(token *jwt.Token, privateKey []byte) {
				capturedToken = token
			}).
			Return(wantSignedToken, nil)

		service := &AuthService{
			appSettings:    appSettings,
			payloadSigner:  payloadSigner,
			jsonSerialiser: jsonSerialiser,
			oauthClient:    oauthClient,
			verifier:       verifier,
			userRepo:       userRepo,
			jwtSigner:      jwtSigner,
		}

		gotResp, err := service.AuthenticateUser(
			ctx,
			authCode,
			returnedState,
			signedOidcState,
		)
		assert.Nil(t, err)
		assert.Equal(t, wantOidcState.RequestedPath, gotResp.RequestedPath)

		gotAppClaims, ok := capturedToken.Claims.(appClaims)
		assert.Equal(t, true, ok)
		assert.Equal(t, wantAppTokenClaims.IdToken, gotAppClaims.IdToken)
		assert.Equal(t, wantAppTokenClaims.UserId, gotAppClaims.UserId)
		assert.Equal(t, wantAppTokenClaims.UserOid, gotAppClaims.UserOid)
		assert.Equal(t, wantAppTokenClaims.UserEmail, gotAppClaims.UserEmail)

	})
}

func TestParseUserJwtCookie(t *testing.T) {
	t.Run("handles parsing errors", func(t *testing.T) {
		wantError := errors.New("failed to parse claims")
		jwtToken := "some-token"

		jwtSigner := NewMockJwtSigner(t)
		jwtSigner.
			EXPECT().ParseWithClaims(
			jwtToken,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		).
			Return(nil, wantError)

		service := &AuthService{
			jwtSigner: jwtSigner,
		}

		claims, err := service.ParseUserJwtCookie(jwtToken)
		assert.Nil(t, claims)
		assert.EqualError(t, err, wantError.Error())
	})

	t.Run("returns app jwt secret in the key func", func(t *testing.T) {
		appSettings := &etc.AppSettings{
			AppJwtSecret: []byte("want-app-jwt-secret"),
		}
		var capturedKeyFuncVal any = nil
		var capturedKeyFuncErr error = nil

		wantAppClaims := appClaims{
			IdToken:   "want-id-token",
			UserId:    "want-user-id",
			UserOid:   "want-user-oid",
			UserEmail: "want-user-email",
		}
		jwtToken := createJwt(t, wantAppClaims, appSettings.AppJwtSecret)

		jwtSigner := NewMockJwtSigner(t)
		jwtSigner.EXPECT().ParseWithClaims(
			jwtToken,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		).Run(func(_ string, claims jwt.Claims, keyFunc jwt.Keyfunc, _ ...jwt.ParserOption) {
			capturedKeyFuncVal, capturedKeyFuncErr = keyFunc(nil)
			if claims, ok := claims.(*appClaims); ok {
				claims.IdToken = wantAppClaims.IdToken
				claims.UserId = wantAppClaims.UserId
				claims.UserOid = wantAppClaims.UserOid
				claims.UserEmail = wantAppClaims.UserEmail
			}
		}).Return(nil, nil)

		service := &AuthService{
			appSettings: appSettings,
			jwtSigner:   jwtSigner,
		}

		_, err := service.ParseUserJwtCookie(jwtToken)
		assert.Nil(t, err)
		assert.Nil(t, capturedKeyFuncErr)
		assert.Equal(t, capturedKeyFuncVal, appSettings.AppJwtSecret)
	})

	t.Run("returns errors if any claims are empty", func(t *testing.T) {
		testData := []struct {
			claims         appClaims
			expectedErrMsg string
		}{
			{
				claims: appClaims{
					UserId:    "want-user-id",
					UserOid:   "want-user-oid",
					UserEmail: "want-user-email",
				},
				expectedErrMsg: constants.ErrMsgParseAppJwtErrUnableToReadIdToken,
			},
			{
				claims: appClaims{
					IdToken:   "want-id-token",
					UserOid:   "want-user-oid",
					UserEmail: "want-user-email",
				},
				expectedErrMsg: constants.ErrMsgParseAppJwtErrUnableToReadUserId,
			},
			{
				claims: appClaims{
					IdToken:   "want-id-token",
					UserId:    "want-user-id",
					UserEmail: "want-user-email",
				},
				expectedErrMsg: constants.ErrMsgParseAppJwtErrUnableToReadUserOid,
			},
			{
				claims: appClaims{
					IdToken: "want-id-token",
					UserId:  "want-user-id",
					UserOid: "want-user-oid",
				},
				expectedErrMsg: constants.ErrMsgParseAppJwtErrUnableToReadUserEmail,
			},
		}

		for _, td := range testData {
			appSettings := &etc.AppSettings{
				AppJwtSecret: []byte("want-app-jwt-secret"),
			}

			jwtToken := createJwt(t, td.claims, appSettings.AppJwtSecret)

			jwtSigner := NewMockJwtSigner(t)
			jwtSigner.EXPECT().ParseWithClaims(
				jwtToken,
				mock.Anything,
				mock.Anything,
				mock.Anything,
			).Run(func(_ string, claims jwt.Claims, keyFunc jwt.Keyfunc, _ ...jwt.ParserOption) {
				if claims, ok := claims.(*appClaims); ok {
					claims.IdToken = td.claims.IdToken
					claims.UserId = td.claims.UserId
					claims.UserOid = td.claims.UserOid
					claims.UserEmail = td.claims.UserEmail
				}
			}).Return(nil, nil)

			service := &AuthService{
				appSettings: appSettings,
				jwtSigner:   jwtSigner,
			}

			claims, err := service.ParseUserJwtCookie(jwtToken)
			assert.Nil(t, claims)
			assert.EqualError(t, err, td.expectedErrMsg)
		}
	})

	t.Run("returns claims from the JWT", func(t *testing.T) {
		appSettings := &etc.AppSettings{
			AppJwtSecret: []byte("want-app-jwt-secret"),
		}

		wantAppClaims := appClaims{
			IdToken:   "want-id-token",
			UserId:    "want-user-id",
			UserOid:   "want-user-oid",
			UserEmail: "want-user-email",
		}
		jwtToken := createJwt(t, wantAppClaims, appSettings.AppJwtSecret)

		jwtSigner := NewMockJwtSigner(t)
		jwtSigner.EXPECT().ParseWithClaims(
			jwtToken,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		).Run(func(_ string, claims jwt.Claims, keyFunc jwt.Keyfunc, _ ...jwt.ParserOption) {
			if claims, ok := claims.(*appClaims); ok {
				claims.IdToken = wantAppClaims.IdToken
				claims.UserId = wantAppClaims.UserId
				claims.UserOid = wantAppClaims.UserOid
				claims.UserEmail = wantAppClaims.UserEmail
			}
		}).Return(nil, nil)

		service := &AuthService{
			appSettings: appSettings,
			jwtSigner:   jwtSigner,
		}

		claims, err := service.ParseUserJwtCookie(jwtToken)
		assert.Nil(t, err)
		assert.Equal(t, wantAppClaims.IdToken, claims.IdToken)
		assert.Equal(t, wantAppClaims.UserId, claims.UserId)
		assert.Equal(t, wantAppClaims.UserOid, claims.UserOid)
		assert.Equal(t, wantAppClaims.UserEmail, claims.UserEmail)
	})
}

func TestRetrieveUserById(t *testing.T) {
	t.Parallel()

	t.Run("is a thin wrapper around user repo method", func(t *testing.T) {
		testData := []struct {
			userId string
			user   *db.User
			err    error
		}{
			{
				userId: "test-id-1",
				user:   &db.User{},
				err:    nil,
			},
			{
				userId: "test-id-2",
				user:   nil,
				err:    errors.New("some retrieve user by id err"),
			},
		}

		for _, td := range testData {
			userRepo := NewMockUserRepo(t)
			userRepo.EXPECT().RetrieveUserById(td.userId).Return(td.user, td.err)

			service := AuthService{
				userRepo: userRepo,
			}

			gotUser, gotErr := service.RetrieveUserById(td.userId)
			assert.Equal(t, td.user, gotUser)
			assert.Equal(t, td.err, gotErr)
		}
	})
}

type buildOidcProviderArgs struct {
	ctx    context.Context
	issuer string
}

type buildOidcProviderResp struct {
	provider OidcProvider
	err      error
}

type oidcProviderFactory struct {
	t            testing.TB
	buildCalls   []buildOidcProviderArgs
	buildReturns buildOidcProviderResp
}

func (m *oidcProviderFactory) Build(
	ctx context.Context,
	issuer string,
) (OidcProvider, error) {
	m.t.Helper()

	m.buildCalls = append(m.buildCalls, buildOidcProviderArgs{
		ctx:    ctx,
		issuer: issuer,
	})
	return m.buildReturns.provider, m.buildReturns.err
}

type deterministicRandReader struct {
	t         testing.TB
	callCount int
	patterns  [][]byte
	returns   []error
}

// ReadNext fills bytes with looping patterns in sequence
func (m *deterministicRandReader) ReadNext(bytes []byte) (n int, err error) {
	m.t.Helper()

	nextPattern := []byte("abcdefg")
	if len(m.patterns) > 0 {
		nextPattern = m.patterns[m.callCount%len(m.patterns)]
	}

	if len(m.returns) > 0 {
		err = m.returns[m.callCount%len(m.returns)]
	}

	m.callCount = m.callCount + 1
	if err != nil {
		return 0, err
	}

	return m.ReadPattern(bytes, nextPattern)
}

// Read pattern fills bytes with the looping pattern
func (m *deterministicRandReader) ReadPattern(bytes []byte, pattern []byte) (n int, err error) {
	plen := len(pattern)

	var i int
	for i = range bytes {
		bytes[i] = pattern[i%plen]
	}
	return i + 1, nil
}

func createJwt(t *testing.T, claims appClaims, secret []byte) string {
	t.Helper()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	s, err := token.SignedString(secret)
	require.NoError(t, err)

	return s
}
