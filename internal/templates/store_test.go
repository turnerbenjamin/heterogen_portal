package templates

import (
	"bytes"
	"fmt"
	"html/template"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/turnerbenjamin/go_gbf/internal/config"
)

func TestMakeTemplateStore_ReturnsStore(t *testing.T) {
	root := "templates"

	testTemplate := config.TemplateData{
		Name:         "test_data_config",
		WebResources: config.WebResourceDependencies{},
		Dependencies: []string{"dep1", "dep2"},
	}

	testFS, templateData := makeTestFileStoreAndData(
		t,
		root,
		map[config.TemplateIdentifier]config.TemplateData{
			config.TMPL_COMPONENT_ERRORS: testTemplate,
		},
	)

	got, err := MakeTemplateStore(testFS, root, templateData)

	if err != nil {
		t.Fatalf("expected error to be nil but got %s", err)
	}

	if got == nil {
		t.Fatal("expected store pointer but got nil")
	}
}

func TestMakeTemplateStore_HandlesNilFS(t *testing.T) {
	root := "test_templates"
	_, templateData := makeTestFileStoreAndData(
		t,
		root,
		map[config.TemplateIdentifier]config.TemplateData{},
	)
	_, err := MakeTemplateStore(nil, root, templateData)
	assertEqualError(t, err, Err_FileSystemIsNil)
}

func TestMakeTemplateStore_HandlesMissingTemplateData(t *testing.T) {
	root := "test_templates"
	fs, templateData := makeTestFileStoreAndData(
		t,
		root,
		map[config.TemplateIdentifier]config.TemplateData{},
	)

	delete(templateData, config.TMPL_PAGE_USER_SIGNED_OUT)

	_, err := MakeTemplateStore(fs, root, templateData)
	assertEqualError(t,
		err,
		fmt.Sprintf(
			"%s%d",
			Err_MissingTemplateDataPrefix,
			config.TMPL_PAGE_USER_SIGNED_OUT,
		),
	)

}

func TestMakeTemplateStore_HandlesMissingTemplate(t *testing.T) {
	root := "test_templates"
	fs, templateData := makeTestFileStoreAndData(
		t,
		root,
		map[config.TemplateIdentifier]config.TemplateData{},
	)

	missingTemplate := templateData[config.TMPL_COMPONENT_TOAST]
	delete(fs, filepath.Join(root, missingTemplate.Name+".tmpl"))

	_, err := MakeTemplateStore(fs, root, templateData)
	assertEqualError(t, err, Err_MissingTemplateFilePrefix+missingTemplate.Name)
}

func TestMakeTemplateStore_HandlesMissingDependency(t *testing.T) {
	root := "templates"

	testTemplate := config.TemplateData{
		Name:         "test_data_config",
		WebResources: config.WebResourceDependencies{},
		Dependencies: []string{"dep1", "dep2"},
	}

	testFS, templateData := makeTestFileStoreAndData(
		t,
		root,
		map[config.TemplateIdentifier]config.TemplateData{
			config.TMPL_PAGE_APP: testTemplate,
		},
	)

	missingDepName := "dep1"
	delete(testFS, filepath.Join(root, missingDepName+".tmpl"))

	_, err := MakeTemplateStore(testFS, root, templateData)

	assertEqualError(t, err, Err_MissingTemplateFilePrefix+missingDepName)
}

func TestMakeTemplateStore_HandlesInvalidTemplateSyntax(t *testing.T) {
	root := "templates"
	testTemplate := config.TemplateData{
		Name:         "invalid_syntax_template",
		WebResources: config.WebResourceDependencies{},
		Dependencies: []string{},
	}

	testFS, templateData := makeTestFileStoreAndData(
		t,
		root,
		map[config.TemplateIdentifier]config.TemplateData{
			config.TMPL_PAGE_APP: testTemplate,
		},
	)

	path := filepath.Join(root, testTemplate.Name+".tmpl")
	testFS[path] = &fstest.MapFile{Data: []byte(`{{define "missing brace"}}{{end`)}

	_, err := MakeTemplateStore(testFS, root, templateData)

	if err == nil {
		t.Fatalf("expected error when invalid template syntax, got nil")
	}
}

func TestExecute_HandlesMissingData(t *testing.T) {
	var w bytes.Buffer
	testStore := makeTestStore(
		t,
		config.TMPL_PAGE_APP,
		config.WebResourceDependencies{},
		"",
	)

	err := testStore.Execute(
		config.TMPL_PAGE_USER_SIGNED_OUT,
		&w,
		config.TemplateArgs{
			PageConfig: config.PageConfig{},
			Data:       nil,
		},
	)

	assertEqualError(t,
		err,
		fmt.Sprintf(
			"%s%d",
			Err_MissingTemplateDataPrefix,
			config.TMPL_PAGE_USER_SIGNED_OUT,
		),
	)
}

