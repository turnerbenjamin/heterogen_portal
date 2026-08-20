package query

import (
	"fmt"
)

type ValueExpression interface {
	GetTypeName() string
	GetType() literalType
	IsCompatibleWithComparisonOperator(op ComparisonOperator) bool
	IsSupportedByDbType(dbType DbDataTypeName) bool
	WriteFilterExpression(w *queryWriter, fieldName string, op ComparisonOperator) error
}

type literalType uint8

const (
	literalTypeNull literalType = iota
	literalTypeString
	literalTypeInt
	literalTypeFloat
	literalTypeStringList
	literalTypeIntList
	literalTypeFloatList
)

type NullLiteral struct {
	Value string
}

func (l *NullLiteral) GetTypeName() string { return "null" }

func (l *NullLiteral) GetType() literalType { return literalTypeNull }

func (l *NullLiteral) IsCompatibleWithComparisonOperator(op ComparisonOperator) bool {
	switch op {
	case ComparisonEq,
		ComparisonNe:
		return true
	default:
		return false
	}
}

func (l *NullLiteral) IsSupportedByDbType(_ DbDataTypeName) bool {
	return true
}

func (l *NullLiteral) WriteFilterExpression(w *queryWriter, fieldName string, op ComparisonOperator) error {
	switch op {
	case ComparisonEq:
		fmt.Fprintf(w.sb, "%s IS NULL", fieldName)
	case ComparisonNe:
		fmt.Fprintf(w.sb, "%s IS NOT NULL", fieldName)
	default:
		return fmt.Errorf("unsupported operation: %s", string(op))
	}
	return nil
}

type StringLiteral struct {
	Value string
}

func (l *StringLiteral) GetTypeName() string { return "string" }

func (l *StringLiteral) GetType() literalType { return literalTypeString }

func (l *StringLiteral) IsCompatibleWithComparisonOperator(op ComparisonOperator) bool {
	switch op {
	case ComparisonEq,
		ComparisonNe,
		ComparisonContains,
		ComparisonStartsWith,
		ComparisonEndsWith,
		comparisonNotContains,
		comparisonNotStartsWith,
		comparisonNotEndsWith,
		ComparisonGe,
		ComparisonGt,
		ComparisonLt,
		ComparisonLe:
		return true
	default:
		return false
	}
}

func (l *StringLiteral) IsSupportedByDbType(dbType DbDataTypeName) bool {
	switch dbType {
	case DbTypeNvarchar,
		DbTypeGeography,
		DbTypeDateTimeOffset:
		return true
	default:
		return false
	}
}

func (l *StringLiteral) WriteFilterExpression(
	w *queryWriter,
	fieldName string,
	op ComparisonOperator,
) error {
	switch op {
	case ComparisonEq:
		fmt.Fprintf(w.sb, "%s = %s", fieldName, w.arg(l.Value))
		w.args = append(w.args, l.Value)
	case ComparisonNe:
		fmt.Fprintf(w.sb, "%s != %s", fieldName, w.arg(l.Value))
		w.args = append(w.args, l.Value)
	case ComparisonGt:
		fmt.Fprintf(w.sb, "%s > %s", fieldName, w.arg(l.Value))
	case ComparisonGe:
		fmt.Fprintf(w.sb, "%s >= %s", fieldName, w.arg(l.Value))
	case ComparisonLt:
		fmt.Fprintf(w.sb, "%s < %s", fieldName, w.arg(l.Value))
	case ComparisonLe:
		fmt.Fprintf(w.sb, "%s <= %s", fieldName, w.arg(l.Value))
	case ComparisonContains, comparisonNotContains:
		modifier := ""
		if op == comparisonNotContains {
			modifier = "NOT "
		}

		pattern := fmt.Sprintf("%%%s%%", l.Value)
		fmt.Fprintf(w.sb, "%s %sLIKE %s", fieldName, modifier, w.arg(pattern))

	case ComparisonStartsWith, comparisonNotStartsWith:
		modifier := ""
		if op == comparisonNotStartsWith {
			modifier = "NOT "
		}

		pattern := fmt.Sprintf("%s%%", l.Value)
		fmt.Fprintf(w.sb, "%s %sLIKE %s", fieldName, modifier, w.arg(pattern))
	case ComparisonEndsWith, comparisonNotEndsWith:
		modifier := ""
		if op == comparisonNotEndsWith {
			modifier = "NOT "
		}

		pattern := fmt.Sprintf("%%%s", l.Value)
		fmt.Fprintf(w.sb, "%s %sLIKE %s", fieldName, modifier, w.arg(pattern))
	default:
		return fmt.Errorf("unsupported string operation: %s", string(op))
	}
	return nil
}

type IntLiteral struct {
	Value int64
}

func (l *IntLiteral) GetTypeName() string { return "int" }

func (l *IntLiteral) GetType() literalType { return literalTypeInt }

func (l *IntLiteral) String() string { return fmt.Sprintf("%d", l.Value) }

func (l *IntLiteral) IsCompatibleWithComparisonOperator(op ComparisonOperator) bool {
	return isNumberCompatibleWith(op)
}

func (l *IntLiteral) IsSupportedByDbType(dbType DbDataTypeName) bool {
	return dbType == DbTypeInt
}

