// GENERATED CODE
// see cmd/modelbuilder/
package model

import (
	"time"

	"github.com/a-h/templ"
	"github.com/turnerbenjamin/querystack/querymodel"
)

type AppInputType string
type AppDataType string

const (
	AppTypeString     AppInputType = "string"
	AppTypeInteger    AppInputType = "integer"
	AppTypeFloat      AppInputType = "float"
	AppTypeDate       AppInputType = "date"
	AppTypeEnum       AppInputType = "enum"
	AppTypeIdentifier AppInputType = "identifier"
	AppTypeLookup     AppInputType = "lookup"
	AppTypePoint      AppInputType = "point"
)

type AppColType interface {
	GetInputType() AppInputType
	GetDataType() AppDataType
}

type AppColTypeString struct {
	MaxLength int16
}

func (t AppColTypeString) GetInputType() AppInputType {
	return AppTypeString
}

func (t AppColTypeString) GetDataType() AppDataType {
	return AppDataType(AppTypeString)
}

type AppColTypeInteger struct{}

func (t AppColTypeInteger) GetInputType() AppInputType {
	return AppTypeInteger
}

func (t AppColTypeInteger) GetDataType() AppDataType {
	return AppDataType(AppTypeInteger)
}

type AppColTypeFloat struct{}

func (t AppColTypeFloat) GetInputType() AppInputType {
	return AppTypeFloat
}

func (t AppColTypeFloat) GetDataType() AppDataType {
	return AppDataType(AppTypeFloat)
}

type AppColTypeDate struct{}

func (t AppColTypeDate) GetInputType() AppInputType {
	return AppTypeDate
}

func (t AppColTypeDate) GetDataType() AppDataType {
	return AppDataType(AppTypeDate)
}

type AppColTypePoint struct{}

func (t AppColTypePoint) GetInputType() AppInputType {
	return AppTypePoint
}

func (t AppColTypePoint) GetDataType() AppDataType {
	return AppDataType(AppTypePoint)
}

type AppColTypeIdentifier struct{}

func (t AppColTypeIdentifier) GetInputType() AppInputType {
	return AppTypeIdentifier
}

func (t AppColTypeIdentifier) GetDataType() AppDataType {
	return AppDataType(AppTypeIdentifier)
}

type AppColTypeEnum struct {
	TypeName string
	Options  []EnumOption
}

func (t AppColTypeEnum) GetInputType() AppInputType {
	return AppTypeEnum
}

func (t AppColTypeEnum) GetDataType() AppDataType {
	return AppDataType(t.TypeName)
}

type AppColTypeLookup struct {
	TableName string
}

func (t AppColTypeLookup) GetInputType() AppInputType {
	return AppTypeLookup
}

func (t AppColTypeLookup) GetDataType() AppDataType {
	return AppDataType(AppTypeLookup)
}

type UiInputControlBuilder interface {
	StringInput(v *string) templ.Component
	DateInput(v *time.Time) templ.Component
	EnumInput(typeName string, options []EnumOption, v *uint8) templ.Component
	IdentifierInput(v *string) templ.Component
	LookupInput(tableName string, v *string) templ.Component
	IntegerInput(v *int64) templ.Component
	FloatInput(v *float64) templ.Component
	PointInput(v *querymodel.Point) templ.Component
}

type UiTableCellBuilder interface {
	StringCell(v *string) templ.Component
	DateCell(v *time.Time) templ.Component
	EnumCell(v *string) templ.Component
	IdentifierCell(v *string) templ.Component
	LookupCell(v *string) templ.Component
	IntegerCell(v *int64) templ.Component
	FloatCell(v *float64) templ.Component
	PointCell(v *querymodel.Point) templ.Component
}

type TableDefinition[T querymodel.TableModel] struct {
	DisplayName        string
	PluralDisplayName  string
	MainGridColumns    []TableColumn[T]
	MainGridColumnsMap map[string]TableColumn[T]
	EnumTypes          []AppColTypeEnum
}

