package valuebuilder

import mdl "github.com/turnerbenjamin/heterogen_portal/internal/query/queryModel"

type intValue struct {
	value int64
}

var supportedComparisonOperatorsInt = map[mdl.ComparisonOperator]struct{}{
	mdl.ComparisonEq: {},
	mdl.ComparisonNe: {},
	mdl.ComparisonGt: {},
	mdl.ComparisonGe: {},
	mdl.ComparisonLt: {},
	mdl.ComparisonLe: {},
}

func (v intValue) Type() mdl.ValueType {
	return mdl.ValueTypeInt
}

func (v intValue) SupportsType(t mdl.DbType) bool {
	return t == mdl.DbTypeInt || t == mdl.DbTypeFloat
}

func (v intValue) SupportsOperator(op mdl.ComparisonOperator) bool {
	_, supported := supportedComparisonOperatorsInt[op]
	return supported
}

func (v intValue) WriteFilterExpression(
	w mdl.QueryWriter,
	fieldName string,
	op mdl.ComparisonOperator,
) error {
	return w.WriteFilterExpressionInt(fieldName, op, v.value)
}

func (v intValue) Serialise(s mdl.Serialiser) {
	s.SerialiseInt(v.value)
}
