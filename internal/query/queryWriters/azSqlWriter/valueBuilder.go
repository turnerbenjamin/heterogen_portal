package azSqlWriter

import (
	mdl "github.com/turnerbenjamin/heterogen_portal/internal/query/queryModel"
)

type valueBuilder struct{}

func NewValueBuilder() *valueBuilder {
	return &valueBuilder{}
}

func (valueBuilder) Null() mdl.ValueExpression {
	return &NullLiteral{}
}

func (valueBuilder) String(s string) mdl.ValueExpression {
	return &StringLiteral{value: s}
}

func (valueBuilder) Int(i int64) mdl.ValueExpression {
	return &IntLiteral{value: i}
}

func (valueBuilder) Float(f float64) mdl.ValueExpression {
	return &FloatLiteral{value: f}
}

func (valueBuilder) List(els []mdl.ValueExpression) (mdl.ValueExpression, error) {
	return newListLiteral(els)
}

func (valueBuilder) Deserialise(
	ds mdl.Deserialiser,
	t mdl.LiteralType,
	d []byte,
) (mdl.ValueExpression, error) {
	switch t {

	case mdl.LiteralTypeNull:
		return &NullLiteral{}, nil

	case mdl.LiteralTypeString:
		v := ds.DeserialiseString(d)
		return &StringLiteral{value: v}, nil

	case mdl.LiteralTypeListString:
		vs := ds.DeserialiseListString(d)

		els := make([]mdl.ValueExpression, len(vs))
		for i, s := range vs {
			els[i] = &StringLiteral{value: s}
		}
		return newListLiteral(els)

	case mdl.LiteralTypeInt:
		v, err := ds.DeserialiseInt(d)
		if err != nil {
			return nil, err
		}
		return &IntLiteral{value: v}, nil

	case mdl.LiteralTypeListInt:
		vs, err := ds.DeserialiseListInt(d)
		if err != nil {
			return nil, err
		}

		els := make([]mdl.ValueExpression, len(vs))
		for i, s := range vs {
			els[i] = &IntLiteral{value: s}
		}
		return newListLiteral(els)

	case mdl.LiteralTypeFloat:
		v, err := ds.DeserialiseFloat(d)
		if err != nil {
			return nil, err
		}
		return &FloatLiteral{value: v}, nil

	case mdl.LiteralTypeListFloat:
		vs, err := ds.DeserialiseListFloat(d)
		if err != nil {
			return nil, err
		}

		els := make([]mdl.ValueExpression, len(vs))
		for i, s := range vs {
			els[i] = &FloatLiteral{value: s}
		}
		return newListLiteral(els)

	}
	return nil, nil
}
