package valuebuilder

import (
	mdl "github.com/turnerbenjamin/heterogen_portal/internal/query/queryModel"
)

type pointValue struct {
	value mdl.Point
}

var supportedComparisonOperatorsPoint = map[mdl.ComparisonOperator]struct{}{}

func (v pointValue) Type() mdl.ValueType {
	return mdl.ValueTypePoint
}

func (v pointValue) SupportsType(t mdl.DbType) bool {
	return t == mdl.DbTypePoint
}

func (v pointValue) SupportsOperator(op mdl.ComparisonOperator) bool {
	_, supported := supportedComparisonOperatorsPoint[op]
	return supported
}

func (v pointValue) WriteFilterExpression(
	w mdl.QueryWriter,
	fieldName string,
	op mdl.ComparisonOperator,
) error {
	return w.WriteFilterExpressionPoint(fieldName, op, v.value)
}

func (v pointValue) Serialise(s mdl.Serialiser) {
	panic("NOT YET IMPLEMENTED")
}
