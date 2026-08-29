// GENERATED CODE
// see cmd/model-builder/
package model

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"iter"
	"strconv"
	"strings"
	"time"

	"github.com/turnerbenjamin/heterogen_portal/internal/query/queryModel"
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
func (s *schemaMetadata) GetTableModel(tableName string) queryModel.TableModel {
	return getTableModel(tableName)
}

// GetTableMetadata returns the metadata for a given table or nill if
// the table name is invalid
func (s *schemaMetadata) GetTableMetadata(tableName string) queryModel.TableMetadata {
	return getTableMetadata(tableName)
}

// columnMetadata describes the metadata associated with a database table column.
type columnMetadata struct {
	name         string
	dbType       queryModel.DbDataTypeName
	maxLength    int
	isPrimaryKey bool
	isRequired   bool
}

// Name returns the database name for the column
func (c *columnMetadata) Name() string {
	return c.name
}

// Type returns the column's type
func (c *columnMetadata) Type() queryModel.DbDataTypeName {
	return c.dbType
}

// relationship describes a foreign key relationship between two database columns.
type relationship struct {
	id                  string
	columnName          string
	expansionColumnName string
	relationshipType    queryModel.RelationshipType
	from                queryModel.TableMetadata
	to                  queryModel.TableMetadata
	fromColumn          queryModel.ColumnMetadata
	toColumn            queryModel.ColumnMetadata
	fromTableName       string
	toTableName         string
	fromColumnName      string
	toColumnName        string
	isInitialised       bool
}

// ExpansionColumnName returns the name of the pseudo relationship column on the
// table used to store expanded results
func (r *relationship) ExpansionColumnName() string {
	return r.expansionColumnName
}

