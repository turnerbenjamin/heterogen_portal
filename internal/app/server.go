package app

import (
	"io/fs"
	"net/http"

	"github.com/turnerbenjamin/heterogen_portal/internal/etc"
	"github.com/turnerbenjamin/heterogen_portal/internal/handlers"
	"github.com/turnerbenjamin/heterogen_portal/internal/templates"
)

func NewServer(
	appSettings *etc.AppSettings,
	ts *templates.Store,
	staticFileSystem fs.FS,
	authService handlers.AuthService,
) (http.Handler, error) {
	mux := http.NewServeMux()
	addRoutes(mux, appSettings, ts, staticFileSystem, authService)
	return mux, nil
}
