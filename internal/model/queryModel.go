// GENERATED CODE
// see cmd/modelbuilder/
package model

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	mssql "github.com/microsoft/go-mssqldb"
	"github.com/turnerbenjamin/querystack/querymodel"
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

type schema struct{}

func NewSchema() schema {
	return schema{}
}

func (s schema) GetResource(resourceName string) (querymodel.Resource, bool) {
	switch resourceName {
	case "businesses":
		return BusinessesResource{}, true
	case "farm_fields":
		return FarmFieldsResource{}, true
	case "users":
		return UsersResource{}, true
	default:
		return nil, false
	}
}

type BusinessesResource struct{}

func (r BusinessesResource) GetMetadata() querymodel.TableMetadata {
	return BusinessesMetadata
}

func (r BusinessesResource) InitProjection() querymodel.Projection {
	return new(BusinessesProjection)
}

// SliceFromJSON unmarshals a json array of businesses and returns it as a slice
func (m BusinessesResource) SliceFromJSON(
	jsonData []byte,
	projectionNode *querymodel.ProjectionNode,
) ([]querymodel.TableModel, error) {
	if len(jsonData) == 0 {
		return []querymodel.TableModel{}, nil
	}

	var concreteSlice []*BusinessesModel
	if err := json.Unmarshal(jsonData, &concreteSlice); err != nil {
		return nil, err
	}
	result := make([]querymodel.TableModel, len(concreteSlice))

	for i := range concreteSlice {
		if err := concreteSlice[i].SetProjection(projectionNode); err != nil {
			return nil, err
		}
		result[i] = concreteSlice[i]
	}
	return result, nil
}

// TypedSliceFromJSON unmarshals a json array of businesses and returns it as a slice

func (m BusinessesResource) TypedSliceFromJSON(
	jsonData []byte,
	projectionNode *querymodel.ProjectionNode,
) ([]*BusinessesModel, error) {
	if len(jsonData) == 0 {
		return []*BusinessesModel{}, nil
	}

	var results []*BusinessesModel
	if err := json.Unmarshal(jsonData, &results); err != nil {
		return nil, err
	}
	for _, result := range results {
		if err := result.SetProjection(projectionNode); err != nil {
			return nil, err
		}
	}
	return results, nil
}

// BusinessesMetadata contains the database metadata for the hg.businesses table.
var BusinessesMetadata = querymodel.TableMetadata{
	Name:               "businesses",
	FullyQualifiedName: "hg.businesses",
	PrimaryKeyColumn: querymodel.ColumnMetadata{
		Name: "id",
		Type: querymodel.DbTypeString,
	},
	Columns: map[string]querymodel.ColumnMetadata{
		"address_line_1": {
			Name: "address_line_1",
			Type: querymodel.DbTypeString,
		},
		"address_line_2": {
			Name: "address_line_2",
			Type: querymodel.DbTypeString,
		},
		"business_type": {
			Name: "business_type",
			Type: querymodel.DbTypeEnum,
		},
		"contact_number": {
			Name: "contact_number",
			Type: querymodel.DbTypeString,
		},
		"country": {
			Name: "country",
			Type: querymodel.DbTypeString,
		},
		"county": {
			Name: "county",
			Type: querymodel.DbTypeString,
		},
		"cph_number": {
			Name: "cph_number",
			Type: querymodel.DbTypeString,
		},
		"created_at": {
			Name: "created_at",
			Type: querymodel.DbTypeDate,
		},
		"created_by_id": {
			Name: "created_by_id",
			Type: querymodel.DbTypeString,
		},
		"description": {
			Name: "description",
			Type: querymodel.DbTypeString,
		},
		"email_address": {
			Name: "email_address",
			Type: querymodel.DbTypeString,
		},
		"id": {
			Name: "id",
			Type: querymodel.DbTypeString,
		},
		"location": {
			Name: "location",
			Type: querymodel.DbTypePoint,
		},
		"logo_url": {
			Name: "logo_url",
			Type: querymodel.DbTypeString,
		},
		"modified_at": {
			Name: "modified_at",
			Type: querymodel.DbTypeDate,
		},
		"modified_by_id": {
			Name: "modified_by_id",
			Type: querymodel.DbTypeString,
		},
		"postcode": {
			Name: "postcode",
			Type: querymodel.DbTypeString,
		},
		"reference": {
			Name: "reference",
			Type: querymodel.DbTypeString,
		},
		"town": {
			Name: "town",
			Type: querymodel.DbTypeString,
		},
		"trading_name": {
			Name: "trading_name",
			Type: querymodel.DbTypeString,
		},
		"website_url": {
			Name: "website_url",
			Type: querymodel.DbTypeString,
		},
	},
	Relationships: map[string]querymodel.RelationshipMetadata{
		"created_by_id": {
			Id:                  "businesses_created_by_id_users_id",
			Type:                "N:1",
			ColumnName:          "created_by_id",
			ExpansionColumnName: "created_by",
			From:                BusinessesResource{},
			To:                  UsersResource{},
			FromColumn: querymodel.ColumnMetadata{
				Name: "created_by_id",
				Type: querymodel.DbTypeString,
			},
			ToColumn: querymodel.ColumnMetadata{
				Name: "id",
				Type: querymodel.DbTypeString,
			},
		},
		"farm_fields_businesses_business_id": {
			Id:                  "farm_fields_business_id_businesses_id",
			Type:                "1:N",
			ColumnName:          "farm_fields_businesses_business_id",
			ExpansionColumnName: "farm_fields_businesses_business_id",
			From:                BusinessesResource{},
			To:                  FarmFieldsResource{},
			FromColumn: querymodel.ColumnMetadata{
				Name: "farm_fields_businesses_business_id",
				Type: querymodel.DbTypeString,
			},
			ToColumn: querymodel.ColumnMetadata{
				Name: "business_id",
				Type: querymodel.DbTypeString,
			},
		},
		"modified_by_id": {
			Id:                  "businesses_modified_by_id_users_id",
			Type:                "N:1",
			ColumnName:          "modified_by_id",
			ExpansionColumnName: "modified_by",
			From:                BusinessesResource{},
			To:                  UsersResource{},
			FromColumn: querymodel.ColumnMetadata{
				Name: "modified_by_id",
				Type: querymodel.DbTypeString,
			},
			ToColumn: querymodel.ColumnMetadata{
				Name: "id",
				Type: querymodel.DbTypeString,
			},
		},
	},
}

