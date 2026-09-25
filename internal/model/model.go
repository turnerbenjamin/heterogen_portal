// GENERATED CODE
// see cmd/model-builder/
package model

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/turnerbenjamin/heterogen_portal/internal/query/queryModel"
)

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

type schema struct{}

func NewSchema() schema {
	return schema{}
}
func (s schema) GetResource(resourceName string) (queryModel.Resource, bool) {
	switch resourceName {
	case "businesses":
		return &BusinessesResource{}, true
	case "farm_fields":
		return &FarmFieldsResource{}, true
	case "users":
		return &UsersResource{}, true
	default:
		return nil, false
	}
}

type BusinessesResource struct{}

func (r *BusinessesResource) GetMetadata() queryModel.TableData {
	return businessesMetadata
}

func (r *BusinessesResource) InitModel() queryModel.TableModel {
	return new(BusinessesModel)
}

func (r *BusinessesResource) InitProjection() queryModel.Projection {
	return new(BusinessesProjection)
}

// businessesMetadata contains the database metadata for the hg.businesses table.
var businessesMetadata = queryModel.TableData{
	Name:               "businesses",
	FullyQualifiedName: "hg.businesses",
	PrimaryKeyColumn: queryModel.ColumnData{
		Name: "id",
		Type: queryModel.DbTypeString,
	},
	Columns: map[string]queryModel.ColumnData{
		"id": {
			Name: "id",
			Type: queryModel.DbTypeString,
		},
		"reference": {
			Name: "reference",
			Type: queryModel.DbTypeString,
		},
		"trading_name": {
			Name: "trading_name",
			Type: queryModel.DbTypeString,
		},
		"logo_url": {
			Name: "logo_url",
			Type: queryModel.DbTypeString,
		},
		"description": {
			Name: "description",
			Type: queryModel.DbTypeString,
		},
		"business_type": {
			Name: "business_type",
			Type: queryModel.DbTypeInt,
		},
		"cph_number": {
			Name: "cph_number",
			Type: queryModel.DbTypeString,
		},
		"email_address": {
			Name: "email_address",
			Type: queryModel.DbTypeString,
		},
		"contact_number": {
			Name: "contact_number",
			Type: queryModel.DbTypeString,
		},
		"website_url": {
			Name: "website_url",
			Type: queryModel.DbTypeString,
		},
		"address_line_1": {
			Name: "address_line_1",
			Type: queryModel.DbTypeString,
		},
		"address_line_2": {
			Name: "address_line_2",
			Type: queryModel.DbTypeString,
		},
		"town": {
			Name: "town",
			Type: queryModel.DbTypeString,
		},
		"county": {
			Name: "county",
			Type: queryModel.DbTypeString,
		},
		"country": {
			Name: "country",
			Type: queryModel.DbTypeString,
		},
		"postcode": {
			Name: "postcode",
			Type: queryModel.DbTypeString,
		},
		"location": {
			Name: "location",
			Type: queryModel.DbTypePoint,
		},
		"created_at": {
			Name: "created_at",
			Type: queryModel.DbTypeDateTime,
		},
		"created_by_id": {
			Name: "created_by_id",
			Type: queryModel.DbTypeString,
		},
		"modified_at": {
			Name: "modified_at",
			Type: queryModel.DbTypeDateTime,
		},
		"modified_by_id": {
			Name: "modified_by_id",
			Type: queryModel.DbTypeString,
		},
	},
	Relationships: map[string]queryModel.RelationshipData{
		"created_by_id": {
			Id:                  "businesses_created_by_id_users_id",
			Type:                queryModel.RelationshipManyToOne,
			ColumnName:          "created_by_id",
			ExpansionColumnName: "created_by",
			From:                &BusinessesResource{},
			To:                  &UsersResource{},
			FromColumn: queryModel.ColumnData{
				Name: "created_by_id",
				Type: queryModel.DbTypeString,
			},
			ToColumn: queryModel.ColumnData{
				Name: "id",
				Type: queryModel.DbTypeString,
			},
		},
		"modified_by_id": {
			Id:                  "businesses_modified_by_id_users_id",
			Type:                queryModel.RelationshipManyToOne,
			ColumnName:          "modified_by_id",
			ExpansionColumnName: "modified_by",
			From:                &BusinessesResource{},
			To:                  &UsersResource{},
			FromColumn: queryModel.ColumnData{
				Name: "modified_by_id",
				Type: queryModel.DbTypeString,
			},
			ToColumn: queryModel.ColumnData{
				Name: "id",
				Type: queryModel.DbTypeString,
			},
		},
		"farm_fields_businesses_business_id": {
			Id:                  "farm_fields_business_id_businesses_id",
			Type:                queryModel.RelationshipOneToMany,
			ColumnName:          "farm_fields_businesses_business_id",
			ExpansionColumnName: "farm_fields_businesses_business_id",
			From:                &BusinessesResource{},
			To:                  &FarmFieldsResource{},
			FromColumn: queryModel.ColumnData{
				Name: "id",
				Type: queryModel.DbTypeString,
			},
			ToColumn: queryModel.ColumnData{
				Name: "business_id",
				Type: queryModel.DbTypeString,
			},
		},
	},
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
	Location                       *queryModel.Point    `json:"location"`
	CreatedAt                      *time.Time           `json:"created_at"`
	CreatedById                    string               `json:"created_by_id"`
	ModifiedAt                     *time.Time           `json:"modified_at"`
	ModifiedById                   string               `json:"modified_by_id"`
	FarmFieldsBusinessesBusinessId FarmFieldsModels     `json:"farm_fields_businesses_business_id"`
	CreatedBy                      *UsersModel          `json:"created_by"`
	ModifiedBy                     *UsersModel          `json:"modified_by"`
	projection                     BusinessesProjection `json:"-"`
}

