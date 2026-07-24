// GENERATED CODE
// see cmd/model-builder/
package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// DbDataTypeName represents a SQL Server data type name supported by the model metadata system.
type DbDataTypeName string

const (
	// DbTypeNvarchar represents the SQL Server nvarchar data type.
	DbTypeNvarchar DbDataTypeName = "nvarchar"

	// DbTypeInt represents the SQL Server int data type.
	DbTypeInt DbDataTypeName = "int"

	// DbTypeFloat represents the SQL Server float data type.
	DbTypeFloat DbDataTypeName = "float"

	// DbTypeGeography represents the SQL Server geography spatial data type.
	DbTypeGeography DbDataTypeName = "geography"

	// DbTypeDateTimeOffset represents the SQL Server datetimeoffset date/time data type.
	DbTypeDateTimeOffset DbDataTypeName = "datetimeoffset"
)

// relationshipId represents a unique identifier for a given relationship
type relationshipId string

// relationshipType represents different table relationships
type relationshipType string

const (
	// RelationshipOneToMany represents a 1:N relationship
	RelationshipOneToMany relationshipType = "1:N"

	// RelationshipManyToOne represents an N:1 relationship
	RelationshipManyToOne relationshipType = "N:1"
)

// SupportedDbTypes contains the SQL Server data types supported by the model
// generation and mapping system.
var SupportedDbTypes = map[DbDataTypeName]bool{
	DbTypeNvarchar:       true,
	DbTypeInt:            true,
	DbTypeFloat:          true,
	DbTypeGeography:      true,
	DbTypeDateTimeOffset: true,
}

// ColumnMetadata describes the metadata associated with a database table column.
type ColumnMetadata struct {
	Name         string
	Type         DbDataTypeName
	MaxLength    int
	IsPrimaryKey bool
	IsRequired   bool
}

// Relationship describes a foreign key relationship between two database columns.
type Relationship struct {
	Id                 relationshipId
	Type               relationshipType
	RelationshipColumn string
	RelatedTable       string
	LocalColumn        string
	ForeignColumn      string
}

// GetTo returns metadata for the related table if it exists else nil
func (r *Relationship) GetTo() *TableMetadata {
	return GetTableMetadata(r.RelatedTable)
}

// TableMetadata describes the structure and relationships of a database table.
type TableMetadata struct {
	TableName          string
	SchemaName         string
	FullyQualifiedName string
	PrimaryKey         string
	Columns            map[string]ColumnMetadata
	Relationships      map[string]Relationship
}

// GetResourceShortName returns the table name without its schema namespace
func (t *TableMetadata) GetResourceShortName() string {
	return t.TableName
}

// GetResourceFullname returns the table name with its schema namespace
func (t *TableMetadata) GetResourceFullname() string {
	return t.FullyQualifiedName
}

// GetColumn returns column metadata if a column exists on the table, else nil
func (t *TableMetadata) GetColumn(columnName string) *ColumnMetadata {
	c, exists := t.Columns[columnName]
	if !exists {
		return nil
	}
	return &c
}

// GetRelationship returns relationship metadata if it exists, else nil
func (t *TableMetadata) GetRelationship(relationshipName string) *Relationship {
	r, exists := t.Relationships[relationshipName]
	if !exists {
		return nil
	}
	return &r
}

// TableModel is an interface shared by all database table models
type TableModel interface {
	// GetMetadata returns metadata for the table
	GetMetadata() TableMetadata

	// GetTableName returns the name of the table in the database
	GetTableName() string

	// NewSlice unmarshals a json array and returns it as a slice
	NewSlice(jsonData []byte) ([]TableModel, error)

	// SetRelationshipField sets a given relationship field
	SetRelationshipField(relationshipId relationshipId, value TableModel) error

	// GetJoinOnValue returns the value of the relevant column for a given relationship
	GetJoinOnValue(relationshipId relationshipId) (string, error)
}

// Point represents a geographic point with a GeoJSON-compatible structure.
type Point struct {
	Type        string     `json:"type"`
	Coordinates [2]float64 `json:"coordinates"`
}

