package handlers

import (
	"fmt"
	"net/http"

	"github.com/turnerbenjamin/heterogen_portal/internal/etc"
	"github.com/turnerbenjamin/heterogen_portal/internal/templates"
)

func GET_ROOT(ts *templates.Store) AppHandler[UserRaft] {
	return func(w http.ResponseWriter, r *http.Request, c *PipelineContext[UserRaft]) *etc.AppError {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return nil
		}

		if c.state.User == nil {
			if r.Header.Get("HX-Request") != "" {
				w.Header().Set("HX-Redirect", "/")
			} else {
				http.Redirect(w, r, "/sign-in", http.StatusSeeOther)
				return nil
			}
		}

		pageConfig := templates.PageConfig{
			ContentOnly: r.Header.Get("HX-Request") != "",
			Title:       "HETEROGEN",
		}

		err := ts.Execute(
			templates.TMPL_PAGE_APP,
			w,
			templates.TemplateArgs{PageConfig: pageConfig, Data: c.state},
		)
		if err != nil {
			fmt.Println(err.Error())
			return ErrServer
		}
		return nil
	}
}

func GET_SIGN_IN(ts *templates.Store) AppHandler[NoPipelineState] {
	return func(w http.ResponseWriter, r *http.Request, c *PipelineContext[NoPipelineState]) *etc.AppError {
		pageConfig := templates.PageConfig{
			ContentOnly: false,
			Title:       "HETEROGEN | SIGN-IN",
		}

		err := ts.Execute(
			templates.TMPL_PAGE_USER_SIGN_IN,
			w,
			templates.TemplateArgs{PageConfig: pageConfig, Data: nil},
		)
		if err != nil {
			fmt.Println(err.Error())
			return ErrServer
		}
		return nil
	}
}

func GET_SIGN_IN_REDIRECT(ts *templates.Store) AppHandler[NoPipelineState] {
	return func(w http.ResponseWriter, r *http.Request, c *PipelineContext[NoPipelineState]) *etc.AppError {
		pageConfig := templates.PageConfig{
			ContentOnly: false,
			Title:       "HETEROGEN | SIGN-IN",
		}

		err := ts.Execute(
			templates.TMPL_PAGE_USER_SIGN_IN_REDIRECT,
			w,
			templates.TemplateArgs{PageConfig: pageConfig, Data: nil},
		)
		if err != nil {
			fmt.Println(err.Error())
			return ErrServer
		}
		return nil
	}
}

func GET_SIGNED_OUT(ts *templates.Store) AppHandler[NoPipelineState] {
	return func(w http.ResponseWriter, r *http.Request, c *PipelineContext[NoPipelineState]) *etc.AppError {
		unsetJWTCookie(w)

		pageConfig := templates.PageConfig{
			ContentOnly: false,
			Title:       "HETEROGEN | SIGNED OUT",
		}

		err := ts.Execute(
			templates.TMPL_PAGE_USER_SIGNED_OUT,
			w,
			templates.TemplateArgs{PageConfig: pageConfig, Data: nil},
		)
		if err != nil {
			fmt.Println(err.Error())
			return ErrServer
		}
		return nil
	}
}
