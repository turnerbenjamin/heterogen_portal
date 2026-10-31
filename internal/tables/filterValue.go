package tables

import (
	"fmt"
	"math"
	"time"

	"github.com/a-h/templ"
	"github.com/turnerbenjamin/heterogen_portal/internal/model"
	"github.com/turnerbenjamin/querystack/querymodel"
	"github.com/turnerbenjamin/querystack/valuebuilder"
)

type FilterValue interface {
	BuildInput(model.UiInputControlBuilder) templ.Component
}

type DateTimeValue struct {
	Column model.AppColTypeDate
	Value  *time.Time
}

func (v DateTimeValue) BuildInput(b model.UiInputControlBuilder) templ.Component {
	return b.DateInput(v.Value)
}

type FloatValue struct {
	Column model.AppColTypeFloat
	Value  *float64
}

func (v FloatValue) BuildInput(b model.UiInputControlBuilder) templ.Component {
	return b.FloatInput(v.Value)
}

type IntegerValue struct {
	Column model.AppColTypeInteger
	Value  *int64
}

func (v IntegerValue) BuildInput(b model.UiInputControlBuilder) templ.Component {
	return b.IntegerInput(v.Value)
}

type PointValue struct {
	Column model.AppColTypePoint
	Value  *querymodel.Point
}

func (v PointValue) BuildInput(b model.UiInputControlBuilder) templ.Component {
	return b.PointInput(v.Value)
}

type StringValue struct {
	Column model.AppColTypeString
	Value  *string
}

func (v StringValue) BuildInput(b model.UiInputControlBuilder) templ.Component {
	return b.StringInput(v.Value)
}

type IdentifierValue struct {
	Column model.AppColTypeIdentifier
	Value  *string
}

func (v IdentifierValue) BuildInput(b model.UiInputControlBuilder) templ.Component {
	return b.IdentifierInput(v.Value)
}

type LookupValue struct {
	Column model.AppColTypeLookup
	Value  *string
}

func (v LookupValue) BuildInput(b model.UiInputControlBuilder) templ.Component {
	return b.LookupInput(v.Column.TableName, v.Value)
}

type EnumValue struct {
	Column model.AppColTypeEnum
	Value  *uint8
}

func (v EnumValue) BuildInput(b model.UiInputControlBuilder) templ.Component {
	return b.EnumInput(v.Column.TypeName, v.Column.Options, v.Value)
}

func parseFilterValue[T querymodel.TableModel](
	column *model.TableColumn[T],
	queryValue querymodel.Value,
) (FilterValue, error) {

	switch c := column.Type.(type) {

	case model.AppColTypeDate:
		switch v := queryValue.(type) {

		case valuebuilder.NullValue:
			return DateTimeValue{Value: nil, Column: c}, nil

		case valuebuilder.DateTimeValue:
			return DateTimeValue{Value: &v.Value, Column: c}, nil

		default:
			return nil, fmt.Errorf(
				"unable to parse filter value: invalid value type for date col: %d",
				v.Type(),
			)
		}

	case model.AppColTypeFloat:
		switch v := queryValue.(type) {

		case valuebuilder.NullValue:
			return FloatValue{Value: nil, Column: c}, nil

		case valuebuilder.FloatValue:
			return FloatValue{Value: &v.Value, Column: c}, nil

		default:
			return nil, fmt.Errorf(
				"unable to parse filter value: invalid value type for float col: %d",
				v.Type(),
			)
		}

	case model.AppColTypeInteger:
		switch v := queryValue.(type) {

		case valuebuilder.NullValue:
			return IntegerValue{Value: nil, Column: c}, nil

		case valuebuilder.IntValue:
			return IntegerValue{Value: &v.Value, Column: c}, nil

		default:
			return nil, fmt.Errorf(
				"unable to parse filter value: invalid value type for int col: %d",
				v.Type(),
			)
		}

	case model.AppColTypePoint:
		switch v := queryValue.(type) {

		case valuebuilder.NullValue:
			return PointValue{Value: nil, Column: c}, nil

		case valuebuilder.PointValue:
			return PointValue{Value: &v.Value, Column: c}, nil

		default:
			return nil, fmt.Errorf(
				"unable to parse filter value: invalid value type for point col: %d",
				v.Type(),
			)
		}

	case model.AppColTypeString:
		switch v := queryValue.(type) {

		case valuebuilder.NullValue:
			return StringValue{Value: nil, Column: c}, nil

		case valuebuilder.StringValue:
			return StringValue{Value: &v.Value, Column: c}, nil

		default:
			return nil, fmt.Errorf(
				"unable to parse filter value: invalid value type for string col: %d",
				v.Type(),
			)
		}

	case model.AppColTypeIdentifier:
		switch v := queryValue.(type) {

		case valuebuilder.NullValue:
			return IdentifierValue{Value: nil, Column: c}, nil

		case valuebuilder.StringValue:
			return IdentifierValue{Value: &v.Value, Column: c}, nil

		default:
			return nil, fmt.Errorf(
				"unable to parse filter value: invalid value type for identifier col: %d",
				v.Type(),
			)
		}

	case model.AppColTypeLookup:
		switch v := queryValue.(type) {

		case valuebuilder.NullValue:
			return LookupValue{Value: nil, Column: c}, nil

		case valuebuilder.StringValue:
			return LookupValue{Value: &v.Value, Column: c}, nil

		default:
			return nil, fmt.Errorf(
				"unable to parse filter value: invalid value type for lookup col: %d",
				v.Type(),
			)
		}

	case model.AppColTypeEnum:
		switch v := queryValue.(type) {

		case valuebuilder.NullValue:
			return EnumValue{Value: nil, Column: c}, nil

		case valuebuilder.IntValue:
			if v.Value < 0 || v.Value > math.MaxUint8 {
				return nil, fmt.Errorf(
					"unable to parse filter value: enum value is out of range: %d",
					v.Type(),
				)
			}
			cv := uint8(v.Value)
			return EnumValue{Value: &cv, Column: c}, nil

		default:
			return nil, fmt.Errorf(
				"unable to parse filter value: invalid value type for lookup col: %d",
				v.Type(),
			)
		}
	default:
		{
			return nil, fmt.Errorf(
				"unable to parse filter value: unsupported app col type: %s",
				column.Type.GetInputType(),
			)
		}
	}
}
