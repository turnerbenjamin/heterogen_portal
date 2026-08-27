package query

// // ColumnValue represents a column, the syntax parser populates the column name
// // directly from the query string. The binder will then associate this with
// // column metadata
// type ColumnValue struct {

// 	// ColumnName is the column name as defined in the database
// 	ColumnName string `json:"columnName"`

// 	// ColumnData contains metadata for the column
// 	ColumnData ColumnMetadata `json:"-"`
// }

// // SelectOperation is used to select specific columns from the resource
// type SelectOperation struct {
// 	Columns map[string]*ColumnValue
// }

// // IsQueryOperation indicates that SelectOperation is a QueryOperation
// func (o *SelectOperation) IsQueryOperation() {}

// parseSelectOperation is responsible for parsing select operations It expects
// a simple list of comma-separated values containing at least one column
// identifier
func parseSelectOperation(
	queryDataStore QueryDataStore,
	t *Tokeniser,
	operationSeparator tokenType,
	operationTerminator tokenType,
) error {
	for {
		tkn := t.Next()
		if tkn.Type != TokenIdentifier {
			return t.TknErr(tkn, "expected a column identifier but received '%s'", tkn.Value)
		}

		if err := queryDataStore.AddSelect(tkn.Value); err != nil {
			return err
		}

		nxt := t.Peek()
		switch nxt.Type {
		case operationSeparator, operationTerminator:
			return nil
		case TokenComma:
			_ = t.Next()
		default:
			return t.TknErr(nxt, "unexpected token encountered '%s'", nxt.Value)
		}
	}
}
