// GENERATED CODE
// see cmd/modelbuilder/
package model

import "github.com/turnerbenjamin/querystack/querymodel"

type AppColType interface {
	IsAppColType()
}

type AppColTypeString struct {
	MaxLength int16
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
	Options []EnumOption
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
	DisplayName: "Id",
	Path:        "id",
	Expansion:   "",
	Type:        AppColTypeIdentifier{},
	GetData: func(r *BusinessesModel) string {
		v := r.Id
		return v.String()
	},
}

var BusinessesReferenceColumn = TableColumn[*BusinessesModel]{
	DisplayName: "Reference",
	Path:        "reference",
	Expansion:   "",
	Type: AppColTypeString{
		MaxLength: int16(8),
	},
	GetData: func(r *BusinessesModel) string {
		v := r.Reference
		return v
	},
}

var BusinessesTradingNameColumn = TableColumn[*BusinessesModel]{
	DisplayName: "Trading Name",
	Path:        "trading_name",
	Expansion:   "",
	Type: AppColTypeString{
		MaxLength: int16(255),
	},
	GetData: func(r *BusinessesModel) string {
		v := r.TradingName
		return v
	},
}

var BusinessesBusinessTypeColumn = TableColumn[*BusinessesModel]{
	DisplayName: "Business Type",
	Path:        "business_type",
	Expansion:   "",
	Type: AppColTypeEnum{
		Options: EnumBusinessesBusinessTypeOptions,
	},
	GetData: func(r *BusinessesModel) string {
		v := r.BusinessType
		label, exists := v.String()
		if !exists {
			return "-"
		}
		return label
	},
}

var BusinessesCphNumberColumn = TableColumn[*BusinessesModel]{
	DisplayName: "CPH Number",
	Path:        "cph_number",
	Expansion:   "",
	Type: AppColTypeString{
		MaxLength: int16(11),
	},
	GetData: func(r *BusinessesModel) string {
		v := r.CphNumber
		return v
	},
}

var BusinessesEmailAddressColumn = TableColumn[*BusinessesModel]{
	DisplayName: "Email Address",
	Path:        "email_address",
	Expansion:   "",
	Type: AppColTypeString{
		MaxLength: int16(320),
	},
	GetData: func(r *BusinessesModel) string {
		v := r.EmailAddress
		return v
	},
}

var BusinessesWebsiteUrlColumn = TableColumn[*BusinessesModel]{
	DisplayName: "Website Url",
	Path:        "website_url",
	Expansion:   "",
	Type: AppColTypeString{
		MaxLength: int16(2048),
	},
	GetData: func(r *BusinessesModel) string {
		v := r.WebsiteUrl
		return v
	},
}

var BusinessesAddressLine1Column = TableColumn[*BusinessesModel]{
	DisplayName: "Address Line 1",
	Path:        "address_line_1",
	Expansion:   "",
	Type: AppColTypeString{
		MaxLength: int16(255),
	},
	GetData: func(r *BusinessesModel) string {
		v := r.AddressLine1
		return v
	},
}

var BusinessesAddressLine2Column = TableColumn[*BusinessesModel]{
	DisplayName: "Address Line 2",
	Path:        "address_line_2",
	Expansion:   "",
	Type: AppColTypeString{
		MaxLength: int16(255),
	},
	GetData: func(r *BusinessesModel) string {
		v := r.AddressLine2
		return v
	},
}

var BusinessesTownColumn = TableColumn[*BusinessesModel]{
	DisplayName: "Town",
	Path:        "town",
	Expansion:   "",
	Type: AppColTypeString{
		MaxLength: int16(100),
	},
	GetData: func(r *BusinessesModel) string {
		v := r.Town
		return v
	},
}

var BusinessesCountyColumn = TableColumn[*BusinessesModel]{
	DisplayName: "County",
	Path:        "county",
	Expansion:   "",
	Type: AppColTypeString{
		MaxLength: int16(100),
	},
	GetData: func(r *BusinessesModel) string {
		v := r.County
		return v
	},
}

var BusinessesCountryColumn = TableColumn[*BusinessesModel]{
	DisplayName: "Country",
	Path:        "country",
	Expansion:   "",
	Type: AppColTypeString{
		MaxLength: int16(100),
	},
	GetData: func(r *BusinessesModel) string {
		v := r.Country
		return v
	},
}

var BusinessesPostcodeColumn = TableColumn[*BusinessesModel]{
	DisplayName: "Postcode",
	Path:        "postcode",
	Expansion:   "",
	Type: AppColTypeString{
		MaxLength: int16(20),
	},
	GetData: func(r *BusinessesModel) string {
		v := r.Postcode
		return v
	},
}

