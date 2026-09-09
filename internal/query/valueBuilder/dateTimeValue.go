package valuebuilder

import (
	"time"

	mdl "github.com/turnerbenjamin/heterogen_portal/internal/query/queryModel"
)

type dateTimeValue struct {
	value time.Time
}

var supportedComparisonOperatorsDateTime = map[mdl.ComparisonOperator]struct{}{
	mdl.ComparisonEq: {},
	mdl.ComparisonNe: {},
	mdl.ComparisonGt: {},
	mdl.ComparisonGe: {},
	mdl.ComparisonLt: {},
	mdl.ComparisonLe: {},
}

func (v dateTimeValue) Type() mdl.ValueType {
	return mdl.ValueTypeDateTime
}

func (v dateTimeValue) SupportsType(t mdl.DbType) bool {
	return t == mdl.DbTypeDateTime
}

func (v dateTimeValue) SupportsOperator(op mdl.ComparisonOperator) bool {
	_, supported := supportedComparisonOperatorsDateTime[op]
	return supported
}

func (v dateTimeValue) WriteFilterExpression(
	w mdl.QueryWriter,
	fieldName string,
	op mdl.ComparisonOperator,
) error {
	return w.WriteFilterExpressionDateTime(fieldName, op, v.value)
}

func (v dateTimeValue) Serialise(s mdl.Serialiser) {
	panic("NOT YET IMPLEMENTED")
}