// ColumnName returns the name of the column, as it should be referenced in a
// query string
func (r *relationship) ColumnName() string {
	return r.columnName
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
func (r *relationship) From() queryModel.TableMetadata {
	return r.from
}

// To returns table metadata for the to table
func (r *relationship) To() queryModel.TableMetadata {
	return r.to
}

// From returns table metadata for the from table
func (r *relationship) FromColumn() queryModel.ColumnMetadata {
	return r.fromColumn
}

// To returns table metadata for the to table
func (r *relationship) ToColumn() queryModel.ColumnMetadata {
	return r.toColumn
}

// Type returns the relationship type
func (r *relationship) Type() queryModel.RelationshipType {
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
func (t *tableMetadata) GetColumnMetadata(columnName string) queryModel.ColumnMetadata {
	c, exists := t.columns[columnName]
	if !exists {
		return nil
	}
	return c
}

// GetRelationshipMetadata returns metadata for a given table column; returns
// nil if the relationship does not exist on the table
func (t *tableMetadata) GetRelationshipMetadata(relationshipName string) queryModel.RelationshipMetadata {
	r, exists := t.relationships[relationshipName]
	if !exists {
		return nil
	}
	return r
}

// GetTableModel returns a model representing the table
func (t *tableMetadata) GetModel() queryModel.TableModel {
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
func (t *tableMetadata) Columns() iter.Seq[queryModel.ColumnMetadata] {
	return func(yield func(queryModel.ColumnMetadata) bool) {
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

// PrimaryKeyField returns the column metadata for the primary key field
func (t *tableMetadata) PrimaryKeyField() queryModel.ColumnMetadata {
	pk, ok := t.columns[t.primaryKey]
	if !ok || pk == nil {
		panic(fmt.Sprintf("unable to access primary key for %s table", t.name))
	}
	return pk
}

// InitProjection initialises a projection object for the table
func (t *tableMetadata) InitProjection() (queryModel.Projection, error) {
	return initTableProjection(t.name)
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

// String returns the string representation of a Point
func (p *Point) String() string {
	return fmt.Sprintf("POINT (%f %f)", p.Coordinates[0], p.Coordinates[1])
}

// marshalProperty is a helper method to marshal an attribute and value to a json string buffer
func marshalProperty(buf *bytes.Buffer, isFirst bool, columnName string, value any) error {
	if !isFirst {
		buf.WriteByte(',')
	}

	buf.WriteString("\"")
	buf.WriteString(columnName)
	buf.WriteString("\":")

	v, err := json.Marshal(value)
	if err != nil {
		return err
	}

	buf.Write(v)
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
			dbType:       queryModel.DbTypeNvarchar,
			maxLength:    72,
			isRequired:   true,
			isPrimaryKey: true,
		},
		"reference": {
			name:         "reference",
			dbType:       queryModel.DbTypeNvarchar,
			maxLength:    16,
			isRequired:   true,
			isPrimaryKey: false,
		},
		"trading_name": {
			name:         "trading_name",
			dbType:       queryModel.DbTypeNvarchar,
			maxLength:    510,
			isRequired:   true,
			isPrimaryKey: false,
		},
		"logo_url": {
			name:         "logo_url",
			dbType:       queryModel.DbTypeNvarchar,
			maxLength:    4096,
			isRequired:   false,
			isPrimaryKey: false,
		},
		"description": {
			name:         "description",
			dbType:       queryModel.DbTypeNvarchar,
			maxLength:    8000,
			isRequired:   false,
			isPrimaryKey: false,
		},
		"business_type": {
			name:         "business_type",
			dbType:       queryModel.DbTypeInt,
			maxLength:    4,
			isRequired:   true,
			isPrimaryKey: false,
		},
		"cph_number": {
			name:         "cph_number",
			dbType:       queryModel.DbTypeNvarchar,
			maxLength:    22,
			isRequired:   false,
			isPrimaryKey: false,
		},
		"email_address": {
			name:         "email_address",
			dbType:       queryModel.DbTypeNvarchar,
			maxLength:    640,
			isRequired:   false,
			isPrimaryKey: false,
		},
		"contact_number": {
			name:         "contact_number",
			dbType:       queryModel.DbTypeNvarchar,
			maxLength:    100,
			isRequired:   false,
			isPrimaryKey: false,
		},
		"website_url": {
			name:         "website_url",
			dbType:       queryModel.DbTypeNvarchar,
			maxLength:    4096,
			isRequired:   false,
			isPrimaryKey: false,
		},
		"address_line_1": {
			name:         "address_line_1",
			dbType:       queryModel.DbTypeNvarchar,
			maxLength:    510,
			isRequired:   true,
			isPrimaryKey: false,
		},
		"address_line_2": {
			name:         "address_line_2",
			dbType:       queryModel.DbTypeNvarchar,
			maxLength:    510,
			isRequired:   false,
			isPrimaryKey: false,
		},
		"town": {
			name:         "town",
			dbType:       queryModel.DbTypeNvarchar,
			maxLength:    200,
			isRequired:   false,
			isPrimaryKey: false,
		},
		"county": {
			name:         "county",
			dbType:       queryModel.DbTypeNvarchar,
			maxLength:    200,
			isRequired:   false,
			isPrimaryKey: false,
		},
		"country": {
			name:         "country",
			dbType:       queryModel.DbTypeNvarchar,
			maxLength:    200,
			isRequired:   false,
			isPrimaryKey: false,
		},
		"postcode": {
			name:         "postcode",
			dbType:       queryModel.DbTypeNvarchar,
			maxLength:    40,
			isRequired:   true,
			isPrimaryKey: false,
		},
		"location": {
			name:         "location",
			dbType:       queryModel.DbTypeGeography,
			maxLength:    -1,
			isRequired:   true,
			isPrimaryKey: false,
		},
		"created_at": {
			name:         "created_at",
			dbType:       queryModel.DbTypeDateTimeOffset,
			maxLength:    10,
			isRequired:   true,
			isPrimaryKey: false,
		},
		"created_by_id": {
			name:         "created_by_id",
			dbType:       queryModel.DbTypeNvarchar,
			maxLength:    72,
			isRequired:   true,
			isPrimaryKey: false,
		},
		"modified_at": {
			name:         "modified_at",
			dbType:       queryModel.DbTypeDateTimeOffset,
			maxLength:    10,
			isRequired:   true,
			isPrimaryKey: false,
		},
		"modified_by_id": {
			name:         "modified_by_id",
			dbType:       queryModel.DbTypeNvarchar,
			maxLength:    72,
			isRequired:   true,
			isPrimaryKey: false,
		},
	},
	relationships: map[string]*relationship{
		"created_by_id": {
			id:                  "businesses_created_by_id_users_id",
			columnName:          "created_by_id",
			expansionColumnName: "created_by",
			relationshipType:    queryModel.RelationshipManyToOne,
			fromTableName:       "businesses",
			toTableName:         "users",
			fromColumnName:      "created_by_id",
			toColumnName:        "id",
		},
		"modified_by_id": {
			id:                  "businesses_modified_by_id_users_id",
			columnName:          "modified_by_id",
			expansionColumnName: "modified_by",
			relationshipType:    queryModel.RelationshipManyToOne,
			fromTableName:       "businesses",
			toTableName:         "users",
			fromColumnName:      "modified_by_id",
			toColumnName:        "id",
		},
		"farm_fields_businesses_business_id": {
			id:                  "farm_fields_business_id_businesses_id",
			columnName:          "farm_fields_businesses_business_id",
			expansionColumnName: "farm_fields_businesses_business_id",
			relationshipType:    queryModel.RelationshipOneToMany,
			fromTableName:       "businesses",
			toTableName:         "farm_fields",
			fromColumnName:      "id",
			toColumnName:        "business_id",
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
	Id                             string               `json:"id"`
	Reference                      string               `json:"reference"`
	TradingName                    string               `json:"trading_name"`
	LogoUrl                        string               `json:"logo_url"`
	Description                    string               `json:"description"`
	BusinessType                   int64                `json:"business_type"`
	CphNumber                      string               `json:"cph_number"`
	EmailAddress                   string               `json:"email_address"`
	ContactNumber                  string               `json:"contact_number"`
	WebsiteUrl                     string               `json:"website_url"`
	AddressLine1                   string               `json:"address_line_1"`
	AddressLine2                   string               `json:"address_line_2"`
	Town                           string               `json:"town"`
	County                         string               `json:"county"`
	Country                        string               `json:"country"`
	Postcode                       string               `json:"postcode"`
	Location                       *Point               `json:"location"`
	CreatedAt                      *time.Time           `json:"created_at"`
	CreatedById                    string               `json:"created_by_id"`
	ModifiedAt                     *time.Time           `json:"modified_at"`
	ModifiedById                   string               `json:"modified_by_id"`
	FarmFieldsBusinessesBusinessId []*FarmFieldsModel   `json:"farm_fields_businesses_business_id"`
	CreatedBy                      *UsersModel          `json:"created_by"`
	ModifiedBy                     *UsersModel          `json:"modified_by"`
	projection                     BusinessesProjection `json:"-"`
}

// NewSlice unmarshals a json array of businesses and returns it as a slice
func (m *BusinessesModel) NewSlice(jsonData []byte, projection queryModel.Projection) ([]queryModel.TableModel, error) {
	if len(jsonData) == 0 {
		return []queryModel.TableModel{}, nil
	}

	var typedProjection *BusinessesProjection
	switch p := projection.(type) {
	case *BusinessesProjection:
		typedProjection = p
	default:
		return nil, fmt.Errorf("unable to create new slice: invalid projection type received")
	}

	var concreteSlice []*BusinessesModel
	if err := json.Unmarshal(jsonData, &concreteSlice); err != nil {
		return nil, err
	}
	result := make([]queryModel.TableModel, len(concreteSlice))

	for i := range concreteSlice {
		concreteSlice[i].projection = *typedProjection
		result[i] = concreteSlice[i]
	}
	return result, nil
}

// MarshalJSON marshals the model to a json string based on the projection
func (m *BusinessesModel) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte('{')

	isFirst := true
	if m.projection.Has(businessesProjectionId) {
		err := marshalProperty(&buf, isFirst, "id", m.Id)
		if err != nil {
			return nil, err
		}
		isFirst = false
	}

	if m.projection.Has(businessesProjectionReference) {
		err := marshalProperty(&buf, isFirst, "reference", m.Reference)
		if err != nil {
			return nil, err
		}
		isFirst = false
	}

	if m.projection.Has(businessesProjectionTradingName) {
		err := marshalProperty(&buf, isFirst, "trading_name", m.TradingName)
		if err != nil {
			return nil, err
		}
		isFirst = false
	}

	if m.projection.Has(businessesProjectionLogoUrl) {
		err := marshalProperty(&buf, isFirst, "logo_url", m.LogoUrl)
		if err != nil {
			return nil, err
		}
		isFirst = false
	}

	if m.projection.Has(businessesProjectionDescription) {
		err := marshalProperty(&buf, isFirst, "description", m.Description)
		if err != nil {
			return nil, err
		}
		isFirst = false
	}

	if m.projection.Has(businessesProjectionBusinessType) {
		err := marshalProperty(&buf, isFirst, "business_type", m.BusinessType)
		if err != nil {
			return nil, err
		}
		isFirst = false
	}

	if m.projection.Has(businessesProjectionCphNumber) {
		err := marshalProperty(&buf, isFirst, "cph_number", m.CphNumber)
		if err != nil {
			return nil, err
		}
		isFirst = false
	}

	if m.projection.Has(businessesProjectionEmailAddress) {
		err := marshalProperty(&buf, isFirst, "email_address", m.EmailAddress)
		if err != nil {
			return nil, err
		}
		isFirst = false
	}

	if m.projection.Has(businessesProjectionContactNumber) {
		err := marshalProperty(&buf, isFirst, "contact_number", m.ContactNumber)
		if err != nil {
			return nil, err
		}
		isFirst = false
	}

	if m.projection.Has(businessesProjectionWebsiteUrl) {
		err := marshalProperty(&buf, isFirst, "website_url", m.WebsiteUrl)
		if err != nil {
			return nil, err
		}
		isFirst = false
	}

	if m.projection.Has(businessesProjectionAddressLine1) {
		err := marshalProperty(&buf, isFirst, "address_line_1", m.AddressLine1)
		if err != nil {
			return nil, err
		}
		isFirst = false
	}

	if m.projection.Has(businessesProjectionAddressLine2) {
		err := marshalProperty(&buf, isFirst, "address_line_2", m.AddressLine2)
		if err != nil {
			return nil, err
		}
		isFirst = false
	}

	if m.projection.Has(businessesProjectionTown) {
		err := marshalProperty(&buf, isFirst, "town", m.Town)
		if err != nil {
			return nil, err
		}
		isFirst = false
	}

	if m.projection.Has(businessesProjectionCounty) {
		err := marshalProperty(&buf, isFirst, "county", m.County)
		if err != nil {
			return nil, err
		}
		isFirst = false
	}

	if m.projection.Has(businessesProjectionCountry) {
		err := marshalProperty(&buf, isFirst, "country", m.Country)
		if err != nil {
			return nil, err
		}
		isFirst = false
	}

	if m.projection.Has(businessesProjectionPostcode) {
		err := marshalProperty(&buf, isFirst, "postcode", m.Postcode)
		if err != nil {
			return nil, err
		}
		isFirst = false
	}

	if m.projection.Has(businessesProjectionLocation) {
		err := marshalProperty(&buf, isFirst, "location", m.Location)
		if err != nil {
			return nil, err
		}
		isFirst = false
	}

	if m.projection.Has(businessesProjectionCreatedAt) {
		err := marshalProperty(&buf, isFirst, "created_at", m.CreatedAt)
		if err != nil {
			return nil, err
		}
		isFirst = false
	}

	if m.projection.Has(businessesProjectionCreatedById) {
		err := marshalProperty(&buf, isFirst, "created_by_id", m.CreatedById)
		if err != nil {
			return nil, err
		}
		isFirst = false
	}

	if m.projection.Has(businessesProjectionModifiedAt) {
		err := marshalProperty(&buf, isFirst, "modified_at", m.ModifiedAt)
		if err != nil {
			return nil, err
		}
		isFirst = false
	}

	if m.projection.Has(businessesProjectionModifiedById) {
		err := marshalProperty(&buf, isFirst, "modified_by_id", m.ModifiedById)
		if err != nil {
			return nil, err
		}
		isFirst = false
	}

	if m.projection.Has(businessesProjectionCreatedBy) {
		err := marshalProperty(&buf, isFirst, "created_by", m.CreatedBy)
		if err != nil {
			return nil, err
		}
		isFirst = false
	}

	if m.projection.Has(businessesProjectionModifiedBy) {
		err := marshalProperty(&buf, isFirst, "modified_by", m.ModifiedBy)
		if err != nil {
			return nil, err
		}
		isFirst = false
	}

	if m.projection.Has(businessesProjectionFarmFieldsBusinessesBusinessId) {
		err := marshalProperty(&buf, isFirst, "farm_fields_businesses_business_id", m.FarmFieldsBusinessesBusinessId)
		if err != nil {
			return nil, err
		}
		isFirst = false
	}

	buf.WriteByte('}')
	return buf.Bytes(), nil
}

// SetRelationshipField sets a given relationship field on the hg.businesses table
func (m *BusinessesModel) SetRelationshipField(relationshipId string, value queryModel.TableModel) error {
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

// InitRelationshipField initialses 1:N relationship fields to empty arrays on
// the hg.businesses table
func (m *BusinessesModel) InitRelationshipField(relationshipId string) error {
	switch relationshipId {
	case "businesses_created_by_id_users_id", "businesses_modified_by_id_users_id":
		return nil
	case "farm_fields_business_id_businesses_id":
		m.FarmFieldsBusinessesBusinessId = []*FarmFieldsModel{}
		return nil
	default:
		return fmt.Errorf("unknown relationship: %s", relationshipId)
	}
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

// GetValueExpression returns the value from a given path as a ValueExpression
func (m *BusinessesModel) GetValueExpression(path []*queryModel.TraversalStep, columnName string, v queryModel.ValueBuilder) (queryModel.ValueExpression, error) {
	if len(path) > 0 {
		nextStep := path[0]
		nextEntity, err := m.getRelatedEntity(nextStep.Relationship)
		if err != nil {
			return nil, err
		}
		if nextEntity.IsNil() {
			return v.Null(), nil
		}
		return nextEntity.GetValueExpression(path[1:], columnName, v)
	}
	val, err := m.getValueExpression(columnName, v)
	if err != nil {
		return nil, err
	}
	if val == nil {
		return v.Null(), nil
	}
	return val, nil
}

// GetRelatedEntity returns the value from N:1/1:1 relationships as a TableModel
// It will return an error for invalid relationships and relationship types
func (m *BusinessesModel) getRelatedEntity(relationship queryModel.RelationshipMetadata) (queryModel.TableModel, error) {
	switch relationship.Id() {
	case "businesses_created_by_id_users_id":
		return m.CreatedBy, nil
	case "businesses_modified_by_id_users_id":
		return m.ModifiedBy, nil
	default:
		return nil, fmt.Errorf("unable to get related entity: unsupported relationship '%s'", relationship.Id())
	}
}

// IsNil is used to determine if a typed nil pointer contains a nil value
func (m *BusinessesModel) IsNil() bool {
	return m == nil
}

// getValueExpression returns the value from a given column as a value expression
func (m *BusinessesModel) getValueExpression(columnName string, v queryModel.ValueBuilder) (queryModel.ValueExpression, error) {
	switch columnName {
	case "id":
		return v.String(m.Id), nil
	case "reference":
		return v.String(m.Reference), nil
	case "trading_name":
		return v.String(m.TradingName), nil
	case "logo_url":
		return v.String(m.LogoUrl), nil
	case "description":
		return v.String(m.Description), nil
	case "business_type":
		return v.Int(m.BusinessType), nil
	case "cph_number":
		return v.String(m.CphNumber), nil
	case "email_address":
		return v.String(m.EmailAddress), nil
	case "contact_number":
		return v.String(m.ContactNumber), nil
	case "website_url":
		return v.String(m.WebsiteUrl), nil
	case "address_line_1":
		return v.String(m.AddressLine1), nil
	case "address_line_2":
		return v.String(m.AddressLine2), nil
	case "town":
		return v.String(m.Town), nil
	case "county":
		return v.String(m.County), nil
	case "country":
		return v.String(m.Country), nil
	case "postcode":
		return v.String(m.Postcode), nil
	case "location":
		if m.Location == nil {
			return v.Null(), nil
		}
		return v.String(m.Location.String()), nil
	case "created_at":
		if m.CreatedAt == nil {
			return v.Null(), nil
		}
		return v.String(m.CreatedAt.String()), nil
	case "created_by_id":
		return v.String(m.CreatedById), nil
	case "modified_at":
		if m.ModifiedAt == nil {
			return v.Null(), nil
		}
		return v.String(m.ModifiedAt.String()), nil
	case "modified_by_id":
		return v.String(m.ModifiedById), nil
	default:
		return nil, fmt.Errorf("unsupported column: '%s'", columnName)
	}
}

// Project adds a column to the model's projection set
func (m *BusinessesModel) Project(columnName string) error {
	return (&m.projection).Add(columnName)
}

// BusinessesProjection represents column projection for the businesses table.
type BusinessesProjection uint64

const (
	businessesProjectionId BusinessesProjection = 1 << iota
	businessesProjectionReference
	businessesProjectionTradingName
	businessesProjectionLogoUrl
	businessesProjectionDescription
	businessesProjectionBusinessType
	businessesProjectionCphNumber
	businessesProjectionEmailAddress
	businessesProjectionContactNumber
	businessesProjectionWebsiteUrl
	businessesProjectionAddressLine1
	businessesProjectionAddressLine2
	businessesProjectionTown
	businessesProjectionCounty
	businessesProjectionCountry
	businessesProjectionPostcode
	businessesProjectionLocation
	businessesProjectionCreatedAt
	businessesProjectionCreatedById
	businessesProjectionModifiedAt
	businessesProjectionModifiedById
	businessesProjectionCreatedBy
	businessesProjectionModifiedBy
	businessesProjectionFarmFieldsBusinessesBusinessId
)

// Add includes a given column within the projection
func (p *BusinessesProjection) Add(columnName string) error {
	switch columnName {
	case "id":
		*p |= businessesProjectionId
	case "reference":
		*p |= businessesProjectionReference
	case "trading_name":
		*p |= businessesProjectionTradingName
	case "logo_url":
		*p |= businessesProjectionLogoUrl
	case "description":
		*p |= businessesProjectionDescription
	case "business_type":
		*p |= businessesProjectionBusinessType
	case "cph_number":
		*p |= businessesProjectionCphNumber
	case "email_address":
		*p |= businessesProjectionEmailAddress
	case "contact_number":
		*p |= businessesProjectionContactNumber
	case "website_url":
		*p |= businessesProjectionWebsiteUrl
	case "address_line_1":
		*p |= businessesProjectionAddressLine1
	case "address_line_2":
		*p |= businessesProjectionAddressLine2
	case "town":
		*p |= businessesProjectionTown
	case "county":
		*p |= businessesProjectionCounty
	case "country":
		*p |= businessesProjectionCountry
	case "postcode":
		*p |= businessesProjectionPostcode
	case "location":
		*p |= businessesProjectionLocation
	case "created_at":
		*p |= businessesProjectionCreatedAt
	case "created_by_id":
		*p |= businessesProjectionCreatedById
	case "modified_at":
		*p |= businessesProjectionModifiedAt
	case "modified_by_id":
		*p |= businessesProjectionModifiedById
	case "farm_fields_businesses_business_id":
		*p |= businessesProjectionFarmFieldsBusinessesBusinessId
	case "created_by":
		*p |= businessesProjectionCreatedBy
	case "modified_by":
		*p |= businessesProjectionModifiedBy
	default:
		return fmt.Errorf("unsupported column: '%s'", columnName)
	}
	return nil
}

// Has is used to determine if a given column is in a projection
func (p BusinessesProjection) Has(projection BusinessesProjection) bool {
	return p&projection != 0
}

// IsEmpty is used to determine if there are no projections
func (p BusinessesProjection) IsEmpty() bool {
	return p == 0
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
			dbType:       queryModel.DbTypeNvarchar,
			maxLength:    72,
			isRequired:   true,
			isPrimaryKey: true,
		},
		"reference": {
			name:         "reference",
			dbType:       queryModel.DbTypeNvarchar,
			maxLength:    510,
			isRequired:   true,
			isPrimaryKey: false,
		},
		"business_id": {
			name:         "business_id",
			dbType:       queryModel.DbTypeNvarchar,
			maxLength:    72,
			isRequired:   true,
			isPrimaryKey: false,
		},
		"location": {
			name:         "location",
			dbType:       queryModel.DbTypeGeography,
			maxLength:    -1,
			isRequired:   true,
			isPrimaryKey: false,
		},
		"created_at": {
			name:         "created_at",
			dbType:       queryModel.DbTypeDateTimeOffset,
			maxLength:    10,
			isRequired:   true,
			isPrimaryKey: false,
		},
		"created_by_id": {
			name:         "created_by_id",
			dbType:       queryModel.DbTypeNvarchar,
			maxLength:    72,
			isRequired:   true,
			isPrimaryKey: false,
		},
		"modified_at": {
			name:         "modified_at",
			dbType:       queryModel.DbTypeDateTimeOffset,
			maxLength:    10,
			isRequired:   true,
			isPrimaryKey: false,
		},
		"modified_by_id": {
			name:         "modified_by_id",
			dbType:       queryModel.DbTypeNvarchar,
			maxLength:    72,
			isRequired:   true,
			isPrimaryKey: false,
		},
	},
	relationships: map[string]*relationship{
		"business_id": {
			id:                  "farm_fields_business_id_businesses_id",
			columnName:          "business_id",
			expansionColumnName: "business",
			relationshipType:    queryModel.RelationshipManyToOne,
			fromTableName:       "farm_fields",
			toTableName:         "businesses",
			fromColumnName:      "business_id",
			toColumnName:        "id",
		},
		"created_by_id": {
			id:                  "farm_fields_created_by_id_users_id",
			columnName:          "created_by_id",
			expansionColumnName: "created_by",
			relationshipType:    queryModel.RelationshipManyToOne,
			fromTableName:       "farm_fields",
			toTableName:         "users",
			fromColumnName:      "created_by_id",
			toColumnName:        "id",
		},
		"modified_by_id": {
			id:                  "farm_fields_modified_by_id_users_id",
			columnName:          "modified_by_id",
			expansionColumnName: "modified_by",
			relationshipType:    queryModel.RelationshipManyToOne,
			fromTableName:       "farm_fields",
			toTableName:         "users",
			fromColumnName:      "modified_by_id",
			toColumnName:        "id",
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
	Id           string               `json:"id"`
	Reference    string               `json:"reference"`
	BusinessId   string               `json:"business_id"`
	Location     *Point               `json:"location"`
	CreatedAt    *time.Time           `json:"created_at"`
	CreatedById  string               `json:"created_by_id"`
	ModifiedAt   *time.Time           `json:"modified_at"`
	ModifiedById string               `json:"modified_by_id"`
	Business     *BusinessesModel     `json:"business"`
	CreatedBy    *UsersModel          `json:"created_by"`
	ModifiedBy   *UsersModel          `json:"modified_by"`
	projection   FarmFieldsProjection `json:"-"`
}

// NewSlice unmarshals a json array of farm_fields and returns it as a slice
func (m *FarmFieldsModel) NewSlice(jsonData []byte, projection queryModel.Projection) ([]queryModel.TableModel, error) {
	if len(jsonData) == 0 {
		return []queryModel.TableModel{}, nil
	}

	var typedProjection *FarmFieldsProjection
	switch p := projection.(type) {
	case *FarmFieldsProjection:
		typedProjection = p
	default:
		return nil, fmt.Errorf("unable to create new slice: invalid projection type received")
	}

	var concreteSlice []*FarmFieldsModel
	if err := json.Unmarshal(jsonData, &concreteSlice); err != nil {
		return nil, err
	}
	result := make([]queryModel.TableModel, len(concreteSlice))

	for i := range concreteSlice {
		concreteSlice[i].projection = *typedProjection
		result[i] = concreteSlice[i]
	}
	return result, nil
}

// MarshalJSON marshals the model to a json string based on the projection
func (m *FarmFieldsModel) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte('{')

	isFirst := true
	if m.projection.Has(farmFieldsProjectionId) {
		err := marshalProperty(&buf, isFirst, "id", m.Id)
		if err != nil {
			return nil, err
		}
		isFirst = false
	}

	if m.projection.Has(farmFieldsProjectionReference) {
		err := marshalProperty(&buf, isFirst, "reference", m.Reference)
		if err != nil {
			return nil, err
		}
		isFirst = false
	}

	if m.projection.Has(farmFieldsProjectionBusinessId) {
		err := marshalProperty(&buf, isFirst, "business_id", m.BusinessId)
		if err != nil {
			return nil, err
		}
		isFirst = false
	}

	if m.projection.Has(farmFieldsProjectionLocation) {
		err := marshalProperty(&buf, isFirst, "location", m.Location)
		if err != nil {
			return nil, err
		}
		isFirst = false
	}

	if m.projection.Has(farmFieldsProjectionCreatedAt) {
		err := marshalProperty(&buf, isFirst, "created_at", m.CreatedAt)
		if err != nil {
			return nil, err
		}
		isFirst = false
	}

	if m.projection.Has(farmFieldsProjectionCreatedById) {
		err := marshalProperty(&buf, isFirst, "created_by_id", m.CreatedById)
		if err != nil {
			return nil, err
		}
		isFirst = false
	}

	if m.projection.Has(farmFieldsProjectionModifiedAt) {
		err := marshalProperty(&buf, isFirst, "modified_at", m.ModifiedAt)
		if err != nil {
			return nil, err
		}
		isFirst = false
	}

	if m.projection.Has(farmFieldsProjectionModifiedById) {
		err := marshalProperty(&buf, isFirst, "modified_by_id", m.ModifiedById)
		if err != nil {
			return nil, err
		}
		isFirst = false
	}

	if m.projection.Has(farmFieldsProjectionBusiness) {
		err := marshalProperty(&buf, isFirst, "business", m.Business)
		if err != nil {
			return nil, err
		}
		isFirst = false
	}

	if m.projection.Has(farmFieldsProjectionCreatedBy) {
		err := marshalProperty(&buf, isFirst, "created_by", m.CreatedBy)
		if err != nil {
			return nil, err
		}
		isFirst = false
	}

	if m.projection.Has(farmFieldsProjectionModifiedBy) {
		err := marshalProperty(&buf, isFirst, "modified_by", m.ModifiedBy)
		if err != nil {
			return nil, err
		}
		isFirst = false
	}

	buf.WriteByte('}')
	return buf.Bytes(), nil
}

// SetRelationshipField sets a given relationship field on the hg.farm_fields table
func (m *FarmFieldsModel) SetRelationshipField(relationshipId string, value queryModel.TableModel) error {
	switch relationshipId {
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

	case "farm_fields_business_id_businesses_id":
		v, ok := value.(*BusinessesModel)
		if !ok {
			return errors.New("unexpected relationship type received")
		}
		m.Business = v

	default:
		return fmt.Errorf("unknown relationship: %s", relationshipId)
	}
	return nil
}

// InitRelationshipField initialses 1:N relationship fields to empty arrays on
// the hg.farm_fields table
func (m *FarmFieldsModel) InitRelationshipField(relationshipId string) error {
	switch relationshipId {
	case "farm_fields_business_id_businesses_id", "farm_fields_created_by_id_users_id", "farm_fields_modified_by_id_users_id":
		return nil
	default:
		return fmt.Errorf("unknown relationship: %s", relationshipId)
	}
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

// GetValueExpression returns the value from a given path as a ValueExpression
func (m *FarmFieldsModel) GetValueExpression(path []*queryModel.TraversalStep, columnName string, v queryModel.ValueBuilder) (queryModel.ValueExpression, error) {
	if len(path) > 0 {
		nextStep := path[0]
		nextEntity, err := m.getRelatedEntity(nextStep.Relationship)
		if err != nil {
			return nil, err
		}
		if nextEntity.IsNil() {
			return v.Null(), nil
		}
		return nextEntity.GetValueExpression(path[1:], columnName, v)
	}
	val, err := m.getValueExpression(columnName, v)
	if err != nil {
		return nil, err
	}
	if val == nil {
		return v.Null(), nil
	}
	return val, nil
}

// GetRelatedEntity returns the value from N:1/1:1 relationships as a TableModel
// It will return an error for invalid relationships and relationship types
func (m *FarmFieldsModel) getRelatedEntity(relationship queryModel.RelationshipMetadata) (queryModel.TableModel, error) {
	switch relationship.Id() {
	case "farm_fields_business_id_businesses_id":
		return m.Business, nil
	case "farm_fields_created_by_id_users_id":
		return m.CreatedBy, nil
	case "farm_fields_modified_by_id_users_id":
		return m.ModifiedBy, nil
	default:
		return nil, fmt.Errorf("unable to get related entity: unsupported relationship '%s'", relationship.Id())
	}
}

// IsNil is used to determine if a typed nil pointer contains a nil value
func (m *FarmFieldsModel) IsNil() bool {
	return m == nil
}

// getValueExpression returns the value from a given column as a value expression
func (m *FarmFieldsModel) getValueExpression(columnName string, v queryModel.ValueBuilder) (queryModel.ValueExpression, error) {
	switch columnName {
	case "id":
		return v.String(m.Id), nil
	case "reference":
		return v.String(m.Reference), nil
	case "business_id":
		return v.String(m.BusinessId), nil
	case "location":
		if m.Location == nil {
			return v.Null(), nil
		}
		return v.String(m.Location.String()), nil
	case "created_at":
		if m.CreatedAt == nil {
			return v.Null(), nil
		}
		return v.String(m.CreatedAt.String()), nil
	case "created_by_id":
		return v.String(m.CreatedById), nil
	case "modified_at":
		if m.ModifiedAt == nil {
			return v.Null(), nil
		}
		return v.String(m.ModifiedAt.String()), nil
	case "modified_by_id":
		return v.String(m.ModifiedById), nil
	default:
		return nil, fmt.Errorf("unsupported column: '%s'", columnName)
	}
}

// Project adds a column to the model's projection set
func (m *FarmFieldsModel) Project(columnName string) error {
	return (&m.projection).Add(columnName)
}

// FarmFieldsProjection represents column projection for the farm_fields table.
type FarmFieldsProjection uint64

const (
	farmFieldsProjectionId FarmFieldsProjection = 1 << iota
	farmFieldsProjectionReference
	farmFieldsProjectionBusinessId
	farmFieldsProjectionLocation
	farmFieldsProjectionCreatedAt
	farmFieldsProjectionCreatedById
	farmFieldsProjectionModifiedAt
	farmFieldsProjectionModifiedById
	farmFieldsProjectionBusiness
	farmFieldsProjectionCreatedBy
	farmFieldsProjectionModifiedBy
)

// Add includes a given column within the projection
func (p *FarmFieldsProjection) Add(columnName string) error {
	switch columnName {
	case "id":
		*p |= farmFieldsProjectionId
	case "reference":
		*p |= farmFieldsProjectionReference
	case "business_id":
		*p |= farmFieldsProjectionBusinessId
	case "location":
		*p |= farmFieldsProjectionLocation
	case "created_at":
		*p |= farmFieldsProjectionCreatedAt
	case "created_by_id":
		*p |= farmFieldsProjectionCreatedById
	case "modified_at":
		*p |= farmFieldsProjectionModifiedAt
	case "modified_by_id":
		*p |= farmFieldsProjectionModifiedById
	case "business":
		*p |= farmFieldsProjectionBusiness
	case "created_by":
		*p |= farmFieldsProjectionCreatedBy
	case "modified_by":
		*p |= farmFieldsProjectionModifiedBy
	default:
		return fmt.Errorf("unsupported column: '%s'", columnName)
	}
	return nil
}

// Has is used to determine if a given column is in a projection
func (p FarmFieldsProjection) Has(projection FarmFieldsProjection) bool {
	return p&projection != 0
}

// IsEmpty is used to determine if there are no projections
func (p FarmFieldsProjection) IsEmpty() bool {
	return p == 0
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
			dbType:       queryModel.DbTypeNvarchar,
			maxLength:    72,
			isRequired:   true,
			isPrimaryKey: true,
		},
		"oid": {
			name:         "oid",
			dbType:       queryModel.DbTypeNvarchar,
			maxLength:    72,
			isRequired:   true,
			isPrimaryKey: false,
		},
		"given_name": {
			name:         "given_name",
			dbType:       queryModel.DbTypeNvarchar,
			maxLength:    128,
			isRequired:   true,
			isPrimaryKey: false,
		},
		"family_name": {
			name:         "family_name",
			dbType:       queryModel.DbTypeNvarchar,
			maxLength:    128,
			isRequired:   true,
			isPrimaryKey: false,
		},
		"user_name": {
			name:         "user_name",
			dbType:       queryModel.DbTypeNvarchar,
			maxLength:    256,
			isRequired:   true,
			isPrimaryKey: false,
		},
		"email_address": {
			name:         "email_address",
			dbType:       queryModel.DbTypeNvarchar,
			maxLength:    640,
			isRequired:   true,
			isPrimaryKey: false,
		},
		"created_at": {
			name:         "created_at",
			dbType:       queryModel.DbTypeDateTimeOffset,
			maxLength:    10,
			isRequired:   true,
			isPrimaryKey: false,
		},
		"modified_at": {
			name:         "modified_at",
			dbType:       queryModel.DbTypeDateTimeOffset,
			maxLength:    10,
			isRequired:   true,
			isPrimaryKey: false,
		},
	},
	relationships: map[string]*relationship{
		"businesses_users_created_by_id": {
			id:                  "businesses_created_by_id_users_id",
			columnName:          "businesses_users_created_by_id",
			expansionColumnName: "businesses_users_created_by_id",
			relationshipType:    queryModel.RelationshipOneToMany,
			fromTableName:       "users",
			toTableName:         "businesses",
			fromColumnName:      "id",
			toColumnName:        "created_by_id",
		},
		"businesses_users_modified_by_id": {
			id:                  "businesses_modified_by_id_users_id",
			columnName:          "businesses_users_modified_by_id",
			expansionColumnName: "businesses_users_modified_by_id",
			relationshipType:    queryModel.RelationshipOneToMany,
			fromTableName:       "users",
			toTableName:         "businesses",
			fromColumnName:      "id",
			toColumnName:        "modified_by_id",
		},
		"farm_fields_users_created_by_id": {
			id:                  "farm_fields_created_by_id_users_id",
			columnName:          "farm_fields_users_created_by_id",
			expansionColumnName: "farm_fields_users_created_by_id",
			relationshipType:    queryModel.RelationshipOneToMany,
			fromTableName:       "users",
			toTableName:         "farm_fields",
			fromColumnName:      "id",
			toColumnName:        "created_by_id",
		},
		"farm_fields_users_modified_by_id": {
			id:                  "farm_fields_modified_by_id_users_id",
			columnName:          "farm_fields_users_modified_by_id",
			expansionColumnName: "farm_fields_users_modified_by_id",
			relationshipType:    queryModel.RelationshipOneToMany,
			fromTableName:       "users",
			toTableName:         "farm_fields",
			fromColumnName:      "id",
			toColumnName:        "modified_by_id",
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
	Id                          string             `json:"id"`
	Oid                         string             `json:"oid"`
	GivenName                   string             `json:"given_name"`
	FamilyName                  string             `json:"family_name"`
	UserName                    string             `json:"user_name"`
	EmailAddress                string             `json:"email_address"`
	CreatedAt                   *time.Time         `json:"created_at"`
	ModifiedAt                  *time.Time         `json:"modified_at"`
	BusinessesUsersCreatedById  []*BusinessesModel `json:"businesses_users_created_by_id"`
	BusinessesUsersModifiedById []*BusinessesModel `json:"businesses_users_modified_by_id"`
	FarmFieldsUsersCreatedById  []*FarmFieldsModel `json:"farm_fields_users_created_by_id"`
	FarmFieldsUsersModifiedById []*FarmFieldsModel `json:"farm_fields_users_modified_by_id"`
	projection                  UsersProjection    `json:"-"`
}

// NewSlice unmarshals a json array of users and returns it as a slice
func (m *UsersModel) NewSlice(jsonData []byte, projection queryModel.Projection) ([]queryModel.TableModel, error) {
	if len(jsonData) == 0 {
		return []queryModel.TableModel{}, nil
	}

	var typedProjection *UsersProjection
	switch p := projection.(type) {
	case *UsersProjection:
		typedProjection = p
	default:
		return nil, fmt.Errorf("unable to create new slice: invalid projection type received")
	}

	var concreteSlice []*UsersModel
	if err := json.Unmarshal(jsonData, &concreteSlice); err != nil {
		return nil, err
	}
	result := make([]queryModel.TableModel, len(concreteSlice))

	for i := range concreteSlice {
		concreteSlice[i].projection = *typedProjection
		result[i] = concreteSlice[i]
	}
	return result, nil
}

// MarshalJSON marshals the model to a json string based on the projection
func (m *UsersModel) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte('{')

	isFirst := true
	if m.projection.Has(usersProjectionId) {
		err := marshalProperty(&buf, isFirst, "id", m.Id)
		if err != nil {
			return nil, err
		}
		isFirst = false
	}

	if m.projection.Has(usersProjectionOid) {
		err := marshalProperty(&buf, isFirst, "oid", m.Oid)
		if err != nil {
			return nil, err
		}
		isFirst = false
	}

	if m.projection.Has(usersProjectionGivenName) {
		err := marshalProperty(&buf, isFirst, "given_name", m.GivenName)
		if err != nil {
			return nil, err
		}
		isFirst = false
	}

	if m.projection.Has(usersProjectionFamilyName) {
		err := marshalProperty(&buf, isFirst, "family_name", m.FamilyName)
		if err != nil {
			return nil, err
		}
		isFirst = false
	}

	if m.projection.Has(usersProjectionUserName) {
		err := marshalProperty(&buf, isFirst, "user_name", m.UserName)
		if err != nil {
			return nil, err
		}
		isFirst = false
	}

	if m.projection.Has(usersProjectionEmailAddress) {
		err := marshalProperty(&buf, isFirst, "email_address", m.EmailAddress)
		if err != nil {
			return nil, err
		}
		isFirst = false
	}

	if m.projection.Has(usersProjectionCreatedAt) {
		err := marshalProperty(&buf, isFirst, "created_at", m.CreatedAt)
		if err != nil {
			return nil, err
		}
		isFirst = false
	}

	if m.projection.Has(usersProjectionModifiedAt) {
		err := marshalProperty(&buf, isFirst, "modified_at", m.ModifiedAt)
		if err != nil {
			return nil, err
		}
		isFirst = false
	}

	if m.projection.Has(usersProjectionBusinessesUsersModifiedById) {
		err := marshalProperty(&buf, isFirst, "businesses_users_modified_by_id", m.BusinessesUsersModifiedById)
		if err != nil {
			return nil, err
		}
		isFirst = false
	}

	if m.projection.Has(usersProjectionFarmFieldsUsersCreatedById) {
		err := marshalProperty(&buf, isFirst, "farm_fields_users_created_by_id", m.FarmFieldsUsersCreatedById)
		if err != nil {
			return nil, err
		}
		isFirst = false
	}

	if m.projection.Has(usersProjectionFarmFieldsUsersModifiedById) {
		err := marshalProperty(&buf, isFirst, "farm_fields_users_modified_by_id", m.FarmFieldsUsersModifiedById)
		if err != nil {
			return nil, err
		}
		isFirst = false
	}

	if m.projection.Has(usersProjectionBusinessesUsersCreatedById) {
		err := marshalProperty(&buf, isFirst, "businesses_users_created_by_id", m.BusinessesUsersCreatedById)
		if err != nil {
			return nil, err
		}
		isFirst = false
	}

	buf.WriteByte('}')
	return buf.Bytes(), nil
}

// SetRelationshipField sets a given relationship field on the hg.users table
func (m *UsersModel) SetRelationshipField(relationshipId string, value queryModel.TableModel) error {
	switch relationshipId {
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

	case "businesses_modified_by_id_users_id":
		v, ok := value.(*BusinessesModel)
		if !ok {
			return errors.New("unexpected relationship type received")
		}
		m.BusinessesUsersModifiedById = append(m.BusinessesUsersModifiedById, v)

	default:
		return fmt.Errorf("unknown relationship: %s", relationshipId)
	}
	return nil
}

// InitRelationshipField initialses 1:N relationship fields to empty arrays on
// the hg.users table
func (m *UsersModel) InitRelationshipField(relationshipId string) error {
	switch relationshipId {
	case "farm_fields_created_by_id_users_id":
		m.FarmFieldsUsersCreatedById = []*FarmFieldsModel{}
		return nil
	case "farm_fields_modified_by_id_users_id":
		m.FarmFieldsUsersModifiedById = []*FarmFieldsModel{}
		return nil
	case "businesses_created_by_id_users_id":
		m.BusinessesUsersCreatedById = []*BusinessesModel{}
		return nil
	case "businesses_modified_by_id_users_id":
		m.BusinessesUsersModifiedById = []*BusinessesModel{}
		return nil
	default:
		return fmt.Errorf("unknown relationship: %s", relationshipId)
	}
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

// GetValueExpression returns the value from a given path as a ValueExpression
func (m *UsersModel) GetValueExpression(path []*queryModel.TraversalStep, columnName string, v queryModel.ValueBuilder) (queryModel.ValueExpression, error) {
	if len(path) > 0 {
		nextStep := path[0]
		nextEntity, err := m.getRelatedEntity(nextStep.Relationship)
		if err != nil {
			return nil, err
		}
		if nextEntity.IsNil() {
			return v.Null(), nil
		}
		return nextEntity.GetValueExpression(path[1:], columnName, v)
	}
	val, err := m.getValueExpression(columnName, v)
	if err != nil {
		return nil, err
	}
	if val == nil {
		return v.Null(), nil
	}
	return val, nil
}

// GetRelatedEntity returns the value from N:1/1:1 relationships as a TableModel
// It will return an error for invalid relationships and relationship types
func (m *UsersModel) getRelatedEntity(relationship queryModel.RelationshipMetadata) (queryModel.TableModel, error) {
	switch relationship.Id() {
	default:
		return nil, fmt.Errorf("unable to get related entity: unsupported relationship '%s'", relationship.Id())
	}
}

// IsNil is used to determine if a typed nil pointer contains a nil value
func (m *UsersModel) IsNil() bool {
	return m == nil
}

// getValueExpression returns the value from a given column as a value expression
func (m *UsersModel) getValueExpression(columnName string, v queryModel.ValueBuilder) (queryModel.ValueExpression, error) {
	switch columnName {
	case "id":
		return v.String(m.Id), nil
	case "oid":
		return v.String(m.Oid), nil
	case "given_name":
		return v.String(m.GivenName), nil
	case "family_name":
		return v.String(m.FamilyName), nil
	case "user_name":
		return v.String(m.UserName), nil
	case "email_address":
		return v.String(m.EmailAddress), nil
	case "created_at":
		if m.CreatedAt == nil {
			return v.Null(), nil
		}
		return v.String(m.CreatedAt.String()), nil
	case "modified_at":
		if m.ModifiedAt == nil {
			return v.Null(), nil
		}
		return v.String(m.ModifiedAt.String()), nil
	default:
		return nil, fmt.Errorf("unsupported column: '%s'", columnName)
	}
}

// Project adds a column to the model's projection set
func (m *UsersModel) Project(columnName string) error {
	return (&m.projection).Add(columnName)
}

// UsersProjection represents column projection for the users table.
type UsersProjection uint64

const (
	usersProjectionId UsersProjection = 1 << iota
	usersProjectionOid
	usersProjectionGivenName
	usersProjectionFamilyName
	usersProjectionUserName
	usersProjectionEmailAddress
	usersProjectionCreatedAt
	usersProjectionModifiedAt
	usersProjectionBusinessesUsersCreatedById
	usersProjectionBusinessesUsersModifiedById
	usersProjectionFarmFieldsUsersCreatedById
	usersProjectionFarmFieldsUsersModifiedById
)

// Add includes a given column within the projection
func (p *UsersProjection) Add(columnName string) error {
	switch columnName {
	case "id":
		*p |= usersProjectionId
	case "oid":
		*p |= usersProjectionOid
	case "given_name":
		*p |= usersProjectionGivenName
	case "family_name":
		*p |= usersProjectionFamilyName
	case "user_name":
		*p |= usersProjectionUserName
	case "email_address":
		*p |= usersProjectionEmailAddress
	case "created_at":
		*p |= usersProjectionCreatedAt
	case "modified_at":
		*p |= usersProjectionModifiedAt
	case "businesses_users_created_by_id":
		*p |= usersProjectionBusinessesUsersCreatedById
	case "businesses_users_modified_by_id":
		*p |= usersProjectionBusinessesUsersModifiedById
	case "farm_fields_users_created_by_id":
		*p |= usersProjectionFarmFieldsUsersCreatedById
	case "farm_fields_users_modified_by_id":
		*p |= usersProjectionFarmFieldsUsersModifiedById
	default:
		return fmt.Errorf("unsupported column: '%s'", columnName)
	}
	return nil
}

// Has is used to determine if a given column is in a projection
func (p UsersProjection) Has(projection UsersProjection) bool {
	return p&projection != 0
}

// IsEmpty is used to determine if there are no projections
func (p UsersProjection) IsEmpty() bool {
	return p == 0
}

// getTableModel returns a new instance of a given table or nill if
// the table name is invalid
func getTableModel(tableName string) queryModel.TableModel {
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
func getTableMetadata(tableName string) queryModel.TableMetadata {
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

// func initTableProjection initialises a projection for the table
func initTableProjection(tableName string) (queryModel.Projection, error) {
	switch tableName {
	case businessesMetadata.name:
		p := BusinessesProjection(0)
		return &p, nil
	case farmFieldsMetadata.name:
		p := FarmFieldsProjection(0)
		return &p, nil
	case usersMetadata.name:
		p := UsersProjection(0)
		return &p, nil
	default:
		return nil, fmt.Errorf("unsupported table: '%s'", tableName)
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
func (p *DatabaseAccessPolicy) GetTableAccessPolicy(tableName string) queryModel.TableAccessPolicy {
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
func (p *BusinessesAccessPolicy) GetColumnAccessPolicy(columnName string) queryModel.ColumnAccessPolicy {
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
func (p *FarmFieldsAccessPolicy) GetColumnAccessPolicy(columnName string) queryModel.ColumnAccessPolicy {
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
func (p *UsersAccessPolicy) GetColumnAccessPolicy(columnName string) queryModel.ColumnAccessPolicy {
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