type Expansion struct {
	Path   string
	Select string
}

type TableColumn[T querymodel.TableModel] struct {
	DisplayName string
	Type        AppColType

	Path      string
	Expansion *Expansion

	BuildInput     func(T, UiInputControlBuilder) templ.Component
	BuildTableCell func(T, UiTableCellBuilder) templ.Component
}

var BusinessesTable = TableDefinition[*BusinessesModel]{
	DisplayName:       "Business",
	PluralDisplayName: "Businesses",
	MainGridColumns: []TableColumn[*BusinessesModel]{
		BusinessesAddressLine1Column,
		BusinessesAddressLine2Column,
		BusinessesBusinessTypeColumn,
		BusinessesCphNumberColumn,
		BusinessesContactNumberColumn,
		BusinessesCountryColumn,
		BusinessesCountyColumn,
		BusinessesCreatedAtColumn,
		BusinessesCreatedByIdColumn,
		BusinessesCreatedByIdEmailAddressColumn,
		BusinessesDescriptionColumn,
		BusinessesEmailAddressColumn,
		BusinessesModifiedAtColumn,
		BusinessesModifiedByIdColumn,
		BusinessesPostcodeColumn,
		BusinessesReferenceColumn,
		BusinessesTownColumn,
		BusinessesTradingNameColumn,
	},
	MainGridColumnsMap: map[string]TableColumn[*BusinessesModel]{
		"address_line_1":              BusinessesAddressLine1Column,
		"address_line_2":              BusinessesAddressLine2Column,
		"business_type":               BusinessesBusinessTypeColumn,
		"cph_number":                  BusinessesCphNumberColumn,
		"contact_number":              BusinessesContactNumberColumn,
		"country":                     BusinessesCountryColumn,
		"county":                      BusinessesCountyColumn,
		"created_at":                  BusinessesCreatedAtColumn,
		"created_by_id":               BusinessesCreatedByIdColumn,
		"created_by_id/email_address": BusinessesCreatedByIdEmailAddressColumn,
		"description":                 BusinessesDescriptionColumn,
		"email_address":               BusinessesEmailAddressColumn,
		"modified_at":                 BusinessesModifiedAtColumn,
		"modified_by_id":              BusinessesModifiedByIdColumn,
		"postcode":                    BusinessesPostcodeColumn,
		"reference":                   BusinessesReferenceColumn,
		"town":                        BusinessesTownColumn,
		"trading_name":                BusinessesTradingNameColumn,
	},
	EnumTypes: []AppColTypeEnum{
		{
			TypeName: "EnumBusinessesBusinessType",
			Options:  EnumBusinessesBusinessTypeOptions,
		}},
}

var BusinessesAddressLine1Column = TableColumn[*BusinessesModel]{
	DisplayName: "Address Line 1",
	Path:        "address_line_1",
	Type: AppColTypeString{
		MaxLength: int16(255),
	},
	BuildInput: func(r *BusinessesModel, b UiInputControlBuilder) templ.Component {
		v := r.AddressLine1
		return b.StringInput(&v)
	},
	BuildTableCell: func(r *BusinessesModel, b UiTableCellBuilder) templ.Component {
		v := r.AddressLine1
		return b.StringCell(&v)
	},
}

var BusinessesAddressLine2Column = TableColumn[*BusinessesModel]{
	DisplayName: "Address Line 2",
	Path:        "address_line_2",
	Type: AppColTypeString{
		MaxLength: int16(255),
	},
	BuildInput: func(r *BusinessesModel, b UiInputControlBuilder) templ.Component {
		v := r.AddressLine2
		return b.StringInput(&v)
	},
	BuildTableCell: func(r *BusinessesModel, b UiTableCellBuilder) templ.Component {
		v := r.AddressLine2
		return b.StringCell(&v)
	},
}

