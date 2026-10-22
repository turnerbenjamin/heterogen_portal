package schemamodel

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/turnerbenjamin/heterogen_portal/cmd/modelbuilder/builderrepo"
)

type relationshipType string

const (
	RelationshipOneToMany relationshipType = "1:N"
	RelationshipManyToOne relationshipType = "N:1"
)

const (
	MsqlTypeUniqueIdentifier builderrepo.MsqlDataTypeName = "uniqueidentifier"
	MsqlTypeTinyInt          builderrepo.MsqlDataTypeName = "tinyint"
	MsqlTypeNvarchar         builderrepo.MsqlDataTypeName = "nvarchar"
	MsqlTypeInt              builderrepo.MsqlDataTypeName = "int"
	MsqlTypeGeography        builderrepo.MsqlDataTypeName = "geography"
	MsqlTypeDateTimeOffset   builderrepo.MsqlDataTypeName = "datetimeoffset"
)

type GoColType string

const (
	GoColTypeId     GoColType = "mssql.UniqueIdentifier"
	GoColTypeString GoColType = "string"
	GoColTypeInt64  GoColType = "int64"
	GoColTypeUInt8  GoColType = "uint8"
	GoColTypeTime   GoColType = "*time.Time"
	GoColTypePoint  GoColType = "*querymodel.Point"
)

func (t GoColType) String() string {
	return string(t)
}

type QueryModelType string

const (
	QueryModelTypeEnum   QueryModelType = "DbTypeEnum"
	QueryModelTypeString QueryModelType = "DbTypeString"
	QueryModelTypeInt    QueryModelType = "DbTypeInt"
	QueryModelTypeTime   QueryModelType = "DbTypeDateTime"
	QueryModelTypePoint  QueryModelType = "DbTypePoint"
)

func (t QueryModelType) String() string {
	return string(t)
}

type AppColType interface {
	IsAppColType()
}

type AppColTypeString struct {
	MaxLenth int16
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
	EnumType string
}

func (t AppColTypeEnum) IsAppColType() {}

type AppColTypeLookup struct {
	Relationship *Relationship
}

func (t AppColTypeLookup) IsAppColType() {}

type ColumnIdentifiers struct {
	StructProperty string
	Projection     string
}

type Column struct {
	Name              string
	DisplayName       string
	Identifiers       ColumnIdentifiers
	TypeData          TypeData
	IsPrimaryKey      bool
	IsExpansionColumn bool
}

type Relationship struct {
	Id   string
	Type relationshipType

	ColumnName string
	Column     *Column

	ExpansionColumn *Column

	RelatedTableName string
	RelatedTable     *Table

	RelatedColumnName string
	RelatedColumn     *Column
}

type TableIdentifiers struct {
	ModelStruct      string
	CollectionType   string
	Resource         string
	Projection       string
	MetadataVar      string
	AccessPolicyName string
}

type TypeData struct {
	DbType                   builderrepo.MsqlDataTypeName
	GoTypeIdentifier         GoColType
	QueryModelTypeIdentifier QueryModelType
	AppType                  AppColType
}

type ColumnData struct {
	// Map is a mapping from a columns database name to the column data
	Map map[string]*Column

	// Ordered is the collection of columns sorted in alphabetical order - This
	// should be used when writing the model to ensure deterministic output
	Ordered []*Column

	All []*Column

	// PrimaryKey is the primary key field for the table
	PrimaryKey *Column

	// NameColumn is the column used as the display name/label for a given record
	// - This model requires that the display name is a string type
	NameColumn *Column
}

type RelationshipData struct {
	// Map is a mapping from a relationship column name to a relationship
	Map map[string]*Relationship

	// Ordered is the collection of relaitonships, sorted by the column name
	Ordered []*Relationship

	OneToManyRelationships []*Relationship
	ManyToOneRelationships []*Relationship
}

type Table struct {
	// Name is the database name of the table
	Name string

	// FullyQualifiedName is the full database name of the table including the
	// schema
	FullyQualifiedName string

	// DisplayName is the display name of the table
	DisplayName string

	// PluralName is the plural display name for the table
	PluralDisplayName string

	// PrimaryColumn is the column used as a display name for individual records,
	// e.g. to use in place of the raw id for lookup columns
	PrimaryColumn *Column

	// Identifiers is the collection of identifiers used when building models and
	// data structures for the table
	Identifiers TableIdentifiers

	ColumnData       ColumnData
	RelationshipData RelationshipData
}

type TableData struct {
	Map     map[string]*Table
	Ordered []*Table
}

