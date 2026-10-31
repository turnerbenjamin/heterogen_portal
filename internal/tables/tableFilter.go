package tables

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/turnerbenjamin/heterogen_portal/internal/model"
	"github.com/turnerbenjamin/querystack/querymodel"
)

func BuildTableFilter[T querymodel.TableModel](
	tableData *model.TableDefinition[T],
	ast querymodel.FilterExpression,
) (FilterItem, error) {
	pathBuilder := strings.Builder{}

	if ast == nil {
		return FilterGroup{
			GroupType: querymodel.LogicalAnd,
			Items:     []FilterItem{},
		}, nil
	}

	filterItems, err := flattenExpression(
		tableData,
		pathBuilder,
		nil,
		ast,
	)
	if err != nil {
		return nil, err
	}

	if len(filterItems) != 1 {
		return nil, fmt.Errorf(
			"unable to build top level table filter: expected top level filter "+
				"item length to be 1 but got %d",
			len(filterItems),
		)
	}

	topLevelItem := filterItems[0]
	switch i := topLevelItem.(type) {
	case FilterGroup:
		return i, nil
	case FilterRule[T]:
		return FilterGroup{
			GroupType: querymodel.LogicalAnd,
			Items:     []FilterItem{topLevelItem},
		}, nil
	}
	return filterItems[0], nil
}

func flattenExpression[T querymodel.TableModel](
	tableData *model.TableDefinition[T],
	pathBuilder strings.Builder,
	parentGroupType *querymodel.LogicalOperator,
	ast querymodel.FilterExpression,
) ([]FilterItem, error) {
	switch typedEx := ast.(type) {
	case *querymodel.LogicalExpression:
		left, err := flattenExpression(
			tableData,
			pathBuilder,
			&typedEx.Operator,
			typedEx.Left,
		)
		if err != nil {
			return nil, err
		}

		right, err := flattenExpression(
			tableData,
			pathBuilder,
			&typedEx.Operator,
			typedEx.Right,
		)
		if err != nil {
			return nil, err
		}

		if parentGroupType != nil && *parentGroupType == typedEx.Operator {
			return slices.Concat(left, right), nil
		} else {
			return []FilterItem{FilterGroup{
				GroupType: typedEx.Operator,
				Items:     slices.Concat(left, right),
			}}, nil
		}

	case *querymodel.ComparisonExpression:
		column, err := getColDefinition(
			tableData,
			pathBuilder,
			typedEx.ResolvedColumn,
		)
		if err != nil {
			return nil, err
		}

		operator, err := getOperator(
			typedEx.Operator,
			typedEx.Value,
		)
		if err != nil {
			return nil, err
		}

		value, err := parseFilterValue(column, typedEx.Value)
		if err != nil {
			return nil, fmt.Errorf("unable to build table filter: %v", err)
		}

		return []FilterItem{FilterRule[T]{
			Column:   column,
			Operator: operator,
			Value:    value,
		}}, nil

	default:
		return nil, errors.New(
			"unable to build table filter: unsupported filter expression received",
		)
	}
}

func getColDefinition[T querymodel.TableModel](
	tableData *model.TableDefinition[T],
	pathBuilder strings.Builder,
	col *querymodel.ResolvedColumn,
) (*model.TableColumn[T], error) {

	pathBuilder.Reset()
	for _, s := range col.ResolvedPath.Steps {
		pathBuilder.WriteString(s.Relationship.ColumnName)
		pathBuilder.WriteByte('/')
	}
	pathBuilder.WriteString(col.Metadata.Name)

	path := pathBuilder.String()

	column, exists := tableData.MainGridColumnsMap[path]
	if !exists {
		return nil, fmt.Errorf(
			"unable to find column definition for %s -> %s",
			tableData.DisplayName,
			path,
		)
	}

	return &column, nil
}

func getOperator(
	rawOp querymodel.ComparisonOperator,
	value querymodel.Value,
) (*ComparisonOperator, error) {
	if value.Type() == querymodel.ValueTypeNull {
		switch rawOp {
		case querymodel.ComparisonEq:
			return ComparisonOperatorEmpty, nil
		case querymodel.ComparisonNe:
			return ComparisonOperatorNotEmpty, nil
		default:
			return nil, fmt.Errorf(
				"unable to build table filter: operator '%s' is not supported for null values",
				rawOp,
			)
		}
	}

	switch rawOp {
	case querymodel.ComparisonEq:
		return ComparisonOperatorEq, nil
	case querymodel.ComparisonNe:
		return ComparisonOperatorNe, nil
	case querymodel.ComparisonGt:
		return ComparisonOperatorGt, nil
	case querymodel.ComparisonGe:
		return ComparisonOperatorGe, nil
	case querymodel.ComparisonLt:
		return ComparisonOperatorLt, nil
	case querymodel.ComparisonLe:
		return ComparisonOperatorLe, nil
	case querymodel.ComparisonContains:
		return ComparisonOperatorContains, nil
	case querymodel.ComparisonStartsWith:
		return ComparisonOperatorStartsWith, nil
	case querymodel.ComparisonEndsWith:
		return ComparisonOperatorEndsWith, nil
	default:
		return nil, fmt.Errorf(
			"unable to build table filter: unsupported operator '%s'",
			rawOp,
		)

	}
}
