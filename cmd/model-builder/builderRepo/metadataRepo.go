package builderRepo

import (
	"database/sql"
	"encoding/json"
)

type MsqlDataTypeName string

type MetadataRepo struct {
	Db *sql.DB
}

type DatabaseSchema struct {
	Tables        []*TableMetadata        `json:"tables"`
	Relationships []*RelationshipMetadata `json:"relationships"`
}

type TableMetadata struct {
	Schema        string            `json:"schema"`
	Name          string            `json:"name"`
	Columns       []*ColumnMetadata `json:"columns"`
	Relationships map[string]*TableRelationship
}

type ColumnMetadata struct {
	Name       string           `json:"name"`
	Type       MsqlDataTypeName `json:"type"`
	MaxLength  int16            `json:"maxLength"`
	PrimaryKey int8             `json:"primaryKey"`
	IsNullable bool             `json:"isNullable"`
	Checks     []Check          `json:"checks"`
}

type Check struct {
	Name         string `json:"name"`
	Definition   string `json:"definition"`
	IsDisabled   bool   `json:"isDisabled"`
	IsNotTrusted bool   `json:"isNotTrusted"`
}

type RelationshipMetadata struct {
	Name             string `json:"name"`
	ParentSchema     string `json:"parentSchema"`
	ParentTable      string `json:"parentTable"`
	ParentColumn     string `json:"parentColumn"`
	ReferencedSchema string `json:"referencedSchema"`
	ReferencedTable  string `json:"referencedTable"`
	ReferencedColumn string `json:"referencedColumn"`
}

type TableRelationship struct {
	Id                     string
	Name                   string
	Type                   string
	ColumnName             string
	ExpansionColumnName    string
	RelationshipColumnType string
	RelatedTable           string
	LocalColumn            string
	ForeignColumn          string
}

func (r *MetadataRepo) GetMetadata() (*DatabaseSchema, error) {
	q := `
SELECT CAST(
(
    SELECT
        JSON_QUERY((
            SELECT
                s.name AS [schema],
                t.name AS [name],
                JSON_QUERY((
                    SELECT
                        c.name AS [name],
                        ty.name AS [type],
                        c.max_length AS [maxLength],
                        c.is_nullable AS [isNullable],
                    CASE
                        WHEN EXISTS (
                            SELECT 1
                            FROM sys.indexes AS i
                            INNER JOIN sys.index_columns AS ic
                                ON ic.object_id = i.object_id
                            AND ic.index_id = i.index_id
                            WHERE i.object_id = c.object_id
                            AND i.is_primary_key = 1
                            AND ic.column_id = c.column_id
                        )
                        THEN 1
                        ELSE 0
                    END AS [primaryKey],

                    JSON_QUERY((
                        SELECT
                            cc.name AS [name],
                            cc.definition AS [definition],
                            cc.is_disabled AS [isDisabled],
                            cc.is_not_trusted AS [isNotTrusted]
                        FROM sys.check_constraints AS cc
                        WHERE cc.parent_object_id = c.object_id
                        AND cc.parent_column_id = c.column_id
                        FOR JSON PATH
                    )) AS [checks]

                    FROM sys.columns AS c
                        INNER JOIN sys.types AS ty
                            ON ty.user_type_id = c.user_type_id
                            AND ty.system_type_id = c.system_type_id
                    WHERE c.object_id = t.object_id
                    ORDER BY c.column_id
                    FOR JSON PATH
                )) AS [columns]
            FROM sys.tables AS t
            INNER JOIN sys.schemas AS s
                ON s.schema_id = t.schema_id
            WHERE t.is_ms_shipped = 0
            ORDER BY
                s.name,
                t.name
            FOR JSON PATH
        )) AS [tables],

        JSON_QUERY((
            SELECT
                fk.name AS [name],
                ps.name AS [parentSchema],
                pt.name AS [parentTable],
                pc.name AS [parentColumn],
                rs.name AS [referencedSchema],
                rt.name AS [referencedTable],
                rc.name AS [referencedColumn]
            FROM sys.foreign_keys AS fk
            INNER JOIN sys.foreign_key_columns AS fkc
                ON fk.object_id = fkc.constraint_object_id
            INNER JOIN sys.tables AS pt
                ON pt.object_id = fk.parent_object_id
            INNER JOIN sys.schemas AS ps
                ON ps.schema_id = pt.schema_id
            INNER JOIN sys.columns AS pc
                ON pc.object_id = pt.object_id
               AND pc.column_id = fkc.parent_column_id
            INNER JOIN sys.tables AS rt
                ON rt.object_id = fk.referenced_object_id
            INNER JOIN sys.schemas AS rs
                ON rs.schema_id = rt.schema_id
            INNER JOIN sys.columns AS rc
                ON rc.object_id = rt.object_id
               AND rc.column_id = fkc.referenced_column_id
            ORDER BY
                ps.name,
                pt.name,
                fk.name
            FOR JSON PATH
        )) AS [relationships]

    FOR JSON PATH, WITHOUT_ARRAY_WRAPPER
) AS nvarchar(max));
    `

	var raw string
	if err := r.Db.QueryRow(q).Scan(&raw); err != nil {
		return nil, err
	}

	var schema DatabaseSchema
	if err := json.Unmarshal([]byte(raw), &schema); err != nil {
		return nil, err
	}

	return &schema, nil

}