var BusinessesBusinessTypeColumn = TableColumn[*BusinessesModel]{
	DisplayName: "Business Type",
	Path:        "business_type",
	Type: AppColTypeEnum{
		TypeName: "EnumBusinessesBusinessType",
		Options:  EnumBusinessesBusinessTypeOptions,
	},
	BuildInput: func(r *BusinessesModel, b UiInputControlBuilder) templ.Component {
		v := r.BusinessType
		_, exists := v.String()
		if !exists {
			return b.EnumInput("EnumBusinessesBusinessType", EnumBusinessesBusinessTypeOptions, nil)
		}
		cv := uint8(v)
		return b.EnumInput("EnumBusinessesBusinessType", EnumBusinessesBusinessTypeOptions, &cv)
	},
	BuildTableCell: func(r *BusinessesModel, b UiTableCellBuilder) templ.Component {
		v := r.BusinessType
		label, exists := v.String()
		if !exists {
			return b.EnumCell(nil)
		}
		return b.EnumCell(&label)
	},
}

var BusinessesCphNumberColumn = TableColumn[*BusinessesModel]{
	DisplayName: "CPH Number",
	Path:        "cph_number",
	Type: AppColTypeString{
		MaxLength: int16(11),
	},
	BuildInput: func(r *BusinessesModel, b UiInputControlBuilder) templ.Component {
		v := r.CphNumber
		return b.StringInput(&v)
	},
	BuildTableCell: func(r *BusinessesModel, b UiTableCellBuilder) templ.Component {
		v := r.CphNumber
		return b.StringCell(&v)
	},
}

var BusinessesContactNumberColumn = TableColumn[*BusinessesModel]{
	DisplayName: "Contact Number",
	Path:        "contact_number",
	Type: AppColTypeString{
		MaxLength: int16(50),
	},
	BuildInput: func(r *BusinessesModel, b UiInputControlBuilder) templ.Component {
		v := r.ContactNumber
		return b.StringInput(&v)
	},
	BuildTableCell: func(r *BusinessesModel, b UiTableCellBuilder) templ.Component {
		v := r.ContactNumber
		return b.StringCell(&v)
	},
}

var BusinessesCountryColumn = TableColumn[*BusinessesModel]{
	DisplayName: "Country",
	Path:        "country",
	Type: AppColTypeString{
		MaxLength: int16(100),
	},
	BuildInput: func(r *BusinessesModel, b UiInputControlBuilder) templ.Component {
		v := r.Country
		return b.StringInput(&v)
	},
	BuildTableCell: func(r *BusinessesModel, b UiTableCellBuilder) templ.Component {
		v := r.Country
		return b.StringCell(&v)
	},
}

var BusinessesCountyColumn = TableColumn[*BusinessesModel]{
	DisplayName: "County",
	Path:        "county",
	Type: AppColTypeString{
		MaxLength: int16(100),
	},
	BuildInput: func(r *BusinessesModel, b UiInputControlBuilder) templ.Component {
		v := r.County
		return b.StringInput(&v)
	},
	BuildTableCell: func(r *BusinessesModel, b UiTableCellBuilder) templ.Component {
		v := r.County
		return b.StringCell(&v)
	},
}

var BusinessesCreatedAtColumn = TableColumn[*BusinessesModel]{
	DisplayName: "Created At",
	Path:        "created_at",
	Type:        AppColTypeDate{},
	BuildInput: func(r *BusinessesModel, b UiInputControlBuilder) templ.Component {
		v := r.CreatedAt
		return b.DateInput(v)
	},
	BuildTableCell: func(r *BusinessesModel, b UiTableCellBuilder) templ.Component {
		v := r.CreatedAt
		if v == nil {
			return b.DateCell(nil)
		}
		return b.DateCell(v)
	},
}