// UnmarshalJSON parses a point from WKT point data.
func (p *Point) UnmarshalJSON(data []byte) error {
	strData := strings.Trim(string(data), "\"")
	if strData == "null" || strData == "" {
		return nil
	}

	if !strings.HasPrefix(strData, "POINT") {
		return fmt.Errorf("invalid json for point: %s", strData)
	}

	start := strings.Index(strData, "(")
	end := strings.Index(strData, ")")
	if start == -1 || end == -1 || end <= start {
		return fmt.Errorf("invalid WKT point format: %s", strData)
	}

	coordsStr := strings.TrimSpace(strData[start+1 : end])
	parts := strings.Fields(coordsStr)
	if len(parts) != 2 {
		return fmt.Errorf("expected 2 coordinates in WKT point, got %d", len(parts))
	}

	long, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return fmt.Errorf("invalid longitude value: %w", err)
	}

	lat, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return fmt.Errorf("invalid latitude value: %w", err)
	}

	p.Type = "Point"
	p.Coordinates = [2]float64{long, lat}
	return nil
}

// Constants representing field names in the database
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
	ColFarmFieldsId            = "id"
	ColFarmFieldsReference     = "reference"
	ColFarmFieldsBusinessId    = "business_id"
	ColFarmFieldsLocation      = "location"
	ColFarmFieldsCreatedAt     = "created_at"
	ColFarmFieldsCreatedById   = "created_by_id"
	ColFarmFieldsModifiedAt    = "modified_at"
	ColFarmFieldsModifiedById  = "modified_by_id"
	ColUsersId                 = "id"
	ColUsersOid                = "oid"
	ColUsersGivenName          = "given_name"
	ColUsersFamilyName         = "family_name"
	ColUsersUserName           = "user_name"
	ColUsersEmailAddress       = "email_address"
	ColUsersCreatedAt          = "created_at"
	ColUsersModifiedAt         = "modified_at"
)

// Constants representing field names in the database
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
	MaxLenUsersUserName           = 256
	MaxLenUsersEmailAddress       = 640
)

