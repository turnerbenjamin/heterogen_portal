// Package queryModel contains models used throughout the query package and by
// consuming packages
//
// This file contains models
package queryModel

// AccessPolicy defines the access policy for a database
type AccessPolicy interface {
	// GetTableAccessPolicy returns an access policy for a given table or nil if
	// the table does not exist
	GetTableAccessPolicy(tableName string) TableAccessPolicy
}

// TableAccessPolicy defines the query access policy for a given database table
type TableAccessPolicy interface {
	// CanAccess defines, at the table level, if a user can perform any
	// operations on that table. ColumnAccessPolicies can restrict access but
	// will override a false value returned from this method
	CanAccess() bool

	// GetColumnAccessPolicy returns an access policy for a specific column or
	// nill if the column cannot be found
	GetColumnAccessPolicy(columnName string) ColumnAccessPolicy
}

// ColumnAccessPolicy defines the query access policy for a given database
// column
type ColumnAccessPolicy interface {
	// CanAccess defines, at the column level, if a user can perform any
	// operations on that column
	CanAccess() bool
}
