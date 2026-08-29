package app

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"net/http"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/golang-jwt/jwt/v5"
	"github.com/turnerbenjamin/heterogen_portal/internal/etc"
	"github.com/turnerbenjamin/heterogen_portal/internal/query"
	"github.com/turnerbenjamin/heterogen_portal/internal/query/paginationTokens"
	"github.com/turnerbenjamin/heterogen_portal/internal/query/queryBuilder"
	"github.com/turnerbenjamin/heterogen_portal/internal/query/queryParser"
	"github.com/turnerbenjamin/heterogen_portal/internal/services"
	"github.com/turnerbenjamin/heterogen_portal/internal/utils"
	"golang.org/x/oauth2"
)

type appDependencies struct {
	jsonSerialiser       *stdJsonSerialiser
	tokenSigner          *jwtTokenSigner
	payloadSigner        *utils.PayloadSigner
	httpClient           *http.Client
	newOidcProvider      func(ctx context.Context, issuer string) (services.OidcProvider, error)
	randReader           services.RandReader
	queryParser          queryBuilder.QueryParser
	nextPageTokenBuilder query.PagingTokenBuilder
}

func initAppDependencies(appSettings *etc.AppSettings) (*appDependencies, error) {
	nextPageTokenBuilder, err := paginationTokens.NewNextPageTokenBuilder(
		&utils.PayloadSigner{},
		appSettings.QueryTokenSecret,
	)
	if err != nil {
		return nil, err
	}

	return &appDependencies{
		jsonSerialiser:       &stdJsonSerialiser{},
		tokenSigner:          &jwtTokenSigner{},
		payloadSigner:        &utils.PayloadSigner{},
		httpClient:           &http.Client{},
		queryParser:          &queryParser.QuerySyntaxParser{},
		newOidcProvider:      oidcNewProvider,
		randReader:           rand.Read,
		nextPageTokenBuilder: nextPageTokenBuilder,
	}, nil
}

type stdJsonSerialiser struct{}

func (s *stdJsonSerialiser) Marshal(v any) ([]byte, error) {
	return json.Marshal(v)
}
func (s *stdJsonSerialiser) Unmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

type idToken struct {
	token *oidc.IDToken
}

func (t *idToken) Claims(v any) error {
	return t.token.Claims(v)
}

type idTokenVerifier struct {
	verifier *oidc.IDTokenVerifier
}

func (v *idTokenVerifier) Verify(
	ctx context.Context,
	rawIDToken string,
) (services.IdToken, error) {
	token, err := v.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, err
	}

	return &idToken{token}, nil
}

type oidcProvider struct {
	provider *oidc.Provider
}

func (p *oidcProvider) Endpoint() oauth2.Endpoint {
	return p.provider.Endpoint()
}

func (p *oidcProvider) Verifier(config *oidc.Config) services.IdTokenVerifier {
	return &idTokenVerifier{
		verifier: p.provider.Verifier(config),
	}
}

func oidcNewProvider(ctx context.Context, issuer string) (services.OidcProvider, error) {
	provider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		return nil, err
	}

	return &oidcProvider{
		provider: provider,
	}, nil
}

type jwtTokenSigner struct{}

func (sp *jwtTokenSigner) Sign(token *jwt.Token, key []byte) (string, error) {
	return token.SignedString(key)
}

func (sp *jwtTokenSigner) ParseWithClaims(
	tokenString string,
	claims jwt.Claims,
	keyFunc jwt.Keyfunc,
	parserOptions ...jwt.ParserOption,
) (*jwt.Token, error) {
	return jwt.ParseWithClaims(
		tokenString,
		claims,
		keyFunc,
		parserOptions...,
	)
}
