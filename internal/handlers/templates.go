// Package handlers contains HTTP handlers for the application
package handlers

import (
	"io"

	"github.com/turnerbenjamin/heterogen_portal/internal/model"
	"github.com/turnerbenjamin/heterogen_portal/internal/templates"
)

// TemplateStore is an interface wrapping a simple Execute method used to
// write a given template to a writer
type TemplateStore interface {
	ExecutePageOutOfAppError(
		w io.Writer,
		config templates.TemplateConfig[templates.ErrorTemplateData],
	) error

	ExecutePageUserSignedOut(
		w io.Writer,
		config templates.TemplateConfig[templates.PageUserSignedOutData],
	) error

	ExecutePageApp(
		w io.Writer,
		config templates.TemplateConfig[templates.PageAppData],
	) error

	ExecutePageBusinesses(
		w io.Writer,
		config templates.TemplateConfig[templates.TableData[*model.BusinessesModel]],
	) error

	ExecutePageUsers(
		w io.Writer,
		config templates.TemplateConfig[templates.TableData[*model.UsersModel]],
	) error
}
