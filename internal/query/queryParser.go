package query

import (
	"errors"
	"fmt"
	"net/url"
)

type QueryOperation struct {
	Identifier string
	Values     []QueryOperationValue
}

type QueryOperationValue struct {
	Value            string
	NestedOperations []QueryOperation
}

type QueryParser struct{}

func (p *QueryParser) Parse(queryString string) ([]QueryOperation, error) {
	decodedQuery, err := url.QueryUnescape(queryString)
	if err != nil {
		return nil, err
	}
	tokeniser := NewNewTokeniser(decodedQuery)
	return ParseOperations(tokeniser, TokenAmper, TokenEOF)

}

func ParseOperations(
	tokeniser *newTokeniser,
	operationSeparator tokenType,
	endOfOperationsSentinal tokenType,
) ([]QueryOperation, error) {
	o := []QueryOperation{}
	for {
		// Parse next token
		tkn, err := tokeniser.NextToken()
		if err != nil {
			return nil, err
		}

		// Expect end of operations sentinal or operation identifier
		if tkn.Type == endOfOperationsSentinal || tkn.Type == TokenEOF {
			break
		}

		if tkn.Type != TokenIdentifier {
			return nil, fmt.Errorf("expected an operator but received '%s'", tkn.Value)
		}
		operator := tkn.Value

		// Expect operator to be followed by equals token
		nextTkn, err := tokeniser.NextToken()
		if err != nil {
			return nil, err
		}

		if nextTkn.Type != TokenEquals {
			return nil, fmt.Errorf("expected '=' but received %s", nextTkn.Value)
		}

		var valueParser func(
			tokeniser *newTokeniser,
			operationSeparator tokenType,
			endOfOperationsSentinal tokenType,
		) ([]QueryOperationValue, error)
		switch operator {
		case "select":
			valueParser = selectValueParser
		case "expand":
			valueParser = expandValueParser
		default:
			return nil, fmt.Errorf("unsupported operator: %s", operator)
		}

		// Parse values
		operationValues, err := valueParser(tokeniser, operationSeparator, endOfOperationsSentinal)
		if err != nil {
			return nil, err
		}
		o = append(o, QueryOperation{Identifier: operator, Values: operationValues})

		if tokeniser.CurrentToken().Type == endOfOperationsSentinal {
			break
		}
	}
	return o, nil

}

func selectValueParser(
	tokeniser *newTokeniser,
	operationSeparator tokenType,
	endOfOperationsSentinal tokenType,
) ([]QueryOperationValue, error) {
	o := make([]QueryOperationValue, 0, 8)
	var lastToken *token = nil
	for {
		token, err := tokeniser.NextToken()
		if err != nil {
			return nil, err
		}

		switch token.Type {
		case operationSeparator, endOfOperationsSentinal:
			return o, nil
		case TokenIdentifier:
			if lastToken != nil && lastToken.Type != TokenComma {
				return nil, fmt.Errorf("unexpected identifier received '%s'", token.Value)
			}
			lastToken = &token
			o = append(o, QueryOperationValue{Value: token.Value})
		case TokenComma:
			if lastToken.Type != TokenIdentifier {
				return nil, errors.New("unexpected comma received")
			}
			lastToken = &token
		default:
			return nil, fmt.Errorf("unexpected token received '%s'", token.Value)
		}
	}
}

func expandValueParser(
	tokeniser *newTokeniser,
	operationSeparator tokenType,
	endOfOperationsSentinal tokenType,
) ([]QueryOperationValue, error) {
	o := make([]QueryOperationValue, 0, 8)
	var lastToken *token = nil
	for {
		token, err := tokeniser.NextToken()
		if err != nil {
			return nil, err
		}

		switch token.Type {
		case operationSeparator, endOfOperationsSentinal:
			return o, nil
		case TokenIdentifier:
			if lastToken != nil && lastToken.Type != TokenComma {
				return nil, fmt.Errorf("unexpected identifier received '%s'", token.Value)
			}
			lastToken = &token
			o = append(o, QueryOperationValue{Value: token.Value})
		case TokenParenL:
			if lastToken == nil || lastToken.Type != TokenIdentifier {
				return nil, fmt.Errorf("unexpected ')' received")
			}
			nestedOperations, err := ParseOperations(tokeniser, TokenSemiColon, TokenParenR)
			if err != nil {
				return nil, err
			}
			o[len(o)-1].NestedOperations = nestedOperations
		case TokenComma:
			if lastToken == nil || lastToken.Type != TokenIdentifier {
				return nil, errors.New("unexpected ',' received")
			}
			lastToken = &token
		default:
			return nil, fmt.Errorf("unexpected token received '%s'", token.Value)
		}
	}
}
