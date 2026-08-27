package query

// // Expand represents the expansion of a specific relationship. The syntax parser
// // will populate the relationship name and any nested query operations. The
// // metadata bind will later populate link to associate the expansion with
// // relationship metadata
// type Expand struct {

// 	// RelationshipName is the name of the relationship at it appears in the
// 	// query string
// 	RelationshipName string `json:"relationshipName"`

// 	// Operations contains any nested operations passed as an argument to an
// 	// expansion
// 	Operations *Operations `json:"operations"`

// 	// Link contains metadata for the relationship and resources
// 	Link *TraversalStep `json:"-"`
// }

// // ExpandOperation is used to retrieve the details of a related record
// type ExpandOperation struct {
// 	Expands map[string]*Expand `json:"expands"`
// }

// IsQueryOperation indicates that ExpandOperation is a QueryOperation
// func (o *ExpandOperation) IsQueryOperation() {}

// parseExpandOperation is responsible for parsing expand operations. It expects
// a comma-separated list of column identifiers containing at least one element.
// Optionally, arguments can be passed in parenthesis to a column identifier
// with a semi-colon-separated list of query operations to perform on the
// expansion
func parseExpandOperation(
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

		expandOperations, err := queryDataStore.AddExpand(tkn.Value)
		if err != nil {
			return err
		}

		nxt := t.Peek()
		// If next is an opening parenthesis, parse the nested operations
		if nxt.Type == TokenParenL {
			// consume the opening parenthesis
			_ = t.Next()

			// parse operations returns operations and consumes closing parenthesis
			err := parseOperations(expandOperations, t, TokenSemiColon, TokenParenR)
			if err != nil {
				return err
			}

			// set nxt again
			nxt = t.Peek()
		}

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
