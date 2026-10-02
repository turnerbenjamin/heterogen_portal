// Package handlers contains HTTP handlers for the application.
//
// This file exposes handlers for the users resource
package handlers

import (
	"net/http"

	"github.com/turnerbenjamin/heterogen_portal/internal/constants"
	"github.com/turnerbenjamin/heterogen_portal/internal/model"
	"github.com/turnerbenjamin/heterogen_portal/internal/templates"
)

type UsersHandler struct {
	templateStore        TemplateStore
	tableColumns         []ColumnData[*model.UsersModel]
	defaultHeaderColumns []HeaderColumn[*model.UsersModel]
}

func NewUsersHandler(templateStore TemplateStore) *UsersHandler {
	return &UsersHandler{
		templateStore: templateStore,
		tableColumns:  OrderedColumnData(userTableColumns),
		defaultHeaderColumns: []HeaderColumn[*model.UsersModel]{
			{Column: userTableColumns[model.ColUsersUserName]},
			{Column: userTableColumns[model.ColUsersGivenName]},
			{Column: userTableColumns[model.ColUsersFamilyName]},
		},
	}
}

// GetRootHandler returns a handler for the application root.
func (h *UsersHandler) GetRoot(w http.ResponseWriter, r *http.Request, c *PipelineContext[*TableState[*model.UsersModel]]) *AppError {
	pageConfig := templates.PageConfig{
		ContentOnly: r.Header.Get(constants.HxRequestHeaderRequest) != "",
		Title:       "HETEROGEN - USERS",
	}

	c.state.TableName = "Users"
	c.state.SchemaVersion = 1
	c.state.ColumnDefinitions = h.tableColumns
	c.state.TableHeader = h.defaultHeaderColumns

	err := h.templateStore.Execute(
		templates.TmplPageUsers,
		w,
		templates.TemplateArgs{PageConfig: pageConfig, Data: c.state},
	)
	if err != nil {
		return NewServerError(err)
	}
	return nil
}
