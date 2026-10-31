package templatestore

import (
	"time"

	"github.com/a-h/templ"
	"github.com/turnerbenjamin/heterogen_portal/internal/model"
	"github.com/turnerbenjamin/heterogen_portal/internal/templates/components/inputs"
	"github.com/turnerbenjamin/querystack/querymodel"
)

var InputComponentBuilder = inputComponentBuilder{}

type inputComponentBuilder struct{}

func (b inputComponentBuilder) StringInput(v *string) templ.Component {
	return inputs.TextInput(v)
}

func (b inputComponentBuilder) DateInput(v *time.Time) templ.Component {
	return inputs.DateInput(v)
}

func (b inputComponentBuilder) EnumInput(
	typeName string,
	options []model.EnumOption,
	v *uint8,
) templ.Component {
	return inputs.Enuminput(typeName, options, v)
}

func (b inputComponentBuilder) IdentifierInput(v *string) templ.Component {
	return inputs.IdentifierInput(v)
}

func (b inputComponentBuilder) LookupInput(tableName string, v *string) templ.Component {
	return inputs.LookupInput(tableName, v)
}

func (b inputComponentBuilder) IntegerInput(v *int64) templ.Component {
	return inputs.IntegerInput(v)
}

func (b inputComponentBuilder) FloatInput(v *float64) templ.Component {
	return inputs.FloatInput(v)
}

func (b inputComponentBuilder) PointInput(_ *querymodel.Point) templ.Component {
	return inputs.PointInput()
}
