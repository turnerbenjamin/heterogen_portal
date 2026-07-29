package query

import "strings"

// QueryOperation represents an operations in a query string such as select,
// expand or filter
type QueryOperation interface {

	// IsQueryOperation is used to allow types to opt into the interface
	IsQueryOperation()
}

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

// Expand represents the expansion of a specific relationship. The syntax parser
// will populate the relationship name and any nested query operations. The
// metadata bind will later populate link to associate the expansion with
// relationship metadata
type Expand struct {

	// RelationshipName is the name of the relationship at it appears in the
	// query string
	RelationshipName string

	// Operations contains any nested operations passed as an argument to an
	// expansion
	Operations []QueryOperation

	// Link contains metadata for the relationship and resources
	Link *TraversalStep
}

// ExpandOperation is used to retrieve the details of a related record
type ExpandOperation struct {
	Expands []*Expand
}

// IsQueryOperation indicates that ExpandOperation is a QueryOperation
func (o *ExpandOperation) IsQueryOperation() {}

// QuerySyntaxParser is a utility for parsing query strings
type QuerySyntaxParser struct{}

// Parse is used to parse query strings as a collection of query operations
func (p *QuerySyntaxParser) Parse(queryString string) ([]QueryOperation, error) {
	if strings.TrimSpace(queryString) == "" {
		return []QueryOperation{}, nil
	}

	tokeniser := NewTokeniser(queryString)
	return parseOperations(tokeniser, TokenAmpersand, TokenEOF)
}

// parseOperations is used to parse operations at the top-level of query strings
// and those passed as arguments to an expand operation
func parseOperations(
	t *Tokeniser,
	operationSeparator tokenType,
	operationTerminator tokenType,
) ([]QueryOperation, error) {
	o := []QueryOperation{}
	seen := map[string]struct{}{}
	for {
		// Parse next token
		tkn := t.Next()
		// if tkn.Type == TokenEOF || tkn.Type == operationTerminator {
		// 	return o, nil
		// }

		if tkn.Type != TokenIdentifier {
			return nil, t.TknErr(tkn, "expected an operator but received '%s'", tkn.Value)
		}
		operator := strings.ToLower(tkn.Value)
		if _, exists := seen[operator]; exists {
			return nil, t.TknErr(tkn, "invalid re-declaration of the '%s' operator", operator)
		}
		seen[operator] = struct{}{}

		var operationParser func(
			tokeniser *Tokeniser,
			operationSeparator tokenType,
			operationTerminator tokenType,
		) (QueryOperation, error)
		switch operator {
		case "select":
			operationParser = parseSelectOperation
		case "expand":
			operationParser = parseExpandOperation
		case "filter":
			operationParser = parseFilterOperation
		default:
			return nil, syntaxErr("unsupported operator: %s", operator)
		}

		// Expect operator to be followed by equals token
		tkn = t.Next()
		if tkn.Type != TokenEquals {
			return nil, t.TknErr(tkn, "expected '=' but received '%s'", tkn.Value)
		}

		operation, err := operationParser(t, operationSeparator, operationTerminator)
		if err != nil {
			return nil, err
		}
		o = append(o, operation)

		tkn = t.Next()

		switch tkn.Type {
		case operationTerminator:
			return o, nil
		case operationSeparator:
			continue
		default:
			return nil, t.TknErr(tkn, "unexpected token received '%s'", tkn.Value)
		}
	}
}

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

// parseExpandOperation is responsible for parsing expand operations. It expects
// a comma-separated list of column identifiers containing at least one element.
// Optionally, arguments can be passed in parenthesis to a column identifier
// with a semi-colon-separated list of query operations to perform on the
// expansion
func parseExpandOperation(
	t *Tokeniser,
	operationSeparator tokenType,
	operationTerminator tokenType,
) (QueryOperation, error) {
	o := &ExpandOperation{
		Expands: make([]*Expand, 0, 4),
	}

	for {
		tkn := t.Next()
		if tkn.Type != TokenIdentifier {
			return o, t.TknErr(tkn, "expected a column identifier but received '%s'", tkn.Value)
		}
		expansion := &Expand{RelationshipName: tkn.Value, Operations: []QueryOperation{}}

		nxt := t.Peek()
		// If next is an opening parenthesis, parse the nested operations
		if nxt.Type == TokenParenL {
			// consume the opening parenthesis
			_ = t.Next()

			// parse operations returns operations and consumes closing parenthesis
			expandOperations, err := parseOperations(t, TokenSemiColon, TokenParenR)
			if err != nil {
				return nil, err
			}
			expansion.Operations = expandOperations

			// set nxt again
			nxt = t.Peek()
		}
		o.Expands = append(o.Expands, expansion)

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
