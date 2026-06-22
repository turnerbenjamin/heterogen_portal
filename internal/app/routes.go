package app

import (
	"io/fs"
	"net/http"
	"os"

	"github.com/turnerbenjamin/heterogen_portal/internal/etc"
	h "github.com/turnerbenjamin/heterogen_portal/internal/handlers"
	"github.com/turnerbenjamin/heterogen_portal/internal/templates"
)

func addRoutes(
	mux *http.ServeMux,
	appSettings *etc.AppSettings,
	ts *templates.Store,
	staticFileSystem fs.FS,
	authService h.AuthService,
) {
	errorHandler := h.NewErrorHandler(ts)

	pipeline := h.NewPipelineBuilder[h.NoState](errorHandler, os.Stdout)

	pipelineWithUserState := h.NewPipelineBuilder[h.UserState](
		errorHandler,
		os.Stdout,
	)

	parseAdminJWT := h.NewParseJwtMiddleware(appSettings, authService)

	if sub, err := fs.Sub(staticFileSystem, "static"); err == nil {
		mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(sub))))
	} else {
		mux.Handle("GET /static/", http.FileServer(http.FS(staticFileSystem)))
	}

	mux.Handle(
		"GET /",
		pipelineWithUserState.New(
			[]h.Middleware[h.UserState]{parseAdminJWT},
			h.GetRootHandler(ts, appSettings, authService),
		),
	)

	mux.Handle(
		"GET /sign-in",
		pipeline.New(
			[]h.Middleware[h.NoState]{},
			h.GetSignInHandler(ts, appSettings),
		),
	)

	mux.Handle(
		"GET /sign-in-redirect",
		pipeline.New(
			[]h.Middleware[h.NoState]{},
			h.GetSignInRedirectHandler(ts, appSettings, authService),
		),
	)

	mux.Handle(
		"GET /sign-out",
		pipeline.New(
			[]h.Middleware[h.NoState]{},
			h.GetSignOutHandler(ts, appSettings),
		),
	)

	mux.Handle(
		"GET /signed-out",
		pipeline.New(
			[]h.Middleware[h.NoState]{},
			h.GetSignedOutHandler(ts),
		),
	)
}