// businessesMetadata contains the database metadata for the hg.businesses table.
var businessesMetadata = TableMetadata{
	TableName:          "businesses",
	SchemaName:         "hg",
	FullyQualifiedName: "hg.businesses",
	PrimaryKey:         "id",
	Columns: map[string]ColumnMetadata{
		"id": {
			Name:         "id",
			Type:         DbTypeNvarchar,
			MaxLength:    72,
			IsRequired:   true,
			IsPrimaryKey: true,
		},
		"reference": {
			Name:         "reference",
			Type:         DbTypeNvarchar,
			MaxLength:    16,
			IsRequired:   true,
			IsPrimaryKey: false,
		},
		"trading_name": {
			Name:         "trading_name",
			Type:         DbTypeNvarchar,
			MaxLength:    510,
			IsRequired:   true,
			IsPrimaryKey: false,
		},
		"logo_url": {
			Name:         "logo_url",
			Type:         DbTypeNvarchar,
			MaxLength:    4096,
			IsRequired:   false,
			IsPrimaryKey: false,
		},
		"description": {
			Name:         "description",
			Type:         DbTypeNvarchar,
			MaxLength:    8000,
			IsRequired:   false,
			IsPrimaryKey: false,
		},
		"business_type": {
			Name:         "business_type",
			Type:         DbTypeInt,
			MaxLength:    4,
			IsRequired:   true,
			IsPrimaryKey: false,
		},
		"cph_number": {
			Name:         "cph_number",
			Type:         DbTypeNvarchar,
			MaxLength:    22,
			IsRequired:   false,
			IsPrimaryKey: false,
		},
		"email_address": {
			Name:         "email_address",
			Type:         DbTypeNvarchar,
			MaxLength:    640,
			IsRequired:   false,
			IsPrimaryKey: false,
		},
		"contact_number": {
			Name:         "contact_number",
			Type:         DbTypeNvarchar,
			MaxLength:    100,
			IsRequired:   false,
			IsPrimaryKey: false,
		},
		"website_url": {
			Name:         "website_url",
			Type:         DbTypeNvarchar,
			MaxLength:    4096,
			IsRequired:   false,
			IsPrimaryKey: false,
		},
		"address_line_1": {
			Name:         "address_line_1",
			Type:         DbTypeNvarchar,
			MaxLength:    510,
			IsRequired:   true,
			IsPrimaryKey: false,
		},
		"address_line_2": {
			Name:         "address_line_2",
			Type:         DbTypeNvarchar,
			MaxLength:    510,
			IsRequired:   false,
			IsPrimaryKey: false,
		},
		"town": {
			Name:         "town",
			Type:         DbTypeNvarchar,
			MaxLength:    200,
			IsRequired:   false,
			IsPrimaryKey: false,
		},
		"county": {
			Name:         "county",
			Type:         DbTypeNvarchar,
			MaxLength:    200,
			IsRequired:   false,
			IsPrimaryKey: false,
		},
		"country": {
			Name:         "country",
			Type:         DbTypeNvarchar,
			MaxLength:    200,
			IsRequired:   false,
			IsPrimaryKey: false,
		},
		"postcode": {
			Name:         "postcode",
			Type:         DbTypeNvarchar,
			MaxLength:    40,
			IsRequired:   true,
			IsPrimaryKey: false,
		},
		"location": {
			Name:         "location",
			Type:         DbTypeGeography,
			MaxLength:    -1,
			IsRequired:   true,
			IsPrimaryKey: false,
		},
		"created_at": {
			Name:         "created_at",
			Type:         DbTypeDateTimeOffset,
			MaxLength:    10,
			IsRequired:   true,
			IsPrimaryKey: false,
		},
		"created_by_id": {
			Name:         "created_by_id",
			Type:         DbTypeNvarchar,
			MaxLength:    72,
			IsRequired:   true,
			IsPrimaryKey: false,
		},
		"modified_at": {
			Name:         "modified_at",
			Type:         DbTypeDateTimeOffset,
			MaxLength:    10,
			IsRequired:   true,
			IsPrimaryKey: false,
		},
		"modified_by_id": {
			Name:         "modified_by_id",
			Type:         DbTypeNvarchar,
			MaxLength:    72,
			IsRequired:   true,
			IsPrimaryKey: false,
		},
	},
	Relationships: map[string]Relationship{
		"created_by_id": {
			Id:                 relationshipId("businesses_created_by_id_users_id"),
			Type:               RelationshipManyToOne,
			RelationshipColumn: "created_by",
			RelatedTable:       "users",
			LocalColumn:        "created_by_id",
			ForeignColumn:      "id",
		},
		"modified_by_id": {
			Id:                 relationshipId("businesses_modified_by_id_users_id"),
			Type:               RelationshipManyToOne,
			RelationshipColumn: "modified_by",
			RelatedTable:       "users",
			LocalColumn:        "modified_by_id",
			ForeignColumn:      "id",
		},
		"farm_fields_businesses_business_id": {
			Id:                 relationshipId("farm_fields_business_id_businesses_id"),
			Type:               RelationshipOneToMany,
			RelationshipColumn: "farm_fields_businesses_business_id",
			RelatedTable:       "farm_fields",
			LocalColumn:        "id",
			ForeignColumn:      "business_id",
		},
	},
}

// GetBusinessesMetadata returns the database metadata associated with the businesses.Businesses table.
func GetBusinessesMetadata() TableMetadata {
	return businessesMetadata
}

// BusinessesModel represents a row from the hg.businesses table.
type BusinessesModel struct {
	Id                             string             `json:"id,omitempty"`
	Reference                      string             `json:"reference,omitempty"`
	TradingName                    string             `json:"trading_name,omitempty"`
	LogoUrl                        string             `json:"logo_url,omitempty"`
	Description                    string             `json:"description,omitempty"`
	BusinessType                   int                `json:"business_type,omitempty"`
	CphNumber                      string             `json:"cph_number,omitempty"`
	EmailAddress                   string             `json:"email_address,omitempty"`
	ContactNumber                  string             `json:"contact_number,omitempty"`
	WebsiteUrl                     string             `json:"website_url,omitempty"`
	AddressLine1                   string             `json:"address_line_1,omitempty"`
	AddressLine2                   string             `json:"address_line_2,omitempty"`
	Town                           string             `json:"town,omitempty"`
	County                         string             `json:"county,omitempty"`
	Country                        string             `json:"country,omitempty"`
	Postcode                       string             `json:"postcode,omitempty"`
	Location                       *Point             `json:"location,omitempty"`
	CreatedAt                      *time.Time         `json:"created_at,omitempty"`
	CreatedById                    string             `json:"created_by_id,omitempty"`
	ModifiedAt                     *time.Time         `json:"modified_at,omitempty"`
	ModifiedById                   string             `json:"modified_by_id,omitempty"`
	CreatedBy                      *UsersModel        `json:"created_by,omitempty"`
	ModifiedBy                     *UsersModel        `json:"modified_by,omitempty"`
	FarmFieldsBusinessesBusinessId []*FarmFieldsModel `json:"farm_fields_businesses_business_id,omitempty"`
}

