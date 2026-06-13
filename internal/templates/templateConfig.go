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
	TmplPageUserSignIn: {
		Name: "page-user-sign-in",
		Dependencies: []string{
			"layout-top",
			"layout-bottom",
			"main-user-dash",
			"main-user-login",
		},
		WebResources: WebResourceDependencies{
			HG_AUTH: true,
		},
	},
	TmpPageUserSignInRedirect: {
		Name: "page-user-sign-in-redirect",
		Dependencies: []string{
			"layout-top",
			"layout-bottom",
			"main-user-dash",
			"main-user-login",
		},
		WebResources: WebResourceDependencies{
			HG_AUTH: true,
		},
	},
	TmplPageUserSignOut: {
		Name: "page-user-sign-out",
		Dependencies: []string{
			"layout-top",
			"layout-bottom",
			"main-user-dash",
			"main-user-login",
		},
		WebResources: WebResourceDependencies{
			HG_AUTH: true,
		},
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
