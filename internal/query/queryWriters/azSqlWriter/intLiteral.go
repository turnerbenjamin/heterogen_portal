package azSqlWriter

import (
	"fmt"

	mdl "github.com/turnerbenjamin/heterogen_portal/internal/query/queryModel"
)

type IntLiteral struct {
	value int64
}

func (l *IntLiteral) Type() mdl.LiteralType { return mdl.LiteralTypeInt }

func (l *IntLiteral) Value() any {
	return l.value
}

func (l *IntLiteral) Serialise(s mdl.Serialiser) {
	s.SerialiseInt(l.value)
}

func (l *IntLiteral) IsCompatibleWithComparisonOperator(op mdl.ComparisonOperator) bool {
	return isNumberCompatibleWith(op)
}

func (l *IntLiteral) IsSupportedByDbType(dbType mdl.DbDataTypeName) bool {
	return dbType == mdl.DbTypeInt
}

func (l *IntLiteral) WriteFilterExpression(
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
