// Package handlers contains HTTP handlers for the application.
//
// This file exposes handlers for the users resource
package handlers

import (
	"context"
	"net/http"

	"github.com/turnerbenjamin/heterogen_portal/internal/etc"
	"github.com/turnerbenjamin/heterogen_portal/internal/model"
	"github.com/turnerbenjamin/heterogen_portal/internal/tables"
	"github.com/turnerbenjamin/heterogen_portal/internal/templates"
	"github.com/turnerbenjamin/querystack/querymodel"
)

type UsersService interface {
	ExecuteQuery(
		ctx context.Context,
		queryString string,
	) (*querymodel.TypedExecuteResult[*model.UsersModel], error)
}

type UsersHandler struct {
	templateStore   TemplateStore
	usersService    UsersService
	TableDefinition model.TableDefinition[*model.UsersModel]

	defaultSelect tables.SelectQuery[*model.UsersModel]
}

func NewUsersHandler(
	templateStore TemplateStore,
	usersService UsersService,
) (*UsersHandler, error) {
	h := &UsersHandler{
		templateStore: templateStore,
		usersService:  usersService,

		TableDefinition: model.UsersTable,
		defaultSelect: tables.BuildSelectQuery([]model.TableColumn[*model.UsersModel]{
			model.UsersDisplayNameColumn,
			model.UsersGivenNameColumn,
			model.UsersFamilyNameColumn,
		}),
	}

	return h, nil
}

// GetRootHandler returns a handler for the application root.
func (h *UsersHandler) GetRoot(w http.ResponseWriter, r *http.Request, c *PipelineContext[*UsersTableState]) *etc.AppError {
	pageConfig := templates.PageConfig{
		ContentOnly: DoSendContentOnly(r),
		Title:       "HETEROGEN - USERS",
	}

	d := c.state.TableData
	d.TableDefinition = h.TableDefinition

	if len(d.Query.SelectQuery.Columns) == 0 {
		d.Query.SetSelect(h.defaultSelect)
	}

	result, err := h.usersService.ExecuteQuery(
		r.Context(),
		d.Query.QueryString,
	)
	if err != nil {
		return etc.NewServerError(err)
	}

	tableFilter, err := tables.BuildTableFilter(
		&h.TableDefinition,
		result.Filter,
	)
	if err != nil {
		return etc.NewServerError(err)
	}
	d.Filter = tableFilter

	d.Rows = result.Data

	err = h.templateStore.ExecutePageUsers(
		r.Context(),
		w,
		pageConfig,
		c.state.TableData,
		r.URL.Path,
	)
	if err != nil {
		return etc.NewServerError(err)
	}

	return nil
}
