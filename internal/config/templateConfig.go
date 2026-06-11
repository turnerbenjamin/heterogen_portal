package config

import (
	"github.com/turnerbenjamin/heterogen_portal/internal/templates"
)

var TemplateDataMap = map[templates.TemplateIdentifier]templates.TemplateData{
	templates.TmplPageApp: {
		Name: "page-app",
		Dependencies: []string{
			"layout-top",
			"layout-bottom",
			"main-user-dash",
			"main-user-login",
		},
		WebResources: templates.WebResourceDependencies{},
	},
	templates.TmplPageUserSignIn: {
		Name: "page-user-sign-in",
		Dependencies: []string{
			"layout-top",
			"layout-bottom",
			"main-user-dash",
			"main-user-login",
		},
		WebResources: templates.WebResourceDependencies{
			HG_AUTH: true,
		},
	},
	templates.TmpPageUserSignInRedirect: {
		Name: "page-user-sign-in-redirect",
		Dependencies: []string{
			"layout-top",
			"layout-bottom",
			"main-user-dash",
			"main-user-login",
		},
		WebResources: templates.WebResourceDependencies{
			HG_AUTH: true,
		},
	},
	templates.TmplPageUserSignOut: {
		Name: "page-user-sign-out",
		Dependencies: []string{
			"layout-top",
			"layout-bottom",
			"main-user-dash",
			"main-user-login",
		},
		WebResources: templates.WebResourceDependencies{
			HG_AUTH: true,
		},
	},
	templates.TmplPageUserSignedOut: {
		Name: "page-user-signed-out",
		Dependencies: []string{
			"layout-top",
			"layout-bottom",
			"main-user-dash",
			"main-user-login",
		},
		WebResources: templates.WebResourceDependencies{},
	},
	templates.TmplComponentErrors: {
		Name:         "component-errors",
		Dependencies: []string{},
	},
	templates.TmplComponentToast: {
		Name:         "component-toast",
		Dependencies: []string{},
	},
}
