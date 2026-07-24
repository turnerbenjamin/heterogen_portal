package query

import (
	"fmt"

	"github.com/turnerbenjamin/heterogen_portal/internal/model"
)

type ValueExpression interface {
	GetTypeName() string
	IsCompatibleWithComparisonOperator(op ComparisonOperator) bool
	IsSupportedByDbType(dbType model.DbDataTypeName) bool
	WriteFilterExpression(o *sqlQuery, fieldName string, op ComparisonOperator) error
}

type NullLiteral struct {
	Value string
}

func (l NullLiteral) GetTypeName() string { return "null" }

func (l NullLiteral) IsCompatibleWithComparisonOperator(op ComparisonOperator) bool {
	switch op {
	case ComparisonEq,
		ComparisonNe:
		return true
	default:
		return false
	}
}

func (l NullLiteral) IsSupportedByDbType(_ model.DbDataTypeName) bool {
	return true
}

func (l NullLiteral) WriteFilterExpression(o *sqlQuery, fieldName string, op ComparisonOperator) error {
	switch op {
	case ComparisonEq:
		fmt.Fprintf(o.sb, "%s IS NULL", fieldName)
	case ComparisonNe:
		fmt.Fprintf(o.sb, "%s IS NOT NULL", fieldName)
	default:
		return fmt.Errorf("unsupported operation: %s", string(op))
	}
	return nil
}

type StringLiteral struct {
	Value string
}

func (l StringLiteral) GetTypeName() string { return "string" }

func (l StringLiteral) IsCompatibleWithComparisonOperator(op ComparisonOperator) bool {
	switch op {
	case ComparisonEq,
		ComparisonNe,
		ComparisonContains,
		ComparisonStartsWith,
		ComparisonEndsWith,
		ComparisonNotContains,
		ComparisonNotStartsWith,
		ComparisonNotEndsWith:
		return true
	default:
		return false
	}
}

func (l StringLiteral) IsSupportedByDbType(dbType model.DbDataTypeName) bool {
	switch dbType {
	case model.DbTypeNvarchar,
		model.DbTypeGeography,
		model.DbTypeDateTimeOffset:
		return true
	default:
		return false
	}
}

func (l StringLiteral) WriteFilterExpression(
	o *sqlQuery,
	fieldName string,
	op ComparisonOperator,
) error {
	switch op {
	case ComparisonEq:
		fmt.Fprintf(o.sb, "%s = %s", fieldName, o.nextPlaceholder())
		o.Args = append(o.Args, l.Value)
	case ComparisonNe:
		fmt.Fprintf(o.sb, "%s != %s", fieldName, o.nextPlaceholder())
		o.Args = append(o.Args, l.Value)
	case ComparisonContains, ComparisonNotContains:
		modifier := ""
		if op == ComparisonNotContains {
			modifier = "NOT "
		}

		fmt.Fprintf(o.sb, "%s %sLIKE %s", fieldName, modifier, o.nextPlaceholder())
		o.Args = append(o.Args, fmt.Sprintf("%%%s%%", l.Value))
	case ComparisonStartsWith, ComparisonNotStartsWith:
		modifier := ""
		if op == ComparisonNotStartsWith {
			modifier = "NOT "
		}

		fmt.Fprintf(o.sb, "%s %sLIKE %s", fieldName, modifier, o.nextPlaceholder())
		o.Args = append(o.Args, fmt.Sprintf("%s%%", l.Value))
	case ComparisonEndsWith, ComparisonNotEndsWith:
		modifier := ""
		if op == ComparisonNotEndsWith {
			modifier = "NOT "
		}

		fmt.Fprintf(o.sb, "%s %sLIKE %s", fieldName, modifier, o.nextPlaceholder())
		o.Args = append(o.Args, fmt.Sprintf("%%%s", l.Value))
	default:
		return fmt.Errorf("unsupported string operation: %s", string(op))
	}
	return nil
}

type IntLiteral struct {
	Value int
}

func (l IntLiteral) GetTypeName() string { return "int" }

func (l IntLiteral) IsCompatibleWithComparisonOperator(op ComparisonOperator) bool {
	return isNumberCompatibleWith(op)
}

func (l IntLiteral) IsSupportedByDbType(dbType model.DbDataTypeName) bool {
	return dbType == model.DbTypeInt
}

func (l IntLiteral) WriteFilterExpression(
	o *sqlQuery,
	fieldName string,
	op ComparisonOperator,
) error {
	switch op {
	case ComparisonEq:
		fmt.Fprintf(o.sb, "%s = %s", fieldName, o.nextPlaceholder())
	case ComparisonNe:
		fmt.Fprintf(o.sb, "%s != %s", fieldName, o.nextPlaceholder())
	case ComparisonGt:
		fmt.Fprintf(o.sb, "%s > %s", fieldName, o.nextPlaceholder())
	case ComparisonGe:
		fmt.Fprintf(o.sb, "%s >= %s", fieldName, o.nextPlaceholder())
	case ComparisonLt:
		fmt.Fprintf(o.sb, "%s < %s", fieldName, o.nextPlaceholder())
	case ComparisonLe:
		fmt.Fprintf(o.sb, "%s <= %s", fieldName, o.nextPlaceholder())
	default:
		return fmt.Errorf("unsupported string operation: %s", string(op))
	}
	o.Args = append(o.Args, l.Value)
	return nil
}

