// GENERATED CODE
// see cmd/model-builder/
package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"iter"
	"strconv"
	"strings"
	"time"

	"github.com/turnerbenjamin/heterogen_portal/internal/query"
)

var metadataIsBound = false

type schemaMetadata struct{}

func NewSchemaMetadata() *schemaMetadata {
	s := &schemaMetadata{}
	if !metadataIsBound {
		bindMetadata()
	}
	return s
}

// GetTableModel returns a new instance of a given table or nill if
// the table name is invalid
func (s *schemaMetadata) GetTableModel(tableName string) query.TableModel {
	return getTableModel(tableName)
}

// GetTableMetadata returns the metadata for a given table or nill if
// the table name is invalid
func (s *schemaMetadata) GetTableMetadata(tableName string) query.TableMetadata {
	return getTableMetadata(tableName)
}

// columnMetadata describes the metadata associated with a database table column.
type columnMetadata struct {
	name         string
	dbType       query.DbDataTypeName
	maxLength    int
	isPrimaryKey bool
	isRequired   bool
}

// Name returns the database name for the column
func (c *columnMetadata) Name() string {
	return c.name
}

// Type returns the column's type
func (c *columnMetadata) Type() query.DbDataTypeName {
	return c.dbType
}

// relationship describes a foreign key relationship between two database columns.
type relationship struct {
	id               string
	name             string
	relationshipType query.RelationshipType
	from             query.TableMetadata
	to               query.TableMetadata
	fromColumn       query.ColumnMetadata
	toColumn         query.ColumnMetadata
	fromTableName    string
	toTableName      string
	fromColumnName   string
	toColumnName     string
	isInitialised    bool
}

// bindMetadata binds metadata references at runtime to avoid invalid initiation
// cycle due to circular references
func (r *relationship) bindMetadata() {
	r.from = getTableMetadata(r.fromTableName)
	r.fromColumn = r.from.GetColumnMetadata(r.fromColumnName)
	r.to = getTableMetadata(r.toTableName)
	r.toColumn = r.to.GetColumnMetadata(r.toColumnName)

	if r.from == nil || r.fromColumn == nil || r.to == nil || r.toColumn == nil {
		panic(fmt.Sprintf("unable to bind metadata for relationship %s", r.id))
	}
}

// Id returns a unique identifier for a table relationship
func (r *relationship) Id() string {
	return r.id
}

// From returns table metadata for the from table
func (r *relationship) From() query.TableMetadata {
	return r.from
}

// To returns table metadata for the to table
func (r *relationship) To() query.TableMetadata {
	return r.to
}

// From returns table metadata for the from table
func (r *relationship) FromColumn() query.ColumnMetadata {
	return r.fromColumn
}

// To returns table metadata for the to table
func (r *relationship) ToColumn() query.ColumnMetadata {
	return r.toColumn
}

// Type returns the relationship type
func (r *relationship) Type() query.RelationshipType {
	return r.relationshipType
}

// tableMetadata describes the structure and relationships of a database table.
type tableMetadata struct {
	name               string
	schemaName         string
	fullyQualifiedName string
	primaryKey         string
	columns            map[string]*columnMetadata
	columnCount        int
	relationships      map[string]*relationship
}

// bindMetadata binds metadata references at runtime to avoid invalid initiation
// cycle due to circular references
func (t *tableMetadata) bindMetadata() {
	for _, relationship := range t.relationships {
		relationship.bindMetadata()
	}
}

// GetColumnMetadata returns metadata for a given table column; returns nil if
// the column does not exist on the table
func (t *tableMetadata) GetColumnMetadata(columnName string) query.ColumnMetadata {
	c, exists := t.columns[columnName]
	if !exists {
		return nil
	}
	return c
}

// GetRelationshipMetadata returns metadata for a given table column; returns
// nil if the relationship does not exist on the table
func (t *tableMetadata) GetRelationshipMetadata(relationshipName string) query.RelationshipMetadata {
	r, exists := t.relationships[relationshipName]
	if !exists {
		return nil
	}
	return r
}

// GetTableModel returns a model representing the table
func (t *tableMetadata) GetModel() query.TableModel {
	return getTableModel(t.name)
}

// Name returns the table name as it appears in the database
func (t *tableMetadata) Name() string {
	return t.name
}

// FullyQualifiedName returns the fully-qualified table name including the
// schema namespace
func (t *tableMetadata) FullyQualifiedName() string {
	return t.fullyQualifiedName
}

