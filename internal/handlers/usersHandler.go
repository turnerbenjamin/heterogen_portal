// Package handlers contains HTTP handlers for the application.
//
// This file exposes handlers for the users resource
package handlers

import (
	"context"
	"net/http"

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
	templateStore        TemplateStore
	usersService         UsersService
	tableColumns         []templates.ColumnData[*model.UsersModel]
	defaultHeaderColumns []templates.HeaderColumn[*model.UsersModel]
}

func NewUsersHandler(
	templateStore TemplateStore,
	usersService UsersService,
) *UsersHandler {
	return &UsersHandler{
		templateStore: templateStore,
		usersService:  usersService,
		tableColumns:  OrderedColumnData(userTableColumns),
		defaultHeaderColumns: []templates.HeaderColumn[*model.UsersModel]{
			{Column: userTableColumns[model.ColUsersDisplayName]},
			{Column: userTableColumns[model.ColUsersGivenName]},
			{Column: userTableColumns[model.ColUsersFamilyName]},
		},
	}
}

// GetRootHandler returns a handler for the application root.
func (h *UsersHandler) GetRoot(w http.ResponseWriter, r *http.Request, c *PipelineContext[*UsersTableState]) *AppError {
	pageConfig := templates.PageConfig{
		ContentOnly: DoSendContentOnly(r),
		Title:       "HETEROGEN - USERS",
	}

	// c.state.TableName = "Users"
	// c.state.SchemaVersion = 1
	// c.state.ColumnDefinitions = h.tableColumns
	// c.state.TableHeader = h.defaultHeaderColumns

	result, err := h.usersService.ExecuteQuery(
		r.Context(),
		"",
	)
	if err != nil {
		return NewServerError(err)
	}

	c.state.TableData = templates.TableData[*model.UsersModel]{
		TableName:         "Users",
		SchemaVersion:     1,
		ColumnDefinitions: h.tableColumns,
		TableHeader:       h.defaultHeaderColumns,
		Rows:              result.Data,
	}

	err = h.templateStore.ExecutePageUsers(
		w,
		templates.TemplateConfig[templates.TableData[*model.UsersModel]]{
			PageConfig: pageConfig,
			Data:       c.state.TableData,
		},
	)
	if err != nil {
		return NewServerError(err)
	}
	return nil
}
