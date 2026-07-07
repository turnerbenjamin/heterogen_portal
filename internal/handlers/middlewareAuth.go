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

// ParseJwtMiddleware parses app jwt cookies, retrieves the user and adds it
// to pipeline context. If there is no cookie, or it cannot be parsed, user will
// not be set on the pipeline context
func ParseJwtMiddleware[T UserState](authService AuthService) Middleware[T] {
	return func(next AppHandler[T]) AppHandler[T] {
		return func(w http.ResponseWriter, r *http.Request, c *PipelineContext[T]) *AppError {
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

// RequireSignInMiddleware accesses user state from the pipeline context. If
// user is nil, they will be redirected to authenticate with the auth provider.
// This middleware should be called after NewParseJwtMiddleware
func RequireSignInMiddleware[T UserState](authService AuthService) Middleware[T] {
	return func(next AppHandler[T]) AppHandler[T] {
		return func(w http.ResponseWriter, r *http.Request, c *PipelineContext[T]) *AppError {
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
