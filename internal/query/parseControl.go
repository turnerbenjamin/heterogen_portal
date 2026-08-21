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
func (p *QuerySyntaxParser) Parse(queryString string) (*Operations, error) {
	if strings.TrimSpace(queryString) == "" {
		return &Operations{}, nil
	}

	tokeniser := NewTokeniser(queryString)

	ops, err := parseOperations(tokeniser, TokenAmpersand, TokenEOF)
	if err != nil {
		return nil, err
	}

	//bind operations to the original query string
	ops.QueryString = queryString
	return ops, nil
}

// parseOperations is used to parse operations at the top-level of query strings
// and those passed as arguments to an expand operation
func parseOperations(
	t *Tokeniser,
	operationSeparator tokenType,
	operationTerminator tokenType,
) (*Operations, error) {
	o := &Operations{}
	for {
		// Parse next token
		tkn := t.Next()
		if tkn.Type != TokenIdentifier {
			return nil, t.TknErr(tkn, "expected an operator but received '%s'", tkn.Value)
		}
		operator := strings.ToLower(tkn.Value)

		// Expect operator to be followed by equals token
		tkn = t.Next()
		if tkn.Type != TokenEquals {
			return nil, t.TknErr(tkn, "expected '=' but received '%s'", tkn.Value)
		}

		var parsingError error = nil
		switch operator {
		case "select":
			if o.SelectOperation != nil {
				return nil, t.TknErr(tkn, "invalid re-declaration of the '%s' operator", operator)
			}
			o.SelectOperation, parsingError = parseSelectOperation(t, operationSeparator, operationTerminator)

		case "expand":
			if o.ExpandOperation != nil {
				return nil, t.TknErr(tkn, "invalid re-declaration of the '%s' operator", operator)
			}
			o.ExpandOperation, parsingError = parseExpandOperation(t, operationSeparator, operationTerminator)

		case "filter":
			if o.FilterOperation != nil {
				return nil, t.TknErr(tkn, "invalid re-declaration of the '%s' operator", operator)
			}
			o.FilterOperation, parsingError = parseFilterOperation(t, operationSeparator, operationTerminator)

		case "orderby":
			if o.OrderByOperation != nil {
				return nil, t.TknErr(tkn, "invalid re-declaration of the '%s' operator", operator)
			}
			o.OrderByOperation, parsingError = parseOrderByOperation(t, operationSeparator, operationTerminator)

		case "limit":
			if o.LimitOperation != nil {
				return nil, t.TknErr(tkn, "invalid re-declaration of the '%s' operator", operator)
			}
			o.LimitOperation, parsingError = parseLimitOperation(t)

		case "count":
			if o.CountOperation != nil {
				return nil, t.TknErr(tkn, "invalid re-declaration of the '%s' operator", operator)
			}
			o.CountOperation, parsingError = parseCountOperation(t)

		case "pagingtoken":
			if o.PagingTokenOperation != nil {
				return nil, t.TknErr(tkn, "invalid re-declaration of the '%s' operator", operator)
			}
			o.PagingTokenOperation, parsingError = parsePagingTokenOperation(t, operationSeparator, operationTerminator)

		default:
			return nil, syntaxErr("unsupported operator: %s", operator)
		}

		if parsingError != nil {
			return nil, internalErr("unable to parse operation: %w", parsingError)
		}

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