// GetMetadata returns metadata for the hg.businesses table.
func (m *BusinessesModel) GetMetadata() TableMetadata {
	return businessesMetadata
}

// GetTableName returns the name of the hg.businesses table.
func (m *BusinessesModel) GetTableName() string {
	return businessesMetadata.TableName
}

// NewSlice unmarshals a json array of businesses and returns it as a slice
func (m *BusinessesModel) NewSlice(jsonData []byte) ([]TableModel, error) {
	var concreteSlice []*BusinessesModel
	if err := json.Unmarshal(jsonData, &concreteSlice); err != nil {
		return nil, err
	}
	result := make([]TableModel, len(concreteSlice))
	for i := range concreteSlice {
		result[i] = concreteSlice[i]
	}
	return result, nil
}

// SetRelationshipField sets a given relationship field on the hg.businesses table
func (m *BusinessesModel) SetRelationshipField(relationshipId relationshipId, value TableModel) error {
	switch relationshipId {
	case "businesses_created_by_id_users_id":
		v, ok := value.(*UsersModel)
		if !ok {
			return errors.New("unexpected relationship type received")
		}
		m.CreatedBy = v

	case "businesses_modified_by_id_users_id":
		v, ok := value.(*UsersModel)
		if !ok {
			return errors.New("unexpected relationship type received")
		}
		m.ModifiedBy = v

	case "farm_fields_business_id_businesses_id":
		v, ok := value.(*FarmFieldsModel)
		if !ok {
			return errors.New("unexpected relationship type received")
		}
		m.FarmFieldsBusinessesBusinessId = append(m.FarmFieldsBusinessesBusinessId, v)

	default:
		return fmt.Errorf("unknown relationship: %s", relationshipId)
	}
	return nil
}

// GetJoinOnValue returns the value of the relevant column for a given relationship
func (m *BusinessesModel) GetJoinOnValue(relationshipId relationshipId) (string, error) {
	switch relationshipId {
	case "businesses_created_by_id_users_id":
		return m.CreatedById, nil
	case "businesses_modified_by_id_users_id":
		return m.ModifiedById, nil
	case "farm_fields_business_id_businesses_id":
		return m.Id, nil
	default:
		return "", fmt.Errorf("unknown relationship: %s", relationshipId)
	}
}

