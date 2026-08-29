package azSqlWriter

import (
	"fmt"

	mdl "github.com/turnerbenjamin/heterogen_portal/internal/query/queryModel"
)

type FloatLiteral struct {
	value float64
}

func (l *FloatLiteral) Type() mdl.LiteralType { return mdl.LiteralTypeFloat }

func (l *FloatLiteral) Value() any {
	return l.value
}

func (l *FloatLiteral) Serialise(s mdl.Serialiser) {
	s.SerialiseFloat(l.value)
}

func (l *FloatLiteral) IsCompatibleWithComparisonOperator(op mdl.ComparisonOperator) bool {
	return isNumberCompatibleWith(op)
}

func (l *FloatLiteral) IsSupportedByDbType(dbType mdl.DbDataTypeName) bool {
	return dbType == mdl.DbTypeFloat
}

func (l *FloatLiteral) WriteFilterExpression(
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
	default:
		return fmt.Errorf("unsupported string operation: %s", string(op))
	}
	return nil
}

var NumberCompatibleComparisonOperators = map[mdl.ComparisonOperator]struct{}{
	mdl.ComparisonEq: {},
	mdl.ComparisonNe: {},
	mdl.ComparisonGt: {},
	mdl.ComparisonGe: {},
	mdl.ComparisonLt: {},
	mdl.ComparisonLe: {},
}

func isNumberCompatibleWith(op mdl.ComparisonOperator) bool {
	switch op {
	case mdl.ComparisonEq,
		mdl.ComparisonNe,
		mdl.ComparisonGt,
		mdl.ComparisonGe,
		mdl.ComparisonLt,
		mdl.ComparisonLe:
		return true
	default:
		return false
	}
}