func (l *IntLiteral) WriteFilterExpression(
	w *queryWriter,
	fieldName string,
	op ComparisonOperator,
) error {
	switch op {
	case ComparisonEq:
		fmt.Fprintf(w.sb, "%s = %s", fieldName, w.arg(l.Value))
	case ComparisonNe:
		fmt.Fprintf(w.sb, "%s != %s", fieldName, w.arg(l.Value))
	case ComparisonGt:
		fmt.Fprintf(w.sb, "%s > %s", fieldName, w.arg(l.Value))
	case ComparisonGe:
		fmt.Fprintf(w.sb, "%s >= %s", fieldName, w.arg(l.Value))
	case ComparisonLt:
		fmt.Fprintf(w.sb, "%s < %s", fieldName, w.arg(l.Value))
	case ComparisonLe:
		fmt.Fprintf(w.sb, "%s <= %s", fieldName, w.arg(l.Value))
	default:
		return fmt.Errorf("unsupported string operation: %s", string(op))
	}
	return nil
}

type FloatLiteral struct {
	Value float64
}

func (l *FloatLiteral) GetTypeName() string { return "float" }

func (l *FloatLiteral) GetType() literalType { return literalTypeFloat }

func (l *FloatLiteral) IsCompatibleWithComparisonOperator(op ComparisonOperator) bool {
	return isNumberCompatibleWith(op)
}

func (l *FloatLiteral) IsSupportedByDbType(dbType DbDataTypeName) bool {
	return dbType == DbTypeFloat
}

func (l *FloatLiteral) WriteFilterExpression(
	w *queryWriter,
	fieldName string,
	op ComparisonOperator,
) error {
	switch op {
	case ComparisonEq:
		fmt.Fprintf(w.sb, "%s = %s", fieldName, w.arg(l.Value))
	case ComparisonNe:
		fmt.Fprintf(w.sb, "%s != %s", fieldName, w.arg(l.Value))
	case ComparisonGt:
		fmt.Fprintf(w.sb, "%s > %s", fieldName, w.arg(l.Value))
	case ComparisonGe:
		fmt.Fprintf(w.sb, "%s >= %s", fieldName, w.arg(l.Value))
	case ComparisonLt:
		fmt.Fprintf(w.sb, "%s < %s", fieldName, w.arg(l.Value))
	case ComparisonLe:
		fmt.Fprintf(w.sb, "%s <= %s", fieldName, w.arg(l.Value))
	default:
		return fmt.Errorf("unsupported string operation: %s", string(op))
	}
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

func (l *StringListLiteral) GetTypeName() string { return "string list" }

func (l *StringListLiteral) GetType() literalType { return literalTypeStringList }

func (l *StringListLiteral) IsCompatibleWithComparisonOperator(op ComparisonOperator) bool {
	return op == ComparisonIn
}

func (l *StringListLiteral) IsSupportedByDbType(dbType DbDataTypeName) bool {
	sl := StringLiteral{}
	return (&sl).IsSupportedByDbType(dbType)
}

func (l *StringListLiteral) WriteFilterExpression(
	w *queryWriter,
	fieldName string,
	op ComparisonOperator,
) error {
	return WriteListFilterExpression(l.Values, w, fieldName, op)
}

type IntListLiteral struct {
	Values []int64
}

func (l *IntListLiteral) GetTypeName() string  { return "int" }
func (l *IntListLiteral) GetType() literalType { return literalTypeIntList }

func (l *IntListLiteral) IsCompatibleWithComparisonOperator(op ComparisonOperator) bool {
	return op == ComparisonIn
}

func (l *IntListLiteral) IsSupportedByDbType(dbType DbDataTypeName) bool {
	il := IntListLiteral{}
	return (&il).IsSupportedByDbType(dbType)
}

func (l *IntListLiteral) WriteFilterExpression(
	w *queryWriter,
	fieldName string,
	op ComparisonOperator,
) error {
	return WriteListFilterExpression(l.Values, w, fieldName, op)
}

type FloatListLiteral struct {
	Values []float64
}

func (l *FloatListLiteral) GetTypeName() string  { return "float list" }
func (l *FloatListLiteral) GetType() literalType { return literalTypeFloatList }

func (l *FloatListLiteral) IsCompatibleWithComparisonOperator(op ComparisonOperator) bool {
	return op == ComparisonIn
}

func (l *FloatListLiteral) IsSupportedByDbType(dbType DbDataTypeName) bool {
	fl := FloatLiteral{}
	return (&fl).IsSupportedByDbType(dbType)
}

func (l *FloatListLiteral) WriteFilterExpression(
	w *queryWriter,
	fieldName string,
	op ComparisonOperator,
) error {
	return WriteListFilterExpression(l.Values, w, fieldName, op)
}

func WriteListFilterExpression[T any](
	values []T,
	w *queryWriter,
	fieldName string,
	op ComparisonOperator,
) error {
	switch op {
	case ComparisonIn, comparisonNotIn:
		modifier := ""
		if op == comparisonNotIn {
			modifier = "NOT "
		}

		fmt.Fprintf(w.sb, "%s %sIN (", fieldName, modifier)
		for i, v := range values {
			if i != 0 {
				w.sb.WriteRune(',')
			}
			w.sb.WriteString(w.arg(v))
		}
		w.sb.WriteRune(')')
	default:
		return fmt.Errorf("unsupported list operation: %s", string(op))
	}
	return nil
}
