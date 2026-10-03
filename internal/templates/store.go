// Package templates is responsible for parsing and executing templates
package templates

import (
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"path/filepath"
	"sort"

	"github.com/turnerbenjamin/heterogen_portal/internal/constants"
	"github.com/turnerbenjamin/heterogen_portal/internal/icons"
)

// Store holds the parsed templates and metadata required to execute them.
type Store struct {
	templates *template.Template
}

const (
	TmplPageApp           = "page-app"
	TmplPageUserSignedOut = "page-user-signed-out"
	TmplPageOutOfAppErr   = "page-out-of-app-err"

	TmplPageUsers      = "page-users"
	TmplPageBusinesses = "page-businesses"

	TmplMainUserDash     = "main-user-dash"
	TmplMainTableContent = "main-table-content"

	TmplComponentErrors = "component-errors"
	TmplComponentToast  = "component-toast"

	TmplDepLayoutTop    = "layout-top"
	TmplDepLayoutBottom = "layout-bottom"
)

var requiredTemplates = []string{
	TmplPageApp,
	TmplPageUserSignedOut,
	TmplPageOutOfAppErr,

	TmplPageUsers,
	TmplPageBusinesses,

	TmplMainUserDash,
	TmplMainTableContent,

	TmplComponentErrors,
	TmplComponentToast,

	TmplDepLayoutTop,
	TmplDepLayoutBottom,
}

type PageConfig struct {
	ContentOnly  bool
	Title        string
	ToastSuccess string
}

type WebResourceDependencies struct {
	HTMX      bool
	HG_COMMON bool
}

// executeTemplateData holds data to be passed to all templates
type TemplateConfig[T any] struct {
	PageConfig PageConfig
	Data       T
}

type templateData[T any] struct {
	PageConfig   PageConfig
	Data         T
	WebResources WebResourceDependencies
	Icons        *icons.IconSet
}

// MakeTemplateStore builds a `Store` by reading and parsing all templates
// found under `root` in `fileSystem`. Validates that all template identifiers
// have corresponding template data and that files exist for all templates and
// their dependencies
//
// It returns a fully-initialized `Store` on success or an error when the
// filesystem is nil, required template files are missing, or parsing fails.
func MakeTemplateStore(fileSystem fs.FS, root string) (*Store, error) {
	if fileSystem == nil {
		return nil, errors.New(constants.ErrMsgFileSystemIsNil)
	}

	buildPaths, err := getTemplatePaths(fileSystem, root)
	if err != nil {
		return nil, err
	}

	t := template.New("")
	for _, p := range buildPaths {
		b, err := fs.ReadFile(fileSystem, p)
		if err != nil {
			return nil, err
		}
		t, err = t.Parse(string(b))
		if err != nil {
			return nil, err
		}
	}

	for _, tName := range requiredTemplates {
		requiredT := t.Lookup(tName)
		if requiredT == nil {
			return nil, fmt.Errorf("missing template: %s", tName)
		}
	}

	return &Store{
		templates: t,
	}, nil
}

// getTemplatePaths returns the .tmpl file paths under root
func getTemplatePaths(fileSystem fs.FS, root string) ([]string, error) {
	templatePaths := make([]string, 0, 64)
	err := fs.WalkDir(fileSystem, root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if filepath.Ext(path) == ".tmpl" {
			templatePaths = append(templatePaths, path)
		}
		return nil
	})

	sort.Strings(templatePaths)
	return templatePaths, err
}