var BusinessesCreatedAtColumn = TableColumn[*BusinessesModel]{
	DisplayName: "Created At",
	Path:        "created_at",
	Expansion:   "",
	Type:        AppColTypeTime{},
	GetData: func(r *BusinessesModel) string {
		v := r.CreatedAt
		return v.Format("2 Jan 06 15:04")
	},
}

var BusinessesCreatedByIdColumn = TableColumn[*BusinessesModel]{
	DisplayName: "Created By Id",
	Path:        "created_by_id",
	Expansion:   "expand=created_by_id(select=display_name)",
	Type: AppColTypeLookup{
		TableName: "users",
	},
	GetData: func(r *BusinessesModel) string {
		intermediateVal1 := r.CreatedBy
		if intermediateVal1 == nil {
			return "-"
		}

		v := intermediateVal1.DisplayName
		return v
	},
}

var BusinessesCreatedByIdDisplayNameColumn = TableColumn[*BusinessesModel]{
	DisplayName: "Created By Id(Display Name)",
	Path:        "created_by_id/display_name",
	Expansion:   "expand=created_by_id(select=display_name)",
	Type: AppColTypeString{
		MaxLength: int16(128),
	},
	GetData: func(r *BusinessesModel) string {
		intermediateVal1 := r.CreatedBy
		if intermediateVal1 == nil {
			return "-"
		}

		v := intermediateVal1.DisplayName
		return v
	},
}

var BusinessesModifiedAtColumn = TableColumn[*BusinessesModel]{
	DisplayName: "Modified At",
	Path:        "modified_at",
	Expansion:   "",
	Type:        AppColTypeTime{},
	GetData: func(r *BusinessesModel) string {
		v := r.ModifiedAt
		return v.Format("2 Jan 06 15:04")
	},
}

var BusinessesModifiedByIdColumn = TableColumn[*BusinessesModel]{
	DisplayName: "Modified By Id",
	Path:        "modified_by_id",
	Expansion:   "expand=modified_by_id(select=display_name)",
	Type: AppColTypeLookup{
		TableName: "users",
	},
	GetData: func(r *BusinessesModel) string {
		intermediateVal1 := r.ModifiedBy
		if intermediateVal1 == nil {
			return "-"
		}

		v := intermediateVal1.DisplayName
		return v
	},
}

var UsersIdColumn = TableColumn[*UsersModel]{
	DisplayName: "Id",
	Path:        "id",
	Expansion:   "",
	Type:        AppColTypeIdentifier{},
	GetData: func(r *UsersModel) string {
		v := r.Id
		return v.String()
	},
}

var UsersGivenNameColumn = TableColumn[*UsersModel]{
	DisplayName: "Given Name",
	Path:        "given_name",
	Expansion:   "",
	Type: AppColTypeString{
		MaxLength: int16(64),
	},
	GetData: func(r *UsersModel) string {
		v := r.GivenName
		return v
	},
}

var UsersFamilyNameColumn = TableColumn[*UsersModel]{
	DisplayName: "Family Name",
	Path:        "family_name",
	Expansion:   "",
	Type: AppColTypeString{
		MaxLength: int16(64),
	},
	GetData: func(r *UsersModel) string {
		v := r.FamilyName
		return v
	},
}

var UsersDisplayNameColumn = TableColumn[*UsersModel]{
	DisplayName: "Display Name",
	Path:        "display_name",
	Expansion:   "",
	Type: AppColTypeString{
		MaxLength: int16(128),
	},
	GetData: func(r *UsersModel) string {
		v := r.DisplayName
		return v
	},
}

var UsersEmailAddressColumn = TableColumn[*UsersModel]{
	DisplayName: "Email Address",
	Path:        "email_address",
	Expansion:   "",
	Type: AppColTypeString{
		MaxLength: int16(320),
	},
	GetData: func(r *UsersModel) string {
		v := r.EmailAddress
		return v
	},
}

var UsersCreatedAtColumn = TableColumn[*UsersModel]{
	DisplayName: "Created At",
	Path:        "created_at",
	Expansion:   "",
	Type:        AppColTypeTime{},
	GetData: func(r *UsersModel) string {
		v := r.CreatedAt
		return v.Format("2 Jan 06 15:04")
	},
}

var UsersModifiedAtColumn = TableColumn[*UsersModel]{
	DisplayName: "Modified At",
	Path:        "modified_at",
	Expansion:   "",
	Type:        AppColTypeTime{},
	GetData: func(r *UsersModel) string {
		v := r.ModifiedAt
		return v.Format("2 Jan 06 15:04")
	},
}