// BusinessesModel represents a row from the hg.businesses table.
type BusinessesModel struct {
	AddressLine1                   string                     `json:"address_line_1"`
	AddressLine2                   string                     `json:"address_line_2"`
	BusinessType                   EnumBusinessesBusinessType `json:"business_type"`
	ContactNumber                  string                     `json:"contact_number"`
	Country                        string                     `json:"country"`
	County                         string                     `json:"county"`
	CphNumber                      string                     `json:"cph_number"`
	CreatedAt                      *time.Time                 `json:"created_at"`
	CreatedBy                      *UsersModel                `json:"created_by"`
	CreatedById                    mssql.UniqueIdentifier     `json:"created_by_id"`
	Description                    string                     `json:"description"`
	EmailAddress                   string                     `json:"email_address"`
	FarmFieldsBusinessesBusinessId *FarmFieldsModel           `json:"farm_fields_businesses_business_id"`
	Id                             mssql.UniqueIdentifier     `json:"id"`
	Location                       *querymodel.Point          `json:"location"`
	LogoUrl                        string                     `json:"logo_url"`
	ModifiedAt                     *time.Time                 `json:"modified_at"`
	ModifiedBy                     *UsersModel                `json:"modified_by"`
	ModifiedById                   mssql.UniqueIdentifier     `json:"modified_by_id"`
	Postcode                       string                     `json:"postcode"`
	Reference                      string                     `json:"reference"`
	Town                           string                     `json:"town"`
	TradingName                    string                     `json:"trading_name"`
	WebsiteUrl                     string                     `json:"website_url"`
	projection                     BusinessesProjection       `json:"-"`
}

type BusinessesModels []*BusinessesModel