var BusinessesCreatedByIdColumn = TableColumn[*BusinessesModel]{
	DisplayName: "Created By",
	Path:        "created_by_id",
	Expansion: &Expansion{
		Path:   "created_by_id",
		Select: "display_name",
	},
	Type: AppColTypeLookup{
		TableName: "users",
	},
	BuildInput: func(r *BusinessesModel, b UiInputControlBuilder) templ.Component {
		intermediateVal1 := r.CreatedBy
		if intermediateVal1 == nil {
			return b.StringInput(nil)
		}

		v := intermediateVal1.DisplayName
		return b.StringInput(&v)
	},
	BuildTableCell: func(r *BusinessesModel, b UiTableCellBuilder) templ.Component {
		intermediateVal1 := r.CreatedBy
		if intermediateVal1 == nil {
			return b.StringCell(nil)
		}

		v := intermediateVal1.DisplayName
		return b.StringCell(&v)
	},
}

var BusinessesCreatedByIdEmailAddressColumn = TableColumn[*BusinessesModel]{
	DisplayName: "Created By (Email Address)",
	Path:        "created_by_id/email_address",
	Expansion: &Expansion{
		Path:   "created_by_id",
		Select: "email_address",
	},
	Type: AppColTypeString{
		MaxLength: int16(320),
	},
	BuildInput: func(r *BusinessesModel, b UiInputControlBuilder) templ.Component {
		intermediateVal1 := r.CreatedBy
		if intermediateVal1 == nil {
			return b.StringInput(nil)
		}

		v := intermediateVal1.EmailAddress
		return b.StringInput(&v)
	},
	BuildTableCell: func(r *BusinessesModel, b UiTableCellBuilder) templ.Component {
		intermediateVal1 := r.CreatedBy
		if intermediateVal1 == nil {
			return b.StringCell(nil)
		}

		v := intermediateVal1.EmailAddress
		return b.StringCell(&v)
	},
}

var BusinessesDescriptionColumn = TableColumn[*BusinessesModel]{
	DisplayName: "Description",
	Path:        "description",
	Type: AppColTypeString{
		MaxLength: int16(4000),
	},
	BuildInput: func(r *BusinessesModel, b UiInputControlBuilder) templ.Component {
		v := r.Description
		return b.StringInput(&v)
	},
	BuildTableCell: func(r *BusinessesModel, b UiTableCellBuilder) templ.Component {
		v := r.Description
		return b.StringCell(&v)
	},
}

var BusinessesEmailAddressColumn = TableColumn[*BusinessesModel]{
	DisplayName: "Email Address",
	Path:        "email_address",
	Type: AppColTypeString{
		MaxLength: int16(320),
	},
	BuildInput: func(r *BusinessesModel, b UiInputControlBuilder) templ.Component {
		v := r.EmailAddress
		return b.StringInput(&v)
	},
	BuildTableCell: func(r *BusinessesModel, b UiTableCellBuilder) templ.Component {
		v := r.EmailAddress
		return b.StringCell(&v)
	},
}

var BusinessesIdColumn = TableColumn[*BusinessesModel]{
	DisplayName: "Id",
	Path:        "id",
	Type:        AppColTypeIdentifier{},
	BuildInput: func(r *BusinessesModel, b UiInputControlBuilder) templ.Component {
		v := r.Id
		s := v.String()
		return b.IdentifierInput(&s)
	},
	BuildTableCell: func(r *BusinessesModel, b UiTableCellBuilder) templ.Component {
		v := r.Id
		s := v.String()
		return b.IdentifierCell(&s)
	},
}

var BusinessesLocationColumn = TableColumn[*BusinessesModel]{
	DisplayName: "Location",
	Path:        "location",
	Type:        AppColTypePoint{},
	BuildInput: func(r *BusinessesModel, b UiInputControlBuilder) templ.Component {
		v := r.Location
		return b.PointInput(v)
	},
	BuildTableCell: func(r *BusinessesModel, b UiTableCellBuilder) templ.Component {
		v := r.Location
		return b.PointCell(v)
	},
}

