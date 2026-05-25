package config

type TemplateIdentifier int

const (
	TMPL_PAGE_APP TemplateIdentifier = iota
	TMPL_PAGE_USER_SIGN_IN
	TMPL_PAGE_USER_SIGN_IN_REDIRECT
	TMPL_PAGE_USER_SIGNED_OUT
	TMPL_COMPONENT_ERRORS
	TMPL_COMPONENT_TOAST
	TMPL_ENUM_END
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

var TemplateDataMap = map[TemplateIdentifier]TemplateData{
	TMPL_PAGE_APP: {
		Name: "page-app",
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
	TMPL_PAGE_USER_SIGN_IN: {
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
	TMPL_PAGE_USER_SIGN_IN_REDIRECT: {
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
	TMPL_PAGE_USER_SIGNED_OUT: {
		Name: "page-user-signed-out",
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
	TMPL_COMPONENT_ERRORS: {
		Name:         "component-errors",
		Dependencies: []string{},
	},
	TMPL_COMPONENT_TOAST: {
		Name:         "component-toast",
		Dependencies: []string{},
	},
}
