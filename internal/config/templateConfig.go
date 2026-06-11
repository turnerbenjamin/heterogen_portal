package config

import (
	"github.com/turnerbenjamin/heterogen_portal/internal/templates"
)

var TemplateDataMap = map[templates.TemplateIdentifier]templates.TemplateData{
	templates.TMPL_PAGE_APP: {
		Name: "page-app",
		Dependencies: []string{
			"layout-top",
			"layout-bottom",
			"main-user-dash",
			"main-user-login",
		},
		WebResources: templates.WebResourceDependencies{},
	},
	templates.TMPL_PAGE_USER_SIGN_IN: {
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
	templates.TMPL_PAGE_USER_SIGN_IN_REDIRECT: {
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
	templates.TMPL_PAGE_USER_SIGN_OUT: {
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
	templates.TMPL_PAGE_USER_SIGNED_OUT: {
		Name: "page-user-signed-out",
		Dependencies: []string{
			"layout-top",
			"layout-bottom",
			"main-user-dash",
			"main-user-login",
		},
		WebResources: templates.WebResourceDependencies{},
	},
	templates.TMPL_COMPONENT_ERRORS: {
		Name:         "component-errors",
		Dependencies: []string{},
	},
	templates.TMPL_COMPONENT_TOAST: {
		Name:         "component-toast",
		Dependencies: []string{},
	},
}