func TestExecute_SetsWebResources(t *testing.T) {
	var w bytes.Buffer
	testStore := makeTestStore(
		t,
		config.TMPL_PAGE_APP,
		config.WebResourceDependencies{
			HG_AUTH: true,
		},
		"{{- if .WebResources.HG_AUTH}}PASS{{- else}}FAIL{{end}}",
	)

	err := testStore.Execute(
		config.TMPL_PAGE_APP,
		&w,
		config.TemplateArgs{
			PageConfig: config.PageConfig{},
			Data:       nil,
		},
	)

	if err != nil {
		t.Fatalf("expected error to be nil but got %s", err)
	}
	assertTemplateContentLooseMatch(t, w.String(), "PASS")
}

func TestExecute_PassesDataCorrectly(t *testing.T) {
	var w bytes.Buffer

	data := config.TemplateArgs{
		PageConfig: config.PageConfig{
			ContentOnly:  true,
			ToastSuccess: "TOAST_SUCCESS",
		},
		Data: struct{ TestData bool }{TestData: true},
	}

	testStore := makeTestStore(
		t,
		config.TMPL_PAGE_APP,
		config.WebResourceDependencies{
			HG_AUTH: true,
		},
		`{{- if .PageConfig.ContentOnly}}PASS{{- else}}FAIL{{end}},
		{{- if eq .PageConfig.ToastSuccess "TOAST_SUCCESS"}}PASS{{- else}}FAIL{{end}},
		{{- if .Data.TestData}}PASS{{- else}}FAIL{{end}}`,
	)

	err := testStore.Execute(
		config.TMPL_PAGE_APP,
		&w,
		data,
	)

	if err != nil {
		t.Fatalf("expected error to be nil but got %s", err)
	}
	assertTemplateContentLooseMatch(t, w.String(), "PASS,PASS,PASS")
}

func makeTestFileStoreAndData(
	t *testing.T,
	root string,
	overrides map[config.TemplateIdentifier]config.TemplateData,
) (fstest.MapFS, map[config.TemplateIdentifier]config.TemplateData) {
	t.Helper()
	mf := fstest.MapFS{}
	td := make(map[config.TemplateIdentifier]config.TemplateData)

	// Create default fs and data for all template identifiers
	for i := range config.TMPL_ENUM_END {
		id := config.TemplateIdentifier(i)
		name := fmt.Sprintf("tmpl_%d", i)
		content := fmt.Sprintf(`{{define "%s"}}TITLE:{{.PageConfig.Title}} ID:%d{{end}}`, name, i)
		path := filepath.Join(root, name+".tmpl")
		mf[path] = &fstest.MapFile{Data: []byte(content)}
		td[id] = config.TemplateData{
			Name:         name,
			Dependencies: []string{},
		}
	}

	// Apply overrides for specific template identifiers
	for id, v := range overrides {
		td[id] = v
		if v.Name != "" {
			path := filepath.Join(root, v.Name+".tmpl")
			if _, ok := mf[path]; !ok {
				content := fmt.Sprintf(`{{define "%s"}}TITLE_OVERRIDE{{end}}`, v.Name)
				mf[path] = &fstest.MapFile{Data: []byte(content)}
			}
		}
		for _, d := range v.Dependencies {
			path := filepath.Join(root, d+".tmpl")
			if _, ok := mf[path]; !ok {
				content := fmt.Sprintf(`{{define "%s"}}DEPENDENCY_OVERRIDE{{end}}`, d)
				mf[path] = &fstest.MapFile{Data: []byte(content)}
			}
		}
	}

	return mf, td
}

func makeTestStore(
	t *testing.T,
	tid config.TemplateIdentifier,
	wr config.WebResourceDependencies,
	templateContent string,
) *Store {
	t.Helper()

	testTemplateName := "test_template"
	content := fmt.Sprintf(`{{define "%s"}}%s{{end}}`, testTemplateName, templateContent)

	fs := fstest.MapFS{}
	fs["test_template.tmpl"] = &fstest.MapFile{Data: []byte(content)}

	tmpl := template.New("")
	// register small helper funcs used by tests (eq)
	tmpl = tmpl.Funcs(template.FuncMap{
		"eq": func(a, b interface{}) bool { return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b) },
	})
	if _, err := tmpl.ParseFS(fs, "test_template.tmpl"); err != nil {
		t.Fatalf("failed to parse test template: %v", err)
	}

	return &Store{
		templateData: map[config.TemplateIdentifier]config.TemplateData{
			tid: {
				Name:         testTemplateName,
				Dependencies: []string{},
				WebResources: wr,
			},
		},
		templates: tmpl,
	}
}

func assertTemplateContentLooseMatch(t *testing.T, got, want string) {
	t.Helper()
	normalise := func(s string) string {
		return strings.Join(strings.Fields(strings.TrimSpace(s)), " ")
	}
	gn := normalise(got)
	wn := normalise(want)

	if !strings.Contains(strings.ToLower(gn), strings.ToLower(wn)) {
		t.Fatalf("expected output to contain:\nwant=%q\ngot =%q", wn, gn)
	}
}

func assertEqualError(t *testing.T, got error, want string) {
	t.Helper()
	if got == nil {
		t.Fatal("expected error but got nil")
	}

	if got.Error() != want {
		t.Fatalf("got %s, but want %s", got.Error(), want)
	}
}
