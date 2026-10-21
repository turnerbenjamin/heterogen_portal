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
	ColBusinessesId            = "id"
	ColBusinessesReference     = "reference"
	ColBusinessesTradingName   = "trading_name"
	ColBusinessesLogoUrl       = "logo_url"
	ColBusinessesDescription   = "description"
	ColBusinessesBusinessType  = "business_type"
	ColBusinessesCphNumber     = "cph_number"
	ColBusinessesEmailAddress  = "email_address"
	ColBusinessesContactNumber = "contact_number"
	ColBusinessesWebsiteUrl    = "website_url"
	ColBusinessesAddressLine1  = "address_line_1"
	ColBusinessesAddressLine2  = "address_line_2"
	ColBusinessesTown          = "town"
	ColBusinessesCounty        = "county"
	ColBusinessesCountry       = "country"
	ColBusinessesPostcode      = "postcode"
	ColBusinessesLocation      = "location"
	ColBusinessesCreatedAt     = "created_at"
	ColBusinessesCreatedById   = "created_by_id"
	ColBusinessesModifiedAt    = "modified_at"
	ColBusinessesModifiedById  = "modified_by_id"

	ColFarmFieldsId           = "id"
	ColFarmFieldsReference    = "reference"
	ColFarmFieldsBusinessId   = "business_id"
	ColFarmFieldsLocation     = "location"
	ColFarmFieldsCreatedAt    = "created_at"
	ColFarmFieldsCreatedById  = "created_by_id"
	ColFarmFieldsModifiedAt   = "modified_at"
	ColFarmFieldsModifiedById = "modified_by_id"

	ColUsersId           = "id"
	ColUsersOid          = "oid"
	ColUsersGivenName    = "given_name"
	ColUsersFamilyName   = "family_name"
	ColUsersDisplayName  = "display_name"
	ColUsersEmailAddress = "email_address"
	ColUsersCreatedAt    = "created_at"
	ColUsersModifiedAt   = "modified_at"
)

// Constants representing column data types
const (
	ColTypeBusinessesId            = DataTypeString
	ColTypeBusinessesReference     = DataTypeString
	ColTypeBusinessesTradingName   = DataTypeString
	ColTypeBusinessesLogoUrl       = DataTypeString
	ColTypeBusinessesDescription   = DataTypeString
	ColTypeBusinessesBusinessType  = DataTypeEnum
	ColTypeBusinessesCphNumber     = DataTypeString
	ColTypeBusinessesEmailAddress  = DataTypeString
	ColTypeBusinessesContactNumber = DataTypeString
	ColTypeBusinessesWebsiteUrl    = DataTypeString
	ColTypeBusinessesAddressLine1  = DataTypeString
	ColTypeBusinessesAddressLine2  = DataTypeString
	ColTypeBusinessesTown          = DataTypeString
	ColTypeBusinessesCounty        = DataTypeString
	ColTypeBusinessesCountry       = DataTypeString
	ColTypeBusinessesPostcode      = DataTypeString
	ColTypeBusinessesLocation      = DataTypePoint
	ColTypeBusinessesCreatedAt     = DataTypeTime
	ColTypeBusinessesCreatedById   = DataTypeString
	ColTypeBusinessesModifiedAt    = DataTypeTime
	ColTypeBusinessesModifiedById  = DataTypeString

	ColTypeFarmFieldsId           = DataTypeString
	ColTypeFarmFieldsReference    = DataTypeString
	ColTypeFarmFieldsBusinessId   = DataTypeString
	ColTypeFarmFieldsLocation     = DataTypePoint
	ColTypeFarmFieldsCreatedAt    = DataTypeTime
	ColTypeFarmFieldsCreatedById  = DataTypeString
	ColTypeFarmFieldsModifiedAt   = DataTypeTime
	ColTypeFarmFieldsModifiedById = DataTypeString

	ColTypeUsersId           = DataTypeString
	ColTypeUsersOid          = DataTypeString
	ColTypeUsersGivenName    = DataTypeString
	ColTypeUsersFamilyName   = DataTypeString
	ColTypeUsersDisplayName  = DataTypeString
	ColTypeUsersEmailAddress = DataTypeString
	ColTypeUsersCreatedAt    = DataTypeTime
	ColTypeUsersModifiedAt   = DataTypeTime
)

