package queryParser

import (
	"strings"

	qstore "github.com/turnerbenjamin/heterogen_portal/internal/query/queryDataStore"
	mdl "github.com/turnerbenjamin/heterogen_portal/internal/query/queryModel"
)

func parseFilterOperation(
	s qstore.QueryDataStore,
	t *Tokeniser,
	operationSeparator tokenType,
	endOfOperationsSentinal tokenType,
) error {
	b := s.FilterExpressionBuilder()
	expression, err := parseFilterExpression(b, t)
	if err != nil {
		return err
	}

	nxtTkn := t.Peek()
	if nxtTkn.Type != operationSeparator && nxtTkn.Type != endOfOperationsSentinal {
		return t.TknErr(nxtTkn, "expected end of filter value but received '%s'", nxtTkn.Value)
	}

	s.SetFilterExpression(expression)
	return nil
}

func parseFilterExpression(b qstore.FilterExpressionBuilder, t *Tokeniser) (mdl.FilterExpression, error) {
	return parseOr(b, t)
}

func parseOr(b qstore.FilterExpressionBuilder, t *Tokeniser) (mdl.FilterExpression, error) {

	left, err := parseAnd(b, t)
	if err != nil {
		return nil, err
	}

	for {
		tkn := t.Peek()
		if tkn.Type != TokenLogicalOperator || tkn.Value != "or" {
			break
		}
		t.Next()

		right, err := parseAnd(b, t)
		if err != nil {
			return nil, err
		}

		if left, err = b.NewLogicalExpression(
			left,
			mdl.LogicalOr,
			right,
		); err != nil {
			return nil, err
		}
	}

	return left, nil
}

func parseAnd(b qstore.FilterExpressionBuilder, t *Tokeniser) (mdl.FilterExpression, error) {

	left, err := parsePrimary(b, t)
	if err != nil {
		return nil, err
	}

	for {
		tkn := t.Peek()
		if tkn.Type != TokenLogicalOperator || tkn.Value != "and" {
			break
		}
		t.Next()

		right, err := parsePrimary(b, t)
		if err != nil {
			return nil, err
		}

		if left, err = b.NewLogicalExpression(
			left,
			mdl.LogicalAnd,
			right,
		); err != nil {
			return nil, err
		}
	}

	return left, nil
}

func parsePrimary(b qstore.FilterExpressionBuilder, t *Tokeniser) (mdl.FilterExpression, error) {

	tkn := t.Peek()

	switch tkn.Type {

	case TokenParenL:
		// consume opening parenthesis
		t.Next()

		expression, err := parseFilterExpression(b, t)
		if err != nil {
			return nil, err
		}

		// consume closing parenthesis
		close := t.Next()
		if close.Type != TokenParenR {
			return nil, t.TknErr(close, "expected ')' but received '%s'", close.Value)
		}

		return expression, nil

	default:
		path, err := parsePath(t)
		if err != nil {
			return nil, err
		}

		nxtTkn := t.Peek()

		if nxtTkn.Type == TokenCollectionOperator {
			return parseCollectionOperator(b, t, path)
		}
		return parseComparison(b, t, path)
	}
}

func parseComparison(
	b qstore.FilterExpressionBuilder,
	t *Tokeniser,
	columnPath string,
) (mdl.FilterExpression, error) {
	operator := t.Next()
	if operator.Type != TokenComparisonOperator {
		return nil, t.TknErr(
			operator,
			"expected comparison operator, got %s",
			operator.Value,
		)
	}

	op, err := parseComparisonOperator(t, operator)
	if err != nil {
		return nil, err
	}

	valueBuilder := b.ValueBuilder()
	right, err := parseValue(t, valueBuilder)
	if err != nil {
		return nil, err
	}

	return b.NewComparisonExpression(columnPath, op, right)
}

func parseComparisonOperator(t *Tokeniser, tkn token) (mdl.ComparisonOperator, error) {
	operator, ok := mdl.SupportedComparisonOperators[tkn.Value]
	if !ok {
		return "", t.TknErr(
			tkn,
			"unknown comparison operator %s",
			tkn.Value,
		)
	}
	return operator, nil
}

func parseCollectionOperator(
	b qstore.FilterExpressionBuilder,
	t *Tokeniser,
	resourcePath string,
) (mdl.FilterExpression, error) {

	tkn := t.Next()
	operator, exists := mdl.SupportedCollectionOperators[tkn.Value]
	if !exists {
		return nil, t.TknErr(tkn, "expected collection operator but received '%s'", tkn.Value)
	}

	if tkn = t.Next(); tkn.Type != TokenParenL {
		return nil, t.TknErr(tkn, "expected '(' but received '%s'", tkn.Value)
	}

	filterExpression, err := parseFilterExpression(b, t)
	if err != nil {
		return nil, err
	}

	if tkn := t.Next(); tkn.Type != TokenParenR {
		return nil, t.TknErr(tkn, "expected ')' but received '%s'", tkn.Value)
	}

	return b.NewCollectionExpression(
		resourcePath,
		operator,
		filterExpression,
	)
}

func parsePath(t *Tokeniser) (string, error) {
	pathBuilder := strings.Builder{}
	for {
		tkn := t.Next()

		switch tkn.Type {
		case TokenIdentifier, TokenLogicalOperator, TokenComparisonOperator, TokenCollectionOperator:
			if pathBuilder.Len() != 0 {
				pathBuilder.WriteByte('/')
			}
			pathBuilder.WriteString(tkn.Value)
		}

		nxtTkn := t.Peek()
		if nxtTkn.Type != TokenSlash {
			break
		}

		// consume slash token
		_ = t.Next()

		// If next element is a collection operator break
		if t.Peek().Type == TokenCollectionOperator &&
			t.PeekN(2).Type == TokenParenL {
			break
		}
	}
	return pathBuilder.String(), nil
}
