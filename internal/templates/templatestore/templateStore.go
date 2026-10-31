package templatestore

import (
	"context"
	"io"

	"github.com/turnerbenjamin/heterogen_portal/internal/etc"
	"github.com/turnerbenjamin/heterogen_portal/internal/icons"
	"github.com/turnerbenjamin/heterogen_portal/internal/model"
	"github.com/turnerbenjamin/heterogen_portal/internal/tables"
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
	page := pages.AppDash(
		pageConfig,
		icons.Set,
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
		icons.Set,
	)
	return page.Render(ctx, w)
}

func (s *store) ExecutePageUsers(
	ctx context.Context,
	w io.Writer,
	pageConfig templates.PageConfig,
	tableData *tables.TableData[*model.UsersModel],
	resourceEndpoint string,
) error {
	page := pages.MainTable(
		pageConfig,
		icons.Set,
		InputComponentBuilder,
		TableCellComponentBuilder,
		*tableData,
		resourceEndpoint,
	)
	return page.Render(ctx, w)
}

func (s *store) ExecutePageBusinesses(
	ctx context.Context,
	w io.Writer,
	pageConfig templates.PageConfig,
	tableData *tables.TableData[*model.BusinessesModel],
	resourceEndpoint string,
) error {
	page := pages.MainTable(
		pageConfig,
		icons.Set,
		InputComponentBuilder,
		TableCellComponentBuilder,
		*tableData,
		resourceEndpoint,
	)
	return page.Render(ctx, w)
}

func (s *store) ExecutePageOutOfAppError(
	ctx context.Context,
	w io.Writer,
	pageConfig templates.PageConfig,
) error {
	page := pages.OutOfAppError(pageConfig, icons.Set)
	return page.Render(ctx, w)
}

func (s *store) ExecuteRenderAppError(
	ctx context.Context,
	w io.Writer,
	errorData *etc.AppError,
) error {
	component := components.Errors(errorData)
	return component.Render(ctx, w)
}
