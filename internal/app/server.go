package app

import (
	"io/fs"
	"net/http"

	"github.com/turnerbenjamin/heterogen_portal/internal/config"
	"github.com/turnerbenjamin/heterogen_portal/internal/db"
	"github.com/turnerbenjamin/heterogen_portal/internal/templates"
)

func NewServer(
	settings config.AppSettings,
	ts *templates.Store,
	staticFileSystem fs.FS,
	adminRepo db.UserRepo,
) (http.Handler, error) {
	mux := http.NewServeMux()
	addRoutes(mux, ts, staticFileSystem, settings, adminRepo)
	return mux, nil
}
