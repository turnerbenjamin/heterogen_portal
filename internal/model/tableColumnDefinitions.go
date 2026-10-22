// GENERATED CODE
// see cmd/modelbuilder/
package model

import "github.com/turnerbenjamin/querystack/querymodel"

type AppColType interface {
	IsAppColType()
}

type AppColTypeString struct {
	MaxLenth int16
}

func (t AppColTypeString) IsAppColType() {}

type AppColTypeNumber struct{}

func (t AppColTypeNumber) IsAppColType() {}

type AppColTypeTime struct{}

func (t AppColTypeTime) IsAppColType() {}

type AppColTypePoint struct{}

func (t AppColTypePoint) IsAppColType() {}

type AppColTypeIdentifier struct{}

func (t AppColTypeIdentifier) IsAppColType() {}

type AppColTypeEnum struct {
	EnumType string
}

func (t AppColTypeEnum) IsAppColType() {}

type AppColTypeLookup struct {
	TableName string
}

func (t AppColTypeLookup) IsAppColType() {}

type TableColumn[T querymodel.TableModel] struct {
	Path        string
	DisplayName string
	Type        AppColType
	Expansion   string
	GetData     func(T) string
}

var BusinessesIdColumn = TableColumn[*BusinessesModel]{
	Path:        "id",
	DisplayName: "Id",
	Expansion:   "",
}

var BusinessesReferenceColumn = TableColumn[*BusinessesModel]{
	Path:        "reference",
	DisplayName: "Reference",
	Expansion:   "",
}

var BusinessesTradingNameColumn = TableColumn[*BusinessesModel]{
	Path:        "trading_name",
	DisplayName: "Trading Name",
	Expansion:   "",
}

var BusinessesBusinessTypeColumn = TableColumn[*BusinessesModel]{
	Path:        "business_type",
	DisplayName: "Business Type",
	Expansion:   "",
}

var BusinessesCphNumberColumn = TableColumn[*BusinessesModel]{
	Path:        "cph_number",
	DisplayName: "Cph Number",
	Expansion:   "",
}

var BusinessesEmailAddressColumn = TableColumn[*BusinessesModel]{
	Path:        "email_address",
	DisplayName: "Email Address",
	Expansion:   "",
}

var BusinessesWebsiteUrlColumn = TableColumn[*BusinessesModel]{
	Path:        "website_url",
	DisplayName: "Website Url",
	Expansion:   "",
}

var BusinessesAddressLine1Column = TableColumn[*BusinessesModel]{
	Path:        "address_line_1",
	DisplayName: "Address Line 1",
	Expansion:   "",
}

var BusinessesAddressLine2Column = TableColumn[*BusinessesModel]{
	Path:        "address_line_2",
	DisplayName: "Address Line 2",
	Expansion:   "",
}

var BusinessesTownColumn = TableColumn[*BusinessesModel]{
	Path:        "town",
	DisplayName: "Town",
	Expansion:   "",
}

var BusinessesCountyColumn = TableColumn[*BusinessesModel]{
	Path:        "county",
	DisplayName: "County",
	Expansion:   "",
}

var BusinessesCountryColumn = TableColumn[*BusinessesModel]{
	Path:        "country",
	DisplayName: "Country",
	Expansion:   "",
}

var BusinessesPostcodeColumn = TableColumn[*BusinessesModel]{
	Path:        "postcode",
	DisplayName: "Postcode",
	Expansion:   "",
}

var BusinessesCreatedAtColumn = TableColumn[*BusinessesModel]{
	Path:        "created_at",
	DisplayName: "Created At",
	Expansion:   "",
}

var BusinessesCreatedByIdDisplayNameColumn = TableColumn[*BusinessesModel]{
	Path:        "created_by_id/display_name",
	DisplayName: "Display Name",
	Expansion:   "expand=created_by_id(select=display_name)",
}

var BusinessesModifiedAtColumn = TableColumn[*BusinessesModel]{
	Path:        "modified_at",
	DisplayName: "Modified At",
	Expansion:   "",
}

var BusinessesModifiedByIdDisplayNameColumn = TableColumn[*BusinessesModel]{
	Path:        "modified_by_id/display_name",
	DisplayName: "Display Name",
	Expansion:   "expand=modified_by_id(select=display_name)",
}

var FarmFieldsBusinessIdCreatedByIdDisplayNameColumn = TableColumn[*FarmFieldsModel]{
	Path:        "business_id/created_by_id/display_name",
	DisplayName: "Display Name",
	Expansion:   "expand=business_id(expand=created_by_id(select=display_name))",
}

var UsersIdColumn = TableColumn[*UsersModel]{
	Path:        "id",
	DisplayName: "Id",
	Expansion:   "",
}

var UsersGivenNameColumn = TableColumn[*UsersModel]{
	Path:        "given_name",
	DisplayName: "Given Name",
	Expansion:   "",
}

var UsersFamilyNameColumn = TableColumn[*UsersModel]{
	Path:        "family_name",
	DisplayName: "Family Name",
	Expansion:   "",
}

var UsersDisplayNameColumn = TableColumn[*UsersModel]{
	Path:        "display_name",
	DisplayName: "Display Name",
	Expansion:   "",
}

var UsersEmailAddressColumn = TableColumn[*UsersModel]{
	Path:        "email_address",
	DisplayName: "Email Address",
	Expansion:   "",
}

var UsersCreatedAtColumn = TableColumn[*UsersModel]{
	Path:        "created_at",
	DisplayName: "Created At",
	Expansion:   "",
}

var UsersModifiedAtColumn = TableColumn[*UsersModel]{
	Path:        "modified_at",
	DisplayName: "Modified At",
	Expansion:   "",
}
