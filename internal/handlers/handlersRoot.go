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
	"github.com/turnerbenjamin/heterogen_portal/internal/etc"
	"github.com/turnerbenjamin/heterogen_portal/internal/services"
	"github.com/turnerbenjamin/heterogen_portal/internal/templates"
)

var errHtmxNotSupported *AppError = NewServerError(
	errors.New(constants.ErrMsgHtmxNotSupported),
)

type AuthService interface {
	ParseUserJwtCookie(tokenString string) (*services.AppClaims, error)
	RetrieveUserById(userId string) (*db.User, error)
	BuildSignInRedirectRequest() (*services.SignInRedirectRequest, error)
	BuildSignOutRedirectRequest(email string) string
	AuthenticateUser(
		ctx context.Context,
		authorisationCode string,
		returnedState string,
		signedOidcState string,
	) (appToken string, err error)
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	IDToken     string `json:"id_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// GetRootHandler returns a handler for the application root.
//
// It redirects unauthenticated users to the sign-in page and renders the
// main app template for authenticated users.
func GetRootHandler(
	ts TemplateStore,
	appSettings *etc.AppSettings,
	authService AuthService,
) AppHandler[UserState] {
	return func(w http.ResponseWriter, r *http.Request, c *PipelineContext[UserState]) *AppError {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return nil
		}

		// If user is nil, redirect to sign-in service
		if c.state.User == nil {
			redirectReq, err := authService.BuildSignInRedirectRequest()
			if err != nil {
				return NewServerError(err)
			}
			setOidcCookie(w, redirectReq.SignedOidcState)
			redirect(w, r, redirectReq.Url)
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
	appSettings *etc.AppSettings,
	authService AuthService,
) AppHandler[NoState] {
	return func(w http.ResponseWriter, r *http.Request, c *PipelineContext[NoState]) *AppError {
		// extract query params
		code := r.URL.Query().Get("code")
		if code == "" {
			return NewServerError(errors.New(constants.ErrMissingOIDCCodeParam))
		}

		returnedState := r.URL.Query().Get("state")
		if code == "" {
			return NewServerError(errors.New(constants.ErrMissingOIDCStateParam))
		}

		// access signed oidc state from the cookie and clear it up
		cookie, err := r.Cookie(constants.IdentifierOidcStateCookie)
		if err != nil {
			return NewServerError(errors.New("missing oidc state cookie"))
		}
		signedOidcState := cookie.Value
		clearOidcCookie(w)

		// authenticate the user
		appToken, err := authService.AuthenticateUser(
			r.Context(),
			code,
			returnedState,
			signedOidcState,
		)

		if err != nil {
			return NewServerError(err)
		}

		setJwtCookie(w, appToken)

		// redirect user to the app
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return nil
	}
}

// GetSignedOutHandler returns the sign-out page handler.
//
// It clears the JWT cookie and renders the sign-out template. The sign-out page
// will then redirect to allow the user to sign-out from EntraId.
func GetSignOutHandler(ts TemplateStore, appSettings *etc.AppSettings, authService AuthService) AppHandler[NoState] {
	return func(w http.ResponseWriter, r *http.Request, c *PipelineContext[NoState]) *AppError {
		if r.Header.Get(constants.HxRequestHeaderRequest) != "" {
			return errHtmxNotSupported
		}

		jwtCookie, err := r.Cookie(constants.IdentifierJwtCookie)
		if err != nil {
			http.Redirect(w, r, "signed-out", http.StatusSeeOther)
		}

		claims, err := authService.ParseUserJwtCookie(jwtCookie.Value)
		if err != nil {
			unsetJwtCookie(w)
			http.Redirect(w, r, "signed-out", http.StatusSeeOther)
		}

		redirectUrl := authService.BuildSignOutRedirectRequest(claims.IdToken)
		unsetJwtCookie(w)

		http.Redirect(w, r, redirectUrl, http.StatusSeeOther)
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
			return errHtmxNotSupported
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
		MaxAge:      int((time.Hour * 24) / time.Second),
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

func setOidcCookie(w http.ResponseWriter, tokenString string) {
	http.SetCookie(w, &http.Cookie{
		Name:     constants.IdentifierOidcStateCookie,
		Value:    tokenString,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/sign-in-redirect",
		MaxAge:   600,
	})
}

func clearOidcCookie(w http.ResponseWriter) {
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
