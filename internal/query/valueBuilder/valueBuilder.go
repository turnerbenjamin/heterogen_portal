package valuebuilder

import (
	"time"

	qerr "github.com/turnerbenjamin/heterogen_portal/internal/query/queryError"
	mdl "github.com/turnerbenjamin/heterogen_portal/internal/query/queryModel"
)

type valueBuilder struct{}

func NewValueBuilder() mdl.ValueBuilder {
	return &valueBuilder{}
}

func (valueBuilder) Null() mdl.Value {
	return nullValue{}
}

func (valueBuilder) String(s string) mdl.Value {
	return stringValue{value: s}
}

func (valueBuilder) Int(i int64) mdl.Value {
	return intValue{value: i}
}

func (valueBuilder) Float(f float64) mdl.Value {
	return floatValue{value: f}
}

func (valueBuilder) Point(longitude float64, latitude float64) mdl.Value {
	return pointValue{value: mdl.NewPoint(longitude, latitude)}
}

func (valueBuilder) DateTime(dt time.Time) mdl.Value {
	return dateTimeValue{value: dt}
}

func (valueBuilder) List(els []mdl.Value) (mdl.Value, error) {
	return buildListValue(els)
}

func (valueBuilder) ExecuteDeserialisation(
	ds mdl.Deserialiser,
	t mdl.ValueType,
	d []byte,
) (mdl.Value, error) {
	switch t {

	// DESERIALISE - NULL
	case mdl.ValueTypeNull:
		return nullValue{}, nil

	// DESERIALISE - STRING
	case mdl.ValueTypeString:
		v := ds.DeserialiseString(d)
		return stringValue{value: v}, nil

	// DESERIALISE - INT
	case mdl.ValueTypeInt:
		v, err := ds.DeserialiseInt(d)
		if err != nil {
			return nil, err
		}
		return intValue{value: v}, nil

	// DESERIALISE - FLOAT
	case mdl.ValueTypeFloat:
		v, err := ds.DeserialiseFloat(d)
		if err != nil {
			return nil, err
		}
		return floatValue{value: v}, nil

	// DESERIALISE - STRING LIST
	case mdl.ValueTypeStringList:
		stringValues := ds.DeserialiseListString(d)
		return stringListValue{value: stringValues}, nil

	// DESERIALISE - INT LIST
	case mdl.ValueTypeIntList:
		values, err := ds.DeserialiseListInt(d)
		return intListValue{value: values}, err

	// DESERIALISE - FLOAT LIST
	case mdl.ValueTypeFloatList:
		values, err := ds.DeserialiseListFloat(d)
		return floatListValue{value: values}, err
	}
	return nil, nil
}

func buildListValue(els []mdl.Value) (mdl.Value, error) {
	if len(els) == 0 {
		return nil, qerr.InternalErr("lists must contain at least one element")
	}

	listType := els[0].Type()
	switch listType {
	case mdl.ValueTypeString:
		return buildStringListValue(els)
	case mdl.ValueTypeInt:
		return buildIntListValue(els)
	case mdl.ValueTypeFloat:
		return buildFloatListValue(els)
	default:
		return nil, qerr.InternalErr(
			"lists of type '%s' are not supported",
			ValueTypeString(listType),
		)
	}
}

func ValueTypeString(t mdl.ValueType) string {
	switch t {
	case mdl.ValueTypeNull:
		return "null"

	case mdl.ValueTypeString:
		return "string"

	case mdl.ValueTypeStringList:
		return "string list"

	case mdl.ValueTypeInt:
		return "integer"

	case mdl.ValueTypeIntList:
		return "integer list"

	case mdl.ValueTypeFloat:
		return "decimal"

	case mdl.ValueTypeDateTime:
		return "date/time"

	case mdl.ValueTypePoint:
		return "point"

	case mdl.ValueTypeFloatList:
		return "decimal list"

	default:
		panic("unexpected literal type received")
	}
}
