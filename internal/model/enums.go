// GENERATED CODE
// see cmd/modelbuilder/
package model

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
