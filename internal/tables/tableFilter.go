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
	columnData map[string]ColumnData[T],
	ast querymodel.FilterExpression,
) (FilterItem, error) {
	pathBuilder := strings.Builder{}

	if ast == nil {
		return nil, nil
	}

	filterItems, err := flattenExpression(columnData, pathBuilder, nil, ast)
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

	return filterItems[0], nil
}

func flattenExpression[T querymodel.TableModel](
	columnData map[string]ColumnData[T],
	pathBuilder strings.Builder,
	parentGroupType *querymodel.LogicalOperator,
	ast querymodel.FilterExpression,
) ([]FilterItem, error) {
	switch typedEx := ast.(type) {
	case *querymodel.LogicalExpression:
		left, err := flattenExpression(
			columnData,
			pathBuilder,
			&typedEx.Operator,
			typedEx.Left,
		)
		if err != nil {
			return nil, err
		}

		right, err := flattenExpression(
			columnData,
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
		path, columnType, err := getColData(
			columnData,
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

		supportedOperators, exists := DataTypeToOperatorMapping[columnType]
		if !exists {
			return nil, fmt.Errorf(
				"unable to build table filter: unable to access supported operators for %s",
				columnType,
			)
		}

		return []FilterItem{FilterRule{
			ColumnType: columnType,
			ColumnPath: path,

			Operator:           operator,
			SupportedOperators: supportedOperators,

			Value: typedEx.Value,
		}}, nil

	default:
		return nil, errors.New(
			"unable to build table filter: unsupported filter expression received",
		)
	}
}

func getColData[T querymodel.TableModel](
	columns map[string]ColumnData[T],
	pathBuilder strings.Builder,
	col *querymodel.ResolvedColumn,
) (path string, columnType model.DataType, err error) {

	pathBuilder.Reset()
	for _, s := range col.ResolvedPath.Steps {
		pathBuilder.WriteString(s.Relationship.ColumnName)
		pathBuilder.WriteByte('/')
	}
	pathBuilder.WriteString(col.Metadata.Name)

	path = pathBuilder.String()

	colData, exists := columns[pathBuilder.String()]
	if !exists {
		return "", "", fmt.Errorf(
			"unable to get col data: unable to access column with path %s",
			pathBuilder.String(),
		)
	}

	return path, colData.DataType, nil
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
