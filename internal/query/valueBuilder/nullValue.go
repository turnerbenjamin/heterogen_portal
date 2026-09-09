package valuebuilder

import mdl "github.com/turnerbenjamin/heterogen_portal/internal/query/queryModel"

type nullValue struct{}

var supportedComparisonOperatorsNull = map[mdl.ComparisonOperator]struct{}{
	mdl.ComparisonEq: {},
	mdl.ComparisonNe: {},
}

func (v nullValue) Type() mdl.ValueType {
	return mdl.ValueTypeNull
}

func (v nullValue) SupportsType(_ mdl.DbType) bool {
	return true
}

func (v nullValue) SupportsOperator(op mdl.ComparisonOperator) bool {
	_, supported := supportedComparisonOperatorsNull[op]
	return supported
}

func (v nullValue) WriteFilterExpression(
	w mdl.QueryWriter,
	fieldName string,
	op mdl.ComparisonOperator,
) error {
	return w.WriteFilterExpressionNull(fieldName, op)
}

func (v nullValue) Serialise(s mdl.Serialiser) {
	s.SerialiseNull()
}
