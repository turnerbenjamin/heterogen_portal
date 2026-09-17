// Package queryModel contains models used throughout the query package and by
// consuming packages
//
// This file contains models used in the API with consuming packages
package queryModel

// ExecuteResult represents the type returned when Executing a Query
type ExecuteResult struct {
	Count         *int64       `json:"count,omitempty"`
	NextPageToken string       `json:"next_page_token,omitempty"`
	Data          []TableModel `json:"data"`
}

// QueryConfig is used to set configuration options on the query
// executor
type QueryConfig struct {
	// DefaultPageSize represents the default limit applied when no limit is
	// specified in the query string - This property defaults to 5,000
	DefaultPageSize uint32

	// MaxRecordsPerPage represents the maximum number of records, including
	// both top-level and nested records, returned in a given query - This
	// property defaults to 25,000
	MaxRecordsPerPage uint32

	// MaxDepth represents the maximum depth of expand statements - A depth of 1
	// would allow a single expand statement on the top-level query with no
	// nested expand statements - The default is 5
	MaxDepth uint8
}

// QueryConfigWithDefaults populates the queryBuilderConfig with default
// values
func QueryConfigWithDefaults(config QueryConfig) QueryConfig {
	if config.DefaultPageSize == 0 {
		config.DefaultPageSize = 5_000
	}

	if config.MaxRecordsPerPage == 0 {
		config.MaxRecordsPerPage = 25_000
	}

	if config.MaxDepth == 0 {
		config.MaxDepth = 5
	}
	return config
}
