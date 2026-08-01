package query

// ColumnValue represents a column, the syntax parser populates the column name
// directly from the query string. The binder will then associate this with
// column metadata
type ColumnValue struct {

	// ColumnName is the column name as defined in the database
	ColumnName string

	// ColumnData contains metadata for the column
	ColumnData ColumnMetadata
}

// SelectOperation is used to select specific columns from the resource
type SelectOperation struct {
	Columns []*ColumnValue
}

// IsQueryOperation indicates that SelectOperation is a QueryOperation
func (o *SelectOperation) IsQueryOperation() {}

// parseSelectOperation is responsible for parsing select operations It expects
// a simple list of comma-separated values containing at least one column
// identifier
func parseSelectOperation(
	t *Tokeniser,
	operationSeparator tokenType,
	operationTerminator tokenType,
) (QueryOperation, error) {
	o := &SelectOperation{
		Columns: make([]*ColumnValue, 0, 8),
	}

	for {
		tkn := t.Next()
		if tkn.Type != TokenIdentifier {
			return o, t.TknErr(tkn, "expected a column identifier but received '%s'", tkn.Value)
		}
		o.Columns = append(o.Columns, &ColumnValue{ColumnName: tkn.Value})

		nxt := t.Peek()
		switch nxt.Type {
		case operationSeparator, operationTerminator:
			return o, nil
		case TokenComma:
			_ = t.Next()
		default:
			return o, t.TknErr(nxt, "unexpected token encountered '%s'", nxt.Value)
		}
	}
}
