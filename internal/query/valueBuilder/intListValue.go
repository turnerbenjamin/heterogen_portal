package valueBuilder

import (
	querybuilder "github.com/turnerbenjamin/heterogen_portal/internal/query/queryBuilder"
	qerr "github.com/turnerbenjamin/heterogen_portal/internal/query/queryError"
	mdl "github.com/turnerbenjamin/heterogen_portal/internal/query/queryModel"
)

type intListValue struct {
	value []int64
}

var supportedComparisonOperatorsIntList = map[mdl.ComparisonOperator]struct{}{
	mdl.ComparisonIn:    {},
	mdl.ComparisonNotIn: {},
}

func (v intListValue) Type() mdl.ValueType {
	return mdl.ValueTypeIntList
}

func (v intListValue) SupportsType(t mdl.DbType) bool {
	return t == mdl.DbTypeInt || t == mdl.DbTypeFloat
}

func (v intListValue) SupportsOperator(op mdl.ComparisonOperator) bool {
	_, supported := supportedComparisonOperatorsIntList[op]
	return supported
}

func (v intListValue) WriteFilterExpression(
	b *querybuilder.Builder,
	w mdl.QueryWriter,
	fieldName string,
	op mdl.ComparisonOperator,
) error {
	return w.WriteFilterExpressionIntList(b, fieldName, op, v.value)
}

func (v intListValue) Serialise(s mdl.Serialiser) {
	s.SerialiseIntList(v.value)
}

func buildIntListValue(els []mdl.Value) (mdl.Value, error) {
	o := intListValue{
		value: make([]int64, len(els)),
	}
	for i, el := range els {
		switch v := el.(type) {
		case intValue:
			o.value[i] = v.value
		default:
			return nil, qerr.InternalErr(
				"int list does not support elements of type %s",
				ValueTypeString(el.Type()),
			)
		}
	}
	return o, nil
}