// SetProjection sets a projection node and sets projection for itself and any
// child nodes
func (m *BusinessesModel) SetProjection(projectionNode *querymodel.ProjectionNode) error {
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
func (ms BusinessesModels) SetProjection(projectionNode *querymodel.ProjectionNode) error {
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
func (m *BusinessesModel) setBusinessesProjection(
	projection BusinessesProjection,
	childNodes map[string]*querymodel.ProjectionNode,
) error {
	m.projection = projection
	if node, exists := childNodes["created_by"]; exists {
		if m.CreatedBy != nil {
			if err := m.CreatedBy.SetProjection(node); err != nil {
				return err
			}
		}
	}

	if node, exists := childNodes["farm_fields_businesses_business_id"]; exists {
		if m.FarmFieldsBusinessesBusinessId != nil {
			if err := m.FarmFieldsBusinessesBusinessId.SetProjection(node); err != nil {
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

// MarshalJSON marshals BusinessesModel to a json string based on the projection
func (m *BusinessesModel) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte('{')

	isFirst := true
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

	if m.projection.Has(businessesProjectionBusinessType) {
		err := marshalProperty(&buf, isFirst, "business_type", m.BusinessType)
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

	if m.projection.Has(businessesProjectionCountry) {
		err := marshalProperty(&buf, isFirst, "country", m.Country)
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

	if m.projection.Has(businessesProjectionCphNumber) {
		err := marshalProperty(&buf, isFirst, "cph_number", m.CphNumber)
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

	if m.projection.Has(businessesProjectionFkBusinessCreatedById) {
		err := marshalProperty(&buf, isFirst, "created_by", m.CreatedBy)
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

	if m.projection.Has(businessesProjectionDescription) {
		err := marshalProperty(&buf, isFirst, "description", m.Description)
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

	if m.projection.Has(businessesProjectionFkFieldBusiness) {
		err := marshalProperty(&buf, isFirst, "farm_fields_businesses_business_id", m.FarmFieldsBusinessesBusinessId)
		if err != nil {
			return nil, err
		}
		isFirst = false
	}

	if m.projection.Has(businessesProjectionId) {
		err := marshalProperty(&buf, isFirst, "id", m.Id)
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

	if m.projection.Has(businessesProjectionLogoUrl) {
		err := marshalProperty(&buf, isFirst, "logo_url", m.LogoUrl)
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

	if m.projection.Has(businessesProjectionFkBusinessModifiedById) {
		err := marshalProperty(&buf, isFirst, "modified_by", m.ModifiedBy)
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

	if m.projection.Has(businessesProjectionPostcode) {
		err := marshalProperty(&buf, isFirst, "postcode", m.Postcode)
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

	if m.projection.Has(businessesProjectionTown) {
		err := marshalProperty(&buf, isFirst, "town", m.Town)
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

	if m.projection.Has(businessesProjectionWebsiteUrl) {
		err := marshalProperty(&buf, isFirst, "website_url", m.WebsiteUrl)
		if err != nil {
			return nil, err
		}
	}

	buf.WriteByte('}')
	return buf.Bytes(), nil
}

// GetValue returns the value from a given path
func (m *BusinessesModel) GetValue(
	v querymodel.ValueBuilder,
	path []*querymodel.TraversalStep,
	columnName string,
) (querymodel.Value, error) {
	if len(path) > 0 {
		nextStep := path[0]
		nextEntity, nextEntityIsNil, err := m.getRelatedEntity(nextStep.Relationship)
		if err != nil {
			return nil, err
		}
		if nextEntityIsNil {
			return v.Null(), nil
		}
		return nextEntity.GetValue(v, path[1:], columnName)
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
func (m *BusinessesModel) getRelatedEntity(
	relationship querymodel.RelationshipMetadata,
) (querymodel.TableModel, bool, error) {
	switch relationship.Id {
	case "businesses_created_by_id_users_id":
		v := m.CreatedBy
		return v, v == nil, nil
	case "businesses_modified_by_id_users_id":
		v := m.ModifiedBy
		return v, v == nil, nil
	default:
		return nil, true, fmt.Errorf(
			"unable to get related entity: unsupported relationship '%s'",
			relationship.Id,
		)
	}
}

// getValue returns the value from a given column
func (m *BusinessesModel) getValue(
	columnName string,
	v querymodel.ValueBuilder,
) (querymodel.Value, error) {
	switch columnName {
	case "address_line_1":
		return v.String(m.AddressLine1), nil
	case "address_line_2":
		return v.String(m.AddressLine2), nil
	case "business_type":
		return v.Int(int64(m.BusinessType)), nil
	case "contact_number":
		return v.String(m.ContactNumber), nil
	case "country":
		return v.String(m.Country), nil
	case "county":
		return v.String(m.County), nil
	case "cph_number":
		return v.String(m.CphNumber), nil
	case "created_at":
		return v.DateTime(*m.CreatedAt), nil
	case "created_by_id":
		return v.String(m.CreatedById.String()), nil
	case "description":
		return v.String(m.Description), nil
	case "email_address":
		return v.String(m.EmailAddress), nil
	case "id":
		return v.String(m.Id.String()), nil
	case "location":
		return v.Point(m.Location.Coordinates[0], m.Location.Coordinates[1]), nil
	case "logo_url":
		return v.String(m.LogoUrl), nil
	case "modified_at":
		return v.DateTime(*m.ModifiedAt), nil
	case "modified_by_id":
		return v.String(m.ModifiedById.String()), nil
	case "postcode":
		return v.String(m.Postcode), nil
	case "reference":
		return v.String(m.Reference), nil
	case "town":
		return v.String(m.Town), nil
	case "trading_name":
		return v.String(m.TradingName), nil
	case "website_url":
		return v.String(m.WebsiteUrl), nil
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
	businessesProjectionAddressLine1 BusinessesProjection = 1 << iota
	businessesProjectionAddressLine2
	businessesProjectionBusinessType
	businessesProjectionContactNumber
	businessesProjectionCountry
	businessesProjectionCounty
	businessesProjectionCphNumber
	businessesProjectionCreatedAt
	businessesProjectionFkBusinessCreatedById
	businessesProjectionCreatedById
	businessesProjectionDescription
	businessesProjectionEmailAddress
	businessesProjectionFkFieldBusiness
	businessesProjectionId
	businessesProjectionLocation
	businessesProjectionLogoUrl
	businessesProjectionModifiedAt
	businessesProjectionFkBusinessModifiedById
	businessesProjectionModifiedById
	businessesProjectionPostcode
	businessesProjectionReference
	businessesProjectionTown
	businessesProjectionTradingName
	businessesProjectionWebsiteUrl
)

// Add includes a given column within the projection
func (p *BusinessesProjection) Add(columnName string) error {
	switch columnName {
	case "address_line_1":
		*p |= businessesProjectionAddressLine1
	case "address_line_2":
		*p |= businessesProjectionAddressLine2
	case "business_type":
		*p |= businessesProjectionBusinessType
	case "contact_number":
		*p |= businessesProjectionContactNumber
	case "country":
		*p |= businessesProjectionCountry
	case "county":
		*p |= businessesProjectionCounty
	case "cph_number":
		*p |= businessesProjectionCphNumber
	case "created_at":
		*p |= businessesProjectionCreatedAt
	case "created_by":
		*p |= businessesProjectionFkBusinessCreatedById
	case "created_by_id":
		*p |= businessesProjectionCreatedById
	case "description":
		*p |= businessesProjectionDescription
	case "email_address":
		*p |= businessesProjectionEmailAddress
	case "farm_fields_businesses_business_id":
		*p |= businessesProjectionFkFieldBusiness
	case "id":
		*p |= businessesProjectionId
	case "location":
		*p |= businessesProjectionLocation
	case "logo_url":
		*p |= businessesProjectionLogoUrl
	case "modified_at":
		*p |= businessesProjectionModifiedAt
	case "modified_by":
		*p |= businessesProjectionFkBusinessModifiedById
	case "modified_by_id":
		*p |= businessesProjectionModifiedById
	case "postcode":
		*p |= businessesProjectionPostcode
	case "reference":
		*p |= businessesProjectionReference
	case "town":
		*p |= businessesProjectionTown
	case "trading_name":
		*p |= businessesProjectionTradingName
	case "website_url":
		*p |= businessesProjectionWebsiteUrl
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

func (r FarmFieldsResource) GetMetadata() querymodel.TableMetadata {
	return FarmFieldsMetadata
}

func (r FarmFieldsResource) InitProjection() querymodel.Projection {
	return new(FarmFieldsProjection)
}

// SliceFromJSON unmarshals a json array of farm_fields and returns it as a slice
func (m FarmFieldsResource) SliceFromJSON(
	jsonData []byte,
	projectionNode *querymodel.ProjectionNode,
) ([]querymodel.TableModel, error) {
	if len(jsonData) == 0 {
		return []querymodel.TableModel{}, nil
	}

	var concreteSlice []*FarmFieldsModel
	if err := json.Unmarshal(jsonData, &concreteSlice); err != nil {
		return nil, err
	}
	result := make([]querymodel.TableModel, len(concreteSlice))

	for i := range concreteSlice {
		if err := concreteSlice[i].SetProjection(projectionNode); err != nil {
			return nil, err
		}
		result[i] = concreteSlice[i]
	}
	return result, nil
}

// TypedSliceFromJSON unmarshals a json array of farm_fields and returns it as a slice

func (m FarmFieldsResource) TypedSliceFromJSON(
	jsonData []byte,
	projectionNode *querymodel.ProjectionNode,
) ([]*FarmFieldsModel, error) {
	if len(jsonData) == 0 {
		return []*FarmFieldsModel{}, nil
	}

	var results []*FarmFieldsModel
	if err := json.Unmarshal(jsonData, &results); err != nil {
		return nil, err
	}
	for _, result := range results {
		if err := result.SetProjection(projectionNode); err != nil {
			return nil, err
		}
	}
	return results, nil
}

// FarmFieldsMetadata contains the database metadata for the hg.farm_fields table.
var FarmFieldsMetadata = querymodel.TableMetadata{
	Name:               "farm_fields",
	FullyQualifiedName: "hg.farm_fields",
	PrimaryKeyColumn: querymodel.ColumnMetadata{
		Name: "id",
		Type: querymodel.DbTypeString,
	},
	Columns: map[string]querymodel.ColumnMetadata{
		"business_id": {
			Name: "business_id",
			Type: querymodel.DbTypeString,
		},
		"created_at": {
			Name: "created_at",
			Type: querymodel.DbTypeDate,
		},
		"created_by_id": {
			Name: "created_by_id",
			Type: querymodel.DbTypeString,
		},
		"id": {
			Name: "id",
			Type: querymodel.DbTypeString,
		},
		"location": {
			Name: "location",
			Type: querymodel.DbTypePoint,
		},
		"modified_at": {
			Name: "modified_at",
			Type: querymodel.DbTypeDate,
		},
		"modified_by_id": {
			Name: "modified_by_id",
			Type: querymodel.DbTypeString,
		},
		"reference": {
			Name: "reference",
			Type: querymodel.DbTypeString,
		},
	},
	Relationships: map[string]querymodel.RelationshipMetadata{
		"business_id": {
			Id:                  "farm_fields_business_id_businesses_id",
			Type:                "N:1",
			ColumnName:          "business_id",
			ExpansionColumnName: "business",
			From:                FarmFieldsResource{},
			To:                  BusinessesResource{},
			FromColumn: querymodel.ColumnMetadata{
				Name: "business_id",
				Type: querymodel.DbTypeString,
			},
			ToColumn: querymodel.ColumnMetadata{
				Name: "id",
				Type: querymodel.DbTypeString,
			},
		},
		"created_by_id": {
			Id:                  "farm_fields_created_by_id_users_id",
			Type:                "N:1",
			ColumnName:          "created_by_id",
			ExpansionColumnName: "created_by",
			From:                FarmFieldsResource{},
			To:                  UsersResource{},
			FromColumn: querymodel.ColumnMetadata{
				Name: "created_by_id",
				Type: querymodel.DbTypeString,
			},
			ToColumn: querymodel.ColumnMetadata{
				Name: "id",
				Type: querymodel.DbTypeString,
			},
		},
		"modified_by_id": {
			Id:                  "farm_fields_modified_by_id_users_id",
			Type:                "N:1",
			ColumnName:          "modified_by_id",
			ExpansionColumnName: "modified_by",
			From:                FarmFieldsResource{},
			To:                  UsersResource{},
			FromColumn: querymodel.ColumnMetadata{
				Name: "modified_by_id",
				Type: querymodel.DbTypeString,
			},
			ToColumn: querymodel.ColumnMetadata{
				Name: "id",
				Type: querymodel.DbTypeString,
			},
		},
	},
}

// FarmFieldsModel represents a row from the hg.farm_fields table.
type FarmFieldsModel struct {
	Business     *BusinessesModel       `json:"business"`
	BusinessId   mssql.UniqueIdentifier `json:"business_id"`
	CreatedAt    *time.Time             `json:"created_at"`
	CreatedBy    *UsersModel            `json:"created_by"`
	CreatedById  mssql.UniqueIdentifier `json:"created_by_id"`
	Id           mssql.UniqueIdentifier `json:"id"`
	Location     *querymodel.Point      `json:"location"`
	ModifiedAt   *time.Time             `json:"modified_at"`
	ModifiedBy   *UsersModel            `json:"modified_by"`
	ModifiedById mssql.UniqueIdentifier `json:"modified_by_id"`
	Reference    string                 `json:"reference"`
	projection   FarmFieldsProjection   `json:"-"`
}

type FarmFieldsModels []*FarmFieldsModel

// SetProjection sets a projection node and sets projection for itself and any
// child nodes
func (m *FarmFieldsModel) SetProjection(projectionNode *querymodel.ProjectionNode) error {
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
func (ms FarmFieldsModels) SetProjection(projectionNode *querymodel.ProjectionNode) error {
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
func (m *FarmFieldsModel) setFarmFieldsProjection(
	projection FarmFieldsProjection,
	childNodes map[string]*querymodel.ProjectionNode,
) error {
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

// MarshalJSON marshals FarmFieldsModel to a json string based on the projection
func (m *FarmFieldsModel) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte('{')

	isFirst := true
	if m.projection.Has(farmFieldsProjectionFkFieldBusiness) {
		err := marshalProperty(&buf, isFirst, "business", m.Business)
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

	if m.projection.Has(farmFieldsProjectionCreatedAt) {
		err := marshalProperty(&buf, isFirst, "created_at", m.CreatedAt)
		if err != nil {
			return nil, err
		}
		isFirst = false
	}

	if m.projection.Has(farmFieldsProjectionFkFieldCreatedById) {
		err := marshalProperty(&buf, isFirst, "created_by", m.CreatedBy)
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

	if m.projection.Has(farmFieldsProjectionId) {
		err := marshalProperty(&buf, isFirst, "id", m.Id)
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

	if m.projection.Has(farmFieldsProjectionModifiedAt) {
		err := marshalProperty(&buf, isFirst, "modified_at", m.ModifiedAt)
		if err != nil {
			return nil, err
		}
		isFirst = false
	}

	if m.projection.Has(farmFieldsProjectionFkFieldModifiedById) {
		err := marshalProperty(&buf, isFirst, "modified_by", m.ModifiedBy)
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

	if m.projection.Has(farmFieldsProjectionReference) {
		err := marshalProperty(&buf, isFirst, "reference", m.Reference)
		if err != nil {
			return nil, err
		}
	}

	buf.WriteByte('}')
	return buf.Bytes(), nil
}

// GetValue returns the value from a given path
func (m *FarmFieldsModel) GetValue(
	v querymodel.ValueBuilder,
	path []*querymodel.TraversalStep,
	columnName string,
) (querymodel.Value, error) {
	if len(path) > 0 {
		nextStep := path[0]
		nextEntity, nextEntityIsNil, err := m.getRelatedEntity(nextStep.Relationship)
		if err != nil {
			return nil, err
		}
		if nextEntityIsNil {
			return v.Null(), nil
		}
		return nextEntity.GetValue(v, path[1:], columnName)
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
func (m *FarmFieldsModel) getRelatedEntity(
	relationship querymodel.RelationshipMetadata,
) (querymodel.TableModel, bool, error) {
	switch relationship.Id {
	case "farm_fields_business_id_businesses_id":
		v := m.Business
		return v, v == nil, nil
	case "farm_fields_created_by_id_users_id":
		v := m.CreatedBy
		return v, v == nil, nil
	case "farm_fields_modified_by_id_users_id":
		v := m.ModifiedBy
		return v, v == nil, nil
	default:
		return nil, true, fmt.Errorf(
			"unable to get related entity: unsupported relationship '%s'",
			relationship.Id,
		)
	}
}

// getValue returns the value from a given column
func (m *FarmFieldsModel) getValue(
	columnName string,
	v querymodel.ValueBuilder,
) (querymodel.Value, error) {
	switch columnName {
	case "business_id":
		return v.String(m.BusinessId.String()), nil
	case "created_at":
		return v.DateTime(*m.CreatedAt), nil
	case "created_by_id":
		return v.String(m.CreatedById.String()), nil
	case "id":
		return v.String(m.Id.String()), nil
	case "location":
		return v.Point(m.Location.Coordinates[0], m.Location.Coordinates[1]), nil
	case "modified_at":
		return v.DateTime(*m.ModifiedAt), nil
	case "modified_by_id":
		return v.String(m.ModifiedById.String()), nil
	case "reference":
		return v.String(m.Reference), nil
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
	farmFieldsProjectionFkFieldBusiness FarmFieldsProjection = 1 << iota
	farmFieldsProjectionBusinessId
	farmFieldsProjectionCreatedAt
	farmFieldsProjectionFkFieldCreatedById
	farmFieldsProjectionCreatedById
	farmFieldsProjectionId
	farmFieldsProjectionLocation
	farmFieldsProjectionModifiedAt
	farmFieldsProjectionFkFieldModifiedById
	farmFieldsProjectionModifiedById
	farmFieldsProjectionReference
)

// Add includes a given column within the projection
func (p *FarmFieldsProjection) Add(columnName string) error {
	switch columnName {
	case "business":
		*p |= farmFieldsProjectionFkFieldBusiness
	case "business_id":
		*p |= farmFieldsProjectionBusinessId
	case "created_at":
		*p |= farmFieldsProjectionCreatedAt
	case "created_by":
		*p |= farmFieldsProjectionFkFieldCreatedById
	case "created_by_id":
		*p |= farmFieldsProjectionCreatedById
	case "id":
		*p |= farmFieldsProjectionId
	case "location":
		*p |= farmFieldsProjectionLocation
	case "modified_at":
		*p |= farmFieldsProjectionModifiedAt
	case "modified_by":
		*p |= farmFieldsProjectionFkFieldModifiedById
	case "modified_by_id":
		*p |= farmFieldsProjectionModifiedById
	case "reference":
		*p |= farmFieldsProjectionReference
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

func (r UsersResource) GetMetadata() querymodel.TableMetadata {
	return UsersMetadata
}

func (r UsersResource) InitProjection() querymodel.Projection {
	return new(UsersProjection)
}

// SliceFromJSON unmarshals a json array of users and returns it as a slice
func (m UsersResource) SliceFromJSON(
	jsonData []byte,
	projectionNode *querymodel.ProjectionNode,
) ([]querymodel.TableModel, error) {
	if len(jsonData) == 0 {
		return []querymodel.TableModel{}, nil
	}

	var concreteSlice []*UsersModel
	if err := json.Unmarshal(jsonData, &concreteSlice); err != nil {
		return nil, err
	}
	result := make([]querymodel.TableModel, len(concreteSlice))

	for i := range concreteSlice {
		if err := concreteSlice[i].SetProjection(projectionNode); err != nil {
			return nil, err
		}
		result[i] = concreteSlice[i]
	}
	return result, nil
}

// TypedSliceFromJSON unmarshals a json array of users and returns it as a slice

func (m UsersResource) TypedSliceFromJSON(
	jsonData []byte,
	projectionNode *querymodel.ProjectionNode,
) ([]*UsersModel, error) {
	if len(jsonData) == 0 {
		return []*UsersModel{}, nil
	}

	var results []*UsersModel
	if err := json.Unmarshal(jsonData, &results); err != nil {
		return nil, err
	}
	for _, result := range results {
		if err := result.SetProjection(projectionNode); err != nil {
			return nil, err
		}
	}
	return results, nil
}

// UsersMetadata contains the database metadata for the hg.users table.
var UsersMetadata = querymodel.TableMetadata{
	Name:               "users",
	FullyQualifiedName: "hg.users",
	PrimaryKeyColumn: querymodel.ColumnMetadata{
		Name: "id",
		Type: querymodel.DbTypeString,
	},
	Columns: map[string]querymodel.ColumnMetadata{
		"created_at": {
			Name: "created_at",
			Type: querymodel.DbTypeDate,
		},
		"display_name": {
			Name: "display_name",
			Type: querymodel.DbTypeString,
		},
		"email_address": {
			Name: "email_address",
			Type: querymodel.DbTypeString,
		},
		"family_name": {
			Name: "family_name",
			Type: querymodel.DbTypeString,
		},
		"given_name": {
			Name: "given_name",
			Type: querymodel.DbTypeString,
		},
		"id": {
			Name: "id",
			Type: querymodel.DbTypeString,
		},
		"modified_at": {
			Name: "modified_at",
			Type: querymodel.DbTypeDate,
		},
		"oid": {
			Name: "oid",
			Type: querymodel.DbTypeString,
		},
	},
	Relationships: map[string]querymodel.RelationshipMetadata{
		"businesses_users_created_by_id": {
			Id:                  "businesses_created_by_id_users_id",
			Type:                "1:N",
			ColumnName:          "businesses_users_created_by_id",
			ExpansionColumnName: "businesses_users_created_by_id",
			From:                UsersResource{},
			To:                  BusinessesResource{},
			FromColumn: querymodel.ColumnMetadata{
				Name: "businesses_users_created_by_id",
				Type: querymodel.DbTypeString,
			},
			ToColumn: querymodel.ColumnMetadata{
				Name: "created_by_id",
				Type: querymodel.DbTypeString,
			},
		},
		"businesses_users_modified_by_id": {
			Id:                  "businesses_modified_by_id_users_id",
			Type:                "1:N",
			ColumnName:          "businesses_users_modified_by_id",
			ExpansionColumnName: "businesses_users_modified_by_id",
			From:                UsersResource{},
			To:                  BusinessesResource{},
			FromColumn: querymodel.ColumnMetadata{
				Name: "businesses_users_modified_by_id",
				Type: querymodel.DbTypeString,
			},
			ToColumn: querymodel.ColumnMetadata{
				Name: "modified_by_id",
				Type: querymodel.DbTypeString,
			},
		},
		"farm_fields_users_created_by_id": {
			Id:                  "farm_fields_created_by_id_users_id",
			Type:                "1:N",
			ColumnName:          "farm_fields_users_created_by_id",
			ExpansionColumnName: "farm_fields_users_created_by_id",
			From:                UsersResource{},
			To:                  FarmFieldsResource{},
			FromColumn: querymodel.ColumnMetadata{
				Name: "farm_fields_users_created_by_id",
				Type: querymodel.DbTypeString,
			},
			ToColumn: querymodel.ColumnMetadata{
				Name: "created_by_id",
				Type: querymodel.DbTypeString,
			},
		},
		"farm_fields_users_modified_by_id": {
			Id:                  "farm_fields_modified_by_id_users_id",
			Type:                "1:N",
			ColumnName:          "farm_fields_users_modified_by_id",
			ExpansionColumnName: "farm_fields_users_modified_by_id",
			From:                UsersResource{},
			To:                  FarmFieldsResource{},
			FromColumn: querymodel.ColumnMetadata{
				Name: "farm_fields_users_modified_by_id",
				Type: querymodel.DbTypeString,
			},
			ToColumn: querymodel.ColumnMetadata{
				Name: "modified_by_id",
				Type: querymodel.DbTypeString,
			},
		},
	},
}

// UsersModel represents a row from the hg.users table.
type UsersModel struct {
	BusinessesUsersCreatedById  *BusinessesModel       `json:"businesses_users_created_by_id"`
	BusinessesUsersModifiedById *BusinessesModel       `json:"businesses_users_modified_by_id"`
	CreatedAt                   *time.Time             `json:"created_at"`
	DisplayName                 string                 `json:"display_name"`
	EmailAddress                string                 `json:"email_address"`
	FamilyName                  string                 `json:"family_name"`
	FarmFieldsUsersCreatedById  *FarmFieldsModel       `json:"farm_fields_users_created_by_id"`
	FarmFieldsUsersModifiedById *FarmFieldsModel       `json:"farm_fields_users_modified_by_id"`
	GivenName                   string                 `json:"given_name"`
	Id                          mssql.UniqueIdentifier `json:"id"`
	ModifiedAt                  *time.Time             `json:"modified_at"`
	Oid                         mssql.UniqueIdentifier `json:"oid"`
	projection                  UsersProjection        `json:"-"`
}

type UsersModels []*UsersModel

// SetProjection sets a projection node and sets projection for itself and any
// child nodes
func (m *UsersModel) SetProjection(projectionNode *querymodel.ProjectionNode) error {
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
func (ms UsersModels) SetProjection(projectionNode *querymodel.ProjectionNode) error {
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
func (m *UsersModel) setUsersProjection(
	projection UsersProjection,
	childNodes map[string]*querymodel.ProjectionNode,
) error {
	m.projection = projection
	if node, exists := childNodes["businesses_users_created_by_id"]; exists {
		if m.BusinessesUsersCreatedById != nil {
			if err := m.BusinessesUsersCreatedById.SetProjection(node); err != nil {
				return err
			}
		}
	}

	if node, exists := childNodes["businesses_users_modified_by_id"]; exists {
		if m.BusinessesUsersModifiedById != nil {
			if err := m.BusinessesUsersModifiedById.SetProjection(node); err != nil {
				return err
			}
		}
	}

	if node, exists := childNodes["farm_fields_users_created_by_id"]; exists {
		if m.FarmFieldsUsersCreatedById != nil {
			if err := m.FarmFieldsUsersCreatedById.SetProjection(node); err != nil {
				return err
			}
		}
	}

	if node, exists := childNodes["farm_fields_users_modified_by_id"]; exists {
		if m.FarmFieldsUsersModifiedById != nil {
			if err := m.FarmFieldsUsersModifiedById.SetProjection(node); err != nil {
				return err
			}
		}
	}

	return nil
}

// MarshalJSON marshals UsersModel to a json string based on the projection
func (m *UsersModel) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte('{')

	isFirst := true
	if m.projection.Has(usersProjectionFkBusinessCreatedById) {
		err := marshalProperty(&buf, isFirst, "businesses_users_created_by_id", m.BusinessesUsersCreatedById)
		if err != nil {
			return nil, err
		}
		isFirst = false
	}

	if m.projection.Has(usersProjectionFkBusinessModifiedById) {
		err := marshalProperty(&buf, isFirst, "businesses_users_modified_by_id", m.BusinessesUsersModifiedById)
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

	if m.projection.Has(usersProjectionDisplayName) {
		err := marshalProperty(&buf, isFirst, "display_name", m.DisplayName)
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

	if m.projection.Has(usersProjectionFamilyName) {
		err := marshalProperty(&buf, isFirst, "family_name", m.FamilyName)
		if err != nil {
			return nil, err
		}
		isFirst = false
	}

	if m.projection.Has(usersProjectionFkFieldCreatedById) {
		err := marshalProperty(&buf, isFirst, "farm_fields_users_created_by_id", m.FarmFieldsUsersCreatedById)
		if err != nil {
			return nil, err
		}
		isFirst = false
	}

	if m.projection.Has(usersProjectionFkFieldModifiedById) {
		err := marshalProperty(&buf, isFirst, "farm_fields_users_modified_by_id", m.FarmFieldsUsersModifiedById)
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

	if m.projection.Has(usersProjectionId) {
		err := marshalProperty(&buf, isFirst, "id", m.Id)
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

	if m.projection.Has(usersProjectionOid) {
		err := marshalProperty(&buf, isFirst, "oid", m.Oid)
		if err != nil {
			return nil, err
		}
	}

	buf.WriteByte('}')
	return buf.Bytes(), nil
}

// GetValue returns the value from a given path
func (m *UsersModel) GetValue(
	v querymodel.ValueBuilder,
	path []*querymodel.TraversalStep,
	columnName string,
) (querymodel.Value, error) {
	if len(path) > 0 {
		return nil, fmt.Errorf("unable to get value: invalid path")
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

// getValue returns the value from a given column
func (m *UsersModel) getValue(
	columnName string,
	v querymodel.ValueBuilder,
) (querymodel.Value, error) {
	switch columnName {
	case "created_at":
		return v.DateTime(*m.CreatedAt), nil
	case "display_name":
		return v.String(m.DisplayName), nil
	case "email_address":
		return v.String(m.EmailAddress), nil
	case "family_name":
		return v.String(m.FamilyName), nil
	case "given_name":
		return v.String(m.GivenName), nil
	case "id":
		return v.String(m.Id.String()), nil
	case "modified_at":
		return v.DateTime(*m.ModifiedAt), nil
	case "oid":
		return v.String(m.Oid.String()), nil
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
	usersProjectionFkBusinessCreatedById UsersProjection = 1 << iota
	usersProjectionFkBusinessModifiedById
	usersProjectionCreatedAt
	usersProjectionDisplayName
	usersProjectionEmailAddress
	usersProjectionFamilyName
	usersProjectionFkFieldCreatedById
	usersProjectionFkFieldModifiedById
	usersProjectionGivenName
	usersProjectionId
	usersProjectionModifiedAt
	usersProjectionOid
)

// Add includes a given column within the projection
func (p *UsersProjection) Add(columnName string) error {
	switch columnName {
	case "businesses_users_created_by_id":
		*p |= usersProjectionFkBusinessCreatedById
	case "businesses_users_modified_by_id":
		*p |= usersProjectionFkBusinessModifiedById
	case "created_at":
		*p |= usersProjectionCreatedAt
	case "display_name":
		*p |= usersProjectionDisplayName
	case "email_address":
		*p |= usersProjectionEmailAddress
	case "family_name":
		*p |= usersProjectionFamilyName
	case "farm_fields_users_created_by_id":
		*p |= usersProjectionFkFieldCreatedById
	case "farm_fields_users_modified_by_id":
		*p |= usersProjectionFkFieldModifiedById
	case "given_name":
		*p |= usersProjectionGivenName
	case "id":
		*p |= usersProjectionId
	case "modified_at":
		*p |= usersProjectionModifiedAt
	case "oid":
		*p |= usersProjectionOid
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
) (querymodel.TableAccessPolicy, bool) {
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
	AddressLine1  bool
	AddressLine2  bool
	BusinessType  bool
	ContactNumber bool
	Country       bool
	County        bool
	CphNumber     bool
	CreatedAt     bool
	CreatedById   bool
	Description   bool
	EmailAddress  bool
	Id            bool
	Location      bool
	LogoUrl       bool
	ModifiedAt    bool
	ModifiedById  bool
	Postcode      bool
	Reference     bool
	Town          bool
	TradingName   bool
	WebsiteUrl    bool
}

// CanAccess defines, at the table level, if a user can perform any
// operations on the businesses table. Specific column access policies can restrict
// access given but cannot override access denied
func (p *BusinessesAccessPolicy) CanAccess() bool {
	return p.UserCanAccess
}

// CanAccessColumn returns true if a column can be accessed else false. An error
// is returned if the column does not exist on the table
func (p *BusinessesAccessPolicy) CanAccessColumn(
	columnName string,
) (bool, error) {
	switch columnName {
	case "address_line_1":
		return p.AddressLine1, nil
	case "address_line_2":
		return p.AddressLine2, nil
	case "business_type":
		return p.BusinessType, nil
	case "contact_number":
		return p.ContactNumber, nil
	case "country":
		return p.Country, nil
	case "county":
		return p.County, nil
	case "cph_number":
		return p.CphNumber, nil
	case "created_at":
		return p.CreatedAt, nil
	case "created_by_id":
		return p.CreatedById, nil
	case "description":
		return p.Description, nil
	case "email_address":
		return p.EmailAddress, nil
	case "id":
		return p.Id, nil
	case "location":
		return p.Location, nil
	case "logo_url":
		return p.LogoUrl, nil
	case "modified_at":
		return p.ModifiedAt, nil
	case "modified_by_id":
		return p.ModifiedById, nil
	case "postcode":
		return p.Postcode, nil
	case "reference":
		return p.Reference, nil
	case "town":
		return p.Town, nil
	case "trading_name":
		return p.TradingName, nil
	case "website_url":
		return p.WebsiteUrl, nil
	default:
		return false, fmt.Errorf(
			"column %s does not exists on the businesses table",
			columnName,
		)
	}
}

// FarmFieldsAccessPolicy defines an access policy for the farm_fields table
type FarmFieldsAccessPolicy struct {
	UserCanAccess bool
	BusinessId    bool
	CreatedAt     bool
	CreatedById   bool
	Id            bool
	Location      bool
	ModifiedAt    bool
	ModifiedById  bool
	Reference     bool
}

// CanAccess defines, at the table level, if a user can perform any
// operations on the farm_fields table. Specific column access policies can restrict
// access given but cannot override access denied
func (p *FarmFieldsAccessPolicy) CanAccess() bool {
	return p.UserCanAccess
}

// CanAccessColumn returns true if a column can be accessed else false. An error
// is returned if the column does not exist on the table
func (p *FarmFieldsAccessPolicy) CanAccessColumn(
	columnName string,
) (bool, error) {
	switch columnName {
	case "business_id":
		return p.BusinessId, nil
	case "created_at":
		return p.CreatedAt, nil
	case "created_by_id":
		return p.CreatedById, nil
	case "id":
		return p.Id, nil
	case "location":
		return p.Location, nil
	case "modified_at":
		return p.ModifiedAt, nil
	case "modified_by_id":
		return p.ModifiedById, nil
	case "reference":
		return p.Reference, nil
	default:
		return false, fmt.Errorf(
			"column %s does not exists on the farm_fields table",
			columnName,
		)
	}
}

// UsersAccessPolicy defines an access policy for the users table
type UsersAccessPolicy struct {
	UserCanAccess bool
	CreatedAt     bool
	DisplayName   bool
	EmailAddress  bool
	FamilyName    bool
	GivenName     bool
	Id            bool
	ModifiedAt    bool
	Oid           bool
}

// CanAccess defines, at the table level, if a user can perform any
// operations on the users table. Specific column access policies can restrict
// access given but cannot override access denied
func (p *UsersAccessPolicy) CanAccess() bool {
	return p.UserCanAccess
}

// CanAccessColumn returns true if a column can be accessed else false. An error
// is returned if the column does not exist on the table
func (p *UsersAccessPolicy) CanAccessColumn(
	columnName string,
) (bool, error) {
	switch columnName {
	case "created_at":
		return p.CreatedAt, nil
	case "display_name":
		return p.DisplayName, nil
	case "email_address":
		return p.EmailAddress, nil
	case "family_name":
		return p.FamilyName, nil
	case "given_name":
		return p.GivenName, nil
	case "id":
		return p.Id, nil
	case "modified_at":
		return p.ModifiedAt, nil
	case "oid":
		return p.Oid, nil
	default:
		return false, fmt.Errorf(
			"column %s does not exists on the users table",
			columnName,
		)
	}
}
