package templates

import (
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"path/filepath"
	"sort"

	"github.com/turnerbenjamin/go_gbf/internal/config"
)

// Store holds the parsed templates and metadata required to execute them.
type Store struct {
	templates    *template.Template
	templateData map[config.TemplateIdentifier]config.TemplateData
}

// executeTemplateData holds data to be passed to all templates
type executeTemplateData struct {
	PageConfig   config.PageConfig
	Data         any
	WebResources config.WebResourceDependencies
}

const (
	// Err_FileSystemIsNil is returned when a nil fs.FS is passed to MakeTemplateStore.
	Err_FileSystemIsNil = "filesystem is nil"

	// Err_MissingTemplateDataPrefix prefixes errors where template-data for an
	// identifier is missing.
	Err_MissingTemplateDataPrefix = "template data not found: "

	// Err_MissingTemplateFilePrefix prefixes errors where a template file or
	// dependency cannot be found in the provided file system.
	Err_MissingTemplateFilePrefix = "template file not found: "
)

// MakeTemplateStore builds a `Store` by reading and parsing all templates
// found under `root` in `fileSystem`. Validates that all template identifiers
// have corresponding template data and that files exist for all templates and
// their dependencies
//
// It returns a fully-initialized `Store` on success or an error when the
// filesystem is nil, required template files are missing, or parsing fails.
func MakeTemplateStore(fileSystem fs.FS, root string, templateData map[config.TemplateIdentifier]config.TemplateData) (*Store, error) {
	if fileSystem == nil {
		return nil, errors.New(Err_FileSystemIsNil)
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

	for i := range config.TMPL_ENUM_END {
		data, ok := templateData[config.TemplateIdentifier(i)]
		if !ok {
			return nil, fmt.Errorf("%s%d", Err_MissingTemplateDataPrefix, i)
		}

		//Validate template identifier found
		currentT := t.Lookup(data.Name)
		if currentT == nil {
			return nil, fmt.Errorf("%s%s", Err_MissingTemplateFilePrefix, data.Name)
		}

		//Validate template dependencies
		for _, dependency := range data.Dependencies {
			tmpl := t.Lookup(dependency)
			if tmpl == nil {
				return nil, fmt.Errorf("%s%s", Err_MissingTemplateFilePrefix, dependency)
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
	id config.TemplateIdentifier,
	w io.Writer,
	data config.TemplateArgs,
) error {
	td, ok := ts.templateData[id]
	if !ok {
		return fmt.Errorf("%s%d", Err_MissingTemplateDataPrefix, id)
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