type Schema struct {
	TableData TableData
}

func BuildModel(metadata *builderrepo.DatabaseSchema) (*Schema, error) {
	schema := &Schema{
		TableData: TableData{
			Map:     make(map[string]*Table, len(metadata.Tables)),
			Ordered: make([]*Table, 0, len(metadata.Tables)),
		},
	}
	// Build Table
	for _, tableData := range metadata.Tables {
		table, err := buildTable(metadata, tableData)
		if err != nil {
			return nil, err
		}
		schema.TableData.Map[table.Name] = table
		schema.TableData.Ordered = append(schema.TableData.Ordered, table)
	}

	slices.SortFunc(schema.TableData.Ordered, func(a *Table, b *Table) int {
		return strings.Compare(a.Name, b.Name)
	})

	if err := schema.enrichRelationshipData(); err != nil {
		return nil, err
	}

	if err := validateSchemaRelationshipData(schema); err != nil {
		return nil, err
	}

	return schema, nil
}

func (s Schema) enrichRelationshipData() error {
	for _, localTable := range s.TableData.Ordered {
		for _, relationship := range localTable.RelationshipData.Ordered {

			localColumn, exists := localTable.ColumnData.Map[relationship.ColumnName]
			if !exists {
				return fmt.Errorf(
					"unable to enrich relationship data: cannot access column %s",
					relationship.ColumnName,
				)
			}

			foreignTable, exists := s.TableData.Map[relationship.RelatedTableName]
			if !exists {
				return fmt.Errorf(
					"unable to enrich relationship data: cannot access table %s",
					relationship.RelatedTableName,
				)
			}

			foreignColumn, exists := foreignTable.ColumnData.Map[relationship.RelatedColumnName]
			if !exists {
				return fmt.Errorf(
					"unable to enrich relationship data: cannot access column %s",
					relationship.RelatedColumnName,
				)
			}

			relationship.ExpansionColumn.TypeData = TypeData{
				DbType: MsqlTypeNvarchar,
				GoTypeIdentifier: GoColType(
					fmt.Sprintf("*%s", foreignTable.Identifiers.ModelStruct),
				),
				QueryModelTypeIdentifier: QueryModelTypeString,
				AppType: AppColTypeLookup{
					Relationship: relationship,
				},
			}

			relationship.Column = localColumn
			relationship.RelatedTable = foreignTable
			relationship.RelatedColumn = foreignColumn
		}
	}
	return nil
}

func buildTable(
	schema *builderrepo.DatabaseSchema,
	tableData *builderrepo.TableMetadata,
) (*Table, error) {

	relationshipData, err := buildRelationshipData(
		schema,
		tableData,
	)
	if err != nil {
		return nil, err
	}

	columnData, err := buildColumnData(tableData, relationshipData)
	if err != nil {
		return nil, err
	}

	err = validateColumnData(columnData)
	if err != nil {
		return nil, err
	}

	tableNamePascal := snakeToPascal(tableData.Name)

	displayName := tableData.DisplayName
	if displayName == "" {
		return nil, fmt.Errorf(
			"unable to build table '%s': display name is not set",
			tableData.Name,
		)
	}

	pluralName := tableData.PluralName
	if pluralName == "" {
		return nil, fmt.Errorf(
			"unable to build table '%s': plural name is not set",
			tableData.Name,
		)
	}

	primaryColumnName := tableData.PrimaryColumn
	if primaryColumnName == "" {
		return nil, fmt.Errorf(
			"unable to build table '%s': primary column is not set",
			tableData.Name,
		)
	}

	primaryColumn, exists := columnData.Map[primaryColumnName]
	if !exists {
		return nil, fmt.Errorf(
			"unable to build table '%s': primary column '%s' does not exist",
			tableData.Name,
			primaryColumnName,
		)
	}

	switch pc := primaryColumn.TypeData.AppType.(type) {
	case AppColTypeString:
		if pc.MaxLenth < 1 || pc.MaxLenth > 255 {
			return nil, fmt.Errorf(
				"unable to build table '%s': primary column '%s' must have a max length of 1-255",
				tableData.Name,
				primaryColumnName,
			)
		}
	default:
		return nil, fmt.Errorf(
			"unable to build table '%s': primary column '%s' must be a string type",
			tableData.Name,
			primaryColumnName,
		)
	}

	return &Table{
		Name:               tableData.Name,
		FullyQualifiedName: fmt.Sprintf("%s.%s", tableData.Schema, tableData.Name),

		DisplayName:       displayName,
		PluralDisplayName: pluralName,
		PrimaryColumn:     primaryColumn,

		Identifiers: TableIdentifiers{
			ModelStruct:      fmt.Sprintf("%sModel", tableNamePascal),
			CollectionType:   fmt.Sprintf("%sModels", tableNamePascal),
			Projection:       fmt.Sprintf("%sProjection", tableNamePascal),
			MetadataVar:      fmt.Sprintf("%sMetadata", tableNamePascal),
			Resource:         fmt.Sprintf("%sResource", tableNamePascal),
			AccessPolicyName: fmt.Sprintf("%sAccessPolicy", tableNamePascal),
		},

		ColumnData:       columnData,
		RelationshipData: relationshipData,
	}, nil
}