// farmFieldsMetadata contains the database metadata for the hg.farm_fields table.
var farmFieldsMetadata = TableMetadata{
	TableName:          "farm_fields",
	SchemaName:         "hg",
	FullyQualifiedName: "hg.farm_fields",
	PrimaryKey:         "id",
	Columns: map[string]ColumnMetadata{
		"id": {
			Name:         "id",
			Type:         DbTypeNvarchar,
			MaxLength:    72,
			IsRequired:   true,
			IsPrimaryKey: true,
		},
		"reference": {
			Name:         "reference",
			Type:         DbTypeNvarchar,
			MaxLength:    510,
			IsRequired:   true,
			IsPrimaryKey: false,
		},
		"business_id": {
			Name:         "business_id",
			Type:         DbTypeNvarchar,
			MaxLength:    72,
			IsRequired:   true,
			IsPrimaryKey: false,
		},
		"location": {
			Name:         "location",
			Type:         DbTypeGeography,
			MaxLength:    -1,
			IsRequired:   true,
			IsPrimaryKey: false,
		},
		"created_at": {
			Name:         "created_at",
			Type:         DbTypeDateTimeOffset,
			MaxLength:    10,
			IsRequired:   true,
			IsPrimaryKey: false,
		},
		"created_by_id": {
			Name:         "created_by_id",
			Type:         DbTypeNvarchar,
			MaxLength:    72,
			IsRequired:   true,
			IsPrimaryKey: false,
		},
		"modified_at": {
			Name:         "modified_at",
			Type:         DbTypeDateTimeOffset,
			MaxLength:    10,
			IsRequired:   true,
			IsPrimaryKey: false,
		},
		"modified_by_id": {
			Name:         "modified_by_id",
			Type:         DbTypeNvarchar,
			MaxLength:    72,
			IsRequired:   true,
			IsPrimaryKey: false,
		},
	},
	Relationships: map[string]Relationship{
		"business_id": {
			Id:                 relationshipId("farm_fields_business_id_businesses_id"),
			Type:               RelationshipManyToOne,
			RelationshipColumn: "business",
			RelatedTable:       "businesses",
			LocalColumn:        "business_id",
			ForeignColumn:      "id",
		},
		"created_by_id": {
			Id:                 relationshipId("farm_fields_created_by_id_users_id"),
			Type:               RelationshipManyToOne,
			RelationshipColumn: "created_by",
			RelatedTable:       "users",
			LocalColumn:        "created_by_id",
			ForeignColumn:      "id",
		},
		"modified_by_id": {
			Id:                 relationshipId("farm_fields_modified_by_id_users_id"),
			Type:               RelationshipManyToOne,
			RelationshipColumn: "modified_by",
			RelatedTable:       "users",
			LocalColumn:        "modified_by_id",
			ForeignColumn:      "id",
		},
	},
}

// GetFarmFieldsMetadata returns the database metadata associated with the farm_fields.FarmFields table.
func GetFarmFieldsMetadata() TableMetadata {
	return farmFieldsMetadata
}

// FarmFieldsModel represents a row from the hg.farm_fields table.
type FarmFieldsModel struct {
	Id           string           `json:"id,omitempty"`
	Reference    string           `json:"reference,omitempty"`
	BusinessId   string           `json:"business_id,omitempty"`
	Location     *Point           `json:"location,omitempty"`
	CreatedAt    *time.Time       `json:"created_at,omitempty"`
	CreatedById  string           `json:"created_by_id,omitempty"`
	ModifiedAt   *time.Time       `json:"modified_at,omitempty"`
	ModifiedById string           `json:"modified_by_id,omitempty"`
	Business     *BusinessesModel `json:"business,omitempty"`
	CreatedBy    *UsersModel      `json:"created_by,omitempty"`
	ModifiedBy   *UsersModel      `json:"modified_by,omitempty"`
}

// GetMetadata returns metadata for the hg.farm_fields table.
func (m *FarmFieldsModel) GetMetadata() TableMetadata {
	return farmFieldsMetadata
}

// GetTableName returns the name of the hg.farm_fields table.
func (m *FarmFieldsModel) GetTableName() string {
	return farmFieldsMetadata.TableName
}

// NewSlice unmarshals a json array of farm_fields and returns it as a slice
func (m *FarmFieldsModel) NewSlice(jsonData []byte) ([]TableModel, error) {
	var concreteSlice []*FarmFieldsModel
	if err := json.Unmarshal(jsonData, &concreteSlice); err != nil {
		return nil, err
	}
	result := make([]TableModel, len(concreteSlice))
	for i := range concreteSlice {
		result[i] = concreteSlice[i]
	}
	return result, nil
}

// SetRelationshipField sets a given relationship field on the hg.farm_fields table
func (m *FarmFieldsModel) SetRelationshipField(relationshipId relationshipId, value TableModel) error {
	switch relationshipId {
	case "farm_fields_business_id_businesses_id":
		v, ok := value.(*BusinessesModel)
		if !ok {
			return errors.New("unexpected relationship type received")
		}
		m.Business = v

	case "farm_fields_created_by_id_users_id":
		v, ok := value.(*UsersModel)
		if !ok {
			return errors.New("unexpected relationship type received")
		}
		m.CreatedBy = v

	case "farm_fields_modified_by_id_users_id":
		v, ok := value.(*UsersModel)
		if !ok {
			return errors.New("unexpected relationship type received")
		}
		m.ModifiedBy = v

	default:
		return fmt.Errorf("unknown relationship: %s", relationshipId)
	}
	return nil
}