// Constants representing column display names in the database
const (
	ColDisplayNameBusinessesId            = "Id"
	ColDisplayNameBusinessesReference     = "Reference"
	ColDisplayNameBusinessesTradingName   = "Trading Name"
	ColDisplayNameBusinessesLogoUrl       = "Logo Url"
	ColDisplayNameBusinessesDescription   = "Description"
	ColDisplayNameBusinessesBusinessType  = "Business Type"
	ColDisplayNameBusinessesCphNumber     = "Cph Number"
	ColDisplayNameBusinessesEmailAddress  = "Email Address"
	ColDisplayNameBusinessesContactNumber = "Contact Number"
	ColDisplayNameBusinessesWebsiteUrl    = "Website Url"
	ColDisplayNameBusinessesAddressLine1  = "Address Line 1"
	ColDisplayNameBusinessesAddressLine2  = "Address Line 2"
	ColDisplayNameBusinessesTown          = "Town"
	ColDisplayNameBusinessesCounty        = "County"
	ColDisplayNameBusinessesCountry       = "Country"
	ColDisplayNameBusinessesPostcode      = "Postcode"
	ColDisplayNameBusinessesLocation      = "Location"
	ColDisplayNameBusinessesCreatedAt     = "Created At"
	ColDisplayNameBusinessesCreatedById   = "Created By Id"
	ColDisplayNameBusinessesModifiedAt    = "Modified At"
	ColDisplayNameBusinessesModifiedById  = "Modified By Id"

	ColDisplayNameFarmFieldsId           = "Id"
	ColDisplayNameFarmFieldsReference    = "Reference"
	ColDisplayNameFarmFieldsBusinessId   = "Business Id"
	ColDisplayNameFarmFieldsLocation     = "Location"
	ColDisplayNameFarmFieldsCreatedAt    = "Created At"
	ColDisplayNameFarmFieldsCreatedById  = "Created By Id"
	ColDisplayNameFarmFieldsModifiedAt   = "Modified At"
	ColDisplayNameFarmFieldsModifiedById = "Modified By Id"

	ColDisplayNameUsersId           = "Id"
	ColDisplayNameUsersOid          = "Oid"
	ColDisplayNameUsersGivenName    = "Given Name"
	ColDisplayNameUsersFamilyName   = "Family Name"
	ColDisplayNameUsersDisplayName  = "Display Name"
	ColDisplayNameUsersEmailAddress = "Email Address"
	ColDisplayNameUsersCreatedAt    = "Created At"
	ColDisplayNameUsersModifiedAt   = "Modified At"
)

// Constants representing relationship columns names
const (
	RelColBusinessesModifiedById = "modified_by_id"
	RelColBusinessesCreatedById  = "created_by_id"

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
	MaxLenBusinessesId            = 72
	MaxLenBusinessesReference     = 16
	MaxLenBusinessesTradingName   = 510
	MaxLenBusinessesLogoUrl       = 4096
	MaxLenBusinessesDescription   = 8000
	MaxLenBusinessesCphNumber     = 22
	MaxLenBusinessesEmailAddress  = 640
	MaxLenBusinessesContactNumber = 100
	MaxLenBusinessesWebsiteUrl    = 4096
	MaxLenBusinessesAddressLine1  = 510
	MaxLenBusinessesAddressLine2  = 510
	MaxLenBusinessesTown          = 200
	MaxLenBusinessesCounty        = 200
	MaxLenBusinessesCountry       = 200
	MaxLenBusinessesPostcode      = 40
	MaxLenBusinessesCreatedById   = 72
	MaxLenBusinessesModifiedById  = 72
	MaxLenFarmFieldsId            = 72
	MaxLenFarmFieldsReference     = 510
	MaxLenFarmFieldsBusinessId    = 72
	MaxLenFarmFieldsCreatedById   = 72
	MaxLenFarmFieldsModifiedById  = 72
	MaxLenUsersId                 = 72
	MaxLenUsersOid                = 72
	MaxLenUsersGivenName          = 128
	MaxLenUsersFamilyName         = 128
	MaxLenUsersDisplayName        = 256
	MaxLenUsersEmailAddress       = 640
)