type BusinessesModels []*BusinessesModel

// NewSlice unmarshals a json array of businesses and returns it as a slice
func (m *BusinessesModel) NewSlice(jsonData []byte, projectionNode *queryModel.ProjectionNode) ([]queryModel.TableModel, error) {
	if len(jsonData) == 0 {
		return []queryModel.TableModel{}, nil
	}

	var concreteSlice []*BusinessesModel
	if err := json.Unmarshal(jsonData, &concreteSlice); err != nil {
		return nil, err
	}
	result := make([]queryModel.TableModel, len(concreteSlice))

	for i := range concreteSlice {
		concreteSlice[i].SetProjection(projectionNode)
		result[i] = concreteSlice[i]
	}
	return result, nil
}

// SetProjection sets a projection node and sets projection for itself and any
// child nodes
func (m *BusinessesModel) SetProjection(projectionNode *queryModel.ProjectionNode) error {
	projection := projectionNode.Projection

	var typedProjection *BusinessesProjection
	switch p := projection.(type) {
	case *BusinessesProjection:
		typedProjection = p
	default:
		return fmt.Errorf("unable to create new slice: invalid projection type received")
	}

	return m.setBusinessesProjection(*typedProjection, projectionNode.Children)
}

// SetProjection sets a projection node and sets projection for itself and any
// child nodes
func (ms BusinessesModels) SetProjection(projectionNode *queryModel.ProjectionNode) error {
	projection := projectionNode.Projection

	var typedProjection *BusinessesProjection
	switch p := projection.(type) {
	case *BusinessesProjection:
		typedProjection = p
	default:
		return fmt.Errorf("unable to create new slice: invalid projection type received")
	}

	for _, m := range ms {
		if err := m.setBusinessesProjection(*typedProjection, projectionNode.Children); err != nil {
			return err
		}
	}
	return nil
}

