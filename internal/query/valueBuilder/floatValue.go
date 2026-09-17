package valueBuilder

import mdl "github.com/turnerbenjamin/heterogen_portal/internal/query/queryModel"

type floatValue struct {
	value float64
}

var supportedComparisonOperatorsFloat = map[mdl.ComparisonOperator]struct{}{
	mdl.ComparisonEq: {},
	mdl.ComparisonNe: {},
	mdl.ComparisonGt: {},
	mdl.ComparisonGe: {},
	mdl.ComparisonLt: {},
	mdl.ComparisonLe: {},
}

func (v floatValue) Type() mdl.ValueType {
	return mdl.ValueTypeFloat
}

func (v floatValue) SupportsType(t mdl.DbType) bool {
	return t == mdl.DbTypeFloat
}

func (v floatValue) SupportsOperator(op mdl.ComparisonOperator) bool {
	_, supported := supportedComparisonOperatorsFloat[op]
	return supported
}

func (v floatValue) WriteFilterExpression(
	w mdl.QueryWriter,
	fieldName string,
	op mdl.ComparisonOperator,
) error {
	return w.WriteFilterExpressionFloat(fieldName, op, v.value)
}

func (v floatValue) Serialise(s mdl.Serialiser) {
	s.SerialiseFloat(v.value)
}
