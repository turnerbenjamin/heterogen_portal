// Package handlers contains HTTP handlers for the application.
//
// This file exposes handlers that render pages via the TemplateStore.
package handlers

import (
	"errors"
	"net/http"

	"github.com/turnerbenjamin/heterogen_portal/internal/constants"
	"github.com/turnerbenjamin/heterogen_portal/internal/etc"
	"github.com/turnerbenjamin/heterogen_portal/internal/templates"
)

var errHtmxNotSupported *etc.AppError = etc.NewServerError(
	errors.New(constants.ErrMsgHtmxNotSupported),
)

// GetRootHandler returns a handler for the application root.
//
// It redirects unauthenticated users to the sign-in page and renders the
// main app template for authenticated users.
func GetRootHandler(ts TemplateStore) AppHandler[UserState] {
	return func(w http.ResponseWriter, r *http.Request, c *PipelineContext[UserState]) *etc.AppError {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return nil
		}

		if c.state.User == nil {
			if r.Header.Get("HX-Request") != "" {
				w.Header().Set("HX-Redirect", "/sign-in")
				w.WriteHeader(http.StatusSeeOther)
			} else {
				http.Redirect(w, r, "/sign-in", http.StatusSeeOther)
			}
			return nil
		}

		pageConfig := templates.PageConfig{
			ContentOnly: r.Header.Get("HX-Request") != "",
			Title:       "HETEROGEN",
		}

		err := ts.Execute(
			templates.TmplPageApp,
			w,
			templates.TemplateArgs{PageConfig: pageConfig, Data: c.state},
		)
		if err != nil {
			return etc.NewServerError(err)
		}
		return nil
	}
}

// GetSignInHandler returns the sign-in page handler.
func GetSignInHandler(ts TemplateStore) AppHandler[NoState] {
	return func(w http.ResponseWriter, r *http.Request, c *PipelineContext[NoState]) *etc.AppError {
		pageConfig := templates.PageConfig{
			ContentOnly: false,
			Title:       "HETEROGEN | SIGN-IN",
		}

		err := ts.Execute(
			templates.TmplPageUserSignIn,
			w,
			templates.TemplateArgs{PageConfig: pageConfig, Data: nil},
		)
		if err != nil {
			return etc.NewServerError(err)
		}
		return nil
	}
}

// GetSignInRedirectHandler returns the sign-in redirect page handler.
func GetSignInRedirectHandler(ts TemplateStore) AppHandler[NoState] {
	return func(w http.ResponseWriter, r *http.Request, c *PipelineContext[NoState]) *etc.AppError {
		if r.Header.Get("HX-Request") != "" {
			return errHtmxNotSupported
		}

		pageConfig := templates.PageConfig{
			ContentOnly: false,
			Title:       "HETEROGEN | SIGN-IN",
		}

		err := ts.Execute(
			templates.TmpPageUserSignInRedirect,
			w,
			templates.TemplateArgs{PageConfig: pageConfig, Data: nil},
		)
		if err != nil {
			return etc.NewServerError(err)
		}
		return nil
	}
}

// GetSignedOutHandler returns the sign-out page handler.
//
// It clears the JWT cookie and renders the sign-out template. The sign-out page
// will then redirect to allow the user to sign-out from EntraId.
func GetSignOutHandler(ts TemplateStore) AppHandler[NoState] {
	return func(w http.ResponseWriter, r *http.Request, c *PipelineContext[NoState]) *etc.AppError {
		if r.Header.Get("HX-Request") != "" {
			return errHtmxNotSupported
		}

		unsetJwtCookie(w)
		pageConfig := templates.PageConfig{
			ContentOnly: false,
			Title:       "HETEROGEN | SIGN-OUT",
		}

		err := ts.Execute(
			templates.TmplPageUserSignOut,
			w,
			templates.TemplateArgs{PageConfig: pageConfig, Data: nil},
		)
		if err != nil {
			return etc.NewServerError(err)
		}
		return nil
	}
}

// GetSignedOutHandler returns the signed-out page handler.
//
// The signed out page can be redirected to by EntraId after the user has
// successfully signed out.
func GetSignedOutHandler(ts TemplateStore) AppHandler[NoState] {
	return func(w http.ResponseWriter, r *http.Request, c *PipelineContext[NoState]) *etc.AppError {
		if r.Header.Get("HX-Request") != "" {
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
			return etc.NewServerError(err)
		}
		return nil
	}
}
