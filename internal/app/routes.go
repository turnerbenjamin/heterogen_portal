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

	pipeline := h.NewPipelineBuilder(errorHandler, os.Stdout, h.NoStateInit)

	pipelineWithUserState := h.NewPipelineBuilder(
		errorHandler,
		os.Stdout,
		h.UserStateInit,
	)

	if sub, err := fs.Sub(staticFileSystem, "static"); err == nil {
		mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(sub))))
	} else {
		mux.Handle("GET /static/", http.FileServer(http.FS(staticFileSystem)))
	}

	mux.Handle(
		"GET /",
		pipelineWithUserState.New(
			[]h.Middleware[h.UserState]{
				h.NewParseJwtMiddleware(authService),
				h.NewRequireSignInMiddleware(authService),
			},
			h.GetRootHandler(ts),
		),
	)

	mux.Handle(
		"GET /sign-in-redirect",
		pipeline.New(
			[]h.Middleware[h.NoState]{},
			h.GetSignInRedirectHandler(ts, authService),
		),
	)

	mux.Handle(
		"GET /sign-out",
		pipeline.New(
			[]h.Middleware[h.NoState]{},
			h.GetSignOutHandler(authService),
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
