package valuebuilder

import mdl "github.com/turnerbenjamin/heterogen_portal/internal/query/queryModel"

type stringValue struct {
	value string
}

var supportedComparisonOperatorsString = map[mdl.ComparisonOperator]struct{}{
	mdl.ComparisonEq:            {},
	mdl.ComparisonNe:            {},
	mdl.ComparisonContains:      {},
	mdl.ComparisonStartsWith:    {},
	mdl.ComparisonEndsWith:      {},
	mdl.ComparisonNotContains:   {},
	mdl.ComparisonNotStartsWith: {},
	mdl.ComparisonNotEndsWith:   {},
	mdl.ComparisonGe:            {},
	mdl.ComparisonGt:            {},
	mdl.ComparisonLt:            {},
	mdl.ComparisonLe:            {},
}

func (v stringValue) Type() mdl.ValueType {
	return mdl.ValueTypeString
}

func (v stringValue) SupportsType(t mdl.DbType) bool {
	return t == mdl.DbTypeString
}

func (v stringValue) SupportsOperator(op mdl.ComparisonOperator) bool {
	_, supported := supportedComparisonOperatorsString[op]
	return supported
}

func (v stringValue) WriteFilterExpression(
	w mdl.QueryWriter,
	fieldName string,
	op mdl.ComparisonOperator,
) error {
	return w.WriteFilterExpressionString(fieldName, op, v.value)
}

func (v stringValue) Serialise(s mdl.Serialiser) {
	s.SerialiseString(v.value)
}
