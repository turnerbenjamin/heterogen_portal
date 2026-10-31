package templatestore

import (
	"time"

	"github.com/a-h/templ"
	tablecells "github.com/turnerbenjamin/heterogen_portal/internal/templates/components/tableCells"
	"github.com/turnerbenjamin/querystack/querymodel"
)

var TableCellComponentBuilder = tableCellComponentBuilder{}

type tableCellComponentBuilder struct{}

func (b tableCellComponentBuilder) StringCell(v *string) templ.Component {
	return tablecells.StringCell(v)
}

func (b tableCellComponentBuilder) DateCell(v *time.Time) templ.Component {
	return tablecells.DateCell(v)
}

func (b tableCellComponentBuilder) EnumCell(label *string) templ.Component {
	return tablecells.EnumCell(label)
}

func (b tableCellComponentBuilder) IdentifierCell(v *string) templ.Component {
	return tablecells.StringCell(v)
}

func (b tableCellComponentBuilder) LookupCell(v *string) templ.Component {
	return tablecells.StringCell(v)
}

func (b tableCellComponentBuilder) IntegerCell(v *int64) templ.Component {
	return tablecells.IntegerCell(v)
}

func (b tableCellComponentBuilder) FloatCell(v *float64) templ.Component {
	return tablecells.FloatCell(v, 2)
}

func (b tableCellComponentBuilder) PointCell(v *querymodel.Point) templ.Component {
	return tablecells.PointCell(v)
}
