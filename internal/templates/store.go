// Package templates is responsible for parsing and executing templates
package templates

import (
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"path/filepath"
	"sort"

	"github.com/turnerbenjamin/heterogen_portal/internal/constants"
)

// Store holds the parsed templates and metadata required to execute them.
type Store struct {
	templates    *template.Template
	templateData map[TemplateIdentifier]TemplateData
}

type TemplateIdentifier int

const (
	TmplPageApp TemplateIdentifier = iota
	TmplPageUserSignedOut
	TmplComponentErrors
	TmplComponentToast
	_tmplEnumEnd
)

type PageConfig struct {
	ContentOnly  bool
	Title        string
	ToastSuccess string
}

type TemplateArgs struct {
	PageConfig PageConfig
	Data       any
}

type WebResourceDependencies struct {
	HG_AUTH bool
}

type TemplateData struct {
	Name         string
	Dependencies []string
	WebResources WebResourceDependencies
}

// executeTemplateData holds data to be passed to all templates
type executeTemplateData struct {
	PageConfig   PageConfig
	Data         any
	WebResources WebResourceDependencies
}

// MakeTemplateStore builds a `Store` by reading and parsing all templates
// found under `root` in `fileSystem`. Validates that all template identifiers
// have corresponding template data and that files exist for all templates and
// their dependencies
//
// It returns a fully-initialized `Store` on success or an error when the
// filesystem is nil, required template files are missing, or parsing fails.
func MakeTemplateStore(fileSystem fs.FS, root string, templateData map[TemplateIdentifier]TemplateData) (*Store, error) {
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

	// Validate for all template constants that:
	// - Template data exists
	// - Template files exist
	// - Template dependencies exist
	for i := range _tmplEnumEnd {
		data, ok := templateData[TemplateIdentifier(i)]
		if !ok {
			return nil, fmt.Errorf("%s%d", constants.ErrMsgPrefixMissingTemplateData, i)
		}

		currentT := t.Lookup(data.Name)
		if currentT == nil {
			return nil, fmt.Errorf("%s%s", constants.ErrMsgPrefixMissingTemplateFile, data.Name)
		}

		for _, dependency := range data.Dependencies {
			tmpl := t.Lookup(dependency)
			if tmpl == nil {
				return nil, fmt.Errorf("%s%s", constants.ErrMsgPrefixMissingTemplateFile, dependency)
			}
		}
	}
	return &Store{
		templates:    t,
		templateData: templateData,
	}, nil
}

// Execute renders the template associated with `id` to the provided
// writer.
func (ts *Store) Execute(
	id TemplateIdentifier,
	w io.Writer,
	data TemplateArgs,
) error {
	td, ok := ts.templateData[id]
	if !ok {
		return fmt.Errorf("%s%d", constants.ErrMsgPrefixMissingTemplateData, id)
	}

	d := &executeTemplateData{
		PageConfig:   data.PageConfig,
		Data:         data.Data,
		WebResources: td.WebResources,
	}

	return ts.templates.ExecuteTemplate(w, td.Name, d)
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
