package app

import (
	"io/fs"
	"net/http"
	"os"

	"github.com/turnerbenjamin/heterogen_portal/internal/config"
	"github.com/turnerbenjamin/heterogen_portal/internal/db"
	h "github.com/turnerbenjamin/heterogen_portal/internal/handlers"
	"github.com/turnerbenjamin/heterogen_portal/internal/templates"
)

func addRoutes(
	mux *http.ServeMux,
	ts *templates.Store,
	staticFileSystem fs.FS,
	settings config.AppSettings,
	adminRepo db.UserRepo,
) {
	errorHandler := h.NewErrorHandler(ts)

	pipeline := h.NewPipelineBuilder[h.NoState](errorHandler, os.Stdout)

	pipelineWithUserState := h.NewPipelineBuilder[h.UserState](
		errorHandler,
		os.Stdout,
	)

	parseAdminJWT := h.NewParseJWTMiddleware(settings, adminRepo)

	if sub, err := fs.Sub(staticFileSystem, "static"); err == nil {
		mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(sub))))
	} else {
		mux.Handle("GET /static/", http.FileServer(http.FS(staticFileSystem)))
	}

	mux.Handle(
		"GET /",
		pipelineWithUserState.New(
			[]h.Middleware[h.UserState]{parseAdminJWT},
			h.GetRootHandler(ts),
		),
	)

	mux.Handle(
		"GET /sign-in",
		pipeline.New(
			[]h.Middleware[h.NoState]{},
			h.GetSignInHandler(ts),
		),
	)

	mux.Handle(
		"GET /sign-in-redirect",
		pipeline.New(
			[]h.Middleware[h.NoState]{},
			h.GetSignInRedirectHandler(ts),
		),
	)

	mux.Handle(
		"GET /sign-out",
		pipeline.New(
			[]h.Middleware[h.NoState]{},
			h.GetSignOutHandler(ts),
		),
	)

	mux.Handle(
		"GET /signed-out",
		pipeline.New(
			[]h.Middleware[h.NoState]{},
			h.GetSignedOutHandler(ts),
		),
	)

	mux.Handle(
		"POST /sign-in",
		pipeline.New(
			[]h.Middleware[h.NoState]{},
			h.POST_UserSignIn(settings, ts, adminRepo),
		),
	)
}
