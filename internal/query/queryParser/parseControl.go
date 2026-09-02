package queryParser

import (
	"strings"

	qcfg "github.com/turnerbenjamin/heterogen_portal/internal/query/queryConfiguration"
	qstore "github.com/turnerbenjamin/heterogen_portal/internal/query/queryDataStore"
	qerr "github.com/turnerbenjamin/heterogen_portal/internal/query/queryError"
)

// QueryOperation represents an operations in a query string such as select,
// expand or filter
type QueryOperation interface {

	// IsQueryOperation is used to allow types to opt into the interface
	IsQueryOperation()
}

// QuerySyntaxParser is a utility for parsing query strings
type QuerySyntaxParser struct{}

func NewQueryParser() qcfg.QueryParser {
	return &QuerySyntaxParser{}
}

// Parse is used to parse query strings as a collection of query operations
func (p *QuerySyntaxParser) Parse(
	queryString string,
	s qstore.QueryDataStore,
) error {
	if strings.TrimSpace(queryString) == "" {
		return nil
	}

	tokeniser := NewTokeniser(queryString)

	return parseOperations(
		s,
		tokeniser,
		TokenAmpersand,
		TokenEOF,
	)
}

// parseOperations is used to parse operations at the top-level of query strings
// and those passed as arguments to an expand operation
func parseOperations(
	s qstore.QueryDataStore,
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
				s,
				t,
				operationSeparator,
				operationTerminator,
			); err != nil {
				return err
			}

		case "expand":
			if err := parseExpandOperation(
				s,
				t,
				operationSeparator,
				operationTerminator,
			); err != nil {
				return err
			}

		case "filter":
			if err := parseFilterOperation(
				s,
				t,
				operationSeparator,
				operationTerminator,
			); err != nil {
				return err
			}

		case "orderby":
			if err := parseOrderByOperation(
				s,
				t,
				operationSeparator,
				operationTerminator,
			); err != nil {
				return err
			}

		case "limit":
			if err := parseLimitOperation(s, t); err != nil {
				return err
			}

		case "count":
			if err := parseCountOperation(s, t); err != nil {
				return err
			}

		case "pagingtoken":
			if err := parsePagingTokenOperation(
				s,
				t,
				operationSeparator,
				operationTerminator,
			); err != nil {
				return err
			}

		default:
			return qerr.SyntaxErr("unsupported operator: %s", operator)
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
