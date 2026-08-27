package query

import "strings"

// QueryOperation represents an operations in a query string such as select,
// expand or filter
type QueryOperation interface {

	// IsQueryOperation is used to allow types to opt into the interface
	IsQueryOperation()
}

// QuerySyntaxParser is a utility for parsing query strings
type QuerySyntaxParser struct{}

// Parse is used to parse query strings as a collection of query operations
func (p *QuerySyntaxParser) Parse(
	queryString string,
	queryDataStore QueryDataStore,
) error {
	if strings.TrimSpace(queryString) == "" {
		return nil
	}

	tokeniser := NewTokeniser(queryString)

	return parseOperations(queryDataStore, tokeniser, TokenAmpersand, TokenEOF)
}

// parseOperations is used to parse operations at the top-level of query strings
// and those passed as arguments to an expand operation
func parseOperations(
	queryDataStore QueryDataStore,
	t *Tokeniser,
	operationSeparator tokenType,
	operationTerminator tokenType,
) error {
	for {
		// Parse next token
		tkn := t.Next()
		if tkn.Type != TokenIdentifier {
			return t.TknErr(tkn, "expected an operator but received '%s'", tkn.Value)
		}
		operator := strings.ToLower(tkn.Value)

		// Expect operator to be followed by equals token
		tkn = t.Next()
		if tkn.Type != TokenEquals {
			return t.TknErr(tkn, "expected '=' but received '%s'", tkn.Value)
		}

		switch operator {
		case "select":
			if err := parseSelectOperation(
				queryDataStore,
				t,
				operationSeparator,
				operationTerminator,
			); err != nil {
				return err
			}

		case "expand":
			if err := parseExpandOperation(
				queryDataStore,
				t,
				operationSeparator,
				operationTerminator,
			); err != nil {
				return err
			}

		case "filter":
			if err := parseFilterOperation(
				queryDataStore,
				t,
				operationSeparator,
				operationTerminator,
			); err != nil {
				return err
			}

		case "orderby":
			if err := parseOrderByOperation(
				queryDataStore,
				t,
				operationSeparator,
				operationTerminator,
			); err != nil {
				return err
			}

		case "limit":
			if err := parseLimitOperation(queryDataStore, t); err != nil {
				return err
			}

		case "count":
			if err := parseCountOperation(queryDataStore, t); err != nil {
				return err
			}

		case "pagingtoken":
			if err := parsePagingTokenOperation(
				queryDataStore,
				t,
				operationSeparator,
				operationTerminator,
			); err != nil {
				return err
			}

		default:
			return syntaxErr("unsupported operator: %s", operator)
		}

		tkn = t.Next()

		switch tkn.Type {
		case operationTerminator:
			return nil
		case operationSeparator:
			continue
		default:
			return t.TknErr(tkn, "unexpected token received '%s'", tkn.Value)
		}
	}
}
