package templates

var TemplateDataMap = map[TemplateIdentifier]TemplateData{
	TmplPageApp: {
		Name: "page-app",
		Dependencies: []string{
			"layout-top",
			"layout-bottom",
			"main-user-dash",
			"main-user-login",
		},
		WebResources: WebResourceDependencies{},
	},
	TmplPageUserSignedOut: {
		Name: "page-user-signed-out",
		Dependencies: []string{
			"layout-top",
			"layout-bottom",
			"main-user-dash",
			"main-user-login",
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