func buildRelationshipData(
	schema *builderrepo.DatabaseSchema,
	table *builderrepo.TableMetadata,
) (RelationshipData, error) {
	data := RelationshipData{
		Map:                    map[string]*Relationship{},
		Ordered:                []*Relationship{},
		ManyToOneRelationships: []*Relationship{},
		OneToManyRelationships: []*Relationship{},
	}

	for _, metadata := range schema.Relationships {
		// check table relates to the table
		if metadata.ParentTable != table.Name &&
			metadata.ReferencedTable != table.Name {
			continue
		}

		relType, err := getRelationshipType(table, metadata)
		if err != nil {
			return data, err
		}

		expansionColName, err := relationshipExpansionColumnName(relType, metadata)
		if err != nil {
			return data, err
		}
		displayName := snakeToDisplayName(expansionColName)

		expansionColumn := &Column{
			Name: expansionColName,
			Identifiers: ColumnIdentifiers{
				StructProperty: snakeToPascal(expansionColName),
				Projection:     getColProjectionId(table.Name, metadata.Name),
			},
			DisplayName:       displayName,
			IsExpansionColumn: true,
		}

		var relationship *Relationship = nil
		if relType == RelationshipManyToOne {
			relationship = &Relationship{
				Id:                relationshipId(metadata),
				Type:              relType,
				ColumnName:        metadata.ParentColumn,
				ExpansionColumn:   expansionColumn,
				RelatedTableName:  metadata.ReferencedTable,
				RelatedColumnName: metadata.ReferencedColumn,
			}
			data.ManyToOneRelationships = append(data.ManyToOneRelationships, relationship)

		} else {
			relationship = &Relationship{
				Id:                relationshipId(metadata),
				Type:              relType,
				ColumnName:        expansionColName,
				ExpansionColumn:   expansionColumn,
				RelatedTableName:  metadata.ParentTable,
				RelatedColumnName: metadata.ParentColumn,
			}
			data.OneToManyRelationships = append(data.OneToManyRelationships, relationship)
		}

		data.Map[relationship.ColumnName] = relationship
		data.Ordered = append(data.Ordered, relationship)
	}

	slices.SortFunc(data.Ordered, func(a *Relationship, b *Relationship) int {
		return strings.Compare(a.ColumnName, b.ColumnName)
	})

	slices.SortFunc(data.ManyToOneRelationships, func(a *Relationship, b *Relationship) int {
		return strings.Compare(a.ColumnName, b.ColumnName)
	})

	slices.SortFunc(data.OneToManyRelationships, func(a *Relationship, b *Relationship) int {
		return strings.Compare(a.ColumnName, b.ColumnName)
	})

	return data, nil
}

func buildColumnData(
	tableData *builderrepo.TableMetadata,
	relationshipData RelationshipData,
) (ColumnData, error) {
	columnData := ColumnData{
		Map:     map[string]*Column{},
		Ordered: make([]*Column, 0, len(tableData.Columns)),
		All:     make([]*Column, 0, len(tableData.Columns)),
	}

	for _, metadata := range tableData.Columns {
		typeData, err := parseTypeData(relationshipData, metadata)
		if err != nil {
			return columnData, err
		}

		displayName := snakeToDisplayName(metadata.Name)
		if metadata.DisplayName != "" {
			displayName = metadata.DisplayName
		}

		col := &Column{
			Name: metadata.Name,
			Identifiers: ColumnIdentifiers{
				StructProperty: snakeToPascal(metadata.Name),
				Projection:     getColProjectionId(tableData.Name, metadata.Name),
			},
			DisplayName: displayName,
			TypeData:    *typeData,
		}

		columnData.Map[col.Name] = col
		columnData.Ordered = append(columnData.Ordered, col)
		columnData.All = append(columnData.All, col)

		if metadata.PrimaryKey == 1 {
			columnData.PrimaryKey = col
			col.IsPrimaryKey = true
		}
	}

	for _, relationship := range relationshipData.Ordered {
		columnData.Map[relationship.ExpansionColumn.Name] = relationship.ExpansionColumn
		columnData.All = append(columnData.All, relationship.ExpansionColumn)
	}

	slices.SortFunc(columnData.Ordered, func(a *Column, b *Column) int {
		return strings.Compare(a.Name, b.Name)
	})

	slices.SortFunc(columnData.All, func(a *Column, b *Column) int {
		return strings.Compare(a.Name, b.Name)
	})

	return columnData, nil
}

