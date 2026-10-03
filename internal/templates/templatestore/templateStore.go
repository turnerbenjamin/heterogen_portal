package templatestore

import (
	"context"
	"io"

	"github.com/turnerbenjamin/heterogen_portal/internal/icons"
	"github.com/turnerbenjamin/heterogen_portal/internal/model"
	"github.com/turnerbenjamin/heterogen_portal/internal/templates"
	"github.com/turnerbenjamin/heterogen_portal/internal/templates/components"
	"github.com/turnerbenjamin/heterogen_portal/internal/templates/pages"
)

type store struct{}

func NewStore() *store {
	return &store{}
}

func (s *store) ExecutePageApp(
	ctx context.Context,
	w io.Writer,
	pageConfig templates.PageConfig,
	user *model.UsersModel,
) error {
	webResourceDependencies := templates.WebResourceDependencies{
		HTMX:      true,
		HG_COMMON: true,
	}
	pageConfig.WebResourceDependencies = webResourceDependencies

	page := pages.AppDash(
		pageConfig,
		user,
	)
	return page.Render(ctx, w)
}

func (s *store) ExecutePageUserSignedOut(
	ctx context.Context,
	w io.Writer,
	pageConfig templates.PageConfig,
) error {
	page := pages.UserSignedOut(
		pageConfig,
	)
	return page.Render(ctx, w)
}

func (s *store) ExecutePageUsers(
	ctx context.Context,
	w io.Writer,
	pageConfig templates.PageConfig,
	tableData templates.TableData[*model.UsersModel],
) error {
	webResourceDependencies := templates.WebResourceDependencies{
		HTMX:      true,
		HG_COMMON: true,
	}
	pageConfig.WebResourceDependencies = webResourceDependencies

	page := pages.Users(
		pageConfig,
		icons.Set,
		tableData,
	)
	return page.Render(ctx, w)
}

func (s *store) ExecutePageBusinesses(
	ctx context.Context,
	w io.Writer,
	pageConfig templates.PageConfig,
	tableData templates.TableData[*model.BusinessesModel],
) error {
	webResourceDependencies := templates.WebResourceDependencies{
		HTMX:      true,
		HG_COMMON: true,
	}
	pageConfig.WebResourceDependencies = webResourceDependencies

	page := pages.Businesses(
		pageConfig,
		icons.Set,
		tableData,
	)
	return page.Render(ctx, w)
}

func (s *store) ExecutePageOutOfAppError(
	ctx context.Context,
	w io.Writer,
	pageConfig templates.PageConfig,
) error {
	page := pages.OutOfAppError(pageConfig)
	return page.Render(ctx, w)
}

func (s *store) ExecuteComponentErrors(
	ctx context.Context,
	w io.Writer,
	errorData templates.ErrorTemplateData,
) error {
	component := components.Errors(errorData)
	return component.Render(ctx, w)
}
