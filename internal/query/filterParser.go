package query

type FilterParser struct{}

func (p *FilterParser) Parse(filterOperation QueryOperation) {
	for _, value := range filterOperation.Values {
		_ = value
	}
}
