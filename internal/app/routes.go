package app

import (
	"fmt"
	"io/fs"
	"net/http"
	"os"

	"github.com/turnerbenjamin/heterogen_portal/internal/constants"
	h "github.com/turnerbenjamin/heterogen_portal/internal/handlers"
	"github.com/turnerbenjamin/heterogen_portal/internal/model"
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

	pipelineWithUserTableState := h.NewPipelineBuilder(
		app.handlers.errorHandler,
		os.Stdout,
		h.UsersTableStateInit,
	)

	pipelineWithBusinessesTableState := h.NewPipelineBuilder(
		app.handlers.errorHandler,
		os.Stdout,
		h.BusinessesTableStateInit,
	)

	if sub, err := fs.Sub(app.staticFileSystem, "static"); err == nil {
		app.addRoute("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(sub))))
	} else {
		app.addRoute("GET /static/", http.FileServer(http.FS(app.staticFileSystem)))
	}

	app.addRoute(
		fmt.Sprintf("GET /api/{%s}", constants.UrlParamResource),
		pipeline.New(
			h.ContentTypeJson,
			[]h.Middleware[h.NoState]{},
			app.handlers.queryApiHandler.ProcessQuery,
		),
	)

	app.addRoute(
		"GET /",
		pipelineWithUserState.New(
			h.ContentTypeHtml,
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
			h.ContentTypeHtml,
			[]h.Middleware[h.NoState]{},
			app.handlers.authHandler.GetSignInRedirect,
		),
	)

	app.addRoute(
		"GET /sign-out",
		pipeline.New(
			h.ContentTypeHtml,
			[]h.Middleware[h.NoState]{},
			app.handlers.authHandler.GetSignOut,
		),
	)

	app.addRoute(
		"GET /signed-out",
		pipeline.New(
			h.ContentTypeHtml,
			[]h.Middleware[h.NoState]{},
			app.handlers.authHandler.GetSignedOut,
		),
	)

	app.addRoute(
		"GET /users",
		pipelineWithUserTableState.New(
			h.ContentTypeHtml,
			[]h.Middleware[*h.UsersTableState]{
				h.ParseJwtMiddleware[*h.UsersTableState](app.services.authService),
				h.RequireSignInMiddleware[*h.UsersTableState](app.services.authService),
				h.MainTableRequestEnrichment[*model.UsersModel, *h.UsersTableState](),
			},
			app.handlers.usersHandler.GetRoot,
		),
	)

	app.addRoute(
		"GET /businesses",
		pipelineWithBusinessesTableState.New(
			h.ContentTypeHtml,
			[]h.Middleware[*h.BusinessesTableState]{
				h.ParseJwtMiddleware[*h.BusinessesTableState](app.services.authService),
				h.RequireSignInMiddleware[*h.BusinessesTableState](app.services.authService),
				h.MainTableRequestEnrichment[*model.BusinessesModel, *h.BusinessesTableState](),
			},
			app.handlers.businessesHandler.GetRoot,
		),
	)
}
