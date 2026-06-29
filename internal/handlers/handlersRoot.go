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
//
// It redirects unauthenticated users to sign-in and renders the main app
// template for authenticated users.
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

func GetSignInRedirectHandler(
	ts TemplateStore,
	authService AuthService,
) AppHandler[NoState] {
	return func(w http.ResponseWriter, r *http.Request, c *PipelineContext[NoState]) *AppError {
		// access signed oidc state from the cookie and clear it up
		cookie, err := r.Cookie(constants.IdentifierOidcStateCookie)
		if err != nil {
			return NewServerError(errors.New(constants.ErrMissingOIDCStateCookie))
		}
		signedOidcState := cookie.Value
		clearOidcStateCookie(w)

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

// GetSignedOutHandler returns the sign-out page handler.
//
// It clears the JWT cookie and renders the sign-out template. The sign-out page
// will then redirect to allow the user to sign-out from EntraId.
func GetSignOutHandler(authService AuthService) AppHandler[NoState] {
	return func(w http.ResponseWriter, r *http.Request, c *PipelineContext[NoState]) *AppError {
		redirectUrl := authService.BuildSignOutRedirectRequest()
		unsetJwtCookie(w)

		redirect(w, r, redirectUrl)
		return nil
	}
}

// GetSignedOutHandler returns the signed-out page handler.
//
// The signed out page can be redirected to by EntraId after the user has
// successfully signed out.
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

// unsetJwtCookie unsets the cookie set by setJwtCookie using a negative max age
// and an expiry date in the past
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

func clearOidcStateCookie(w http.ResponseWriter) {
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

func redirect(w http.ResponseWriter, r *http.Request, url string) {
	if r.Header.Get(constants.HxRequestHeaderRequest) != "" {
		w.Header().Set(constants.HxResponseHeaderRedirect, url)
		w.WriteHeader(http.StatusSeeOther)
	} else {
		http.Redirect(w, r, url, http.StatusSeeOther)
	}
}
