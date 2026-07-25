package query

import (
	"net/url"
)

type QueryOperation interface {
	IsQueryOperation() bool
}

type ColumnValue struct {
	columnName string
	columnData ColumnMetadata
}

type RelationshipValue struct {
	relationshipName string
	relationshipData RelationshipMetadata
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
	t *Tokeniser,
	operationSeparator tokenType,
	endOfOperationsSentinal tokenType,
) ([]QueryOperation, error) {
	o := []QueryOperation{}
	seen := map[string]struct{}{}
	for {
		// Parse next token
		tkn := t.Next()
		if tkn.Type != TokenIdentifier {
			return nil, t.tknErr(tkn, "expected an operator but received '%s'", tkn.Value)
		}
		operator := tkn.Value
		if _, exists := seen[operator]; exists {
			return nil, t.tknErr(tkn, "invalid re-declaration of the '%s' operator", operator)
		}
		seen[operator] = struct{}{}

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

		// Expect operator to be followed by equals token
		tkn = t.Next()
		if tkn.Type != TokenEquals {
			return nil, t.tknErr(tkn, "expected '=' but received '%s'", tkn.Value)
		}

		operation, err := operationParser(t, operationSeparator, endOfOperationsSentinal)
		if err != nil {
			return nil, err
		}
		o = append(o, operation)

		tkn = t.Next()

		switch tkn.Type {
		case endOfOperationsSentinal:
			return o, nil
		case operationSeparator:
			continue
		default:
			return nil, t.tknErr(tkn, "unexpected token received '%s'", tkn.Value)
		}
	}
}

func parseSelectOperation(
	t *Tokeniser,
	operationSeparator tokenType,
	endOfOperationsSentinal tokenType,
) (QueryOperation, error) {
	o := &SelectOperation{
		Columns: make([]*ColumnValue, 0, 8),
	}

	for {
		tkn := t.Next()
		if tkn.Type != TokenIdentifier {
			return o, t.tknErr(tkn, "expected a column identifier but received '%s'", tkn.Value)
		}
		o.Columns = append(o.Columns, &ColumnValue{columnName: tkn.Value})

		nxt := t.Peek()
		switch nxt.Type {
		case operationSeparator, endOfOperationsSentinal:
			return o, nil
		case TokenComma:
			_ = t.Next()
		default:
			return o, t.tknErr(nxt, "unexpected token encountered '%s'", nxt.Value)
		}
	}
}

func parseExpandOperation(
	t *Tokeniser,
	operationSeparator tokenType,
	endOfOperationsSentinal tokenType,
) (QueryOperation, error) {
	o := &ExpandOperation{
		Expands: make([]*Expand, 0, 4),
	}

	for {
		tkn := t.Next()
		if tkn.Type != TokenIdentifier {
			return o, t.tknErr(tkn, "expected a column identifier but received '%s'", tkn.Value)
		}
		operator := &Expand{Relationship: &RelationshipValue{relationshipName: tkn.Value}}

		nxt := t.Peek()
		// If next is an opening parenthesis, parse the nested operations
		if nxt.Type == TokenParenL {
			// consume the opening parenthesis
			_ = t.Next()

			// parse operations returns operatiosn and consumes closing parenthesis
			expandOperations, err := ParseOperations(t, TokenSemiColon, TokenParenR)
			if err != nil {
				return nil, err
			}
			operator.Operations = expandOperations

			// set nxt again
			nxt = t.Peek()
		}
		o.Expands = append(o.Expands, operator)

		switch nxt.Type {
		case operationSeparator, endOfOperationsSentinal:
			return o, nil
		case TokenComma:
			_ = t.Next()
		case TokenParenL:

		default:
			return o, t.tknErr(nxt, "unexpected value encountered '%s'", nxt.Value)
		}
	}
}