type FloatLiteral struct {
	Value float64
}

func (l FloatLiteral) GetTypeName() string { return "float" }

func (l FloatLiteral) IsCompatibleWithComparisonOperator(op ComparisonOperator) bool {
	return isNumberCompatibleWith(op)
}

func (l FloatLiteral) IsSupportedByDbType(dbType model.DbDataTypeName) bool {
	return dbType == model.DbTypeFloat
}

func (l FloatLiteral) WriteFilterExpression(
	o *sqlQuery,
	fieldName string,
	op ComparisonOperator,
) error {
	switch op {
	case ComparisonEq:
		fmt.Fprintf(o.sb, "%s = %s", fieldName, o.nextPlaceholder())
	case ComparisonNe:
		fmt.Fprintf(o.sb, "%s != %s", fieldName, o.nextPlaceholder())
	case ComparisonGt:
		fmt.Fprintf(o.sb, "%s > %s", fieldName, o.nextPlaceholder())
	case ComparisonGe:
		fmt.Fprintf(o.sb, "%s >= %s", fieldName, o.nextPlaceholder())
	case ComparisonLt:
		fmt.Fprintf(o.sb, "%s < %s", fieldName, o.nextPlaceholder())
	case ComparisonLe:
		fmt.Fprintf(o.sb, "%s <= %s", fieldName, o.nextPlaceholder())
	default:
		return fmt.Errorf("unsupported string operation: %s", string(op))
	}
	o.Args = append(o.Args, l.Value)
	return nil
}

var NumberCompatibleComparisonOperators = map[ComparisonOperator]struct{}{
	ComparisonEq: {},
	ComparisonNe: {},
	ComparisonGt: {},
	ComparisonGe: {},
	ComparisonLt: {},
	ComparisonLe: {},
}

func isNumberCompatibleWith(op ComparisonOperator) bool {
	switch op {
	case ComparisonEq,
		ComparisonNe,
		ComparisonGt,
		ComparisonGe,
		ComparisonLt,
		ComparisonLe:
		return true
	default:
		return false
	}
}

type StringListLiteral struct {
	Values []string
}

func (l StringListLiteral) GetTypeName() string { return "string list" }

func (l StringListLiteral) IsCompatibleWithComparisonOperator(op ComparisonOperator) bool {
	return op == ComparisonIn
}

func (l StringListLiteral) IsSupportedByDbType(dbType model.DbDataTypeName) bool {
	return StringLiteral{}.IsSupportedByDbType(dbType)
}

func (l StringListLiteral) WriteFilterExpression(
	o *sqlQuery,
	fieldName string,
	op ComparisonOperator,
) error {
	return WriteListFilterExpression(l.Values, o, fieldName, op)
}

type IntListLiteral struct {
	Values []int
}

func (l IntListLiteral) GetTypeName() string { return "int list" }

func (l IntListLiteral) IsCompatibleWithComparisonOperator(op ComparisonOperator) bool {
	return op == ComparisonIn
}

func (l IntListLiteral) IsSupportedByDbType(dbType model.DbDataTypeName) bool {
	return IntLiteral{}.IsSupportedByDbType(dbType)
}

func (l IntListLiteral) WriteFilterExpression(
	o *sqlQuery,
	fieldName string,
	op ComparisonOperator,
) error {
	return WriteListFilterExpression(l.Values, o, fieldName, op)
}

type FloatListLiteral struct {
	Values []float64
}

func (l FloatListLiteral) GetTypeName() string { return "float list" }

func (l FloatListLiteral) IsCompatibleWithComparisonOperator(op ComparisonOperator) bool {
	return op == ComparisonIn
}

func (l FloatListLiteral) IsSupportedByDbType(dbType model.DbDataTypeName) bool {
	return FloatListLiteral{}.IsSupportedByDbType(dbType)
}

func (l FloatListLiteral) WriteFilterExpression(
	o *sqlQuery,
	fieldName string,
	op ComparisonOperator,
) error {
	return WriteListFilterExpression(l.Values, o, fieldName, op)
}

func WriteListFilterExpression[T any](
	values []T,
	o *sqlQuery,
	fieldName string,
	op ComparisonOperator,
) error {
	switch op {
	case ComparisonIn, ComparisonNotIn:
		modifier := ""
		if op == ComparisonNotIn {
			modifier = "NOT "
		}

		fmt.Fprintf(o.sb, "%s %sIN (", fieldName, modifier)
		for i, v := range values {
			if i != 0 {
				o.sb.WriteRune(',')
			}
			o.sb.WriteString(o.nextPlaceholder())
			o.Args = append(o.Args, v)
		}
		o.sb.WriteRune(')')
	default:
		return fmt.Errorf("unsupported list operation: %s", string(op))
	}
	return nil
}
