package app

import (
	"io/fs"
	"net/http"

	"github.com/turnerbenjamin/heterogen_portal/internal/handlers"
	"github.com/turnerbenjamin/heterogen_portal/internal/templates"
)

func NewServer(
	ts *templates.Store,
	staticFileSystem fs.FS,
	tokenValidator handlers.TokenValidator,
	tokenSignerAndParser handlers.TokenSignerAndParser,
	userRepo handlers.UserRepo,
) (http.Handler, error) {
	mux := http.NewServeMux()
	addRoutes(mux, ts, staticFileSystem, tokenValidator, tokenSignerAndParser, userRepo)
	return mux, nil
}
