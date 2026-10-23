// GENERATED CODE
// see cmd/modelbuilder/
package model

import (
	"fmt"

	"github.com/turnerbenjamin/querystack/querymodel"
)

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

type TableDefinition[T querymodel.TableModel] struct {
	DisplayName       string
	PluralDisplayName string
	MainGridColumns   []TableColumn[T]
}

type TableColumn[T querymodel.TableModel] struct {
	Path        string
	DisplayName string
	Type        AppColType
	Expansion   string
	GetData     func(T) string
}

var BusinessesTable = TableDefinition[*BusinessesModel]{
	DisplayName:       "Business",
	PluralDisplayName: "Businesses",
	MainGridColumns: []TableColumn[*BusinessesModel]{
		BusinessesAddressLine1Column,
		BusinessesAddressLine2Column,
		BusinessesBusinessTypeColumn,
		BusinessesContactNumberColumn,
		BusinessesCountryColumn,
		BusinessesCountyColumn,
		BusinessesCphNumberColumn,
		BusinessesCreatedAtColumn,
		BusinessesCreatedByIdColumn,
		BusinessesDescriptionColumn,
		BusinessesEmailAddressColumn,
		BusinessesModifiedAtColumn,
		BusinessesModifiedByIdColumn,
		BusinessesPostcodeColumn,
		BusinessesReferenceColumn,
		BusinessesTownColumn,
		BusinessesTradingNameColumn,
		BusinessesCreatedByIdDisplayNameColumn,
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

var BusinessesContactNumberColumn = TableColumn[*BusinessesModel]{
	DisplayName: "Contact Number",
	Path:        "contact_number",
	Expansion:   "",
	Type: AppColTypeString{
		MaxLength: int16(50),
	},
	GetData: func(r *BusinessesModel) string {
		v := r.ContactNumber
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

var BusinessesDescriptionColumn = TableColumn[*BusinessesModel]{
	DisplayName: "Description",
	Path:        "description",
	Expansion:   "",
	Type: AppColTypeString{
		MaxLength: int16(4000),
	},
	GetData: func(r *BusinessesModel) string {
		v := r.Description
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

var BusinessesLocationColumn = TableColumn[*BusinessesModel]{
	DisplayName: "Location",
	Path:        "location",
	Expansion:   "",
	Type:        AppColTypePoint{},
	GetData: func(r *BusinessesModel) string {
		v := r.Location
		return fmt.Sprintf("[%f, %f]", v.Coordinates[0], v.Coordinates[1])
	},
}

var BusinessesLogoUrlColumn = TableColumn[*BusinessesModel]{
	DisplayName: "Logo Url",
	Path:        "logo_url",
	Expansion:   "",
	Type: AppColTypeString{
		MaxLength: int16(2048),
	},
	GetData: func(r *BusinessesModel) string {
		v := r.LogoUrl
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

var FarmFieldsTable = TableDefinition[*FarmFieldsModel]{
	DisplayName:       "Field",
	PluralDisplayName: "Fields",
	MainGridColumns: []TableColumn[*FarmFieldsModel]{
		FarmFieldsBusinessIdColumn,
		FarmFieldsCreatedAtColumn,
		FarmFieldsCreatedByIdColumn,
		FarmFieldsModifiedAtColumn,
		FarmFieldsModifiedByIdColumn,
		FarmFieldsReferenceColumn,
	},
}

var FarmFieldsBusinessIdColumn = TableColumn[*FarmFieldsModel]{
	DisplayName: "Business Id",
	Path:        "business_id",
	Expansion:   "expand=business_id(select=trading_name)",
	Type: AppColTypeLookup{
		TableName: "businesses",
	},
	GetData: func(r *FarmFieldsModel) string {
		intermediateVal1 := r.Business
		if intermediateVal1 == nil {
			return "-"
		}

		v := intermediateVal1.TradingName
		return v
	},
}

var FarmFieldsCreatedAtColumn = TableColumn[*FarmFieldsModel]{
	DisplayName: "Created At",
	Path:        "created_at",
	Expansion:   "",
	Type:        AppColTypeTime{},
	GetData: func(r *FarmFieldsModel) string {
		v := r.CreatedAt
		return v.Format("2 Jan 06 15:04")
	},
}

var FarmFieldsCreatedByIdColumn = TableColumn[*FarmFieldsModel]{
	DisplayName: "Created By Id",
	Path:        "created_by_id",
	Expansion:   "expand=created_by_id(select=display_name)",
	Type: AppColTypeLookup{
		TableName: "users",
	},
	GetData: func(r *FarmFieldsModel) string {
		intermediateVal1 := r.CreatedBy
		if intermediateVal1 == nil {
			return "-"
		}

		v := intermediateVal1.DisplayName
		return v
	},
}

var FarmFieldsIdColumn = TableColumn[*FarmFieldsModel]{
	DisplayName: "Id",
	Path:        "id",
	Expansion:   "",
	Type:        AppColTypeIdentifier{},
	GetData: func(r *FarmFieldsModel) string {
		v := r.Id
		return v.String()
	},
}

var FarmFieldsLocationColumn = TableColumn[*FarmFieldsModel]{
	DisplayName: "Location",
	Path:        "location",
	Expansion:   "",
	Type:        AppColTypePoint{},
	GetData: func(r *FarmFieldsModel) string {
		v := r.Location
		return fmt.Sprintf("[%f, %f]", v.Coordinates[0], v.Coordinates[1])
	},
}

var FarmFieldsModifiedAtColumn = TableColumn[*FarmFieldsModel]{
	DisplayName: "Modified At",
	Path:        "modified_at",
	Expansion:   "",
	Type:        AppColTypeTime{},
	GetData: func(r *FarmFieldsModel) string {
		v := r.ModifiedAt
		return v.Format("2 Jan 06 15:04")
	},
}

var FarmFieldsModifiedByIdColumn = TableColumn[*FarmFieldsModel]{
	DisplayName: "Modified By Id",
	Path:        "modified_by_id",
	Expansion:   "expand=modified_by_id(select=display_name)",
	Type: AppColTypeLookup{
		TableName: "users",
	},
	GetData: func(r *FarmFieldsModel) string {
		intermediateVal1 := r.ModifiedBy
		if intermediateVal1 == nil {
			return "-"
		}

		v := intermediateVal1.DisplayName
		return v
	},
}

var FarmFieldsReferenceColumn = TableColumn[*FarmFieldsModel]{
	DisplayName: "Reference",
	Path:        "reference",
	Expansion:   "",
	Type: AppColTypeString{
		MaxLength: int16(255),
	},
	GetData: func(r *FarmFieldsModel) string {
		v := r.Reference
		return v
	},
}

var UsersTable = TableDefinition[*UsersModel]{
	DisplayName:       "User",
	PluralDisplayName: "Users",
	MainGridColumns: []TableColumn[*UsersModel]{
		UsersCreatedAtColumn,
		UsersDisplayNameColumn,
		UsersEmailAddressColumn,
		UsersFamilyNameColumn,
		UsersGivenNameColumn,
		UsersModifiedAtColumn,
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

var UsersOidColumn = TableColumn[*UsersModel]{
	DisplayName: "Oid",
	Path:        "oid",
	Expansion:   "",
	Type:        AppColTypeIdentifier{},
	GetData: func(r *UsersModel) string {
		v := r.Oid
		return v.String()
	},
}