var BusinessesLogoUrlColumn = TableColumn[*BusinessesModel]{
	DisplayName: "Logo Url",
	Path:        "logo_url",
	Type: AppColTypeString{
		MaxLength: int16(2048),
	},
	BuildInput: func(r *BusinessesModel, b UiInputControlBuilder) templ.Component {
		v := r.LogoUrl
		return b.StringInput(&v)
	},
	BuildTableCell: func(r *BusinessesModel, b UiTableCellBuilder) templ.Component {
		v := r.LogoUrl
		return b.StringCell(&v)
	},
}

var BusinessesModifiedAtColumn = TableColumn[*BusinessesModel]{
	DisplayName: "Modified At",
	Path:        "modified_at",
	Type:        AppColTypeDate{},
	BuildInput: func(r *BusinessesModel, b UiInputControlBuilder) templ.Component {
		v := r.ModifiedAt
		return b.DateInput(v)
	},
	BuildTableCell: func(r *BusinessesModel, b UiTableCellBuilder) templ.Component {
		v := r.ModifiedAt
		if v == nil {
			return b.DateCell(nil)
		}
		return b.DateCell(v)
	},
}

var BusinessesModifiedByIdColumn = TableColumn[*BusinessesModel]{
	DisplayName: "Modified By",
	Path:        "modified_by_id",
	Expansion: &Expansion{
		Path:   "modified_by_id",
		Select: "display_name",
	},
	Type: AppColTypeLookup{
		TableName: "users",
	},
	BuildInput: func(r *BusinessesModel, b UiInputControlBuilder) templ.Component {
		intermediateVal1 := r.ModifiedBy
		if intermediateVal1 == nil {
			return b.StringInput(nil)
		}

		v := intermediateVal1.DisplayName
		return b.StringInput(&v)
	},
	BuildTableCell: func(r *BusinessesModel, b UiTableCellBuilder) templ.Component {
		intermediateVal1 := r.ModifiedBy
		if intermediateVal1 == nil {
			return b.StringCell(nil)
		}

		v := intermediateVal1.DisplayName
		return b.StringCell(&v)
	},
}

var BusinessesPostcodeColumn = TableColumn[*BusinessesModel]{
	DisplayName: "Postcode",
	Path:        "postcode",
	Type: AppColTypeString{
		MaxLength: int16(20),
	},
	BuildInput: func(r *BusinessesModel, b UiInputControlBuilder) templ.Component {
		v := r.Postcode
		return b.StringInput(&v)
	},
	BuildTableCell: func(r *BusinessesModel, b UiTableCellBuilder) templ.Component {
		v := r.Postcode
		return b.StringCell(&v)
	},
}

var BusinessesReferenceColumn = TableColumn[*BusinessesModel]{
	DisplayName: "Reference",
	Path:        "reference",
	Type: AppColTypeString{
		MaxLength: int16(8),
	},
	BuildInput: func(r *BusinessesModel, b UiInputControlBuilder) templ.Component {
		v := r.Reference
		return b.StringInput(&v)
	},
	BuildTableCell: func(r *BusinessesModel, b UiTableCellBuilder) templ.Component {
		v := r.Reference
		return b.StringCell(&v)
	},
}

var BusinessesTownColumn = TableColumn[*BusinessesModel]{
	DisplayName: "Town",
	Path:        "town",
	Type: AppColTypeString{
		MaxLength: int16(100),
	},
	BuildInput: func(r *BusinessesModel, b UiInputControlBuilder) templ.Component {
		v := r.Town
		return b.StringInput(&v)
	},
	BuildTableCell: func(r *BusinessesModel, b UiTableCellBuilder) templ.Component {
		v := r.Town
		return b.StringCell(&v)
	},
}

var BusinessesTradingNameColumn = TableColumn[*BusinessesModel]{
	DisplayName: "Trading Name",
	Path:        "trading_name",
	Type: AppColTypeString{
		MaxLength: int16(255),
	},
	BuildInput: func(r *BusinessesModel, b UiInputControlBuilder) templ.Component {
		v := r.TradingName
		return b.StringInput(&v)
	},
	BuildTableCell: func(r *BusinessesModel, b UiTableCellBuilder) templ.Component {
		v := r.TradingName
		return b.StringCell(&v)
	},
}

