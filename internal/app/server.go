package app

import (
	"io/fs"
	"net/http"

	"github.com/turnerbenjamin/go_gbf/internal/config"
	"github.com/turnerbenjamin/go_gbf/internal/db"
	"github.com/turnerbenjamin/go_gbf/internal/templates"
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
