// Package handlers contains HTTP handlers for the application.
//
// This file exposes handlers for the users resource
package handlers

import (
	"context"
	"net/http"

	"github.com/turnerbenjamin/heterogen_portal/internal/etc"
	"github.com/turnerbenjamin/heterogen_portal/internal/model"
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
	templateStore       TemplateStore
	usersService        UsersService
	orderedTableColumns []templates.ColumnData[*model.UsersModel]
	tableColumns        map[string]templates.ColumnData[*model.UsersModel]

	defaultQuery *templates.TableQuery[*model.UsersModel]
}

func NewUsersHandler(
	templateStore TemplateStore,
	usersService UsersService,
) (*UsersHandler, error) {
	h := &UsersHandler{
		templateStore: templateStore,
		usersService:  usersService,

		orderedTableColumns: OrderedColumnData(userTableColumns),
		tableColumns:        userTableColumns,

		defaultQuery: newTableQuery([]templates.ColumnData[*model.UsersModel]{
			columnDataUsersDisplayName,
			columnDataUsersGivenName,
			columnDataUsersFamilyName,
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

	d.TableName = "Users"
	d.SchemaVersion = 1

	d.OrderedColumnDefinitions = h.orderedTableColumns
	d.ColumnDefinitions = h.tableColumns

	if d.Query == nil {
		d.Query = h.defaultQuery
	}

	result, err := h.usersService.ExecuteQuery(
		r.Context(),
		d.Query.QueryString,
	)
	if err != nil {
		return etc.NewServerError(err)
	}

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
