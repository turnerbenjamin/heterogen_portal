package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/MicahParks/keyfunc"
	"github.com/golang-jwt/jwt/v4"
)

var (
	jwksMu    sync.RWMutex
	jwks      *keyfunc.JWKS
	issuer    string
	audience  string
	jwksURL   string
	initOnce  sync.Once
	initErr   error
	refreshIn = time.Hour
)

type PortalTokenClaims struct {
	Oid          string
	GivenName    string
	FamilyName   string
	UserName     string
	EmailAddress string
}

func initOIDC() error {
	initOnce.Do(func() {
		authority := "https://heterogenportalusers.ciamlogin.com/09af08d8-4dbc-4605-b414-2c6d9c7a0e70/v2.0"
		audience = "bb3c49a3-171c-49ab-a96b-8b247f600c44"

		resp, err := http.Get(authority + "/.well-known/openid-configuration")
		if err != nil {
			initErr = err
			return
		}
		defer resp.Body.Close()

		var cfg struct {
			JwksURI string `json:"jwks_uri"`
			Issuer  string `json:"issuer"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&cfg); err != nil {
			initErr = err
			return
		}
		if cfg.JwksURI == "" {
			initErr = errors.New("jwks_uri missing from openid config")
			return
		}
		jwksURL = cfg.JwksURI
		issuer = cfg.Issuer

		j, err := keyfunc.Get(jwksURL, keyfunc.Options{
			RefreshInterval:   refreshIn,
			RefreshUnknownKID: true,
			RefreshTimeout:    10 * time.Second,
		})
		if err != nil {
			initErr = err
			return
		}
		jwksMu.Lock()
		jwks = j
		jwksMu.Unlock()
	})
	return initErr
}

// ValidateToken validates and returns claims; uses cached JWKS.
func ValidateToken(ctx context.Context, tokenString string) (*PortalTokenClaims, error) {
	if tokenString == "" {
		return nil, errors.New("invalid token string")
	}
	fields := strings.Fields(tokenString)
	if len(fields) == 0 {
		return nil, errors.New("invalid token string")
	}

	if len(fields) == 1 {
		tokenString = fields[0]
	} else {
		key := fields[0]
		tokenString = fields[1]
		if !strings.EqualFold(key, "bearer") {
			return nil, errors.New("invalid token prefix")
		}
	}

	if err := initOIDC(); err != nil {
		return nil, err
	}
	jwksMu.RLock()
	kf := jwks.Keyfunc
	jwksMu.RUnlock()

	// Parse using jwt v4 and the keyfunc returned by keyfunc.Get
	token, err := jwt.ParseWithClaims(tokenString, jwt.MapClaims{}, kf)
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	// validate audience and issuer
	if !claims.VerifyAudience(audience, true) {
		if a, ok := claims["aud"]; ok {
			return nil, fmt.Errorf("invalid audience; token aud=%v expected=%q", a, audience)
		}
		return nil, errors.New("invalid audience")
	}
	if !claims.VerifyIssuer(issuer, true) {
		return nil, errors.New("invalid issuer")
	}
	// optional: enforce algorithm
	if token.Method.Alg() != "RS256" {
		return nil, errors.New("unexpected signing algorithm")
	}
	return parsePortalTokenClaims(claims)
}

// GetUserIDFromClaims returns a stable user identifier from the token claims.
// Checks oid (preferred for AAD), then sub, preferred_username, upn, email.
func parsePortalTokenClaims(claims jwt.MapClaims) (*PortalTokenClaims, error) {
	if claims == nil {
		return nil, errors.New("no claims")
	}

	for c := range claims {
		println(c)
	}

	oid, err := parseClaim(claims, "oid")
	if err != nil {
		return nil, err
	}

	givenName, err := parseClaim(claims, "given_name")
	if err != nil {
		return nil, err
	}

	lastName, err := parseClaim(claims, "family_name")
	if err != nil {
		return nil, err
	}

	emailAddress, err := parseClaim(claims, "email")
	if err != nil {
		return nil, err
	}

	userName, err := parseClaim(claims, "name")
	if err != nil {
		return nil, err
	}

	return &PortalTokenClaims{
		Oid:          oid,
		GivenName:    givenName,
		FamilyName:   lastName,
		UserName:     userName,
		EmailAddress: emailAddress,
	}, nil
}

func parseClaim(claims jwt.MapClaims, claim string) (string, error) {
	value, ok := claims[claim].(string)
	if !ok {
		return "", errors.New(fmt.Sprintf("unable to parse %s from token claims", claim))
	}
	return value, nil
}
