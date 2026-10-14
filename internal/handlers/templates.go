// Package handlers contains HTTP handlers for the application
package handlers

import (
	"context"
	"io"

	"github.com/turnerbenjamin/heterogen_portal/internal/etc"
	"github.com/turnerbenjamin/heterogen_portal/internal/model"
	"github.com/turnerbenjamin/heterogen_portal/internal/templates"
)

// TemplateStore is an interface wrapping a simple Execute method used to
// write a given template to a writer
type TemplateStore interface {
	ExecutePageApp(
		ctx context.Context,
		w io.Writer,
		pageConfig templates.PageConfig,
		user *model.UsersModel,
	) error

	ExecutePageUserSignedOut(
		ctx context.Context,
		w io.Writer,
		pageConfig templates.PageConfig,
	) error

	ExecutePageUsers(
		ctx context.Context,
		w io.Writer,
		pageConfig templates.PageConfig,
		tableData *templates.TableData[*model.UsersModel],
		resourceEndpoint string,
	) error

	ExecutePageBusinesses(
		ctx context.Context,
		w io.Writer,
		pageConfig templates.PageConfig,
		tableData *templates.TableData[*model.BusinessesModel],
		resourceEndpoint string,
	) error

	ExecutePageOutOfAppError(
		ctx context.Context,
		w io.Writer,
		pageConfig templates.PageConfig,
	) error

	ExecuteRenderAppError(
		ctx context.Context,
		w io.Writer,
		errorData *etc.AppError,
	) error
}
