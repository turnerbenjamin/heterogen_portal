// Package handlers contains HTTP handlers for the application.
//
// This file provides handlers responsible for authentication
package handlers

import (
	"log/slog"
	"net/http"

	"github.com/turnerbenjamin/heterogen_portal/internal/constants"
	"github.com/turnerbenjamin/heterogen_portal/internal/db"
)

// NewParseJwtMiddleware will, on the happy path, access and parse a JWT cookie
// from the request, retrieve the user from the token's subject and attach the
// user record to the pipeline context. For all other paths, it will log any
// errors, unset invalid cookies and continue, unfazed, to the next handler
func NewParseJwtMiddleware(authService AuthService) Middleware[UserState] {
	return func(next AppHandler[UserState]) AppHandler[UserState] {
		return func(w http.ResponseWriter, r *http.Request, c *PipelineContext[UserState]) *AppError {
			jwtCookie, err := r.Cookie(constants.IdentifierJwtCookie)
			if err != nil {
				return next(w, r, c)
			}

			var user *db.User = nil
			cookieClaims, err := authService.ParseUserJwtCookie(jwtCookie.Value)
			if err == nil {
				user, err = authService.RetrieveUserById(cookieClaims.UserId)
			}

			if err != nil {
				c.AddLoggerKV(slog.String(
					constants.SlogKeyNonFatalErrParseAppJWT,
					err.Error(),
				))
				unsetJwtCookie(w)
			} else {
				c.state.SetUser(user)
			}

			return next(w, r, c)
		}
	}
}

func NewRequireSignInMiddleware(authService AuthService) Middleware[UserState] {
	return func(next AppHandler[UserState]) AppHandler[UserState] {
		return func(w http.ResponseWriter, r *http.Request, c *PipelineContext[UserState]) *AppError {
			// If user is nil, redirect to sign-in service
			if c.state.GetUser() == nil {
				redirectReq, err := authService.BuildSignInRedirectRequest(r.URL.Path)
				if err != nil {
					return NewServerError(err)
				}

				// store oidc state in a cookie so it can be retrieved on the
				// redirect post sign-in for validation/authorisation when
				// exchanging the returned code for an id token
				setOidcStateCookie(w, redirectReq.SignedOidcState)
				redirect(w, r, redirectReq.Url)
				return nil
			}
			return next(w, r, c)
		}
	}
}