func validateSchemaRelationshipData(schema *Schema) error {
	for _, table := range schema.TableData.Ordered {
		if err := validateRelationshipData(table.RelationshipData); err != nil {
			return err
		}
	}
	return nil
}

func validateRelationshipData(relationshipData RelationshipData) error {

	for _, nToOneRel := range relationshipData.ManyToOneRelationships {
		if nToOneRel.Column.TypeData.DbType != MsqlTypeUniqueIdentifier {
			return fmt.Errorf(
				"invalid relationship: '%s': local column must be of type '%s'",
				nToOneRel.Id,
				MsqlTypeUniqueIdentifier,
			)
		}

		if !nToOneRel.RelatedColumn.IsPrimaryKey {
			return fmt.Errorf(
				"invalid relationship: '%s': foreign column must be a primary key",
				nToOneRel.Id,
			)
		}
	}

	return nil
}

func validateColumnData(columnData ColumnData) error {
	if columnData.PrimaryKey == nil {
		return fmt.Errorf("invalid column data: unable to access the primary key")
	}

	if columnData.PrimaryKey.Name != "id" {
		return fmt.Errorf("invalid column data: the primary key must be named \"id\"")
	}

	if columnData.PrimaryKey.TypeData.DbType != MsqlTypeUniqueIdentifier {
		return fmt.Errorf(
			"invalid column data: the primary key must be of type %s",
			MsqlTypeUniqueIdentifier,
		)
	}

	if len(columnData.All) > 64 {
		return fmt.Errorf(
			"model writer currently only supports tables with up to 64 columns",
		)
	}

	return nil
}

func parseTypeData(
	relationships RelationshipData,
	columnData *builderrepo.ColumnMetadata,
) (*TypeData, error) {
	switch columnData.Type {

	case MsqlTypeUniqueIdentifier:
		return parseUniqueIdentifierType(relationships, columnData)

	case MsqlTypeNvarchar:
		return &TypeData{
			DbType:                   MsqlTypeNvarchar,
			GoTypeIdentifier:         GoColTypeString,
			QueryModelTypeIdentifier: QueryModelTypeString,
			AppType: AppColTypeString{
				MaxLenth: columnData.MaxLength / 2, // convert from bytes to chars(2 bytes)
			},
		}, nil

	case MsqlTypeInt:
		return &TypeData{
			DbType:                   MsqlTypeInt,
			GoTypeIdentifier:         GoColTypeInt64,
			QueryModelTypeIdentifier: QueryModelTypeInt,
			AppType:                  AppColTypeNumber{},
		}, nil

	case MsqlTypeTinyInt:
		if columnData.EnumName != "" {
			return &TypeData{
				DbType:                   MsqlTypeTinyInt,
				GoTypeIdentifier:         enumTypeName(columnData.EnumName),
				QueryModelTypeIdentifier: QueryModelTypeEnum,
				AppType: AppColTypeEnum{
					EnumType: columnData.EnumName,
				},
			}, nil
		} else {
			return &TypeData{
				DbType:                   MsqlTypeTinyInt,
				GoTypeIdentifier:         GoColTypeUInt8,
				QueryModelTypeIdentifier: QueryModelTypeInt,
				AppType:                  AppColTypeNumber{},
			}, nil
		}

	case MsqlTypeDateTimeOffset:
		return &TypeData{
			DbType:                   MsqlTypeDateTimeOffset,
			GoTypeIdentifier:         GoColTypeTime,
			QueryModelTypeIdentifier: QueryModelTypeTime,
			AppType:                  AppColTypeTime{},
		}, nil

	case MsqlTypeGeography:
		return parseGeographyType(columnData)

	default:
		return nil, fmt.Errorf("unable to map '%s' to go type", columnData.Type)
	}
}

