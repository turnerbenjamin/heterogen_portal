// Package handlers contains HTTP handlers for the application.
//
// This file provides handlers responsible for authentication
package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/turnerbenjamin/heterogen_portal/internal/constants"
	"github.com/turnerbenjamin/heterogen_portal/internal/db"
)

// UserState is passed in PipelineContext in pipelines using
// NewParseJwtMiddleware to allow the user state to flow down the pipeline
type UserState struct {
	User *db.User
}

// TokenValidator validates auth tokens
type TokenValidator interface {
	ValidatePortalToken(
		ctx context.Context,
		tokenString string,
	) (*PortalTokenClaims, error)
}

// UserRepo performs database operations on the user table
type UserRepo interface {
	UpsertUser(
		ctx context.Context,
		oid, givenName, familyName, userName, emailAddress string,
	) (*db.User, error)
	RetrieveUserById(id string) (*db.User, error)
	RetrieveUserByOid(id string) (*db.User, error)
}

// TokenSigner signs JWT Tokens
type TokenSigner interface {
	Sign(token *jwt.Token) (string, error)
}

// TokenParser parses JWT Tokens
type TokenParser interface {
	ParseWithClaims(
		jswtString string,
		claims *jwt.RegisteredClaims,
	) (*jwt.Token, error)
}

// TokenSignerAndParser can sign and parse JWT tokens
type TokenSignerAndParser interface {
	TokenSigner
	TokenParser
}

// PostSignInHandler handles POST requests to the /sign-in endpoint. It parses
// and validates any Authorization Token (this will be a token issued by MSAL).
// If the token is valid:
// - a user record is upserted based on the token claims
// - a jwt cookie is issued for the app
// - the user is redirected to root
func PostSignInHandler(
	tokenValidator TokenValidator,
	tokenSigner TokenSigner,
	ts TemplateStore,
	userRepo UserRepo,
) AppHandler[NoState] {
	return func(w http.ResponseWriter, r *http.Request, c *PipelineContext[NoState]) *AppError {
		bearerToken := r.Header.Get("Authorization")
		tokenClaims, err := tokenValidator.ValidatePortalToken(r.Context(), bearerToken)
		if err != nil {
			return &AppError{
				Code:       http.StatusUnauthorized,
				ToastError: constants.ErrMsgUnauthorised,
				PageErrors: []string{constants.ErrMsgUnauthorised},
				innerError: err,
			}
		}

		user, err := userRepo.UpsertUser(
			r.Context(),
			tokenClaims.Oid,
			tokenClaims.GivenName,
			tokenClaims.FamilyName,
			tokenClaims.UserName,
			tokenClaims.EmailAddress,
		)
		if err != nil {
			return NewServerError(err)
		}

		// Create new JWT Token
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
			Subject:   user.Id,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		})

		tokenString, err := tokenSigner.Sign(token)
		if err != nil {
			return NewServerError(err)
		}
		setJwtCookie(w, tokenString)

		if r.Header.Get("HX-Request") != "" {
			w.Header().Set("HX-Redirect", "/")
			w.WriteHeader(http.StatusSeeOther)
		} else {
			http.Redirect(w, r, "/", http.StatusSeeOther)
		}
		return nil
	}
}

// NewParseJwtMiddleware will, on the happy path, access and parse a JWT cookie
// from the request, retrieve the user from the token's subject and attach the
// user record to the pipeline context. For all other paths, it will log any
// errors, unset invalid cookies and continue, unfazed, to the next handler
func NewParseJwtMiddleware(tokenParser TokenParser, userRepo UserRepo) Middleware[UserState] {
	return func(next AppHandler[UserState]) AppHandler[UserState] {
		return func(w http.ResponseWriter, r *http.Request, c *PipelineContext[UserState]) *AppError {
			jwtCookie, err := r.Cookie(constants.IdentifierJwtCookie)
			if err != nil {
				return next(w, r, c)
			}

			userId, ok := parseUserJwtCookie(jwtCookie, tokenParser, c)
			if ok {
				user, err := userRepo.RetrieveUserById(userId)
				if err != nil {
					c.AddLoggerKV(slog.String(
						constants.SlogKeyNonFatalErrRetrieveUserById,
						err.Error(),
					))
				} else {
					c.state.User = user
				}
			}

			if c.state.User == nil {
				unsetJwtCookie(w)
			}

			return next(w, r, c)
		}
	}
}

// parseUserJwtCookie is a helper method for parsing a jwt token stored in a
// jwt cookie and returning the user id, contained in the token's subject claim.
// It will handle any errors with logging and return false if there is any issue
// parsing the cookie.
func parseUserJwtCookie(
	jwtCookie *http.Cookie,
	tokenParser TokenParser,
	c *PipelineContext[UserState],

) (string, bool) {
	token, err := tokenParser.ParseWithClaims(
		jwtCookie.Value,
		&jwt.RegisteredClaims{},
	)
	if err != nil {
		c.AddLoggerKV(slog.String(
			constants.SlogKeyNonFatalErrParseWithClaims,
			err.Error(),
		))
		return "", false
	}

	id, err := token.Claims.GetSubject()
	if err != nil {
		c.AddLoggerKV(slog.String(
			constants.SlogKeyNonFatalErrClaimsGetSubject,
			err.Error(),
		))
	}
	if err != nil || id == "" {
		return "", false
	}
	return id, true
}

// setJwtCookie sets a cookie on the response with a jwt token as the value
func setJwtCookie(w http.ResponseWriter, tokenString string) {
	http.SetCookie(w, &http.Cookie{
		Name:        constants.IdentifierJwtCookie,
		Value:       tokenString,
		SameSite:    http.SameSiteStrictMode,
		MaxAge:      int((time.Hour * 24) / time.Second),
		Secure:      true,
		Partitioned: true,
		HttpOnly:    true,
	})
}

// unsetJwtCookie unsets the cookie set by setJwtCookie using a negative max age
// and an expiry date in the past
func unsetJwtCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:        constants.IdentifierJwtCookie,
		SameSite:    http.SameSiteStrictMode,
		MaxAge:      -1,
		Expires:     time.Unix(0, 0).UTC(),
		Secure:      true,
		Partitioned: true,
		HttpOnly:    true,
	})
}
