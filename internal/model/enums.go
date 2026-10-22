// GENERATED CODE
// see cmd/modelbuilder/
package model

type EnumOption struct {
	DisplayName string
	Value       uint8
}

type Enum interface {
	String() (string, bool)
	Options() []EnumOption
}

type EnumBusinessesBusinessType uint8

const (
	EnumBusinessesBusinessTypeMerchant EnumBusinessesBusinessType = 1
	EnumBusinessesBusinessTypeGrower   EnumBusinessesBusinessType = 2
)

func (e EnumBusinessesBusinessType) String() (string, bool) {
	switch e {
	case 1:
		return "Merchant", true
	case 2:
		return "Grower", true
	default:
		return "", false
	}
}

var EnumBusinessesBusinessTypeOptions = []EnumOption{
	{
		DisplayName: "Merchant",
		Value:       1,
	},
	{
		DisplayName: "Grower",
		Value:       2,
	},
}

func (e EnumBusinessesBusinessType) Options() []EnumOption {
	return EnumBusinessesBusinessTypeOptions
}
