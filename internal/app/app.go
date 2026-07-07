package app

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"

	"github.com/turnerbenjamin/heterogen_portal/internal/db"
	"github.com/turnerbenjamin/heterogen_portal/internal/etc"
	"github.com/turnerbenjamin/heterogen_portal/internal/handlers"
	"github.com/turnerbenjamin/heterogen_portal/internal/services"
	"github.com/turnerbenjamin/heterogen_portal/internal/templates"
)

type appRepos struct {
	userRepo *db.UserRepo
}

type appServices struct {
	authService *services.AuthService
}

type appHandlers struct {
	authHandler  *handlers.AuthHandler
	errorHandler *handlers.ErrorHandler
}

type application struct {
	appContext       context.Context
	dbConnection     *sql.DB
	staticFileSystem fs.FS
	templateStore    *templates.Store
	repos            *appRepos
	services         *appServices
	handlers         *appHandlers
	serveMux         *http.ServeMux
	server           *http.Server
}

func (a *application) Start() error {
	return a.server.ListenAndServe()
}

func (a *application) Shutdown(shutdownContext context.Context) error {
	return a.server.Shutdown(shutdownContext)
}

func (a *application) Address() string {
	return a.server.Addr
}

func Init(
	ctx context.Context,
	appSettings *etc.AppSettings,
	staticFileSystem fs.FS,
) (*application, error) {
	var err error
	app := &application{
		appContext:       ctx,
		staticFileSystem: staticFileSystem,
	}

	app.dbConnection, err = db.SetUpDB(appSettings.SqlServerDsn)
	if err != nil {
		app.CleanUp()
		return nil, err
	}

	app.templateStore, err = templates.MakeTemplateStore(
		staticFileSystem,
		"templates",
		templates.TemplateDataMap,
	)
	if err != nil {
		app.CleanUp()
		return nil, err
	}

	dependencies := initAppDependencies()
	repos := initRepos(ctx, app.dbConnection)

	app.services, err = initServices(
		ctx,
		appSettings,
		dependencies,
		repos,
	)
	if err != nil {
		app.CleanUp()
		return nil, err
	}

	app.handlers = initHandlers(
		app.templateStore,
		app.services,
	)

	app.serveMux = http.NewServeMux()
	app.addRoutes()

	app.server = &http.Server{
		Addr:    net.JoinHostPort("", "8080"),
		Handler: app.serveMux,
	}
	return app, nil
}

func initRepos(ctx context.Context, dbConnection *sql.DB) *appRepos {
	return &appRepos{
		userRepo: db.BuildUserRepo(ctx, dbConnection),
	}
}

func initServices(
	ctx context.Context,
	appSettings *etc.AppSettings,
	dependencies *appDependencies,
	repos *appRepos,
) (*appServices, error) {
	authService, err := services.NewAuthService(
		ctx,
		appSettings,
		dependencies.httpClient,
		dependencies.jsonSerialiser,
		dependencies.randReader,
		dependencies.tokenSigner,
		dependencies.payloadSigner,
		dependencies.newOidcProvider,
		repos.userRepo,
	)
	if err != nil {
		return nil, err
	}

	return &appServices{
		authService: authService,
	}, nil
}

func initHandlers(
	templateStore *templates.Store,
	services *appServices,
) *appHandlers {
	return &appHandlers{
		authHandler: handlers.NewAuthHandler(
			templateStore,
			services.authService,
		),
		errorHandler: handlers.NewErrorHandler(templateStore),
	}
}

func (app *application) CleanUp() {
	fmt.Println("closing application")

	if app.repos != nil && app.repos.userRepo != nil {
		fmt.Println("closing user repo")
		app.repos.userRepo.Close()
	}

	if app.dbConnection != nil {
		fmt.Println("closing DB connection")
		err := app.dbConnection.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error closing DB connection: %s\n", err)
		}
	}
}
