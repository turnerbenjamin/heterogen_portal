package app

import (
	"io/fs"
	"net/http"
	"os"

	h "github.com/turnerbenjamin/heterogen_portal/internal/handlers"
)

func (app *application) addRoute(pattern string, handler http.Handler) {
	app.serveMux.Handle(pattern, handler)
}

func (app *application) addRoutes() {
	pipeline := h.NewPipelineBuilder(
		app.handlers.errorHandler,
		os.Stdout,
		h.NoStateInit,
	)

	pipelineWithUserState := h.NewPipelineBuilder(
		app.handlers.errorHandler,
		os.Stdout,
		h.UserStateInit,
	)

	if sub, err := fs.Sub(app.staticFileSystem, "static"); err == nil {
		app.addRoute("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(sub))))
	} else {
		app.addRoute("GET /static/", http.FileServer(http.FS(app.staticFileSystem)))
	}

	app.addRoute(
		"GET /",
		pipelineWithUserState.New(
			[]h.Middleware[h.UserState]{
				h.ParseJwtMiddleware[h.UserState](app.services.authService),
				h.RequireSignInMiddleware[h.UserState](app.services.authService),
			},
			app.handlers.authHandler.GetRoot,
		),
	)

	app.addRoute(
		"GET /sign-in-redirect",
		pipeline.New(
			[]h.Middleware[h.NoState]{},
			app.handlers.authHandler.GetSignInRedirect,
		),
	)

	app.addRoute(
		"GET /sign-out",
		pipeline.New(
			[]h.Middleware[h.NoState]{},
			app.handlers.authHandler.GetSignOut,
		),
	)

	app.addRoute(
		"GET /signed-out",
		pipeline.New(
			[]h.Middleware[h.NoState]{},
			app.handlers.authHandler.GetSignedOut,
		),
	)
}