// GetJoinOnValue returns the value of the relevant column for a given relationship
func (m *FarmFieldsModel) GetJoinOnValue(relationshipId relationshipId) (string, error) {
	switch relationshipId {
	case "farm_fields_business_id_businesses_id":
		return m.BusinessId, nil
	case "farm_fields_created_by_id_users_id":
		return m.CreatedById, nil
	case "farm_fields_modified_by_id_users_id":
		return m.ModifiedById, nil
	default:
		return "", fmt.Errorf("unknown relationship: %s", relationshipId)
	}
}

// usersMetadata contains the database metadata for the hg.users table.
var usersMetadata = TableMetadata{
	TableName:          "users",
	SchemaName:         "hg",
	FullyQualifiedName: "hg.users",
	PrimaryKey:         "id",
	Columns: map[string]ColumnMetadata{
		"id": {
			Name:         "id",
			Type:         DbTypeNvarchar,
			MaxLength:    72,
			IsRequired:   true,
			IsPrimaryKey: true,
		},
		"oid": {
			Name:         "oid",
			Type:         DbTypeNvarchar,
			MaxLength:    72,
			IsRequired:   true,
			IsPrimaryKey: false,
		},
		"given_name": {
			Name:         "given_name",
			Type:         DbTypeNvarchar,
			MaxLength:    128,
			IsRequired:   true,
			IsPrimaryKey: false,
		},
		"family_name": {
			Name:         "family_name",
			Type:         DbTypeNvarchar,
			MaxLength:    128,
			IsRequired:   true,
			IsPrimaryKey: false,
		},
		"user_name": {
			Name:         "user_name",
			Type:         DbTypeNvarchar,
			MaxLength:    256,
			IsRequired:   true,
			IsPrimaryKey: false,
		},
		"email_address": {
			Name:         "email_address",
			Type:         DbTypeNvarchar,
			MaxLength:    640,
			IsRequired:   true,
			IsPrimaryKey: false,
		},
		"created_at": {
			Name:         "created_at",
			Type:         DbTypeDateTimeOffset,
			MaxLength:    10,
			IsRequired:   true,
			IsPrimaryKey: false,
		},
		"modified_at": {
			Name:         "modified_at",
			Type:         DbTypeDateTimeOffset,
			MaxLength:    10,
			IsRequired:   true,
			IsPrimaryKey: false,
		},
	},
	Relationships: map[string]Relationship{
		"businesses_users_modified_by_id": {
			Id:                 relationshipId("businesses_modified_by_id_users_id"),
			Type:               RelationshipOneToMany,
			RelationshipColumn: "businesses_users_modified_by_id",
			RelatedTable:       "businesses",
			LocalColumn:        "id",
			ForeignColumn:      "modified_by_id",
		},
		"farm_fields_users_created_by_id": {
			Id:                 relationshipId("farm_fields_created_by_id_users_id"),
			Type:               RelationshipOneToMany,
			RelationshipColumn: "farm_fields_users_created_by_id",
			RelatedTable:       "farm_fields",
			LocalColumn:        "id",
			ForeignColumn:      "created_by_id",
		},
		"farm_fields_users_modified_by_id": {
			Id:                 relationshipId("farm_fields_modified_by_id_users_id"),
			Type:               RelationshipOneToMany,
			RelationshipColumn: "farm_fields_users_modified_by_id",
			RelatedTable:       "farm_fields",
			LocalColumn:        "id",
			ForeignColumn:      "modified_by_id",
		},
		"businesses_users_created_by_id": {
			Id:                 relationshipId("businesses_created_by_id_users_id"),
			Type:               RelationshipOneToMany,
			RelationshipColumn: "businesses_users_created_by_id",
			RelatedTable:       "businesses",
			LocalColumn:        "id",
			ForeignColumn:      "created_by_id",
		},
	},
}

// GetUsersMetadata returns the database metadata associated with the users.Users table.
func GetUsersMetadata() TableMetadata {
	return usersMetadata
}