// Columns iterates over all columns associated with the table
func (t *tableMetadata) Columns() iter.Seq[query.ColumnMetadata] {
	return func(yield func(query.ColumnMetadata) bool) {
		for _, col := range t.columns {
			if !yield(col) {
				return
			}
		}
	}
}

// ColumnCount returns the total number of columns associated with the table
func (t *tableMetadata) ColumnCount() int {
	if t.columnCount == -1 {
		t.columnCount = len(t.columns)
	}
	return t.columnCount
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

// ColumnAccessPolicy defines the query access policy for a given database
// column
type ColumnAccessPolicy struct {
	UserCanAccess bool
}

// CanAccess defines, at the column level, if a user can perform any
// operations on that column
func (p ColumnAccessPolicy) CanAccess() bool {
	return p.UserCanAccess
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
var businessesMetadata = &tableMetadata{
	name:               "businesses",
	schemaName:         "hg",
	fullyQualifiedName: "hg.businesses",
	primaryKey:         "id",
	columns: map[string]*columnMetadata{
		"id": {
			name:         "id",
			dbType:       query.DbTypeNvarchar,
			maxLength:    72,
			isRequired:   true,
			isPrimaryKey: true,
		},
		"reference": {
			name:         "reference",
			dbType:       query.DbTypeNvarchar,
			maxLength:    16,
			isRequired:   true,
			isPrimaryKey: false,
		},
		"trading_name": {
			name:         "trading_name",
			dbType:       query.DbTypeNvarchar,
			maxLength:    510,
			isRequired:   true,
			isPrimaryKey: false,
		},
		"logo_url": {
			name:         "logo_url",
			dbType:       query.DbTypeNvarchar,
			maxLength:    4096,
			isRequired:   false,
			isPrimaryKey: false,
		},
		"description": {
			name:         "description",
			dbType:       query.DbTypeNvarchar,
			maxLength:    8000,
			isRequired:   false,
			isPrimaryKey: false,
		},
		"business_type": {
			name:         "business_type",
			dbType:       query.DbTypeInt,
			maxLength:    4,
			isRequired:   true,
			isPrimaryKey: false,
		},
		"cph_number": {
			name:         "cph_number",
			dbType:       query.DbTypeNvarchar,
			maxLength:    22,
			isRequired:   false,
			isPrimaryKey: false,
		},
		"email_address": {
			name:         "email_address",
			dbType:       query.DbTypeNvarchar,
			maxLength:    640,
			isRequired:   false,
			isPrimaryKey: false,
		},
		"contact_number": {
			name:         "contact_number",
			dbType:       query.DbTypeNvarchar,
			maxLength:    100,
			isRequired:   false,
			isPrimaryKey: false,
		},
		"website_url": {
			name:         "website_url",
			dbType:       query.DbTypeNvarchar,
			maxLength:    4096,
			isRequired:   false,
			isPrimaryKey: false,
		},
		"address_line_1": {
			name:         "address_line_1",
			dbType:       query.DbTypeNvarchar,
			maxLength:    510,
			isRequired:   true,
			isPrimaryKey: false,
		},
		"address_line_2": {
			name:         "address_line_2",
			dbType:       query.DbTypeNvarchar,
			maxLength:    510,
			isRequired:   false,
			isPrimaryKey: false,
		},
		"town": {
			name:         "town",
			dbType:       query.DbTypeNvarchar,
			maxLength:    200,
			isRequired:   false,
			isPrimaryKey: false,
		},
		"county": {
			name:         "county",
			dbType:       query.DbTypeNvarchar,
			maxLength:    200,
			isRequired:   false,
			isPrimaryKey: false,
		},
		"country": {
			name:         "country",
			dbType:       query.DbTypeNvarchar,
			maxLength:    200,
			isRequired:   false,
			isPrimaryKey: false,
		},
		"postcode": {
			name:         "postcode",
			dbType:       query.DbTypeNvarchar,
			maxLength:    40,
			isRequired:   true,
			isPrimaryKey: false,
		},
		"location": {
			name:         "location",
			dbType:       query.DbTypeGeography,
			maxLength:    -1,
			isRequired:   true,
			isPrimaryKey: false,
		},
		"created_at": {
			name:         "created_at",
			dbType:       query.DbTypeDateTimeOffset,
			maxLength:    10,
			isRequired:   true,
			isPrimaryKey: false,
		},
		"created_by_id": {
			name:         "created_by_id",
			dbType:       query.DbTypeNvarchar,
			maxLength:    72,
			isRequired:   true,
			isPrimaryKey: false,
		},
		"modified_at": {
			name:         "modified_at",
			dbType:       query.DbTypeDateTimeOffset,
			maxLength:    10,
			isRequired:   true,
			isPrimaryKey: false,
		},
		"modified_by_id": {
			name:         "modified_by_id",
			dbType:       query.DbTypeNvarchar,
			maxLength:    72,
			isRequired:   true,
			isPrimaryKey: false,
		},
	},
	relationships: map[string]*relationship{
		"farm_fields_businesses_business_id": {
			id:               "farm_fields_business_id_businesses_id",
			name:             "farm_fields_businesses_business_id",
			relationshipType: query.RelationshipOneToMany,
			fromTableName:    "businesses",
			toTableName:      "farm_fields",
			fromColumnName:   "id",
			toColumnName:     "business_id",
		},
		"created_by_id": {
			id:               "businesses_created_by_id_users_id",
			name:             "created_by",
			relationshipType: query.RelationshipManyToOne,
			fromTableName:    "businesses",
			toTableName:      "users",
			fromColumnName:   "created_by_id",
			toColumnName:     "id",
		},
		"modified_by_id": {
			id:               "businesses_modified_by_id_users_id",
			name:             "modified_by",
			relationshipType: query.RelationshipManyToOne,
			fromTableName:    "businesses",
			toTableName:      "users",
			fromColumnName:   "modified_by_id",
			toColumnName:     "id",
		},
	},
	columnCount: -1,
}

// GetBusinessesMetadata returns the database metadata associated with the businesses.Businesses table.
func GetBusinessesMetadata() *tableMetadata {
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
	FarmFieldsBusinessesBusinessId []*FarmFieldsModel `json:"farm_fields_businesses_business_id,omitempty"`
	CreatedBy                      *UsersModel        `json:"created_by,omitempty"`
	ModifiedBy                     *UsersModel        `json:"modified_by,omitempty"`
}

// NewSlice unmarshals a json array of businesses and returns it as a slice
func (m *BusinessesModel) NewSlice(jsonData []byte) ([]query.TableModel, error) {
	var concreteSlice []*BusinessesModel
	if err := json.Unmarshal(jsonData, &concreteSlice); err != nil {
		return nil, err
	}
	result := make([]query.TableModel, len(concreteSlice))
	for i := range concreteSlice {
		result[i] = concreteSlice[i]
	}
	return result, nil
}

// SetRelationshipField sets a given relationship field on the hg.businesses table
func (m *BusinessesModel) SetRelationshipField(relationshipId string, value query.TableModel) error {
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
func (m *BusinessesModel) GetJoinOnValue(relationshipId string) (string, error) {
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
var farmFieldsMetadata = &tableMetadata{
	name:               "farm_fields",
	schemaName:         "hg",
	fullyQualifiedName: "hg.farm_fields",
	primaryKey:         "id",
	columns: map[string]*columnMetadata{
		"id": {
			name:         "id",
			dbType:       query.DbTypeNvarchar,
			maxLength:    72,
			isRequired:   true,
			isPrimaryKey: true,
		},
		"reference": {
			name:         "reference",
			dbType:       query.DbTypeNvarchar,
			maxLength:    510,
			isRequired:   true,
			isPrimaryKey: false,
		},
		"business_id": {
			name:         "business_id",
			dbType:       query.DbTypeNvarchar,
			maxLength:    72,
			isRequired:   true,
			isPrimaryKey: false,
		},
		"location": {
			name:         "location",
			dbType:       query.DbTypeGeography,
			maxLength:    -1,
			isRequired:   true,
			isPrimaryKey: false,
		},
		"created_at": {
			name:         "created_at",
			dbType:       query.DbTypeDateTimeOffset,
			maxLength:    10,
			isRequired:   true,
			isPrimaryKey: false,
		},
		"created_by_id": {
			name:         "created_by_id",
			dbType:       query.DbTypeNvarchar,
			maxLength:    72,
			isRequired:   true,
			isPrimaryKey: false,
		},
		"modified_at": {
			name:         "modified_at",
			dbType:       query.DbTypeDateTimeOffset,
			maxLength:    10,
			isRequired:   true,
			isPrimaryKey: false,
		},
		"modified_by_id": {
			name:         "modified_by_id",
			dbType:       query.DbTypeNvarchar,
			maxLength:    72,
			isRequired:   true,
			isPrimaryKey: false,
		},
	},
	relationships: map[string]*relationship{
		"modified_by_id": {
			id:               "farm_fields_modified_by_id_users_id",
			name:             "modified_by",
			relationshipType: query.RelationshipManyToOne,
			fromTableName:    "farm_fields",
			toTableName:      "users",
			fromColumnName:   "modified_by_id",
			toColumnName:     "id",
		},
		"business_id": {
			id:               "farm_fields_business_id_businesses_id",
			name:             "business",
			relationshipType: query.RelationshipManyToOne,
			fromTableName:    "farm_fields",
			toTableName:      "businesses",
			fromColumnName:   "business_id",
			toColumnName:     "id",
		},
		"created_by_id": {
			id:               "farm_fields_created_by_id_users_id",
			name:             "created_by",
			relationshipType: query.RelationshipManyToOne,
			fromTableName:    "farm_fields",
			toTableName:      "users",
			fromColumnName:   "created_by_id",
			toColumnName:     "id",
		},
	},
	columnCount: -1,
}

// GetFarmFieldsMetadata returns the database metadata associated with the farm_fields.FarmFields table.
func GetFarmFieldsMetadata() *tableMetadata {
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

// NewSlice unmarshals a json array of farm_fields and returns it as a slice
func (m *FarmFieldsModel) NewSlice(jsonData []byte) ([]query.TableModel, error) {
	var concreteSlice []*FarmFieldsModel
	if err := json.Unmarshal(jsonData, &concreteSlice); err != nil {
		return nil, err
	}
	result := make([]query.TableModel, len(concreteSlice))
	for i := range concreteSlice {
		result[i] = concreteSlice[i]
	}
	return result, nil
}

// SetRelationshipField sets a given relationship field on the hg.farm_fields table
func (m *FarmFieldsModel) SetRelationshipField(relationshipId string, value query.TableModel) error {
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
func (m *FarmFieldsModel) GetJoinOnValue(relationshipId string) (string, error) {
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
var usersMetadata = &tableMetadata{
	name:               "users",
	schemaName:         "hg",
	fullyQualifiedName: "hg.users",
	primaryKey:         "id",
	columns: map[string]*columnMetadata{
		"id": {
			name:         "id",
			dbType:       query.DbTypeNvarchar,
			maxLength:    72,
			isRequired:   true,
			isPrimaryKey: true,
		},
		"oid": {
			name:         "oid",
			dbType:       query.DbTypeNvarchar,
			maxLength:    72,
			isRequired:   true,
			isPrimaryKey: false,
		},
		"given_name": {
			name:         "given_name",
			dbType:       query.DbTypeNvarchar,
			maxLength:    128,
			isRequired:   true,
			isPrimaryKey: false,
		},
		"family_name": {
			name:         "family_name",
			dbType:       query.DbTypeNvarchar,
			maxLength:    128,
			isRequired:   true,
			isPrimaryKey: false,
		},
		"user_name": {
			name:         "user_name",
			dbType:       query.DbTypeNvarchar,
			maxLength:    256,
			isRequired:   true,
			isPrimaryKey: false,
		},
		"email_address": {
			name:         "email_address",
			dbType:       query.DbTypeNvarchar,
			maxLength:    640,
			isRequired:   true,
			isPrimaryKey: false,
		},
		"created_at": {
			name:         "created_at",
			dbType:       query.DbTypeDateTimeOffset,
			maxLength:    10,
			isRequired:   true,
			isPrimaryKey: false,
		},
		"modified_at": {
			name:         "modified_at",
			dbType:       query.DbTypeDateTimeOffset,
			maxLength:    10,
			isRequired:   true,
			isPrimaryKey: false,
		},
	},
	relationships: map[string]*relationship{
		"businesses_users_modified_by_id": {
			id:               "businesses_modified_by_id_users_id",
			name:             "businesses_users_modified_by_id",
			relationshipType: query.RelationshipOneToMany,
			fromTableName:    "users",
			toTableName:      "businesses",
			fromColumnName:   "id",
			toColumnName:     "modified_by_id",
		},
		"farm_fields_users_created_by_id": {
			id:               "farm_fields_created_by_id_users_id",
			name:             "farm_fields_users_created_by_id",
			relationshipType: query.RelationshipOneToMany,
			fromTableName:    "users",
			toTableName:      "farm_fields",
			fromColumnName:   "id",
			toColumnName:     "created_by_id",
		},
		"farm_fields_users_modified_by_id": {
			id:               "farm_fields_modified_by_id_users_id",
			name:             "farm_fields_users_modified_by_id",
			relationshipType: query.RelationshipOneToMany,
			fromTableName:    "users",
			toTableName:      "farm_fields",
			fromColumnName:   "id",
			toColumnName:     "modified_by_id",
		},
		"businesses_users_created_by_id": {
			id:               "businesses_created_by_id_users_id",
			name:             "businesses_users_created_by_id",
			relationshipType: query.RelationshipOneToMany,
			fromTableName:    "users",
			toTableName:      "businesses",
			fromColumnName:   "id",
			toColumnName:     "created_by_id",
		},
	},
	columnCount: -1,
}

// GetUsersMetadata returns the database metadata associated with the users.Users table.
func GetUsersMetadata() *tableMetadata {
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

// NewSlice unmarshals a json array of users and returns it as a slice
func (m *UsersModel) NewSlice(jsonData []byte) ([]query.TableModel, error) {
	var concreteSlice []*UsersModel
	if err := json.Unmarshal(jsonData, &concreteSlice); err != nil {
		return nil, err
	}
	result := make([]query.TableModel, len(concreteSlice))
	for i := range concreteSlice {
		result[i] = concreteSlice[i]
	}
	return result, nil
}

// SetRelationshipField sets a given relationship field on the hg.users table
func (m *UsersModel) SetRelationshipField(relationshipId string, value query.TableModel) error {
	switch relationshipId {
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

	case "businesses_created_by_id_users_id":
		v, ok := value.(*BusinessesModel)
		if !ok {
			return errors.New("unexpected relationship type received")
		}
		m.BusinessesUsersCreatedById = append(m.BusinessesUsersCreatedById, v)

	default:
		return fmt.Errorf("unknown relationship: %s", relationshipId)
	}
	return nil
}

// GetJoinOnValue returns the value of the relevant column for a given relationship
func (m *UsersModel) GetJoinOnValue(relationshipId string) (string, error) {
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

// getTableModel returns a new instance of a given table or nill if
// the table name is invalid
func getTableModel(tableName string) query.TableModel {
	switch tableName {
	case businessesMetadata.name:
		return new(BusinessesModel)
	case farmFieldsMetadata.name:
		return new(FarmFieldsModel)
	case usersMetadata.name:
		return new(UsersModel)
	default:
		return nil
	}
}

// getTableMetadata returns the metadata for a given table or nill if
// the table name is invalid
func getTableMetadata(tableName string) query.TableMetadata {
	switch tableName {
	case businessesMetadata.name:
		return businessesMetadata
	case farmFieldsMetadata.name:
		return farmFieldsMetadata
	case usersMetadata.name:
		return usersMetadata
	default:
		return nil
	}
}

// bindMetadata binds table references at runtime to avoid invalid initiation
// cycle due to circular references
func bindMetadata() {
	businessesMetadata.bindMetadata()
	farmFieldsMetadata.bindMetadata()
	usersMetadata.bindMetadata()
}

// AccessPolicy defines the access policy for the database
type DatabaseAccessPolicy struct {
	BusinessesAccessPolicy *BusinessesAccessPolicy
	FarmFieldsAccessPolicy *FarmFieldsAccessPolicy
	UsersAccessPolicy      *UsersAccessPolicy
}

// GetTableAccessPolicy returns an access policy for a given table for nil
// if the table does not exist
func (p *DatabaseAccessPolicy) GetTableAccessPolicy(tableName string) query.TableAccessPolicy {
	switch tableName {
	case "businesses":
		return p.BusinessesAccessPolicy
	case "farm_fields":
		return p.FarmFieldsAccessPolicy
	case "users":
		return p.UsersAccessPolicy
	default:
		return nil
	}
}

// BusinessesAccessPolicy defines an access policy for the businesses table
type BusinessesAccessPolicy struct {
	UserCanAccess bool
	Id            ColumnAccessPolicy
	Reference     ColumnAccessPolicy
	TradingName   ColumnAccessPolicy
	LogoUrl       ColumnAccessPolicy
	Description   ColumnAccessPolicy
	BusinessType  ColumnAccessPolicy
	CphNumber     ColumnAccessPolicy
	EmailAddress  ColumnAccessPolicy
	ContactNumber ColumnAccessPolicy
	WebsiteUrl    ColumnAccessPolicy
	AddressLine1  ColumnAccessPolicy
	AddressLine2  ColumnAccessPolicy
	Town          ColumnAccessPolicy
	County        ColumnAccessPolicy
	Country       ColumnAccessPolicy
	Postcode      ColumnAccessPolicy
	Location      ColumnAccessPolicy
	CreatedAt     ColumnAccessPolicy
	CreatedById   ColumnAccessPolicy
	ModifiedAt    ColumnAccessPolicy
	ModifiedById  ColumnAccessPolicy
}

// CanAccess defines, at the table level, if a user can perform any
// operations on the businesses table. Specific column access policies can restrict
// access given but cannot override access denied
func (p *BusinessesAccessPolicy) CanAccess() bool {
	return p.UserCanAccess
}

// GetColumnAccessPolicy returns an access policy for a specific column on
// the businesses table. Returns nil if the column does not exist
func (p *BusinessesAccessPolicy) GetColumnAccessPolicy(columnName string) query.ColumnAccessPolicy {
	switch columnName {
	case "id":
		return p.Id
	case "reference":
		return p.Reference
	case "trading_name":
		return p.TradingName
	case "logo_url":
		return p.LogoUrl
	case "description":
		return p.Description
	case "business_type":
		return p.BusinessType
	case "cph_number":
		return p.CphNumber
	case "email_address":
		return p.EmailAddress
	case "contact_number":
		return p.ContactNumber
	case "website_url":
		return p.WebsiteUrl
	case "address_line_1":
		return p.AddressLine1
	case "address_line_2":
		return p.AddressLine2
	case "town":
		return p.Town
	case "county":
		return p.County
	case "country":
		return p.Country
	case "postcode":
		return p.Postcode
	case "location":
		return p.Location
	case "created_at":
		return p.CreatedAt
	case "created_by_id":
		return p.CreatedById
	case "modified_at":
		return p.ModifiedAt
	case "modified_by_id":
		return p.ModifiedById
	default:
		return nil
	}
}

// FarmFieldsAccessPolicy defines an access policy for the farm_fields table
type FarmFieldsAccessPolicy struct {
	UserCanAccess bool
	Id            ColumnAccessPolicy
	Reference     ColumnAccessPolicy
	BusinessId    ColumnAccessPolicy
	Location      ColumnAccessPolicy
	CreatedAt     ColumnAccessPolicy
	CreatedById   ColumnAccessPolicy
	ModifiedAt    ColumnAccessPolicy
	ModifiedById  ColumnAccessPolicy
}

// CanAccess defines, at the table level, if a user can perform any
// operations on the farm_fields table. Specific column access policies can restrict
// access given but cannot override access denied
func (p *FarmFieldsAccessPolicy) CanAccess() bool {
	return p.UserCanAccess
}

// GetColumnAccessPolicy returns an access policy for a specific column on
// the farm_fields table. Returns nil if the column does not exist
func (p *FarmFieldsAccessPolicy) GetColumnAccessPolicy(columnName string) query.ColumnAccessPolicy {
	switch columnName {
	case "id":
		return p.Id
	case "reference":
		return p.Reference
	case "business_id":
		return p.BusinessId
	case "location":
		return p.Location
	case "created_at":
		return p.CreatedAt
	case "created_by_id":
		return p.CreatedById
	case "modified_at":
		return p.ModifiedAt
	case "modified_by_id":
		return p.ModifiedById
	default:
		return nil
	}
}

// UsersAccessPolicy defines an access policy for the users table
type UsersAccessPolicy struct {
	UserCanAccess bool
	Id            ColumnAccessPolicy
	Oid           ColumnAccessPolicy
	GivenName     ColumnAccessPolicy
	FamilyName    ColumnAccessPolicy
	UserName      ColumnAccessPolicy
	EmailAddress  ColumnAccessPolicy
	CreatedAt     ColumnAccessPolicy
	ModifiedAt    ColumnAccessPolicy
}

// CanAccess defines, at the table level, if a user can perform any
// operations on the users table. Specific column access policies can restrict
// access given but cannot override access denied
func (p *UsersAccessPolicy) CanAccess() bool {
	return p.UserCanAccess
}

// GetColumnAccessPolicy returns an access policy for a specific column on
// the users table. Returns nil if the column does not exist
func (p *UsersAccessPolicy) GetColumnAccessPolicy(columnName string) query.ColumnAccessPolicy {
	switch columnName {
	case "id":
		return p.Id
	case "oid":
		return p.Oid
	case "given_name":
		return p.GivenName
	case "family_name":
		return p.FamilyName
	case "user_name":
		return p.UserName
	case "email_address":
		return p.EmailAddress
	case "created_at":
		return p.CreatedAt
	case "modified_at":
		return p.ModifiedAt
	default:
		return nil
	}
}
