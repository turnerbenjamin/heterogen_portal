package templates

import (
	"bytes"
	"fmt"
	"html/template"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

func TestMakeTemplateStore_ReturnsStore(t *testing.T) {
	root := "templates"

	testTemplate := TemplateData{
		Name:         "test_data_config",
		WebResources: WebResourceDependencies{},
		Dependencies: []string{"dep1", "dep2"},
	}

	testFS, templateData := makeTestFileStoreAndData(
		t,
		root,
		map[TemplateIdentifier]TemplateData{
			TMPL_COMPONENT_ERRORS: testTemplate,
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
		map[TemplateIdentifier]TemplateData{},
	)
	_, err := MakeTemplateStore(nil, root, templateData)
	assertEqualError(t, err, Err_FileSystemIsNil)
}

func TestMakeTemplateStore_HandlesMissingTemplateData(t *testing.T) {
	root := "test_templates"
	fs, templateData := makeTestFileStoreAndData(
		t,
		root,
		map[TemplateIdentifier]TemplateData{},
	)

	delete(templateData, TMPL_PAGE_USER_SIGNED_OUT)

	_, err := MakeTemplateStore(fs, root, templateData)
	assertEqualError(t,
		err,
		fmt.Sprintf(
			"%s%d",
			Err_MissingTemplateDataPrefix,
			TMPL_PAGE_USER_SIGNED_OUT,
		),
	)

}

func TestMakeTemplateStore_HandlesMissingTemplate(t *testing.T) {
	root := "test_templates"
	fs, templateData := makeTestFileStoreAndData(
		t,
		root,
		map[TemplateIdentifier]TemplateData{},
	)

	missingTemplate := templateData[TMPL_COMPONENT_TOAST]
	delete(fs, filepath.Join(root, missingTemplate.Name+".tmpl"))

	_, err := MakeTemplateStore(fs, root, templateData)
	assertEqualError(t, err, Err_MissingTemplateFilePrefix+missingTemplate.Name)
}

func TestMakeTemplateStore_HandlesMissingDependency(t *testing.T) {
	root := "templates"

	testTemplate := TemplateData{
		Name:         "test_data_config",
		WebResources: WebResourceDependencies{},
		Dependencies: []string{"dep1", "dep2"},
	}

	testFS, templateData := makeTestFileStoreAndData(
		t,
		root,
		map[TemplateIdentifier]TemplateData{
			TMPL_PAGE_APP: testTemplate,
		},
	)

	missingDepName := "dep1"
	delete(testFS, filepath.Join(root, missingDepName+".tmpl"))

	_, err := MakeTemplateStore(testFS, root, templateData)

	assertEqualError(t, err, Err_MissingTemplateFilePrefix+missingDepName)
}

func TestMakeTemplateStore_HandlesInvalidTemplateSyntax(t *testing.T) {
	root := "templates"
	testTemplate := TemplateData{
		Name:         "invalid_syntax_template",
		WebResources: WebResourceDependencies{},
		Dependencies: []string{},
	}

	testFS, templateData := makeTestFileStoreAndData(
		t,
		root,
		map[TemplateIdentifier]TemplateData{
			TMPL_PAGE_APP: testTemplate,
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
		TMPL_PAGE_APP,
		WebResourceDependencies{},
		"",
	)

	err := testStore.Execute(
		TMPL_PAGE_USER_SIGNED_OUT,
		&w,
		TemplateArgs{
			PageConfig: PageConfig{},
			Data:       nil,
		},
	)

	assertEqualError(t,
		err,
		fmt.Sprintf(
			"%s%d",
			Err_MissingTemplateDataPrefix,
			TMPL_PAGE_USER_SIGNED_OUT,
		),
	)
}

func TestExecute_SetsWebResources(t *testing.T) {
	var w bytes.Buffer
	testStore := makeTestStore(
		t,
		TMPL_PAGE_APP,
		WebResourceDependencies{
			HG_AUTH: true,
		},
		"{{- if .WebResources.HG_AUTH}}PASS{{- else}}FAIL{{end}}",
	)

	err := testStore.Execute(
		TMPL_PAGE_APP,
		&w,
		TemplateArgs{
			PageConfig: PageConfig{},
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

	data := TemplateArgs{
		PageConfig: PageConfig{
			ContentOnly:  true,
			ToastSuccess: "TOAST_SUCCESS",
		},
		Data: struct{ TestData bool }{TestData: true},
	}

	testStore := makeTestStore(
		t,
		TMPL_PAGE_APP,
		WebResourceDependencies{
			HG_AUTH: true,
		},
		`{{- if .PageConfig.ContentOnly}}PASS{{- else}}FAIL{{end}},
		{{- if eq .PageConfig.ToastSuccess "TOAST_SUCCESS"}}PASS{{- else}}FAIL{{end}},
		{{- if .Data.TestData}}PASS{{- else}}FAIL{{end}}`,
	)

	err := testStore.Execute(
		TMPL_PAGE_APP,
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
	overrides map[TemplateIdentifier]TemplateData,
) (fstest.MapFS, map[TemplateIdentifier]TemplateData) {
	t.Helper()
	mf := fstest.MapFS{}
	td := make(map[TemplateIdentifier]TemplateData)

	// Create default fs and data for all template identifiers
	for i := range TMPL_ENUM_END {
		id := TemplateIdentifier(i)
		name := fmt.Sprintf("tmpl_%d", i)
		content := fmt.Sprintf(`{{define "%s"}}TITLE:{{.PageConfig.Title}} ID:%d{{end}}`, name, i)
		path := filepath.Join(root, name+".tmpl")
		mf[path] = &fstest.MapFile{Data: []byte(content)}
		td[id] = TemplateData{
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
	tid TemplateIdentifier,
	wr WebResourceDependencies,
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
		templateData: map[TemplateIdentifier]TemplateData{
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
