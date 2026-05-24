package templates

import (
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"path/filepath"
)

type TemplateIdentifier int

const (
	TMPL_PAGE_APP TemplateIdentifier = iota
	TMPL_PAGE_USER_SIGN_IN
	TMPL_PAGE_USER_SIGN_IN_REDIRECT
	TMPL_PAGE_USER_SIGNED_OUT
	TMPL_COMPONENT_ERRORS
	TMPL_COMPONENT_TOAST
	templateIdentifierEnumEnd
)

type webResourceDependencies struct {
	HG_AUTH bool
}

type PageConfig struct {
	ContentOnly  bool
	Title        string
	ToastSuccess string
}

type TemplateArgs struct {
	PageConfig PageConfig
	Data       any
}

type executeTemplateData struct {
	PageConfig   PageConfig
	Data         any
	WebResources webResourceDependencies
}

type Store struct {
	templates *template.Template
}

type templateData struct {
	name         string
	dependencies []string
	webResources webResourceDependencies
}

var idToTmplDataMap = map[TemplateIdentifier]templateData{
	TMPL_PAGE_APP: {
		name: "page-app",
		dependencies: []string{
			"layout-top",
			"layout-bottom",
			"main-user-dash",
			"main-user-login",
		},
		webResources: webResourceDependencies{
			HG_AUTH: true,
		},
	},
	TMPL_PAGE_USER_SIGN_IN: {
		name: "page-user-sign-in",
		dependencies: []string{
			"layout-top",
			"layout-bottom",
			"main-user-dash",
			"main-user-login",
		},
		webResources: webResourceDependencies{
			HG_AUTH: true,
		},
	},
	TMPL_PAGE_USER_SIGN_IN_REDIRECT: {
		name: "page-user-sign-in-redirect",
		dependencies: []string{
			"layout-top",
			"layout-bottom",
			"main-user-dash",
			"main-user-login",
		},
		webResources: webResourceDependencies{
			HG_AUTH: true,
		},
	},
	TMPL_PAGE_USER_SIGNED_OUT: {
		name: "page-user-signed-out",
		dependencies: []string{
			"layout-top",
			"layout-bottom",
			"main-user-dash",
			"main-user-login",
		},
		webResources: webResourceDependencies{
			HG_AUTH: true,
		},
	},
	TMPL_COMPONENT_ERRORS: {
		name:         "component-errors",
		dependencies: []string{},
	},
	TMPL_COMPONENT_TOAST: {
		name:         "component-toast",
		dependencies: []string{},
	},
}

func MakeTemplateStore(fileSystem fs.FS, root string) (*Store, error) {
	if fileSystem == nil {
		return nil, errors.New("templates: fileSystem is nil")
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
		t = template.Must(t.Parse(string(b)))
	}

	for i := range templateIdentifierEnumEnd {
		data := idToTmplDataMap[TemplateIdentifier(i)]
		for _, dependency := range data.dependencies {
			tmpl := t.Lookup(dependency)
			if tmpl == nil {
				return nil, fmt.Errorf("template not found: %s", dependency)
			}
		}
	}
	return &Store{
		templates: t,
	}, nil
}

func (ts *Store) Execute(
	id TemplateIdentifier,
	w io.Writer,
	data TemplateArgs,
) error {
	td := idToTmplDataMap[id]
	d := executeTemplateData{
		PageConfig:   data.PageConfig,
		Data:         data.Data,
		WebResources: td.webResources,
	}
	return (*ts).templates.ExecuteTemplate(w, td.name, d)
}

func getTemplatePaths(fileSystem fs.FS, root string) ([]string, error) {
	templatePaths := make([]string, 0, 64)
	err := fs.WalkDir(fileSystem, root, func(path string, d fs.DirEntry, err error) error {
		if filepath.Ext(path) == ".tmpl" {
			templatePaths = append(templatePaths, path)
		}
		return nil
	})
	return templatePaths, err
}
