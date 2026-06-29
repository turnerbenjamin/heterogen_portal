// Package handlers contains HTTP handlers for the application.
//
// This file exposes handlers that render pages via the TemplateStore.
package handlers

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/turnerbenjamin/heterogen_portal/internal/constants"
	"github.com/turnerbenjamin/heterogen_portal/internal/db"
	"github.com/turnerbenjamin/heterogen_portal/internal/services"
	"github.com/turnerbenjamin/heterogen_portal/internal/templates"
)

type AuthService interface {
	ParseUserJwtCookie(tokenString string) (*services.AppClaims, error)
	RetrieveUserById(userId string) (*db.User, error)
	BuildSignInRedirectRequest(requestedPath string) (*services.SignInRedirectRequest, error)
	BuildSignOutRedirectRequest() string
	AuthenticateUser(
		ctx context.Context,
		authorisationCode string,
		returnedState string,
		signedOidcState string,
	) (resp *services.AuthenticateUserResponse, err error)
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	IDToken     string `json:"id_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// GetRootHandler returns a handler for the application root.
func GetRootHandler(ts TemplateStore) AppHandler[UserState] {
	return func(w http.ResponseWriter, r *http.Request, c *PipelineContext[UserState]) *AppError {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return nil
		}

		pageConfig := templates.PageConfig{
			ContentOnly: r.Header.Get(constants.HxRequestHeaderRequest) != "",
			Title:       "HETEROGEN",
		}

		err := ts.Execute(
			templates.TmplPageApp,
			w,
			templates.TemplateArgs{PageConfig: pageConfig, Data: c.state},
		)
		if err != nil {
			return NewServerError(err)
		}
		return nil
	}
}

// GetSignInRedirectHandler handles redirects from the auth provider. It will
// authenticate the user and then redirects to the path requested by the user
func GetSignInRedirectHandler(authService AuthService) AppHandler[NoState] {
	return func(w http.ResponseWriter, r *http.Request, c *PipelineContext[NoState]) *AppError {
		// access signed oidc state from the cookie and clear it up
		cookie, err := r.Cookie(constants.IdentifierOidcStateCookie)
		if err != nil {
			return NewServerError(errors.New(constants.ErrMissingOIDCStateCookie))
		}
		signedOidcState := cookie.Value
		unsetOidcStateCookie(w)

		// extract query params
		code := r.URL.Query().Get("code")
		if code == "" {
			return NewServerError(errors.New(constants.ErrMissingOIDCCodeParam))
		}

		returnedState := r.URL.Query().Get("state")
		if returnedState == "" {
			return NewServerError(errors.New(constants.ErrMissingOIDCStateParam))
		}

		// authenticate the user
		authenticateUserResponse, err := authService.AuthenticateUser(
			r.Context(),
			code,
			returnedState,
			signedOidcState,
		)
		if err != nil {
			return NewServerError(err)
		}

		setJwtCookie(w, authenticateUserResponse.AppToken)

		// redirect user to the app
		redirect(w, r, authenticateUserResponse.RequestedPath)
		return nil
	}
}

// GetSignOutHandler unsets the app jwt cookie and redirects the user to sign
// out from the auth provider
func GetSignOutHandler(authService AuthService) AppHandler[NoState] {
	return func(w http.ResponseWriter, r *http.Request, c *PipelineContext[NoState]) *AppError {
		unsetJwtCookie(w)
		redirectUrl := authService.BuildSignOutRedirectRequest()

		redirect(w, r, redirectUrl)
		return nil
	}
}

// GetSignedOutHandler returns the signed-out page handler
func GetSignedOutHandler(ts TemplateStore) AppHandler[NoState] {
	return func(w http.ResponseWriter, r *http.Request, c *PipelineContext[NoState]) *AppError {
		if r.Header.Get(constants.HxRequestHeaderRequest) != "" {
			return NewServerError(errors.New(constants.ErrMsgHtmxNotSupported))
		}

		pageConfig := templates.PageConfig{
			ContentOnly: false,
			Title:       "HETEROGEN | SIGNED-OUT",
		}

		err := ts.Execute(
			templates.TmplPageUserSignedOut,
			w,
			templates.TemplateArgs{PageConfig: pageConfig, Data: nil},
		)
		if err != nil {
			return NewServerError(err)
		}
		return nil
	}
}

// setJwtCookie sets a cookie on the response with a jwt token as the value
func setJwtCookie(w http.ResponseWriter, tokenString string) {
	http.SetCookie(w, &http.Cookie{
		Name:        constants.IdentifierJwtCookie,
		Value:       tokenString,
		HttpOnly:    true,
		Secure:      true,
		Partitioned: true,
		SameSite:    http.SameSiteLaxMode,
		Path:        "/",
		MaxAge:      int(time.Second) * 60 * 60,
	})
}

// unsetJwtCookie unsets the cookie set by setJwtCookie
func unsetJwtCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:        constants.IdentifierJwtCookie,
		SameSite:    http.SameSiteLaxMode,
		Path:        "/",
		MaxAge:      -1,
		Expires:     time.Unix(0, 0).UTC(),
		Secure:      true,
		Partitioned: true,
		HttpOnly:    true,
	})
}

// setOidcStateCookie sets an oidc state cookie with the given signed state
// string
func setOidcStateCookie(w http.ResponseWriter, oidcState string) {
	http.SetCookie(w, &http.Cookie{
		Name:     constants.IdentifierOidcStateCookie,
		Value:    oidcState,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/sign-in-redirect",
		MaxAge:   600,
	})
}

// unsetOidcStateCookie unsets the cookie set by setOidcStateCookie
func unsetOidcStateCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     constants.IdentifierOidcStateCookie,
		Value:    "",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/sign-in-redirect",
		MaxAge:   -1,
	})
}

// redirect redirects to the given url for both htmx and non-htmx requests
func redirect(w http.ResponseWriter, r *http.Request, url string) {
	if r.Header.Get(constants.HxRequestHeaderRequest) != "" {
		w.Header().Set(constants.HxResponseHeaderRedirect, url)
		w.WriteHeader(http.StatusSeeOther)
	} else {
		http.Redirect(w, r, url, http.StatusSeeOther)
	}
}
