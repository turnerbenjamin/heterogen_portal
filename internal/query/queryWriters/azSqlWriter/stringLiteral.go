package azSqlWriter

import (
	"fmt"

	mdl "github.com/turnerbenjamin/heterogen_portal/internal/query/queryModel"
)

type StringLiteral struct {
	value string
}

func (l *StringLiteral) Type() mdl.LiteralType { return mdl.LiteralTypeString }

func (l *StringLiteral) Value() any {
	return l.value
}

func (l *StringLiteral) Serialise(s mdl.Serialiser) {
	s.SerialiseString(l.value)
}

func (l *StringLiteral) IsCompatibleWithComparisonOperator(op mdl.ComparisonOperator) bool {
	switch op {
	case mdl.ComparisonEq,
		mdl.ComparisonNe,
		mdl.ComparisonContains,
		mdl.ComparisonStartsWith,
		mdl.ComparisonEndsWith,
		mdl.ComparisonNotContains,
		mdl.ComparisonNotStartsWith,
		mdl.ComparisonNotEndsWith,
		mdl.ComparisonGe,
		mdl.ComparisonGt,
		mdl.ComparisonLt,
		mdl.ComparisonLe:
		return true
	default:
		return false
	}
}

func (l *StringLiteral) IsSupportedByDbType(dbType mdl.DbDataTypeName) bool {
	switch dbType {
	case mdl.DbTypeNvarchar,
		mdl.DbTypeGeography,
		mdl.DbTypeDateTimeOffset:
		return true
	default:
		return false
	}
}

func (l *StringLiteral) WriteFilterExpression(
	w mdl.QueryWriter,
	fieldName string,
	op mdl.ComparisonOperator,
) error {
	switch op {
	case mdl.ComparisonEq:
		w.Write("%s = %s", fieldName, w.Placeholder(l.value))
	case mdl.ComparisonNe:
		w.Write("%s != %s", fieldName, w.Placeholder(l.value))
	case mdl.ComparisonGt:
		w.Write("%s > %s", fieldName, w.Placeholder(l.value))
	case mdl.ComparisonGe:
		w.Write("%s >= %s", fieldName, w.Placeholder(l.value))
	case mdl.ComparisonLt:
		w.Write("%s < %s", fieldName, w.Placeholder(l.value))
	case mdl.ComparisonLe:
		w.Write("%s <= %s", fieldName, w.Placeholder(l.value))
	case mdl.ComparisonContains, mdl.ComparisonNotContains:
		modifier := ""
		if op == mdl.ComparisonNotContains {
			modifier = "NOT "
		}

		pattern := fmt.Sprintf("%%%s%%", l.value)
		w.Write("%s %sLIKE %s", fieldName, modifier, w.Placeholder(pattern))

	case mdl.ComparisonStartsWith, mdl.ComparisonNotStartsWith:
		modifier := ""
		if op == mdl.ComparisonNotStartsWith {
			modifier = "NOT "
		}

		pattern := fmt.Sprintf("%s%%", l.value)
		w.Write("%s %sLIKE %s", fieldName, modifier, w.Placeholder(pattern))
	case mdl.ComparisonEndsWith, mdl.ComparisonNotEndsWith:
		modifier := ""
		if op == mdl.ComparisonNotEndsWith {
			modifier = "NOT "
		}

		pattern := fmt.Sprintf("%%%s", l.value)
		w.Write("%s %sLIKE %s", fieldName, modifier, w.Placeholder(pattern))
	default:
		return fmt.Errorf("unsupported string operation: %s", string(op))
	}
	return nil
}