// setBusinessesProjection sets a projection node and sets projection for itself and any
// child nodes
func (m *BusinessesModel) setBusinessesProjection(projection BusinessesProjection, childNodes map[string]*queryModel.ProjectionNode) error {
	m.projection = projection
	if node, exists := childNodes["created_by"]; exists {
		if m.CreatedBy != nil {
			if err := m.CreatedBy.SetProjection(node); err != nil {
				return err
			}
		}
	}

	if node, exists := childNodes["modified_by"]; exists {
		if m.ModifiedBy != nil {
			if err := m.ModifiedBy.SetProjection(node); err != nil {
				return err
			}
		}
	}

	if node, exists := childNodes["farm_fields_businesses_business_id"]; exists {
		if m.FarmFieldsBusinessesBusinessId != nil {
			if err := m.FarmFieldsBusinessesBusinessId.SetProjection(node); err != nil {
				return err
			}
		} else {
			m.FarmFieldsBusinessesBusinessId = FarmFieldsModels{}
		}
	}

	return nil
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

// GetValue returns the value from a given path
func (m *BusinessesModel) GetValue(path []*queryModel.TraversalStep, columnName string, v queryModel.ValueBuilder) (queryModel.Value, error) {
	if len(path) > 0 {
		nextStep := path[0]
		nextEntity, nextEntityIsNil, err := m.getRelatedEntity(nextStep.Relationship)
		if err != nil {
			return nil, err
		}
		if nextEntityIsNil {
			return v.Null(), nil
		}
		return nextEntity.GetValue(path[1:], columnName, v)
	}
	val, err := m.getValue(columnName, v)
	if err != nil {
		return nil, err
	}
	if val == nil {
		return v.Null(), nil
	}
	return val, nil
}

// getRelatedEntity returns the value from N:1/1:1 relationships as a TableModel
// It will return an error for invalid relationships and relationship types
func (m *BusinessesModel) getRelatedEntity(relationship queryModel.RelationshipData) (queryModel.TableModel, bool, error) {
	switch relationship.Id {
	case "businesses_created_by_id_users_id":
		v := m.CreatedBy
		return v, v == nil, nil
	case "businesses_modified_by_id_users_id":
		v := m.ModifiedBy
		return v, v == nil, nil
	default:
		return nil, true, fmt.Errorf("unable to get related entity: unsupported relationship '%s'", relationship.Id)
	}
}

// getValue returns the value from a given column
func (m *BusinessesModel) getValue(columnName string, v queryModel.ValueBuilder) (queryModel.Value, error) {
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
		return v.Point(m.Location.Coordinates[0], m.Location.Coordinates[1]), nil
	case "created_at":
		return v.DateTime(*m.CreatedAt), nil
	case "created_by_id":
		return v.String(m.CreatedById), nil
	case "modified_at":
		return v.DateTime(*m.ModifiedAt), nil
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
	case "created_by":
		*p |= businessesProjectionCreatedBy
	case "modified_by":
		*p |= businessesProjectionModifiedBy
	case "farm_fields_businesses_business_id":
		*p |= businessesProjectionFarmFieldsBusinessesBusinessId
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

type FarmFieldsResource struct{}

func (r *FarmFieldsResource) GetMetadata() queryModel.TableData {
	return farmFieldsMetadata
}

func (r *FarmFieldsResource) InitModel() queryModel.TableModel {
	return new(FarmFieldsModel)
}

func (r *FarmFieldsResource) InitProjection() queryModel.Projection {
	return new(FarmFieldsProjection)
}

// farmFieldsMetadata contains the database metadata for the hg.farm_fields table.
var farmFieldsMetadata = queryModel.TableData{
	Name:               "farm_fields",
	FullyQualifiedName: "hg.farm_fields",
	PrimaryKeyColumn: queryModel.ColumnData{
		Name: "id",
		Type: queryModel.DbTypeString,
	},
	Columns: map[string]queryModel.ColumnData{
		"id": {
			Name: "id",
			Type: queryModel.DbTypeString,
		},
		"reference": {
			Name: "reference",
			Type: queryModel.DbTypeString,
		},
		"business_id": {
			Name: "business_id",
			Type: queryModel.DbTypeString,
		},
		"location": {
			Name: "location",
			Type: queryModel.DbTypePoint,
		},
		"created_at": {
			Name: "created_at",
			Type: queryModel.DbTypeDateTime,
		},
		"created_by_id": {
			Name: "created_by_id",
			Type: queryModel.DbTypeString,
		},
		"modified_at": {
			Name: "modified_at",
			Type: queryModel.DbTypeDateTime,
		},
		"modified_by_id": {
			Name: "modified_by_id",
			Type: queryModel.DbTypeString,
		},
	},
	Relationships: map[string]queryModel.RelationshipData{
		"business_id": {
			Id:                  "farm_fields_business_id_businesses_id",
			Type:                queryModel.RelationshipManyToOne,
			ColumnName:          "business_id",
			ExpansionColumnName: "business",
			From:                &FarmFieldsResource{},
			To:                  &BusinessesResource{},
			FromColumn: queryModel.ColumnData{
				Name: "business_id",
				Type: queryModel.DbTypeString,
			},
			ToColumn: queryModel.ColumnData{
				Name: "id",
				Type: queryModel.DbTypeString,
			},
		},
		"created_by_id": {
			Id:                  "farm_fields_created_by_id_users_id",
			Type:                queryModel.RelationshipManyToOne,
			ColumnName:          "created_by_id",
			ExpansionColumnName: "created_by",
			From:                &FarmFieldsResource{},
			To:                  &UsersResource{},
			FromColumn: queryModel.ColumnData{
				Name: "created_by_id",
				Type: queryModel.DbTypeString,
			},
			ToColumn: queryModel.ColumnData{
				Name: "id",
				Type: queryModel.DbTypeString,
			},
		},
		"modified_by_id": {
			Id:                  "farm_fields_modified_by_id_users_id",
			Type:                queryModel.RelationshipManyToOne,
			ColumnName:          "modified_by_id",
			ExpansionColumnName: "modified_by",
			From:                &FarmFieldsResource{},
			To:                  &UsersResource{},
			FromColumn: queryModel.ColumnData{
				Name: "modified_by_id",
				Type: queryModel.DbTypeString,
			},
			ToColumn: queryModel.ColumnData{
				Name: "id",
				Type: queryModel.DbTypeString,
			},
		},
	},
}

// FarmFieldsModel represents a row from the hg.farm_fields table.
type FarmFieldsModel struct {
	Id           string               `json:"id"`
	Reference    string               `json:"reference"`
	BusinessId   string               `json:"business_id"`
	Location     *queryModel.Point    `json:"location"`
	CreatedAt    *time.Time           `json:"created_at"`
	CreatedById  string               `json:"created_by_id"`
	ModifiedAt   *time.Time           `json:"modified_at"`
	ModifiedById string               `json:"modified_by_id"`
	Business     *BusinessesModel     `json:"business"`
	CreatedBy    *UsersModel          `json:"created_by"`
	ModifiedBy   *UsersModel          `json:"modified_by"`
	projection   FarmFieldsProjection `json:"-"`
}

type FarmFieldsModels []*FarmFieldsModel

// NewSlice unmarshals a json array of farm_fields and returns it as a slice
func (m *FarmFieldsModel) NewSlice(jsonData []byte, projectionNode *queryModel.ProjectionNode) ([]queryModel.TableModel, error) {
	if len(jsonData) == 0 {
		return []queryModel.TableModel{}, nil
	}

	var concreteSlice []*FarmFieldsModel
	if err := json.Unmarshal(jsonData, &concreteSlice); err != nil {
		return nil, err
	}
	result := make([]queryModel.TableModel, len(concreteSlice))

	for i := range concreteSlice {
		concreteSlice[i].SetProjection(projectionNode)
		result[i] = concreteSlice[i]
	}
	return result, nil
}

// SetProjection sets a projection node and sets projection for itself and any
// child nodes
func (m *FarmFieldsModel) SetProjection(projectionNode *queryModel.ProjectionNode) error {
	projection := projectionNode.Projection

	var typedProjection *FarmFieldsProjection
	switch p := projection.(type) {
	case *FarmFieldsProjection:
		typedProjection = p
	default:
		return fmt.Errorf("unable to create new slice: invalid projection type received")
	}

	return m.setFarmFieldsProjection(*typedProjection, projectionNode.Children)
}

// SetProjection sets a projection node and sets projection for itself and any
// child nodes
func (ms FarmFieldsModels) SetProjection(projectionNode *queryModel.ProjectionNode) error {
	projection := projectionNode.Projection

	var typedProjection *FarmFieldsProjection
	switch p := projection.(type) {
	case *FarmFieldsProjection:
		typedProjection = p
	default:
		return fmt.Errorf("unable to create new slice: invalid projection type received")
	}

	for _, m := range ms {
		if err := m.setFarmFieldsProjection(*typedProjection, projectionNode.Children); err != nil {
			return err
		}
	}
	return nil
}

// setFarmFieldsProjection sets a projection node and sets projection for itself and any
// child nodes
func (m *FarmFieldsModel) setFarmFieldsProjection(projection FarmFieldsProjection, childNodes map[string]*queryModel.ProjectionNode) error {
	m.projection = projection
	if node, exists := childNodes["business"]; exists {
		if m.Business != nil {
			if err := m.Business.SetProjection(node); err != nil {
				return err
			}
		}
	}

	if node, exists := childNodes["created_by"]; exists {
		if m.CreatedBy != nil {
			if err := m.CreatedBy.SetProjection(node); err != nil {
				return err
			}
		}
	}

	if node, exists := childNodes["modified_by"]; exists {
		if m.ModifiedBy != nil {
			if err := m.ModifiedBy.SetProjection(node); err != nil {
				return err
			}
		}
	}

	return nil
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

	if m.projection.Has(farmFieldsProjectionBusiness) {
		err := marshalProperty(&buf, isFirst, "business", m.Business)
		if err != nil {
			return nil, err
		}
		isFirst = false
	}

	buf.WriteByte('}')
	return buf.Bytes(), nil
}

// GetValue returns the value from a given path
func (m *FarmFieldsModel) GetValue(path []*queryModel.TraversalStep, columnName string, v queryModel.ValueBuilder) (queryModel.Value, error) {
	if len(path) > 0 {
		nextStep := path[0]
		nextEntity, nextEntityIsNil, err := m.getRelatedEntity(nextStep.Relationship)
		if err != nil {
			return nil, err
		}
		if nextEntityIsNil {
			return v.Null(), nil
		}
		return nextEntity.GetValue(path[1:], columnName, v)
	}
	val, err := m.getValue(columnName, v)
	if err != nil {
		return nil, err
	}
	if val == nil {
		return v.Null(), nil
	}
	return val, nil
}

// getRelatedEntity returns the value from N:1/1:1 relationships as a TableModel
// It will return an error for invalid relationships and relationship types
func (m *FarmFieldsModel) getRelatedEntity(relationship queryModel.RelationshipData) (queryModel.TableModel, bool, error) {
	switch relationship.Id {
	case "farm_fields_created_by_id_users_id":
		v := m.CreatedBy
		return v, v == nil, nil
	case "farm_fields_modified_by_id_users_id":
		v := m.ModifiedBy
		return v, v == nil, nil
	case "farm_fields_business_id_businesses_id":
		v := m.Business
		return v, v == nil, nil
	default:
		return nil, true, fmt.Errorf("unable to get related entity: unsupported relationship '%s'", relationship.Id)
	}
}

// getValue returns the value from a given column
func (m *FarmFieldsModel) getValue(columnName string, v queryModel.ValueBuilder) (queryModel.Value, error) {
	switch columnName {
	case "id":
		return v.String(m.Id), nil
	case "reference":
		return v.String(m.Reference), nil
	case "business_id":
		return v.String(m.BusinessId), nil
	case "location":
		return v.Point(m.Location.Coordinates[0], m.Location.Coordinates[1]), nil
	case "created_at":
		return v.DateTime(*m.CreatedAt), nil
	case "created_by_id":
		return v.String(m.CreatedById), nil
	case "modified_at":
		return v.DateTime(*m.ModifiedAt), nil
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

type UsersResource struct{}

func (r *UsersResource) GetMetadata() queryModel.TableData {
	return usersMetadata
}

func (r *UsersResource) InitModel() queryModel.TableModel {
	return new(UsersModel)
}

func (r *UsersResource) InitProjection() queryModel.Projection {
	return new(UsersProjection)
}

// usersMetadata contains the database metadata for the hg.users table.
var usersMetadata = queryModel.TableData{
	Name:               "users",
	FullyQualifiedName: "hg.users",
	PrimaryKeyColumn: queryModel.ColumnData{
		Name: "id",
		Type: queryModel.DbTypeString,
	},
	Columns: map[string]queryModel.ColumnData{
		"id": {
			Name: "id",
			Type: queryModel.DbTypeString,
		},
		"oid": {
			Name: "oid",
			Type: queryModel.DbTypeString,
		},
		"given_name": {
			Name: "given_name",
			Type: queryModel.DbTypeString,
		},
		"family_name": {
			Name: "family_name",
			Type: queryModel.DbTypeString,
		},
		"user_name": {
			Name: "user_name",
			Type: queryModel.DbTypeString,
		},
		"email_address": {
			Name: "email_address",
			Type: queryModel.DbTypeString,
		},
		"created_at": {
			Name: "created_at",
			Type: queryModel.DbTypeDateTime,
		},
		"modified_at": {
			Name: "modified_at",
			Type: queryModel.DbTypeDateTime,
		},
	},
	Relationships: map[string]queryModel.RelationshipData{
		"businesses_users_created_by_id": {
			Id:                  "businesses_created_by_id_users_id",
			Type:                queryModel.RelationshipOneToMany,
			ColumnName:          "businesses_users_created_by_id",
			ExpansionColumnName: "businesses_users_created_by_id",
			From:                &UsersResource{},
			To:                  &BusinessesResource{},
			FromColumn: queryModel.ColumnData{
				Name: "id",
				Type: queryModel.DbTypeString,
			},
			ToColumn: queryModel.ColumnData{
				Name: "created_by_id",
				Type: queryModel.DbTypeString,
			},
		},
		"businesses_users_modified_by_id": {
			Id:                  "businesses_modified_by_id_users_id",
			Type:                queryModel.RelationshipOneToMany,
			ColumnName:          "businesses_users_modified_by_id",
			ExpansionColumnName: "businesses_users_modified_by_id",
			From:                &UsersResource{},
			To:                  &BusinessesResource{},
			FromColumn: queryModel.ColumnData{
				Name: "id",
				Type: queryModel.DbTypeString,
			},
			ToColumn: queryModel.ColumnData{
				Name: "modified_by_id",
				Type: queryModel.DbTypeString,
			},
		},
		"farm_fields_users_created_by_id": {
			Id:                  "farm_fields_created_by_id_users_id",
			Type:                queryModel.RelationshipOneToMany,
			ColumnName:          "farm_fields_users_created_by_id",
			ExpansionColumnName: "farm_fields_users_created_by_id",
			From:                &UsersResource{},
			To:                  &FarmFieldsResource{},
			FromColumn: queryModel.ColumnData{
				Name: "id",
				Type: queryModel.DbTypeString,
			},
			ToColumn: queryModel.ColumnData{
				Name: "created_by_id",
				Type: queryModel.DbTypeString,
			},
		},
		"farm_fields_users_modified_by_id": {
			Id:                  "farm_fields_modified_by_id_users_id",
			Type:                queryModel.RelationshipOneToMany,
			ColumnName:          "farm_fields_users_modified_by_id",
			ExpansionColumnName: "farm_fields_users_modified_by_id",
			From:                &UsersResource{},
			To:                  &FarmFieldsResource{},
			FromColumn: queryModel.ColumnData{
				Name: "id",
				Type: queryModel.DbTypeString,
			},
			ToColumn: queryModel.ColumnData{
				Name: "modified_by_id",
				Type: queryModel.DbTypeString,
			},
		},
	},
}

// UsersModel represents a row from the hg.users table.
type UsersModel struct {
	Id                          string           `json:"id"`
	Oid                         string           `json:"oid"`
	GivenName                   string           `json:"given_name"`
	FamilyName                  string           `json:"family_name"`
	UserName                    string           `json:"user_name"`
	EmailAddress                string           `json:"email_address"`
	CreatedAt                   *time.Time       `json:"created_at"`
	ModifiedAt                  *time.Time       `json:"modified_at"`
	BusinessesUsersModifiedById BusinessesModels `json:"businesses_users_modified_by_id"`
	FarmFieldsUsersCreatedById  FarmFieldsModels `json:"farm_fields_users_created_by_id"`
	FarmFieldsUsersModifiedById FarmFieldsModels `json:"farm_fields_users_modified_by_id"`
	BusinessesUsersCreatedById  BusinessesModels `json:"businesses_users_created_by_id"`
	projection                  UsersProjection  `json:"-"`
}

type UsersModels []*UsersModel

// NewSlice unmarshals a json array of users and returns it as a slice
func (m *UsersModel) NewSlice(jsonData []byte, projectionNode *queryModel.ProjectionNode) ([]queryModel.TableModel, error) {
	if len(jsonData) == 0 {
		return []queryModel.TableModel{}, nil
	}

	var concreteSlice []*UsersModel
	if err := json.Unmarshal(jsonData, &concreteSlice); err != nil {
		return nil, err
	}
	result := make([]queryModel.TableModel, len(concreteSlice))

	for i := range concreteSlice {
		concreteSlice[i].SetProjection(projectionNode)
		result[i] = concreteSlice[i]
	}
	return result, nil
}

// SetProjection sets a projection node and sets projection for itself and any
// child nodes
func (m *UsersModel) SetProjection(projectionNode *queryModel.ProjectionNode) error {
	projection := projectionNode.Projection

	var typedProjection *UsersProjection
	switch p := projection.(type) {
	case *UsersProjection:
		typedProjection = p
	default:
		return fmt.Errorf("unable to create new slice: invalid projection type received")
	}

	return m.setUsersProjection(*typedProjection, projectionNode.Children)
}

// SetProjection sets a projection node and sets projection for itself and any
// child nodes
func (ms UsersModels) SetProjection(projectionNode *queryModel.ProjectionNode) error {
	projection := projectionNode.Projection

	var typedProjection *UsersProjection
	switch p := projection.(type) {
	case *UsersProjection:
		typedProjection = p
	default:
		return fmt.Errorf("unable to create new slice: invalid projection type received")
	}

	for _, m := range ms {
		if err := m.setUsersProjection(*typedProjection, projectionNode.Children); err != nil {
			return err
		}
	}
	return nil
}

// setUsersProjection sets a projection node and sets projection for itself and any
// child nodes
func (m *UsersModel) setUsersProjection(projection UsersProjection, childNodes map[string]*queryModel.ProjectionNode) error {
	m.projection = projection
	if node, exists := childNodes["businesses_users_modified_by_id"]; exists {
		if m.BusinessesUsersModifiedById != nil {
			if err := m.BusinessesUsersModifiedById.SetProjection(node); err != nil {
				return err
			}
		} else {
			m.BusinessesUsersModifiedById = BusinessesModels{}
		}
	}

	if node, exists := childNodes["farm_fields_users_created_by_id"]; exists {
		if m.FarmFieldsUsersCreatedById != nil {
			if err := m.FarmFieldsUsersCreatedById.SetProjection(node); err != nil {
				return err
			}
		} else {
			m.FarmFieldsUsersCreatedById = FarmFieldsModels{}
		}
	}

	if node, exists := childNodes["farm_fields_users_modified_by_id"]; exists {
		if m.FarmFieldsUsersModifiedById != nil {
			if err := m.FarmFieldsUsersModifiedById.SetProjection(node); err != nil {
				return err
			}
		} else {
			m.FarmFieldsUsersModifiedById = FarmFieldsModels{}
		}
	}

	if node, exists := childNodes["businesses_users_created_by_id"]; exists {
		if m.BusinessesUsersCreatedById != nil {
			if err := m.BusinessesUsersCreatedById.SetProjection(node); err != nil {
				return err
			}
		} else {
			m.BusinessesUsersCreatedById = BusinessesModels{}
		}
	}

	return nil
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

	if m.projection.Has(usersProjectionBusinessesUsersCreatedById) {
		err := marshalProperty(&buf, isFirst, "businesses_users_created_by_id", m.BusinessesUsersCreatedById)
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

	buf.WriteByte('}')
	return buf.Bytes(), nil
}

// GetValue returns the value from a given path
func (m *UsersModel) GetValue(path []*queryModel.TraversalStep, columnName string, v queryModel.ValueBuilder) (queryModel.Value, error) {
	if len(path) > 0 {
		nextStep := path[0]
		nextEntity, nextEntityIsNil, err := m.getRelatedEntity(nextStep.Relationship)
		if err != nil {
			return nil, err
		}
		if nextEntityIsNil {
			return v.Null(), nil
		}
		return nextEntity.GetValue(path[1:], columnName, v)
	}
	val, err := m.getValue(columnName, v)
	if err != nil {
		return nil, err
	}
	if val == nil {
		return v.Null(), nil
	}
	return val, nil
}

// getRelatedEntity returns the value from N:1/1:1 relationships as a TableModel
// It will return an error for invalid relationships and relationship types
func (m *UsersModel) getRelatedEntity(relationship queryModel.RelationshipData) (queryModel.TableModel, bool, error) {
	switch relationship.Id {
	default:
		return nil, true, fmt.Errorf("unable to get related entity: unsupported relationship '%s'", relationship.Id)
	}
}

// getValue returns the value from a given column
func (m *UsersModel) getValue(columnName string, v queryModel.ValueBuilder) (queryModel.Value, error) {
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
		return v.DateTime(*m.CreatedAt), nil
	case "modified_at":
		return v.DateTime(*m.ModifiedAt), nil
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

// AccessPolicy defines the access policy for the database
type DatabaseAccessPolicy struct {
	BusinessesAccessPolicy *BusinessesAccessPolicy
	FarmFieldsAccessPolicy *FarmFieldsAccessPolicy
	UsersAccessPolicy      *UsersAccessPolicy
}

// GetTableAccessPolicy returns an access policy for a given table for nil
// if the table does not exist
func (p *DatabaseAccessPolicy) GetTableAccessPolicy(
	tableName string,
) (queryModel.TableAccessPolicy, bool) {
	switch tableName {
	case "businesses":
		return p.BusinessesAccessPolicy, true
	case "farm_fields":
		return p.FarmFieldsAccessPolicy, true
	case "users":
		return p.UsersAccessPolicy, true
	default:
		return nil, false
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
func (p *BusinessesAccessPolicy) GetColumnAccessPolicy(
	columnName string,
) (queryModel.ColumnAccessPolicy, bool) {
	switch columnName {
	case "id":
		return p.Id, true
	case "reference":
		return p.Reference, true
	case "trading_name":
		return p.TradingName, true
	case "logo_url":
		return p.LogoUrl, true
	case "description":
		return p.Description, true
	case "business_type":
		return p.BusinessType, true
	case "cph_number":
		return p.CphNumber, true
	case "email_address":
		return p.EmailAddress, true
	case "contact_number":
		return p.ContactNumber, true
	case "website_url":
		return p.WebsiteUrl, true
	case "address_line_1":
		return p.AddressLine1, true
	case "address_line_2":
		return p.AddressLine2, true
	case "town":
		return p.Town, true
	case "county":
		return p.County, true
	case "country":
		return p.Country, true
	case "postcode":
		return p.Postcode, true
	case "location":
		return p.Location, true
	case "created_at":
		return p.CreatedAt, true
	case "created_by_id":
		return p.CreatedById, true
	case "modified_at":
		return p.ModifiedAt, true
	case "modified_by_id":
		return p.ModifiedById, true
	default:
		return nil, false
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
func (p *FarmFieldsAccessPolicy) GetColumnAccessPolicy(
	columnName string,
) (queryModel.ColumnAccessPolicy, bool) {
	switch columnName {
	case "id":
		return p.Id, true
	case "reference":
		return p.Reference, true
	case "business_id":
		return p.BusinessId, true
	case "location":
		return p.Location, true
	case "created_at":
		return p.CreatedAt, true
	case "created_by_id":
		return p.CreatedById, true
	case "modified_at":
		return p.ModifiedAt, true
	case "modified_by_id":
		return p.ModifiedById, true
	default:
		return nil, false
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
func (p *UsersAccessPolicy) GetColumnAccessPolicy(
	columnName string,
) (queryModel.ColumnAccessPolicy, bool) {
	switch columnName {
	case "id":
		return p.Id, true
	case "oid":
		return p.Oid, true
	case "given_name":
		return p.GivenName, true
	case "family_name":
		return p.FamilyName, true
	case "user_name":
		return p.UserName, true
	case "email_address":
		return p.EmailAddress, true
	case "created_at":
		return p.CreatedAt, true
	case "modified_at":
		return p.ModifiedAt, true
	default:
		return nil, false
	}
}