var BusinessesWebsiteUrlColumn = TableColumn[*BusinessesModel]{
	DisplayName: "Website Url",
	Path:        "website_url",
	Type: AppColTypeString{
		MaxLength: int16(2048),
	},
	BuildInput: func(r *BusinessesModel, b UiInputControlBuilder) templ.Component {
		v := r.WebsiteUrl
		return b.StringInput(&v)
	},
	BuildTableCell: func(r *BusinessesModel, b UiTableCellBuilder) templ.Component {
		v := r.WebsiteUrl
		return b.StringCell(&v)
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
	MainGridColumnsMap: map[string]TableColumn[*FarmFieldsModel]{
		"business_id":    FarmFieldsBusinessIdColumn,
		"created_at":     FarmFieldsCreatedAtColumn,
		"created_by_id":  FarmFieldsCreatedByIdColumn,
		"modified_at":    FarmFieldsModifiedAtColumn,
		"modified_by_id": FarmFieldsModifiedByIdColumn,
		"reference":      FarmFieldsReferenceColumn,
	},
	EnumTypes: []AppColTypeEnum{},
}

var FarmFieldsBusinessIdColumn = TableColumn[*FarmFieldsModel]{
	DisplayName: "Business",
	Path:        "business_id",
	Expansion: &Expansion{
		Path:   "business_id",
		Select: "trading_name",
	},
	Type: AppColTypeLookup{
		TableName: "businesses",
	},
	BuildInput: func(r *FarmFieldsModel, b UiInputControlBuilder) templ.Component {
		intermediateVal1 := r.Business
		if intermediateVal1 == nil {
			return b.StringInput(nil)
		}

		v := intermediateVal1.TradingName
		return b.StringInput(&v)
	},
	BuildTableCell: func(r *FarmFieldsModel, b UiTableCellBuilder) templ.Component {
		intermediateVal1 := r.Business
		if intermediateVal1 == nil {
			return b.StringCell(nil)
		}

		v := intermediateVal1.TradingName
		return b.StringCell(&v)
	},
}

var FarmFieldsCreatedAtColumn = TableColumn[*FarmFieldsModel]{
	DisplayName: "Created At",
	Path:        "created_at",
	Type:        AppColTypeDate{},
	BuildInput: func(r *FarmFieldsModel, b UiInputControlBuilder) templ.Component {
		v := r.CreatedAt
		return b.DateInput(v)
	},
	BuildTableCell: func(r *FarmFieldsModel, b UiTableCellBuilder) templ.Component {
		v := r.CreatedAt
		if v == nil {
			return b.DateCell(nil)
		}
		return b.DateCell(v)
	},
}

var FarmFieldsCreatedByIdColumn = TableColumn[*FarmFieldsModel]{
	DisplayName: "Created By",
	Path:        "created_by_id",
	Expansion: &Expansion{
		Path:   "created_by_id",
		Select: "display_name",
	},
	Type: AppColTypeLookup{
		TableName: "users",
	},
	BuildInput: func(r *FarmFieldsModel, b UiInputControlBuilder) templ.Component {
		intermediateVal1 := r.CreatedBy
		if intermediateVal1 == nil {
			return b.StringInput(nil)
		}

		v := intermediateVal1.DisplayName
		return b.StringInput(&v)
	},
	BuildTableCell: func(r *FarmFieldsModel, b UiTableCellBuilder) templ.Component {
		intermediateVal1 := r.CreatedBy
		if intermediateVal1 == nil {
			return b.StringCell(nil)
		}

		v := intermediateVal1.DisplayName
		return b.StringCell(&v)
	},
}

var FarmFieldsIdColumn = TableColumn[*FarmFieldsModel]{
	DisplayName: "Id",
	Path:        "id",
	Type:        AppColTypeIdentifier{},
	BuildInput: func(r *FarmFieldsModel, b UiInputControlBuilder) templ.Component {
		v := r.Id
		s := v.String()
		return b.IdentifierInput(&s)
	},
	BuildTableCell: func(r *FarmFieldsModel, b UiTableCellBuilder) templ.Component {
		v := r.Id
		s := v.String()
		return b.IdentifierCell(&s)
	},
}

var FarmFieldsLocationColumn = TableColumn[*FarmFieldsModel]{
	DisplayName: "Location",
	Path:        "location",
	Type:        AppColTypePoint{},
	BuildInput: func(r *FarmFieldsModel, b UiInputControlBuilder) templ.Component {
		v := r.Location
		return b.PointInput(v)
	},
	BuildTableCell: func(r *FarmFieldsModel, b UiTableCellBuilder) templ.Component {
		v := r.Location
		return b.PointCell(v)
	},
}

var FarmFieldsModifiedAtColumn = TableColumn[*FarmFieldsModel]{
	DisplayName: "Modified At",
	Path:        "modified_at",
	Type:        AppColTypeDate{},
	BuildInput: func(r *FarmFieldsModel, b UiInputControlBuilder) templ.Component {
		v := r.ModifiedAt
		return b.DateInput(v)
	},
	BuildTableCell: func(r *FarmFieldsModel, b UiTableCellBuilder) templ.Component {
		v := r.ModifiedAt
		if v == nil {
			return b.DateCell(nil)
		}
		return b.DateCell(v)
	},
}

var FarmFieldsModifiedByIdColumn = TableColumn[*FarmFieldsModel]{
	DisplayName: "Modified By",
	Path:        "modified_by_id",
	Expansion: &Expansion{
		Path:   "modified_by_id",
		Select: "display_name",
	},
	Type: AppColTypeLookup{
		TableName: "users",
	},
	BuildInput: func(r *FarmFieldsModel, b UiInputControlBuilder) templ.Component {
		intermediateVal1 := r.ModifiedBy
		if intermediateVal1 == nil {
			return b.StringInput(nil)
		}

		v := intermediateVal1.DisplayName
		return b.StringInput(&v)
	},
	BuildTableCell: func(r *FarmFieldsModel, b UiTableCellBuilder) templ.Component {
		intermediateVal1 := r.ModifiedBy
		if intermediateVal1 == nil {
			return b.StringCell(nil)
		}

		v := intermediateVal1.DisplayName
		return b.StringCell(&v)
	},
}

var FarmFieldsReferenceColumn = TableColumn[*FarmFieldsModel]{
	DisplayName: "Reference",
	Path:        "reference",
	Type: AppColTypeString{
		MaxLength: int16(255),
	},
	BuildInput: func(r *FarmFieldsModel, b UiInputControlBuilder) templ.Component {
		v := r.Reference
		return b.StringInput(&v)
	},
	BuildTableCell: func(r *FarmFieldsModel, b UiTableCellBuilder) templ.Component {
		v := r.Reference
		return b.StringCell(&v)
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
	MainGridColumnsMap: map[string]TableColumn[*UsersModel]{
		"created_at":    UsersCreatedAtColumn,
		"display_name":  UsersDisplayNameColumn,
		"email_address": UsersEmailAddressColumn,
		"family_name":   UsersFamilyNameColumn,
		"given_name":    UsersGivenNameColumn,
		"modified_at":   UsersModifiedAtColumn,
	},
	EnumTypes: []AppColTypeEnum{},
}

var UsersCreatedAtColumn = TableColumn[*UsersModel]{
	DisplayName: "Created At",
	Path:        "created_at",
	Type:        AppColTypeDate{},
	BuildInput: func(r *UsersModel, b UiInputControlBuilder) templ.Component {
		v := r.CreatedAt
		return b.DateInput(v)
	},
	BuildTableCell: func(r *UsersModel, b UiTableCellBuilder) templ.Component {
		v := r.CreatedAt
		if v == nil {
			return b.DateCell(nil)
		}
		return b.DateCell(v)
	},
}

var UsersDisplayNameColumn = TableColumn[*UsersModel]{
	DisplayName: "Display Name",
	Path:        "display_name",
	Type: AppColTypeString{
		MaxLength: int16(128),
	},
	BuildInput: func(r *UsersModel, b UiInputControlBuilder) templ.Component {
		v := r.DisplayName
		return b.StringInput(&v)
	},
	BuildTableCell: func(r *UsersModel, b UiTableCellBuilder) templ.Component {
		v := r.DisplayName
		return b.StringCell(&v)
	},
}

var UsersEmailAddressColumn = TableColumn[*UsersModel]{
	DisplayName: "Email Address",
	Path:        "email_address",
	Type: AppColTypeString{
		MaxLength: int16(320),
	},
	BuildInput: func(r *UsersModel, b UiInputControlBuilder) templ.Component {
		v := r.EmailAddress
		return b.StringInput(&v)
	},
	BuildTableCell: func(r *UsersModel, b UiTableCellBuilder) templ.Component {
		v := r.EmailAddress
		return b.StringCell(&v)
	},
}

var UsersFamilyNameColumn = TableColumn[*UsersModel]{
	DisplayName: "Family Name",
	Path:        "family_name",
	Type: AppColTypeString{
		MaxLength: int16(64),
	},
	BuildInput: func(r *UsersModel, b UiInputControlBuilder) templ.Component {
		v := r.FamilyName
		return b.StringInput(&v)
	},
	BuildTableCell: func(r *UsersModel, b UiTableCellBuilder) templ.Component {
		v := r.FamilyName
		return b.StringCell(&v)
	},
}

var UsersGivenNameColumn = TableColumn[*UsersModel]{
	DisplayName: "Given Name",
	Path:        "given_name",
	Type: AppColTypeString{
		MaxLength: int16(64),
	},
	BuildInput: func(r *UsersModel, b UiInputControlBuilder) templ.Component {
		v := r.GivenName
		return b.StringInput(&v)
	},
	BuildTableCell: func(r *UsersModel, b UiTableCellBuilder) templ.Component {
		v := r.GivenName
		return b.StringCell(&v)
	},
}

var UsersIdColumn = TableColumn[*UsersModel]{
	DisplayName: "Id",
	Path:        "id",
	Type:        AppColTypeIdentifier{},
	BuildInput: func(r *UsersModel, b UiInputControlBuilder) templ.Component {
		v := r.Id
		s := v.String()
		return b.IdentifierInput(&s)
	},
	BuildTableCell: func(r *UsersModel, b UiTableCellBuilder) templ.Component {
		v := r.Id
		s := v.String()
		return b.IdentifierCell(&s)
	},
}

var UsersModifiedAtColumn = TableColumn[*UsersModel]{
	DisplayName: "Modified At",
	Path:        "modified_at",
	Type:        AppColTypeDate{},
	BuildInput: func(r *UsersModel, b UiInputControlBuilder) templ.Component {
		v := r.ModifiedAt
		return b.DateInput(v)
	},
	BuildTableCell: func(r *UsersModel, b UiTableCellBuilder) templ.Component {
		v := r.ModifiedAt
		if v == nil {
			return b.DateCell(nil)
		}
		return b.DateCell(v)
	},
}

var UsersOidColumn = TableColumn[*UsersModel]{
	DisplayName: "Oid",
	Path:        "oid",
	Type:        AppColTypeIdentifier{},
	BuildInput: func(r *UsersModel, b UiInputControlBuilder) templ.Component {
		v := r.Oid
		s := v.String()
		return b.IdentifierInput(&s)
	},
	BuildTableCell: func(r *UsersModel, b UiTableCellBuilder) templ.Component {
		v := r.Oid
		s := v.String()
		return b.IdentifierCell(&s)
	},
}
