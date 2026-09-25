package valueBuilder

import (
	querybuilder "github.com/turnerbenjamin/heterogen_portal/internal/query/queryBuilder"
	qerr "github.com/turnerbenjamin/heterogen_portal/internal/query/queryError"
	mdl "github.com/turnerbenjamin/heterogen_portal/internal/query/queryModel"
)

type floatListValue struct {
	value []float64
}

var supportedComparisonOperatorsFloatList = map[mdl.ComparisonOperator]struct{}{
	mdl.ComparisonIn:    {},
	mdl.ComparisonNotIn: {},
}

func (v floatListValue) Type() mdl.ValueType {
	return mdl.ValueTypeIntList
}

func (v floatListValue) SupportsType(t mdl.DbType) bool {
	return t == mdl.DbTypeFloat
}

func (v floatListValue) SupportsOperator(op mdl.ComparisonOperator) bool {
	_, supported := supportedComparisonOperatorsFloatList[op]
	return supported
}

func (v floatListValue) WriteFilterExpression(
	b *querybuilder.Builder,
	w mdl.QueryWriter,
	fieldName string,
	op mdl.ComparisonOperator,
) error {
	return w.WriteFilterExpressionFloatList(b, fieldName, op, v.value)
}

func (v floatListValue) Serialise(s mdl.Serialiser) {
	s.SerialiseFloatList(v.value)
}

func buildFloatListValue(els []mdl.Value) (mdl.Value, error) {
	o := floatListValue{
		value: make([]float64, len(els)),
	}
	for i, el := range els {
		switch v := el.(type) {
		case floatValue:
			o.value[i] = v.value
		default:
			return nil, qerr.InternalErr(
				"float list does not support elements of type %s",
				ValueTypeString(el.Type()),
			)
		}
	}
	return o, nil
}
