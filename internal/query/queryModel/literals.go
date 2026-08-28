package queryModel

import (
	"fmt"
)

/*
LITERALS ARE TIED TO SYNTAX - WE NEED A LITERAL BUILDER INTERFACE TO BE RETURNED
FROM THE WRITER - THE BELOW WILL USED FOR THE AZURE SQL IMPLEMENTATION OF THE
INTERFACE - WE WILL HAVE FROMSTRING,INT METHODS

LITERALS NEED TO BE RESPONSIBLE FOR WRITING SQL, SERIALISING FOR TOKEN STRINGS
AND COMPATIBILITY TO ISOLATE THEM FROM THE REST OF THE QUERY PACKAGE
*/

type NullLiteral struct {
	Value string
}

func (l *NullLiteral) GetTypeName() string { return "null" }

func (l *NullLiteral) GetType() LiteralType { return LiteralTypeNull }

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

func (l *NullLiteral) WriteFilterExpression(w QueryWriter, fieldName string, op ComparisonOperator) error {
	switch op {
	case ComparisonEq:
		w.Write("%s IS NULL", fieldName)
	case ComparisonNe:
		w.Write("%s IS NOT NULL", fieldName)
	default:
		return fmt.Errorf("unsupported operation: %s", string(op))
	}
	return nil
}

type StringLiteral struct {
	Value string
}

func (l *StringLiteral) GetTypeName() string { return "string" }

func (l *StringLiteral) GetType() LiteralType { return LiteralTypeString }

