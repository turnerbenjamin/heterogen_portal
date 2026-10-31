package tables

import "github.com/turnerbenjamin/heterogen_portal/internal/model"

type ComparisonOperator struct {
	Operator       string
	DisplayName    string
	SupportedTypes map[model.AppInputType]struct{}
	IsUnary        bool
}

var (
	ComparisonOperatorEmpty = &ComparisonOperator{
		Operator:    "empty",
		DisplayName: "Empty",
		IsUnary:     true,
		SupportedTypes: map[model.AppInputType]struct{}{
			model.AppTypeString:     {},
			model.AppTypeInteger:    {},
			model.AppTypeFloat:      {},
			model.AppTypeDate:       {},
			model.AppTypeEnum:       {},
			model.AppTypeIdentifier: {},
			model.AppTypeLookup:     {},
		},
	}

	ComparisonOperatorNotEmpty = &ComparisonOperator{
		Operator:    "notempty",
		DisplayName: "Not Empty",
		IsUnary:     true,
		SupportedTypes: map[model.AppInputType]struct{}{
			model.AppTypeString:     {},
			model.AppTypeInteger:    {},
			model.AppTypeFloat:      {},
			model.AppTypeDate:       {},
			model.AppTypeEnum:       {},
			model.AppTypeIdentifier: {},
			model.AppTypeLookup:     {},
		},
	}

	ComparisonOperatorEq = &ComparisonOperator{
		Operator:    "eq",
		DisplayName: "Equals",
		SupportedTypes: map[model.AppInputType]struct{}{
			model.AppTypeString:     {},
			model.AppTypeInteger:    {},
			model.AppTypeFloat:      {},
			model.AppTypeDate:       {},
			model.AppTypeEnum:       {},
			model.AppTypeIdentifier: {},
			model.AppTypeLookup:     {},
		},
	}

	ComparisonOperatorNe = &ComparisonOperator{
		Operator:    "ne",
		DisplayName: "Not Equal",
		SupportedTypes: map[model.AppInputType]struct{}{
			model.AppTypeString:     {},
			model.AppTypeInteger:    {},
			model.AppTypeFloat:      {},
			model.AppTypeDate:       {},
			model.AppTypeEnum:       {},
			model.AppTypeIdentifier: {},
			model.AppTypeLookup:     {},
		},
	}

	ComparisonOperatorGt = &ComparisonOperator{
		Operator:    "gt",
		DisplayName: "Greater Than",
		SupportedTypes: map[model.AppInputType]struct{}{
			model.AppTypeInteger: {},
			model.AppTypeFloat:   {},
			model.AppTypeDate:    {},
		},
	}

	ComparisonOperatorGe = &ComparisonOperator{
		Operator:    "ge",
		DisplayName: "Greater Than or Equal",
		SupportedTypes: map[model.AppInputType]struct{}{
			model.AppTypeInteger: {},
			model.AppTypeFloat:   {},
			model.AppTypeDate:    {},
		},
	}
	ComparisonOperatorLt = &ComparisonOperator{
		Operator:    "lt",
		DisplayName: "Less Than",
		SupportedTypes: map[model.AppInputType]struct{}{
			model.AppTypeInteger: {},
			model.AppTypeFloat:   {},
			model.AppTypeDate:    {},
		},
	}
	ComparisonOperatorLe = &ComparisonOperator{
		Operator:    "le",
		DisplayName: "Less Than or Equal",
		SupportedTypes: map[model.AppInputType]struct{}{
			model.AppTypeInteger: {},
			model.AppTypeFloat:   {},
			model.AppTypeDate:    {},
		},
	}
	ComparisonOperatorContains = &ComparisonOperator{
		Operator:    "contains",
		DisplayName: "Contains",
		SupportedTypes: map[model.AppInputType]struct{}{
			model.AppTypeString: {},
		},
	}
	ComparisonOperatorStartsWith = &ComparisonOperator{
		Operator:    "startswith",
		DisplayName: "Starts With",
		SupportedTypes: map[model.AppInputType]struct{}{
			model.AppTypeString: {},
		},
	}
	ComparisonOperatorEndsWith = &ComparisonOperator{
		Operator:    "endswith",
		DisplayName: "Ends With",
		SupportedTypes: map[model.AppInputType]struct{}{
			model.AppTypeString: {},
		},
	}

	ComparisonOperators = []*ComparisonOperator{
		ComparisonOperatorEmpty,
		ComparisonOperatorNotEmpty,
		ComparisonOperatorEq,
		ComparisonOperatorNe,
		ComparisonOperatorGt,
		ComparisonOperatorGe,
		ComparisonOperatorLt,
		ComparisonOperatorLe,
		ComparisonOperatorContains,
		ComparisonOperatorStartsWith,
		ComparisonOperatorEndsWith,
	}
)
