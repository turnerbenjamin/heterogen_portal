package query

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/turnerbenjamin/heterogen_portal/internal/etc"
	"github.com/turnerbenjamin/heterogen_portal/internal/model"
	"golang.org/x/sync/errgroup"
)

type Query struct {
	ctx                  context.Context
	executeQuery         func(ctx context.Context, query string) (jsonResult []byte, err error)
	tableData            model.TableMetadata
	model                model.TableModel
	selectFields         map[string]string
	requiredSelectFields map[string]string
	nestedQueries        map[string]*nestedQuery
	filterExp            string
}

type nestedQuery struct {
	query        *Query
	relationship model.Relationship
}

func NewQuery(
	ctx context.Context,
	exectuteQuery func(ctx context.Context, query string) (jsonResult []byte, err error),
	resource string,
	operations []QueryOperation,
) (*Query, *etc.AppError) {
	resourceModel := model.GetTableModel(resource)
	if resourceModel == nil {
		return nil, &etc.AppError{
			Code:         http.StatusNotFound,
			ErrorMessage: fmt.Sprintf("no resource found for %s", resource),
		}
	}

	q := &Query{
		ctx:                  ctx,
		executeQuery:         exectuteQuery,
		model:                resourceModel,
		tableData:            resourceModel.GetMetadata(),
		selectFields:         map[string]string{},
		requiredSelectFields: map[string]string{},
		nestedQueries:        map[string]*nestedQuery{},
	}

	for _, operation := range operations {
		switch operation.Identifier {
		case "select":
			if err := q.processSelectOperation(operation); err != nil {
				return nil, err
			}
		case "expand":
			if err := q.processExpandOperation(operation); err != nil {
				return nil, err
			}
		case "filter":
			filterParser := &FilterParser{}
			filterParser.Parse(operation)
		default:
			return nil, &etc.AppError{
				Code:         http.StatusBadRequest,
				ErrorMessage: fmt.Sprintf("Unsupported query parameter %s", operation.Identifier),
			}
		}
	}

	return q, nil
}

func (nq *nestedQuery) GetNestedResults(
	ctx context.Context,
	parentResults []model.TableModel,
) ([]model.TableModel, error) {
	seen := map[string]struct{}{}
	joinValues := make([]string, 0, len(parentResults))
	for _, parent := range parentResults {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		value, err := parent.GetJoinOnValue(nq.relationship.Id)
		if err != nil {
			return nil, err
		}

		if _, exists := seen[value]; !exists {
			joinValues = append(joinValues, value)
		}
		seen[value] = struct{}{}
	}

	if len(joinValues) == 0 {
		return []model.TableModel{}, nil
	}

	results, err := nq.execute(ctx, joinValues)
	if err != nil {
		return nil, err
	}

	return results, nil
}

func (q Query) Execute() ([]model.TableModel, error) {
	selectExp := q.buildSelectExpression()
	fromExp := q.buildFromExpression()
	jsonFormattingExp := "FOR JSON AUTO"

	topLevelQuery := fmt.Sprintf("%s %s %s %s;", selectExp, fromExp, q.filterExp, jsonFormattingExp)
	json, err := q.executeQuery(q.ctx, topLevelQuery)
	if err != nil {
		return nil, err
	}

	topLevelResults, err := q.model.NewSlice(json)
	if err != nil {
		return nil, err
	}

	// Fetch results of any nested queries concurrently
	if len(topLevelResults) > 0 && len(q.nestedQueries) > 0 {
		g, ctx := errgroup.WithContext(q.ctx)
		nestedQueryResults := map[string][]model.TableModel{}
		for queryKey, nestedQuery := range q.nestedQueries {
			g.Go(func() error {
				res, err := nestedQuery.GetNestedResults(ctx, topLevelResults)
				if err != nil {
					return err
				}
				nestedQueryResults[queryKey] = res
				return nil
			})
		}

		if err := g.Wait(); err != nil {
			return nil, err
		}

		// Attach nested queries to main results
		for queryKey, nestedQuery := range q.nestedQueries {
			nestedResults, exists := nestedQueryResults[queryKey]
			if !exists {
				continue
			}
			switch nestedQuery.relationship.Type {
			case model.RelationshipManyToOne:
				err := attachNestedResultsForManyToOneQuery(nestedQuery.relationship, topLevelResults, nestedResults)
				if err != nil {
					return nil, err
				}
			case model.RelationshipOneToMany:
				err := attachNestedResultsForOneToManyQuery(nestedQuery.relationship, topLevelResults, nestedResults)
				if err != nil {
					return nil, err
				}
			default:
				return nil, err
			}
		}
	}
	return topLevelResults, nil
}

func (nq *nestedQuery) execute(ctx context.Context, joinOnValues []string) ([]model.TableModel, error) {
	nq.query.filterExp = nq.buildJoinFilterExp(joinOnValues)
	nq.query.ctx = ctx
	return nq.query.Execute()
}

