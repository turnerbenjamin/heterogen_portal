package handlers

import (
	"fmt"
	"net/http"

	"github.com/turnerbenjamin/go_gbf/internal/config"
	"github.com/turnerbenjamin/go_gbf/internal/etc"
	"github.com/turnerbenjamin/go_gbf/internal/logging"
	"github.com/turnerbenjamin/go_gbf/internal/templates"
)

func GET_ROOT(ts *templates.Store) AppHandlerWithRaft[UserRaft] {
	return func(w http.ResponseWriter, r *http.Request, l logging.Logger, raft UserRaft) etc.AppError {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return nil
		}

		if raft.User == nil {
			http.Redirect(w, r, "/sign-in", http.StatusFound)
			return nil
		}

		pageConfig := config.PageConfig{
			ContentOnly: r.Header.Get("HX-Request") != "",
			Title:       "HETEROGEN",
		}

		err := ts.Execute(
			config.TMPL_PAGE_APP,
			w,
			config.TemplateArgs{PageConfig: pageConfig, Data: raft},
		)
		if err != nil {
			fmt.Println(err.Error())
			return ErrServer
		}
		return nil
	}
}

func GET_SIGN_IN(ts *templates.Store) AppHandler {
	return func(w http.ResponseWriter, r *http.Request, l logging.Logger) etc.AppError {
		pageConfig := config.PageConfig{
			ContentOnly: false,
			Title:       "HETEROGEN | SIGN-IN",
		}

		err := ts.Execute(
			config.TMPL_PAGE_USER_SIGN_IN,
			w,
			config.TemplateArgs{PageConfig: pageConfig, Data: nil},
		)
		if err != nil {
			fmt.Println(err.Error())
			return ErrServer
		}
		return nil
	}
}

func GET_SIGN_IN_REDIRECT(ts *templates.Store) AppHandler {
	return func(w http.ResponseWriter, r *http.Request, l logging.Logger) etc.AppError {
		pageConfig := config.PageConfig{
			ContentOnly: false,
			Title:       "HETEROGEN | SIGN-IN",
		}

		err := ts.Execute(
			config.TMPL_PAGE_USER_SIGN_IN_REDIRECT,
			w,
			config.TemplateArgs{PageConfig: pageConfig, Data: nil},
		)
		if err != nil {
			fmt.Println(err.Error())
			return ErrServer
		}
		return nil
	}
}

func GET_SIGNED_OUT(ts *templates.Store) AppHandler {
	return func(w http.ResponseWriter, r *http.Request, l logging.Logger) etc.AppError {
		unsetJWTCookie(w)

		pageConfig := config.PageConfig{
			ContentOnly: false,
			Title:       "HETEROGEN | SIGNED OUT",
		}

		err := ts.Execute(
			config.TMPL_PAGE_USER_SIGNED_OUT,
			w,
			config.TemplateArgs{PageConfig: pageConfig, Data: nil},
		)
		if err != nil {
			fmt.Println(err.Error())
			return ErrServer
		}
		return nil
	}
}
