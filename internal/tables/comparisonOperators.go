package tables

import "github.com/turnerbenjamin/heterogen_portal/internal/model"

type ComparisonOperator struct {
	Operator    string
	DisplayName string
}

var (
	ComparisonOperatorEq = &ComparisonOperator{
		Operator:    "eq",
		DisplayName: "Equals",
	}
	ComparisonOperatorNe = &ComparisonOperator{
		Operator:    "ne",
		DisplayName: "Not Equal",
	}
	ComparisonOperatorGt = &ComparisonOperator{
		Operator:    "gt",
		DisplayName: "Greater Than",
	}
	ComparisonOperatorGe = &ComparisonOperator{
		Operator:    "ge",
		DisplayName: "Greater Than or Equal",
	}
	ComparisonOperatorLt = &ComparisonOperator{
		Operator:    "lt",
		DisplayName: "Less Than",
	}
	ComparisonOperatorLe = &ComparisonOperator{
		Operator:    "le",
		DisplayName: "Less Than or Equal",
	}
	ComparisonOperatorContains = &ComparisonOperator{
		Operator:    "contains",
		DisplayName: "Contains",
	}
	ComparisonOperatorStartsWith = &ComparisonOperator{
		Operator:    "startswith",
		DisplayName: "Starts With",
	}
	ComparisonOperatorEndsWith = &ComparisonOperator{
		Operator:    "endswith",
		DisplayName: "Ends With",
	}
	ComparisonOperatorEmpty = &ComparisonOperator{
		Operator:    "empty",
		DisplayName: "Empty",
	}
	ComparisonOperatorNotEmpty = &ComparisonOperator{
		Operator:    "notempty",
		DisplayName: "Not Empty",
	}
)

var DataTypeToOperatorMapping = map[model.DataType][]*ComparisonOperator{
	model.DataTypeString: {
		ComparisonOperatorEq,
		ComparisonOperatorNe,
		ComparisonOperatorEmpty,
		ComparisonOperatorNotEmpty,
		ComparisonOperatorContains,
		ComparisonOperatorStartsWith,
		ComparisonOperatorEndsWith,
	},
	model.DataTypeNumber: {
		ComparisonOperatorEq,
		ComparisonOperatorNe,
		ComparisonOperatorEmpty,
		ComparisonOperatorNotEmpty,
		ComparisonOperatorGt,
		ComparisonOperatorGe,
		ComparisonOperatorLt,
		ComparisonOperatorLe,
	},
	model.DataTypeTime: {
		ComparisonOperatorEq,
		ComparisonOperatorNe,
		ComparisonOperatorEmpty,
		ComparisonOperatorNotEmpty,
		ComparisonOperatorGt,
		ComparisonOperatorGe,
		ComparisonOperatorLt,
		ComparisonOperatorLe,
	},
	model.DataTypeEnum: {
		ComparisonOperatorEq,
		ComparisonOperatorNe,
		ComparisonOperatorEmpty,
		ComparisonOperatorNotEmpty,
	},
}