// UsersModel represents a row from the hg.users table.
type UsersModel struct {
	Id                          string             `json:"id,omitempty"`
	Oid                         string             `json:"oid,omitempty"`
	GivenName                   string             `json:"given_name,omitempty"`
	FamilyName                  string             `json:"family_name,omitempty"`
	UserName                    string             `json:"user_name,omitempty"`
	EmailAddress                string             `json:"email_address,omitempty"`
	CreatedAt                   *time.Time         `json:"created_at,omitempty"`
	ModifiedAt                  *time.Time         `json:"modified_at,omitempty"`
	BusinessesUsersCreatedById  []*BusinessesModel `json:"businesses_users_created_by_id,omitempty"`
	BusinessesUsersModifiedById []*BusinessesModel `json:"businesses_users_modified_by_id,omitempty"`
	FarmFieldsUsersCreatedById  []*FarmFieldsModel `json:"farm_fields_users_created_by_id,omitempty"`
	FarmFieldsUsersModifiedById []*FarmFieldsModel `json:"farm_fields_users_modified_by_id,omitempty"`
}

// GetMetadata returns metadata for the hg.users table.
func (m *UsersModel) GetMetadata() TableMetadata {
	return usersMetadata
}

// GetTableName returns the name of the hg.users table.
func (m *UsersModel) GetTableName() string {
	return usersMetadata.TableName
}

// NewSlice unmarshals a json array of users and returns it as a slice
func (m *UsersModel) NewSlice(jsonData []byte) ([]TableModel, error) {
	var concreteSlice []*UsersModel
	if err := json.Unmarshal(jsonData, &concreteSlice); err != nil {
		return nil, err
	}
	result := make([]TableModel, len(concreteSlice))
	for i := range concreteSlice {
		result[i] = concreteSlice[i]
	}
	return result, nil
}

// SetRelationshipField sets a given relationship field on the hg.users table
func (m *UsersModel) SetRelationshipField(relationshipId relationshipId, value TableModel) error {
	switch relationshipId {
	case "businesses_created_by_id_users_id":
		v, ok := value.(*BusinessesModel)
		if !ok {
			return errors.New("unexpected relationship type received")
		}
		m.BusinessesUsersCreatedById = append(m.BusinessesUsersCreatedById, v)

	case "businesses_modified_by_id_users_id":
		v, ok := value.(*BusinessesModel)
		if !ok {
			return errors.New("unexpected relationship type received")
		}
		m.BusinessesUsersModifiedById = append(m.BusinessesUsersModifiedById, v)

	case "farm_fields_created_by_id_users_id":
		v, ok := value.(*FarmFieldsModel)
		if !ok {
			return errors.New("unexpected relationship type received")
		}
		m.FarmFieldsUsersCreatedById = append(m.FarmFieldsUsersCreatedById, v)

	case "farm_fields_modified_by_id_users_id":
		v, ok := value.(*FarmFieldsModel)
		if !ok {
			return errors.New("unexpected relationship type received")
		}
		m.FarmFieldsUsersModifiedById = append(m.FarmFieldsUsersModifiedById, v)

	default:
		return fmt.Errorf("unknown relationship: %s", relationshipId)
	}
	return nil
}

// GetJoinOnValue returns the value of the relevant column for a given relationship
func (m *UsersModel) GetJoinOnValue(relationshipId relationshipId) (string, error) {
	switch relationshipId {
	case "businesses_created_by_id_users_id":
		return m.Id, nil
	case "businesses_modified_by_id_users_id":
		return m.Id, nil
	case "farm_fields_created_by_id_users_id":
		return m.Id, nil
	case "farm_fields_modified_by_id_users_id":
		return m.Id, nil
	default:
		return "", fmt.Errorf("unknown relationship: %s", relationshipId)
	}
}

// GetTableModel returns a new instance of a given table or nill if
// the table name is invalid
func GetTableModel(tableName string) TableModel {
	switch tableName {
	case businessesMetadata.TableName:
		return new(BusinessesModel)
	case farmFieldsMetadata.TableName:
		return new(FarmFieldsModel)
	case usersMetadata.TableName:
		return new(UsersModel)
	default:
		return nil
	}
}

// GetTableMetadata returns the metadata for a given table or nill if
// the table name is invalid
func GetTableMetadata(tableName string) *TableMetadata {
	switch tableName {
	case businessesMetadata.TableName:
		return &businessesMetadata
	case farmFieldsMetadata.TableName:
		return &farmFieldsMetadata
	case usersMetadata.TableName:
		return &usersMetadata
	default:
		return nil
	}
}
