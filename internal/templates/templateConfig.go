package templates

var TemplateDataMap = map[TemplateIdentifier]TemplateData{
	TmplPageApp: {
		Name: "page-app",
		Dependencies: []string{
			"layout-top",
			"layout-bottom",
			"main-user-dash",
		},
		WebResources: WebResourceDependencies{
			HTMX:      true,
			HG_COMMON: true,
		},
	},
	TmplPageUserSignedOut: {
		Name: "page-user-signed-out",
		Dependencies: []string{
			"layout-top",
			"layout-bottom",
		},
		WebResources: WebResourceDependencies{},
	},
	TmplPageOutOfAppErr: {
		Name: "page-out-of-app-err",
		Dependencies: []string{
			"layout-top",
			"layout-bottom",
		},
		WebResources: WebResourceDependencies{},
	},
	TmplComponentErrors: {
		Name:         "component-errors",
		Dependencies: []string{},
	},
	TmplComponentToast: {
		Name:         "component-toast",
		Dependencies: []string{},
	},
}
