// GENERATED CODE
// see cmd/modelbuilder/
package model

type DataType string

const (
	DataTypeString DataType = "string"
	DataTypeNumber DataType = "number"
	DataTypeEnum   DataType = "enum"
	DataTypeTime   DataType = "time"
	DataTypePoint  DataType = "point"
)

// Constants representing column names in the database
const (
	ColBusinessesAddressLine1  = "address_line_1"
	ColBusinessesAddressLine2  = "address_line_2"
	ColBusinessesBusinessType  = "business_type"
	ColBusinessesContactNumber = "contact_number"
	ColBusinessesCountry       = "country"
	ColBusinessesCounty        = "county"
	ColBusinessesCphNumber     = "cph_number"
	ColBusinessesCreatedAt     = "created_at"
	ColBusinessesCreatedById   = "created_by_id"
	ColBusinessesDescription   = "description"
	ColBusinessesEmailAddress  = "email_address"
	ColBusinessesId            = "id"
	ColBusinessesLocation      = "location"
	ColBusinessesLogoUrl       = "logo_url"
	ColBusinessesModifiedAt    = "modified_at"
	ColBusinessesModifiedById  = "modified_by_id"
	ColBusinessesPostcode      = "postcode"
	ColBusinessesReference     = "reference"
	ColBusinessesTown          = "town"
	ColBusinessesTradingName   = "trading_name"
	ColBusinessesWebsiteUrl    = "website_url"

	ColFarmFieldsBusinessId   = "business_id"
	ColFarmFieldsCreatedAt    = "created_at"
	ColFarmFieldsCreatedById  = "created_by_id"
	ColFarmFieldsId           = "id"
	ColFarmFieldsLocation     = "location"
	ColFarmFieldsModifiedAt   = "modified_at"
	ColFarmFieldsModifiedById = "modified_by_id"
	ColFarmFieldsReference    = "reference"

	ColUsersCreatedAt    = "created_at"
	ColUsersDisplayName  = "display_name"
	ColUsersEmailAddress = "email_address"
	ColUsersFamilyName   = "family_name"
	ColUsersGivenName    = "given_name"
	ColUsersId           = "id"
	ColUsersModifiedAt   = "modified_at"
	ColUsersOid          = "oid"
)

// Constants representing column display names in the database
const (
	ColDisplayNameBusinessesAddressLine1  = "Address Line 1"
	ColDisplayNameBusinessesAddressLine2  = "Address Line 2"
	ColDisplayNameBusinessesBusinessType  = "Business Type"
	ColDisplayNameBusinessesContactNumber = "Contact Number"
	ColDisplayNameBusinessesCountry       = "Country"
	ColDisplayNameBusinessesCounty        = "County"
	ColDisplayNameBusinessesCphNumber     = "Cph Number"
	ColDisplayNameBusinessesCreatedAt     = "Created At"
	ColDisplayNameBusinessesCreatedById   = "Created By Id"
	ColDisplayNameBusinessesDescription   = "Description"
	ColDisplayNameBusinessesEmailAddress  = "Email Address"
	ColDisplayNameBusinessesId            = "Id"
	ColDisplayNameBusinessesLocation      = "Location"
	ColDisplayNameBusinessesLogoUrl       = "Logo Url"
	ColDisplayNameBusinessesModifiedAt    = "Modified At"
	ColDisplayNameBusinessesModifiedById  = "Modified By Id"
	ColDisplayNameBusinessesPostcode      = "Postcode"
	ColDisplayNameBusinessesReference     = "Reference"
	ColDisplayNameBusinessesTown          = "Town"
	ColDisplayNameBusinessesTradingName   = "Trading Name"
	ColDisplayNameBusinessesWebsiteUrl    = "Website Url"

	ColDisplayNameFarmFieldsBusinessId   = "Business Id"
	ColDisplayNameFarmFieldsCreatedAt    = "Created At"
	ColDisplayNameFarmFieldsCreatedById  = "Created By Id"
	ColDisplayNameFarmFieldsId           = "Id"
	ColDisplayNameFarmFieldsLocation     = "Location"
	ColDisplayNameFarmFieldsModifiedAt   = "Modified At"
	ColDisplayNameFarmFieldsModifiedById = "Modified By Id"
	ColDisplayNameFarmFieldsReference    = "Reference"

	ColDisplayNameUsersCreatedAt    = "Created At"
	ColDisplayNameUsersDisplayName  = "Display Name"
	ColDisplayNameUsersEmailAddress = "Email Address"
	ColDisplayNameUsersFamilyName   = "Family Name"
	ColDisplayNameUsersGivenName    = "Given Name"
	ColDisplayNameUsersId           = "Id"
	ColDisplayNameUsersModifiedAt   = "Modified At"
	ColDisplayNameUsersOid          = "Oid"
)

// Constants representing relationship columns names
const (
	RelColBusinessesCreatedById  = "created_by_id"
	RelColBusinessesModifiedById = "modified_by_id"

	RelColFarmFieldsBusinessId   = "business_id"
	RelColFarmFieldsCreatedById  = "created_by_id"
	RelColFarmFieldsModifiedById = "modified_by_id"
)

// Constants representing relationship columns display names
const (
	RelColDisplayNameBusinessesCreatedById  = "Created By"
	RelColDisplayNameBusinessesModifiedById = "Modified By"

	RelColDisplayNameFarmFieldsBusinessId   = "Business"
	RelColDisplayNameFarmFieldsCreatedById  = "Created By"
	RelColDisplayNameFarmFieldsModifiedById = "Modified By"
)

// Constants representing column max lengths
const (
	MaxLenBusinessesAddressLine1  = 510
	MaxLenBusinessesAddressLine2  = 510
	MaxLenBusinessesContactNumber = 100
	MaxLenBusinessesCountry       = 200
	MaxLenBusinessesCounty        = 200
	MaxLenBusinessesCphNumber     = 22
	MaxLenBusinessesCreatedById   = 72
	MaxLenBusinessesDescription   = 8000
	MaxLenBusinessesEmailAddress  = 640
	MaxLenBusinessesId            = 72
	MaxLenBusinessesLogoUrl       = 4096
	MaxLenBusinessesModifiedById  = 72
	MaxLenBusinessesPostcode      = 40
	MaxLenBusinessesReference     = 16
	MaxLenBusinessesTown          = 200
	MaxLenBusinessesTradingName   = 510
	MaxLenBusinessesWebsiteUrl    = 4096
	MaxLenFarmFieldsBusinessId    = 72
	MaxLenFarmFieldsCreatedById   = 72
	MaxLenFarmFieldsId            = 72
	MaxLenFarmFieldsModifiedById  = 72
	MaxLenFarmFieldsReference     = 510
	MaxLenUsersDisplayName        = 256
	MaxLenUsersEmailAddress       = 640
	MaxLenUsersFamilyName         = 128
	MaxLenUsersGivenName          = 128
	MaxLenUsersId                 = 72
	MaxLenUsersOid                = 72
)
