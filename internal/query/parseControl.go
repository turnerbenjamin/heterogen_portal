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
		case "orderby":
			operationParser = parseOrderByOperation
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
