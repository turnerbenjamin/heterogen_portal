// Package handlers contains HTTP handlers for the application.
//
// This file provides handlers responsible for authentication
package handlers

import (
	"log/slog"
	"net/http"

	"github.com/turnerbenjamin/heterogen_portal/internal/constants"
	"github.com/turnerbenjamin/heterogen_portal/internal/db"
	"github.com/turnerbenjamin/heterogen_portal/internal/etc"
)

// UserState is passed in PipelineContext in pipelines using
// NewParseJwtMiddleware to allow the user state to flow down the pipeline
type UserState struct {
	User *db.User
}

// NewParseJwtMiddleware will, on the happy path, access and parse a JWT cookie
// from the request, retrieve the user from the token's subject and attach the
// user record to the pipeline context. For all other paths, it will log any
// errors, unset invalid cookies and continue, unfazed, to the next handler
func NewParseJwtMiddleware(
	appSettings *etc.AppSettings,
	authService AuthService,
) Middleware[UserState] {
	return func(next AppHandler[UserState]) AppHandler[UserState] {
		return func(w http.ResponseWriter, r *http.Request, c *PipelineContext[UserState]) *AppError {
			jwtCookie, err := r.Cookie(constants.IdentifierJwtCookie)
			if err != nil {
				return next(w, r, c)
			}

			user, err := authService.ParseUserJwtCookie(jwtCookie.Value)
			if err != nil {
				c.AddLoggerKV(slog.String(
					constants.EmptyAppErrorString,
					err.Error(),
				))
				unsetJwtCookie(w)
			} else {
				c.state.User = user
			}

			return next(w, r, c)
		}
	}
}
