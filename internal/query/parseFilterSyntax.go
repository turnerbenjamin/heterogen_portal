package query

import (
	"slices"
	"strconv"
	"strings"
)

// type FilterOperation struct {
// 	FilterExpression FilterExpression
// }

// func (o *FilterOperation) IsQueryOperation() {}

func parseFilterOperation(
	s QueryDataStore,
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

func parseFilterExpression(b FilterExpressionBuilder, t *Tokeniser) (FilterExpression, error) {
	return parseOr(b, t)
}

func parseOr(b FilterExpressionBuilder, t *Tokeniser) (FilterExpression, error) {

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
			LogicalOr,
			right,
		); err != nil {
			return nil, err
		}
	}

	return left, nil
}

func parseAnd(b FilterExpressionBuilder, t *Tokeniser) (FilterExpression, error) {

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
			LogicalAnd,
			right,
		); err != nil {
			return nil, err
		}
	}

	return left, nil
}

func parsePrimary(b FilterExpressionBuilder, t *Tokeniser) (FilterExpression, error) {

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
	b FilterExpressionBuilder,
	t *Tokeniser,
	columnPath string,
) (FilterExpression, error) {
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

	right, err := parseValue(t)
	if err != nil {
		return nil, err
	}

	return b.NewComparisonExpression(columnPath, op, right)
}

func parseComparisonOperator(t *Tokeniser, tkn token) (ComparisonOperator, error) {
	operator, ok := supportedComparisonOperators[tkn.Value]
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
	b FilterExpressionBuilder,
	t *Tokeniser,
	resourcePath string,
) (FilterExpression, error) {

	tkn := t.Next()
	operator, exists := supportedCollectionOperators[tkn.Value]
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

func parseValue(t *Tokeniser) (ValueExpression, error) {

	tkn := t.Next()

	switch tkn.Type {
	case TokenNull:
		return &NullLiteral{}, nil
	case TokenStringRaw:
		str, err := t.processRawStringToken(tkn)
		if err != nil {
			return &StringLiteral{}, err
		}
		return &StringLiteral{
			Value: str,
		}, nil
	case TokenNumberRaw:
		return processRawNumberToken(tkn)
	case TokenParenL:
		return parseList(t)

	default:
		return nil, t.TknErr(tkn, "unexpected value %s", tkn.Value)
	}
}

func (t *Tokeniser) processRawStringToken(raw token) (string, error) {
	if raw.Type != TokenStringRaw {
		return "", t.TknErr(raw, "unexpected token type received: '%s'", raw.Value)
	}

	if len(raw.Value) < 2 {
		return "", syntaxErr("raw string token should be at least 2 characters")
	}

	openingQuotationMark := raw.Value[0]
	closingQuotationMark := raw.Value[len(raw.Value)-1]

	if openingQuotationMark != '\'' && openingQuotationMark != '"' {
		return "", syntaxErr("raw string token should be prefixed with a ' or \"")
	}

	if closingQuotationMark != openingQuotationMark {
		return "", syntaxErr("unterminated string literal %s", raw.Value)
	}

	return raw.Value[1 : len(raw.Value)-1], nil
}

func parseList(t *Tokeniser) (ValueExpression, error) {
	first, err := parseValue(t)
	if err != nil {
		return nil, err
	}

	if first == nil {
		return nil, syntaxErr("a list literal must have at least 1 element")
	}

	switch first := first.(type) {
	case *StringLiteral:
		vs, err := parseListElements(
			t,
			first.GetTypeName(),
			func(v *StringLiteral) string {
				return v.Value
			},
		)
		if err != nil {
			return nil, err
		}
		return &StringListLiteral{
			Values: slices.Concat([]string{first.Value}, vs),
		}, nil
	case *IntLiteral:
		vs, err := parseListElements(
			t,
			first.GetTypeName(),
			func(v *IntLiteral) int64 {
				return v.Value
			},
		)
		if err != nil {
			return nil, err
		}
		return &IntListLiteral{
			Values: slices.Concat([]int64{first.Value}, vs),
		}, nil
	case *FloatLiteral:
		vs, err := parseListElements(
			t,
			first.GetTypeName(),
			func(v *FloatLiteral) float64 {
				return v.Value
			},
		)
		if err != nil {
			return nil, err
		}
		return &FloatListLiteral{
			Values: slices.Concat([]float64{first.Value}, vs),
		}, nil
	default:
		return nil, syntaxErr("invalid list element type '%s'", first.GetTypeName())
	}
}

func parseListElements[WT ValueExpression, RT any](
	t *Tokeniser,
	listType string,
	unwrap func(WT) RT,
) ([]RT, error) {
	o := []RT{}
	for {
		if t.Peek().Type == TokenParenR {
			_ = t.Next()
			return o, nil
		}

		if t.Peek().Type == TokenComma {
			_ = t.Next()
			continue
		}

		v, err := parseValue(t)
		if err != nil {
			return nil, err
		}

		wv, ok := v.(WT)
		if !ok {
			return nil, syntaxErr("%s lists cannot contain elements of type %s", listType, v.GetTypeName())
		}
		o = append(o, unwrap(wv))
	}
}

func processRawNumberToken(tkn token) (ValueExpression, error) {
	dpCount := 0
	for _, c := range tkn.Value {
		if c == '.' {
			dpCount++
		}
		if dpCount > 1 {
			break
		}
	}

	switch dpCount {
	case 0:
		i, err := strconv.Atoi(tkn.Value)
		if err != nil {
			return &IntLiteral{}, internalErr("unable to convert string to int: %v", err)
		}
		return &IntLiteral{
			Value: int64(i),
		}, nil
	case 1:
		f, err := strconv.ParseFloat(tkn.Value, 64)
		if err != nil {
			return &FloatLiteral{}, internalErr("unable to convert string to float: %v", err)
		}
		return &FloatLiteral{
			Value: f,
		}, nil
	default:
		return &IntLiteral{}, internalErr("invalid number value: %s", tkn.Value)
	}
}
