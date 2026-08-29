package azSqlWriter

import (
	"fmt"

	mdl "github.com/turnerbenjamin/heterogen_portal/internal/query/queryModel"
)

type NullLiteral struct{}

func (l *NullLiteral) Type() mdl.LiteralType { return mdl.LiteralTypeNull }

func (l *NullLiteral) Value() any {
	return nil
}

func (l *NullLiteral) Serialise(_ mdl.Serialiser) {}

func (l *NullLiteral) IsSupportedByDbType(_ mdl.DbDataTypeName) bool {
	return true
}

func (l *NullLiteral) IsCompatibleWithComparisonOperator(op mdl.ComparisonOperator) bool {
	switch op {
	case mdl.ComparisonEq,
		mdl.ComparisonNe:
		return true
	default:
		return false
	}
}

func (l *NullLiteral) WriteFilterExpression(w mdl.QueryWriter, fieldName string, op mdl.ComparisonOperator) error {
	switch op {
	case mdl.ComparisonEq:
		w.Write("%s IS NULL", fieldName)
	case mdl.ComparisonNe:
		w.Write("%s IS NOT NULL", fieldName)
	default:
		return fmt.Errorf("unsupported operation: %s", string(op))
	}
	return nil
}