func parseUniqueIdentifierType(
	relationships RelationshipData,
	columnData *builderrepo.ColumnMetadata,
) (*TypeData, error) {

	if relationship, exists := relationships.Map[columnData.Name]; exists {
		return &TypeData{
			DbType:                   MsqlTypeUniqueIdentifier,
			GoTypeIdentifier:         GoColTypeId,
			QueryModelTypeIdentifier: QueryModelTypeString,
			AppType: AppColTypeLookup{
				Relationship: relationship,
			},
		}, nil
	}

	return &TypeData{
		DbType:                   MsqlTypeUniqueIdentifier,
		GoTypeIdentifier:         GoColTypeId,
		QueryModelTypeIdentifier: QueryModelTypeString,
		AppType: AppColTypeString{
			MaxLenth: columnData.MaxLength,
		},
	}, nil
}

func parseGeographyType(
	columnData *builderrepo.ColumnMetadata,
) (*TypeData, error) {

	var geographyTypeRegex = regexp.MustCompile(
		`(?i)STGeometryType\]\(\)\s*=\s*'([^']+)'`,
	)

	geographyType := ""
	for _, currentCheck := range columnData.Checks {
		matches := geographyTypeRegex.FindStringSubmatch(currentCheck.Definition)

		if len(matches) == 2 {
			geographyType = matches[1]
		}
	}
	if geographyType == "" {
		return nil, fmt.Errorf("unable to parse geography type: type check not defined")
	}

	switch geographyType {

	case "Point":
		return &TypeData{
			DbType:                   MsqlTypeGeography,
			GoTypeIdentifier:         GoColTypePoint,
			QueryModelTypeIdentifier: QueryModelTypePoint,
			AppType:                  AppColTypePoint{},
		}, nil

	default:
		return nil, fmt.Errorf(
			"unable to parse geography type: unrecognised type: '%s'",
			geographyType,
		)
	}
}

func getRelationshipType(
	tableData *builderrepo.TableMetadata,
	relationship *builderrepo.RelationshipMetadata,
) (relationshipType, error) {
	if relationship.ParentTable == tableData.Name {
		return RelationshipManyToOne, nil
	}

	if relationship.ReferencedTable == tableData.Name {
		return RelationshipOneToMany, nil
	}

	return RelationshipManyToOne, fmt.Errorf(
		"unable to build relationship type: table is not a party to the relationship",
	)
}

func relationshipExpansionColumnName(
	relationshipType relationshipType,
	relationship *builderrepo.RelationshipMetadata,
) (string, error) {
	switch relationshipType {
	case RelationshipManyToOne:
		col := relationship.ParentColumn
		if !strings.HasSuffix(col, "_id") {
			return "", fmt.Errorf("unable to build relationship: 'N' columns must have a _id suffix")
		}
		return strings.TrimSuffix(col, "_id"), nil

	case RelationshipOneToMany:
		return fmt.Sprintf(
			"%s_%s_%s",
			relationship.ParentTable,
			relationship.ReferencedTable,
			relationship.ParentColumn,
		), nil

	default:
		return "", fmt.Errorf("unable to build relationship: invalid relationship type")
	}
}

func relationshipId(relationship *builderrepo.RelationshipMetadata) string {
	return fmt.Sprintf(
		"%s_%s_%s_%s",
		relationship.ParentTable,
		relationship.ParentColumn,
		relationship.ReferencedTable,
		relationship.ReferencedColumn,
	)
}

func enumTypeName(enumTableName string) GoColType {
	return GoColType(fmt.Sprintf("Enum%s", snakeToPascal(enumTableName)))
}

func snakeToDisplayName(s string) string {
	parts := strings.Split(s, "_")

	for i, part := range parts {
		if part == "" {
			continue
		}
		parts[i] = strings.ToUpper(part[:1]) + strings.ToLower(part[1:])
	}

	return strings.Join(parts, " ")
}

func getColProjectionId(tableName, columnName string) string {
	return fmt.Sprintf("%sProjection%s", snakeToCamel(tableName), snakeToPascal(columnName))
}

func snakeToPascal(s string) string {
	parts := strings.Split(s, "_")

	for i, part := range parts {
		if part == "" {
			continue
		}
		parts[i] = strings.ToUpper(part[:1]) + strings.ToLower(part[1:])
	}

	return strings.Join(parts, "")
}

func snakeToCamel(s string) string {
	parts := strings.Split(s, "_")

	for i, part := range parts {
		if part == "" {
			continue
		}
		if i == 0 {
			parts[i] = strings.ToLower(part[:1]) + strings.ToLower(part[1:])
		} else {
			parts[i] = strings.ToUpper(part[:1]) + strings.ToLower(part[1:])
		}
	}

	return strings.Join(parts, "")
}
