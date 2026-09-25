package valueBuilder

import (
	querybuilder "github.com/turnerbenjamin/heterogen_portal/internal/query/queryBuilder"
	qerr "github.com/turnerbenjamin/heterogen_portal/internal/query/queryError"
	mdl "github.com/turnerbenjamin/heterogen_portal/internal/query/queryModel"
)

type stringListValue struct {
	value []string
}

var supportedComparisonOperatorsStringList = map[mdl.ComparisonOperator]struct{}{
	mdl.ComparisonIn:    {},
	mdl.ComparisonNotIn: {},
}

func (v stringListValue) Type() mdl.ValueType {
	return mdl.ValueTypeStringList
}

func (v stringListValue) SupportsType(t mdl.DbType) bool {
	return t == mdl.DbTypeString
}

func (v stringListValue) SupportsOperator(op mdl.ComparisonOperator) bool {
	_, supported := supportedComparisonOperatorsStringList[op]
	return supported
}

func (v stringListValue) WriteFilterExpression(
	b *querybuilder.Builder,
	w mdl.QueryWriter,
	fieldName string,
	op mdl.ComparisonOperator,
) error {
	return w.WriteFilterExpressionStringList(b, fieldName, op, v.value)
}

func buildStringListValue(els []mdl.Value) (mdl.Value, error) {
	o := stringListValue{
		value: make([]string, len(els)),
	}
	for i, el := range els {
		switch v := el.(type) {
		case stringValue:
			o.value[i] = v.value
		default:
			return nil, qerr.InternalErr(
				"string list does not support elements of type %s",
				ValueTypeString(el.Type()),
			)
		}
	}
	return o, nil
}

func (v stringListValue) Serialise(s mdl.Serialiser) {
	s.SerialiseStringList(v.value)
}
