package queryDataStore

import (
	"github.com/turnerbenjamin/heterogen_portal/internal/query/metadataStore"
	qerr "github.com/turnerbenjamin/heterogen_portal/internal/query/queryError"
	mdl "github.com/turnerbenjamin/heterogen_portal/internal/query/queryModel"
	"github.com/turnerbenjamin/heterogen_portal/internal/query/relationships"
)

type filterExpressionBuilder struct {
	rootResource        mdl.TableMetadata
	metadataBinder      *metadataStore.MetadataBinder
	relationshipPlanner *relationships.RelationshipPlanner
	valueBuilder        mdl.ValueBuilder
}

func (eb filterExpressionBuilder) ValueBuilder() mdl.ValueBuilder {
	return eb.valueBuilder
}

func (eb filterExpressionBuilder) NewLogicalExpression(
	left mdl.FilterExpression,
	operator mdl.LogicalOperator,
	right mdl.FilterExpression,
) (*mdl.LogicalExpression, error) {
	if left == nil || right == nil {
		return nil, qerr.InternalErr(
			"unable to build logical expression: left and right must not be nil",
		)
	}

	return &mdl.LogicalExpression{
		Left:     left,
		Operator: operator,
		Right:    right,
	}, nil
}

func (eb filterExpressionBuilder) NewComparisonExpression(
	columnPath string,
	operator mdl.ComparisonOperator,
	value mdl.ValueExpression,
) (*mdl.ComparisonExpression, error) {
	column, err := eb.metadataBinder.ResolveColumn(columnPath)
	if err != nil {
		return nil, err
	}

	if len(column.ResolvedPath.Steps) > 0 && column.ResolvedPath.Type != mdl.RelationshipManyToOne {
		return nil, qerr.BindingErr(
			"invalid value - '%s'. Comparison operations are only supported "+
				"for N:1 and 1:1 relationships. Please use a collection "+
				"operator",
			columnPath,
		)
	}

	return eb.NewComparisonExpressionFromResolvedColumn(column, operator, value)
}


func (eb filterExpressionBuilder) NewComparisonExpressionFromResolvedColumn(
	column mdl.ResolvedColumn,
	operator mdl.ComparisonOperator,
	value mdl.ValueExpression,
) (*mdl.ComparisonExpression, error) {
	existsNodes := eb.relationshipPlanner.ProcessExists(column.ResolvedPath)

	return &mdl.ComparisonExpression{
		ResolvedColumn: &column,
		Operator:       operator,
		Value:          value,
		ExistsNodes:    existsNodes,
	}, nil
}


func (eb filterExpressionBuilder) NewCollectionExpression(
	resourcePath string,
	operator mdl.CollectionOperator,
	filterExpression mdl.FilterExpression,
) (*mdl.CollectionExpression, error) {
	if filterExpression == nil {
		return nil, qerr.InternalErr(
			"unable to build collection expression: filter expression cannot be nil",
		)
	}

	resolvedPath, err := eb.metadataBinder.ResolvePath(resourcePath)
	if err != nil {
		return nil, qerr.InternalErr("unable to build collection expression path: %w", err)
	}

	if resolvedPath.Type != mdl.RelationshipOneToMany {
		return nil, qerr.BindingErr(
			"invalid value - '%s'. Collection operations are only supported "+
				"for 1:N relationships",
			resolvedPath.Id,
		)
	}

	existsNodes := eb.relationshipPlanner.ProcessExists(resolvedPath)

	return &mdl.CollectionExpression{
		ResolvedPath:     &resolvedPath,
		Operator:         operator,
		FilterExpression: filterExpression,
		ExistsNodes:      existsNodes,
	}, nil
}
