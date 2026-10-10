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
	templateStore       TemplateStore
	usersService        UsersService
	orderedTableColumns []templates.ColumnData[*model.UsersModel]
	tableColumns        map[string]templates.ColumnData[*model.UsersModel]

	defaultHeaderColumns    []templates.ColumnData[*model.UsersModel]
	defaultHeaderColumnsSet map[string]struct{}
}

func NewUsersHandler(
	templateStore TemplateStore,
	usersService UsersService,
) (*UsersHandler, error) {
	defaultHeaderColumns := []templates.ColumnData[*model.UsersModel]{
		userTableColumns[model.ColUsersDisplayName],
		userTableColumns[model.ColUsersGivenName],
		userTableColumns[model.ColUsersFamilyName],
	}

	h := &UsersHandler{
		templateStore: templateStore,
		usersService:  usersService,

		orderedTableColumns: OrderedColumnData(userTableColumns),
		tableColumns:        userTableColumns,

		defaultHeaderColumns:    defaultHeaderColumns,
		defaultHeaderColumnsSet: make(map[string]struct{}, len(defaultHeaderColumns)),
	}

	for _, c := range h.defaultHeaderColumns {
		h.defaultHeaderColumnsSet[c.Path] = struct{}{}
	}

	return h, nil
}

// GetRootHandler returns a handler for the application root.
func (h *UsersHandler) GetRoot(w http.ResponseWriter, r *http.Request, c *PipelineContext[*UsersTableState]) *AppError {
	pageConfig := templates.PageConfig{
		ContentOnly: DoSendContentOnly(r),
		Title:       "HETEROGEN - USERS",
	}

	result, err := h.usersService.ExecuteQuery(
		r.Context(),
		"",
	)
	if err != nil {
		return NewServerError(err)
	}

	c.state.TableData = templates.TableData[*model.UsersModel]{
		TableName:     "Users",
		SchemaVersion: 1,

		OrderedColumnDefinitions: h.orderedTableColumns,
		ColumnDefinitions:        h.tableColumns,

		HeaderColumns:   h.defaultHeaderColumns,
		HeaderColumnSet: h.defaultHeaderColumnsSet,

		Rows: result.Data,
	}

	err = h.templateStore.ExecutePageUsers(
		r.Context(),
		w,
		pageConfig,
		c.state.TableData,
	)
	if err != nil {
		return NewServerError(err)
	}
	return nil
}