func (nq *nestedQuery) buildJoinFilterExp(joinOnValues []string) string {
	formattedValues := make([]string, len(joinOnValues))
	for i, v := range joinOnValues {
		formattedValues[i] = fmt.Sprintf("'%s'", v)
	}
	return fmt.Sprintf("WHERE %s IN (%s)", nq.relationship.ForeignColumn, strings.Join(formattedValues, ","))
}

func (q Query) processExpandOperation(operation QueryOperation) *etc.AppError {
	for _, operationValue := range operation.Values {
		if _, ok := q.nestedQueries[operationValue.Value]; ok {
			continue
		}

		relationship, ok := q.tableData.Relationships[operationValue.Value]
		if !ok {
			return newColumnNotSupportedError(q.tableData.TableName, operationValue.Value)
		}

		expandQuery, err := NewQuery(
			q.ctx,
			q.executeQuery,
			relationship.RelatedTable,
			operationValue.NestedOperations,
		)
		if err != nil {
			return err
		}

		q.requiredSelectFields[relationship.LocalColumn] = relationship.LocalColumn
		expandQuery.requiredSelectFields[relationship.ForeignColumn] = relationship.ForeignColumn

		q.nestedQueries[operationValue.Value] = &nestedQuery{
			query:        expandQuery,
			relationship: relationship,
		}
	}
	return nil
}

func (q Query) buildSelectExpression() string {
	selectValues := []string{}
	if len(q.selectFields) == 0 {
		for _, col := range q.tableData.Columns {
			selectValues = append(selectValues, q.getSelectValue(col))
		}
	} else {
		seen := map[string]struct{}{}
		for _, fieldValue := range q.selectFields {
			selectValues = append(selectValues, fieldValue)
			seen[fieldValue] = struct{}{}
		}

		for _, fieldValue := range q.requiredSelectFields {
			if _, exists := seen[fieldValue]; exists {
				continue
			}
			selectValues = append(selectValues, fieldValue)
		}
	}

	return fmt.Sprintf("SELECT %s", strings.Join(selectValues, ","))
}

func (q Query) processSelectOperation(operation QueryOperation) *etc.AppError {
	for _, operationValue := range operation.Values {
		colName := operationValue.Value
		if _, ok := q.selectFields[colName]; ok {
			continue
		}

		colData, ok := q.tableData.Columns[colName]
		if !ok {
			return newColumnNotSupportedError(q.tableData.TableName, colName)
		}
		if len(operationValue.NestedOperations) != 0 {
			return newOperationsDoesNotSupportValueArgsError(colName)
		}

		q.selectFields[colName] = q.getSelectValue(colData)
	}
	return nil
}

func (q Query) getSelectValue(col model.ColumnMetadata) string {
	switch col.Type {
	case model.DbTypeGeography:
		return fmt.Sprintf("%s.STAsText() AS %s", col.Name, col.Name)
	default:
		return col.Name
	}
}

func (q Query) buildFromExpression() string {
	fullTableName := fmt.Sprintf("%s.%s", q.tableData.SchemaName, q.tableData.TableName)
	return fmt.Sprintf("FROM %s", fullTableName)
}

func attachNestedResultsForManyToOneQuery(relationship model.Relationship, parentResults, nestedResults []model.TableModel) error {
	nestedResultsMap := map[string]model.TableModel{}
	for _, nestedResult := range nestedResults {
		joinOnValue, err := nestedResult.GetJoinOnValue(relationship.Id)
		if err != nil {
			return err
		}
		nestedResultsMap[joinOnValue] = nestedResult
	}

	for _, result := range parentResults {
		joinOnValue, err := result.GetJoinOnValue(relationship.Id)
		if err != nil {
			return err
		}
		related, ok := nestedResultsMap[joinOnValue]
		if !ok || related == nil {
			return errors.New("unable to join results")
		}
		result.SetRelationshipField(relationship.Id, related)
	}
	return nil
}

func attachNestedResultsForOneToManyQuery(relationship model.Relationship, parentResults, nestedResults []model.TableModel) error {
	parentResultsMap := map[string]model.TableModel{}
	for _, parentResult := range parentResults {
		joinOnValue, err := parentResult.GetJoinOnValue(relationship.Id)
		if err != nil {
			return err
		}
		parentResultsMap[joinOnValue] = parentResult
	}

	for _, nestedResult := range nestedResults {
		joinOnValue, err := nestedResult.GetJoinOnValue(relationship.Id)
		if err != nil {
			return err
		}
		parent, ok := parentResultsMap[joinOnValue]
		if !ok || parent == nil {
			return errors.New("unable to join results")
		}
		parent.SetRelationshipField(relationship.Id, nestedResult)
	}
	return nil
}

func newColumnNotSupportedError(tablename, columnName string) *etc.AppError {
	return &etc.AppError{
		Code: http.StatusBadRequest,
		ErrorMessage: fmt.Sprintf(
			"%s does not contain a column definition for %s",
			tablename,
			columnName,
		)}
}

func newOperationsDoesNotSupportValueArgsError(columnName string) *etc.AppError {
	return &etc.AppError{
		Code:         http.StatusBadRequest,
		ErrorMessage: fmt.Sprintf("%s does not support value arguments", columnName),
	}
}
