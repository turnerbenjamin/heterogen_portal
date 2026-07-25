package query

import (
	"net/url"

	"github.com/turnerbenjamin/heterogen_portal/internal/model"
)

type QueryOperation interface {
	IsQueryOperation() bool
}

type ColumnValue struct {
	columnName string
	columnData *model.ColumnMetadata
}

type RelationshipValue struct {
	relationshipName string
	relationshipData *model.Relationship
}

type SelectOperation struct {
	Columns []*ColumnValue
}

func (o *SelectOperation) IsQueryOperation() bool { return true }

type Expand struct {
	Relationship *RelationshipValue
	Operations   []QueryOperation
	Link         *TraversalStep
}

type ExpandOperation struct {
	Expands []*Expand
}

func (o *ExpandOperation) IsQueryOperation() bool { return true }

type QuerySyntaxParser struct{}

func (p *QuerySyntaxParser) Parse(queryString string) ([]QueryOperation, error) {
	decodedQuery, err := url.QueryUnescape(queryString)
	if err != nil {
		return nil, internalErr("failed to decode query string: %v", err)
	}
	tokeniser := NewTokeniser(decodedQuery)
	return ParseOperations(tokeniser, TokenAmper, TokenEOF)

}

func ParseOperations(
	tokeniser *Tokeniser,
	operationSeparator tokenType,
	endOfOperationsSentinal tokenType,
) ([]QueryOperation, error) {
	o := []QueryOperation{}
	seen := map[string]struct{}{}
	for {
		// Parse next token
		tkn := tokeniser.Next()
		if tkn.Type != TokenIdentifier {
			return nil, syntaxErr("expected an operator but received '%s'", tkn.Value)
		}
		operator := tkn.Value
		if _, exists := seen[operator]; exists {
			return nil, syntaxErr("invalid redeclaration of the %s operator", operator)
		}

		// Expect operator to be followed by equals token
		tkn = tokeniser.Next()
		if tkn.Type != TokenEquals {
			return nil, syntaxErr("expected '=' but received %s", tkn.Value)
		}

		var operationParser func(
			tokeniser *Tokeniser,
			operationSeparator tokenType,
			endOfOperationsSentinal tokenType,
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

		operation, err := operationParser(tokeniser, operationSeparator, endOfOperationsSentinal)
		if err != nil {
			return nil, err
		}
		o = append(o, operation)

		tkn = tokeniser.Next()

		switch tkn.Type {
		case endOfOperationsSentinal:
			return o, nil
		case operationSeparator:
			continue
		default:
			return nil, syntaxErr("unexpected token received '%s'", tkn.Value)
		}
	}
}

func parseSelectOperation(
	tokeniser *Tokeniser,
	operationSeparator tokenType,
	endOfOperationsSentinal tokenType,
) (QueryOperation, error) {
	o := &SelectOperation{
		Columns: make([]*ColumnValue, 0, 8),
	}

	for {
		token := tokeniser.Next()
		if token.Type != TokenIdentifier {
			return o, syntaxErr("expected a field identifier but received '%s'", token.Value)
		}
		o.Columns = append(o.Columns, &ColumnValue{columnName: token.Value})

		nxt := tokeniser.Peek()
		switch nxt.Type {
		case operationSeparator, endOfOperationsSentinal:
			return o, nil
		case TokenComma:
			_ = tokeniser.Next()
		default:
			return o, syntaxErr("unexpected token encountered '%s'", nxt.Value)
		}
	}
}

func parseExpandOperation(
	tokeniser *Tokeniser,
	operationSeparator tokenType,
	endOfOperationsSentinal tokenType,
) (QueryOperation, error) {
	o := &ExpandOperation{
		Expands: make([]*Expand, 0, 4),
	}

	for {
		token := tokeniser.Next()
		if token.Type != TokenIdentifier {
			return o, syntaxErr("expected a field identifier but received '%s'", token.Value)
		}
		operator := &Expand{Relationship: &RelationshipValue{relationshipName: token.Value}}

		nxt := tokeniser.Peek()
		// If next is an opening parenthesis, parse the nested operations
		if nxt.Type == TokenParenL {
			// consume the opening parenthesis
			_ = tokeniser.Next()

			// parse operations returns operatiosn and consumes closing parenthesis
			expandOperations, err := ParseOperations(tokeniser, TokenSemiColon, TokenParenR)
			if err != nil {
				return nil, err
			}
			operator.Operations = expandOperations

			// set nxt again
			nxt = tokeniser.Peek()
		}
		o.Expands = append(o.Expands, operator)

		switch nxt.Type {
		case operationSeparator, endOfOperationsSentinal:
			return o, nil
		case TokenComma:
			_ = tokeniser.Next()
		case TokenParenL:

		default:
			return o, syntaxErr("unexpected value encountered '%s'", nxt.Value)
		}
	}
}
