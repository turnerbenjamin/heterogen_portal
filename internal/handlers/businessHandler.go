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

type BusinessesService interface {
	ExecuteQuery(
		ctx context.Context,
		queryString string,
	) (*querymodel.TypedExecuteResult[*model.BusinessesModel], error)
}

type BusinessesHandler struct {
	templateStore     TemplateStore
	businessesService BusinessesService

	TableDefinition model.TableDefinition[*model.BusinessesModel]

	defaultSelect tables.SelectQuery[*model.BusinessesModel]
}

func NewBusinessesHandler(
	templateStore TemplateStore,
	businessesService BusinessesService,
) (*BusinessesHandler, error) {
	h := &BusinessesHandler{
		templateStore:     templateStore,
		businessesService: businessesService,

		TableDefinition: model.BusinessesTable,

		defaultSelect: tables.BuildSelectQuery([]model.TableColumn[*model.BusinessesModel]{
			model.BusinessesTradingNameColumn,
			model.BusinessesReferenceColumn,
			model.BusinessesBusinessTypeColumn,
			model.BusinessesCountyColumn,
			model.BusinessesCreatedByIdColumn,
		}),
	}

	return h, nil
}

// GetRootHandler returns a handler for the application root.
func (h *BusinessesHandler) GetRoot(
	w http.ResponseWriter,
	r *http.Request,
	c *PipelineContext[*BusinessesTableState],
) *etc.AppError {
	pageConfig := templates.PageConfig{
		ContentOnly: DoSendContentOnly(r),
		Title:       "HETEROGEN - BUSINESSES",
	}

	d := c.state.TableData
	d.TableDefinition = h.TableDefinition

	if len(d.Query.SelectQuery.Columns) == 0 {
		d.Query.SetSelect(h.defaultSelect)
	}

	// d.Query.QueryString +=
	// 	"&filter=reference contains '0' and (county eq 'Dorset' or county " +
	// 		"contains 'shire' or business_type eq 2) and created_at lt " +
	// 		"datetime(2029-10-15T07:30:44.958590557+01:00) and " +
	// 		"(created_by_id/email_address startswith 'info') and (trading_name ne null " +
	// 		" or (created_by_id ne null and trading_name endswith 'turn'))"

	result, err := h.businessesService.ExecuteQuery(
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

	err = h.templateStore.ExecutePageBusinesses(
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