func (l *StringLiteral) IsCompatibleWithComparisonOperator(op ComparisonOperator) bool {
	switch op {
	case ComparisonEq,
		ComparisonNe,
		ComparisonContains,
		ComparisonStartsWith,
		ComparisonEndsWith,
		ComparisonNotContains,
		ComparisonNotStartsWith,
		ComparisonNotEndsWith,
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
	w QueryWriter,
	fieldName string,
	op ComparisonOperator,
) error {
	switch op {
	case ComparisonEq:
		w.Write("%s = %s", fieldName, w.Placeholder(l.Value))
	case ComparisonNe:
		w.Write("%s != %s", fieldName, w.Placeholder(l.Value))
	case ComparisonGt:
		w.Write("%s > %s", fieldName, w.Placeholder(l.Value))
	case ComparisonGe:
		w.Write("%s >= %s", fieldName, w.Placeholder(l.Value))
	case ComparisonLt:
		w.Write("%s < %s", fieldName, w.Placeholder(l.Value))
	case ComparisonLe:
		w.Write("%s <= %s", fieldName, w.Placeholder(l.Value))
	case ComparisonContains, ComparisonNotContains:
		modifier := ""
		if op == ComparisonNotContains {
			modifier = "NOT "
		}

		pattern := fmt.Sprintf("%%%s%%", l.Value)
		w.Write("%s %sLIKE %s", fieldName, modifier, w.Placeholder(pattern))

	case ComparisonStartsWith, ComparisonNotStartsWith:
		modifier := ""
		if op == ComparisonNotStartsWith {
			modifier = "NOT "
		}

		pattern := fmt.Sprintf("%s%%", l.Value)
		w.Write("%s %sLIKE %s", fieldName, modifier, w.Placeholder(pattern))
	case ComparisonEndsWith, ComparisonNotEndsWith:
		modifier := ""
		if op == ComparisonNotEndsWith {
			modifier = "NOT "
		}

		pattern := fmt.Sprintf("%%%s", l.Value)
		w.Write("%s %sLIKE %s", fieldName, modifier, w.Placeholder(pattern))
	default:
		return fmt.Errorf("unsupported string operation: %s", string(op))
	}
	return nil
}

type IntLiteral struct {
	Value int64
}

func (l *IntLiteral) GetTypeName() string { return "int" }

func (l *IntLiteral) GetType() LiteralType { return LiteralTypeInt }

func (l *IntLiteral) String() string { return fmt.Sprintf("%d", l.Value) }

func (l *IntLiteral) IsCompatibleWithComparisonOperator(op ComparisonOperator) bool {
	return isNumberCompatibleWith(op)
}

func (l *IntLiteral) IsSupportedByDbType(dbType DbDataTypeName) bool {
	return dbType == DbTypeInt
}

func (l *IntLiteral) WriteFilterExpression(
	w QueryWriter,
	fieldName string,
	op ComparisonOperator,
) error {
	switch op {
	case ComparisonEq:
		w.Write("%s = %s", fieldName, w.Placeholder(l.Value))
	case ComparisonNe:
		w.Write("%s != %s", fieldName, w.Placeholder(l.Value))
	case ComparisonGt:
		w.Write("%s > %s", fieldName, w.Placeholder(l.Value))
	case ComparisonGe:
		w.Write("%s >= %s", fieldName, w.Placeholder(l.Value))
	case ComparisonLt:
		w.Write("%s < %s", fieldName, w.Placeholder(l.Value))
	case ComparisonLe:
		w.Write("%s <= %s", fieldName, w.Placeholder(l.Value))
	default:
		return fmt.Errorf("unsupported string operation: %s", string(op))
	}
	return nil
}

type FloatLiteral struct {
	Value float64
}

func (l *FloatLiteral) GetTypeName() string { return "float" }

func (l *FloatLiteral) GetType() LiteralType { return LiteralTypeFloat }

func (l *FloatLiteral) IsCompatibleWithComparisonOperator(op ComparisonOperator) bool {
	return isNumberCompatibleWith(op)
}

func (l *FloatLiteral) IsSupportedByDbType(dbType DbDataTypeName) bool {
	return dbType == DbTypeFloat
}

func (l *FloatLiteral) WriteFilterExpression(
	w QueryWriter,
	fieldName string,
	op ComparisonOperator,
) error {
	switch op {
	case ComparisonEq:
		w.Write("%s = %s", fieldName, w.Placeholder(l.Value))
	case ComparisonNe:
		w.Write("%s != %s", fieldName, w.Placeholder(l.Value))
	case ComparisonGt:
		w.Write("%s > %s", fieldName, w.Placeholder(l.Value))
	case ComparisonGe:
		w.Write("%s >= %s", fieldName, w.Placeholder(l.Value))
	case ComparisonLt:
		w.Write("%s < %s", fieldName, w.Placeholder(l.Value))
	case ComparisonLe:
		w.Write("%s <= %s", fieldName, w.Placeholder(l.Value))
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

func (l *StringListLiteral) GetType() LiteralType { return LiteralTypeStringList }

func (l *StringListLiteral) IsCompatibleWithComparisonOperator(op ComparisonOperator) bool {
	return op == ComparisonIn
}

func (l *StringListLiteral) IsSupportedByDbType(dbType DbDataTypeName) bool {
	sl := StringLiteral{}
	return (&sl).IsSupportedByDbType(dbType)
}

func (l *StringListLiteral) WriteFilterExpression(
	w QueryWriter,
	fieldName string,
	op ComparisonOperator,
) error {
	return WriteListFilterExpression(l.Values, w, fieldName, op)
}

type IntListLiteral struct {
	Values []int64
}

func (l *IntListLiteral) GetTypeName() string  { return "int" }
func (l *IntListLiteral) GetType() LiteralType { return LiteralTypeIntList }

func (l *IntListLiteral) IsCompatibleWithComparisonOperator(op ComparisonOperator) bool {
	return op == ComparisonIn
}

func (l *IntListLiteral) IsSupportedByDbType(dbType DbDataTypeName) bool {
	il := IntListLiteral{}
	return (&il).IsSupportedByDbType(dbType)
}

func (l *IntListLiteral) WriteFilterExpression(
	w QueryWriter,
	fieldName string,
	op ComparisonOperator,
) error {
	return WriteListFilterExpression(l.Values, w, fieldName, op)
}

type FloatListLiteral struct {
	Values []float64
}

func (l *FloatListLiteral) GetTypeName() string  { return "float list" }
func (l *FloatListLiteral) GetType() LiteralType { return LiteralTypeFloatList }

func (l *FloatListLiteral) IsCompatibleWithComparisonOperator(op ComparisonOperator) bool {
	return op == ComparisonIn
}

func (l *FloatListLiteral) IsSupportedByDbType(dbType DbDataTypeName) bool {
	fl := FloatLiteral{}
	return (&fl).IsSupportedByDbType(dbType)
}

func (l *FloatListLiteral) WriteFilterExpression(
	w QueryWriter,
	fieldName string,
	op ComparisonOperator,
) error {
	return WriteListFilterExpression(l.Values, w, fieldName, op)
}

func WriteListFilterExpression[T any](
	values []T,
	w QueryWriter,
	fieldName string,
	op ComparisonOperator,
) error {
	switch op {
	case ComparisonIn, ComparisonNotIn:
		modifier := ""
		if op == ComparisonNotIn {
			modifier = "NOT "
		}

		w.Write("%s %sIN (", fieldName, modifier)
		for i, v := range values {
			if i != 0 {
				w.Write(",")
			}
			w.Write(w.Placeholder(v))
		}
		w.Write(")")
	default:
		return fmt.Errorf("unsupported list operation: %s", string(op))
	}
	return nil
}
