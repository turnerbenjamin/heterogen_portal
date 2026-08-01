package query

import (
	"slices"
	"strconv"
)

type FilterOperation struct {
	FilterExpression FilterExpression
}

func (o *FilterOperation) IsQueryOperation() {}

type FilterExpression interface {
	IsFilterExpression()
}

type LogicalOperator int8

const (
	LogicalAnd LogicalOperator = iota
	LogicalOr
)

var supportedLogicalOperators = map[string]LogicalOperator{
	"and": LogicalAnd,
	"or":  LogicalOr,
}

type SortDirectionOperator string

const (
	SortDirectionAsc  SortDirectionOperator = "asc"
	SortDirectionDesc SortDirectionOperator = "desc"
)

type ComparisonOperator string

const (
	ComparisonEq         ComparisonOperator = "eq"
	ComparisonNe         ComparisonOperator = "ne"
	ComparisonGt         ComparisonOperator = "gt"
	ComparisonGe         ComparisonOperator = "ge"
	ComparisonLt         ComparisonOperator = "lt"
	ComparisonLe         ComparisonOperator = "le"
	ComparisonIn         ComparisonOperator = "in"
	ComparisonContains   ComparisonOperator = "contains"
	ComparisonStartsWith ComparisonOperator = "startswith"
	ComparisonEndsWith   ComparisonOperator = "endswith"

	// not supported - included for internal negation logic only:
	comparisonNotIn         ComparisonOperator = "notin"
	comparisonNotContains   ComparisonOperator = "notcontains"
	comparisonNotStartsWith ComparisonOperator = "notstartswith"
	comparisonNotEndsWith   ComparisonOperator = "notendswith"
)

var supportedComparisonOperators = map[string]ComparisonOperator{
	"eq":         ComparisonEq,
	"ne":         ComparisonNe,
	"gt":         ComparisonGt,
	"ge":         ComparisonGe,
	"lt":         ComparisonLt,
	"le":         ComparisonLe,
	"in":         ComparisonIn,
	"contains":   ComparisonContains,
	"startswith": ComparisonStartsWith,
	"endswith":   ComparisonEndsWith,
}

type CollectionOperator string

const (
	CollectionAny CollectionOperator = "any"
	CollectionAll CollectionOperator = "all"
)

var supportedCollectionOperators = map[string]CollectionOperator{
	"any": CollectionAny,
	"all": CollectionAll,
}

type LogicalExpression struct {
	Left     FilterExpression
	Operator LogicalOperator
	Right    FilterExpression
}

func (e *LogicalExpression) IsFilterExpression() {}

type ComparisonExpression struct {
	Path           PropertyPath
	Operator       ComparisonOperator
	Value          ValueExpression
	ResolvedPath   *ResolvedPath
	ResolvedColumn *ColumnValue
	ExistsPlan     *Exists
}

func (e *ComparisonExpression) IsFilterExpression() {}

type CollectionExpression struct {
	Path             PropertyPath
	Operator         CollectionOperator
	FilterExpression FilterExpression
	ResolvedPath     *ResolvedPath
	ExistsPlan       *Exists
}

func (e *CollectionExpression) IsFilterExpression() {}

type PropertyPath struct {
	Segments []string
}

type parser struct {
	t *Tokeniser
}

func parseFilterOperation(
	t *Tokeniser,
	operationSeparator tokenType,
	endOfOperationsSentinal tokenType,
) (QueryOperation, error) {
	p := &parser{
		t: t,
	}

	expression, err := parseFilterExpression(p)
	if err != nil {
		return nil, err
	}

	nxtTkn := t.Peek()
	if nxtTkn.Type != operationSeparator && nxtTkn.Type != endOfOperationsSentinal {
		return nil, t.TknErr(nxtTkn, "expected end of filter value but received '%s'", nxtTkn.Value)
	}

	return &FilterOperation{
		FilterExpression: expression,
	}, nil
}

func parseFilterExpression(p *parser) (FilterExpression, error) {
	return parseOr(p)
}

func parseOr(p *parser) (FilterExpression, error) {

	left, err := parseAnd(p)
	if err != nil {
		return nil, err
	}

	for {
		tkn := p.t.Peek()
		if tkn.Type != TokenLogicalOperator || tkn.Value != "or" {
			break
		}
		p.t.Next()

		right, err := parseAnd(p)
		if err != nil {
			return nil, err
		}

		left = &LogicalExpression{
			Left:     left,
			Operator: LogicalOr,
			Right:    right,
		}
	}

	return left, nil
}

func parseAnd(p *parser) (FilterExpression, error) {

	left, err := parsePrimary(p)
	if err != nil {
		return nil, err
	}

	for {
		tkn := p.t.Peek()
		if tkn.Type != TokenLogicalOperator || tkn.Value != "and" {
			break
		}
		p.t.Next()

		right, err := parsePrimary(p)
		if err != nil {
			return nil, err
		}

		left = &LogicalExpression{
			Left:     left,
			Operator: LogicalAnd,
			Right:    right,
		}
	}

	return left, nil
}

func parsePrimary(p *parser) (FilterExpression, error) {

	tkn := p.t.Peek()

	switch tkn.Type {

	case TokenParenL:
		// consume opening parenthesis
		p.t.Next()

		expression, err := parseFilterExpression(p)
		if err != nil {
			return nil, err
		}

		// consume closing parenthesis
		close := p.t.Next()
		if close.Type != TokenParenR {
			return nil, p.t.TknErr(close, "expected ')' but received '%s'", close.Value)
		}

		return expression, nil

	default:
		path, err := parsePath(p.t)
		if err != nil {
			return nil, err
		}

		nxtTkn := p.t.Peek()

		if nxtTkn.Type == TokenCollectionOperator {
			return parseCollectionOperator(p, path)
		}
		return parseComparison(p, path)
	}
}

func parseComparison(
	p *parser,
	path PropertyPath,
) (FilterExpression, error) {

	if len(path.Segments) == 0 {
		return nil, internalErr("invalid path with a length of 0")
	}

	operator := p.t.Next()
	if operator.Type != TokenComparisonOperator {
		return nil, p.t.TknErr(
			operator,
			"expected comparison operator, got %s",
			operator.Value,
		)
	}

	op, err := parseComparisonOperator(p, operator)
	if err != nil {
		return nil, err
	}

	right, err := parseValue(p)
	if err != nil {
		return nil, err
	}

	return &ComparisonExpression{
		Path:     path,
		Operator: op,
		Value:    right,
	}, nil
}

func parseComparisonOperator(p *parser, tkn token) (ComparisonOperator, error) {
	operator, ok := supportedComparisonOperators[tkn.Value]
	if !ok {
		return "", p.t.TknErr(
			tkn,
			"unknown comparison operator %s",
			tkn.Value,
		)
	}
	return operator, nil
}

func parseCollectionOperator(
	p *parser,
	collection PropertyPath,
) (FilterExpression, error) {

	tkn := p.t.Next()
	operator, exists := supportedCollectionOperators[tkn.Value]
	if !exists {
		return nil, p.t.TknErr(tkn, "expected collection operator but received '%s'", tkn.Value)
	}

	cp := &parser{
		t: p.t,
	}

	if tkn = p.t.Next(); tkn.Type != TokenParenL {
		return nil, p.t.TknErr(tkn, "expected '(' but received '%s'", tkn.Value)
	}

	filterExpression, err := parseFilterExpression(cp)
	if err != nil {
		return nil, err
	}

	if tkn := p.t.Next(); tkn.Type != TokenParenR {
		return nil, p.t.TknErr(tkn, "expected ')' but received '%s'", tkn.Value)
	}

	return &CollectionExpression{
		Path:             collection,
		Operator:         operator,
		FilterExpression: filterExpression,
	}, nil
}

func parsePath(t *Tokeniser) (PropertyPath, error) {
	path := []string{}
	for {
		tkn := t.Next()

		switch tkn.Type {
		case TokenIdentifier, TokenLogicalOperator, TokenComparisonOperator, TokenCollectionOperator:
			path = append(path, tkn.Value)
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
	return PropertyPath{
		Segments: path,
	}, nil
}

func parseValue(p *parser) (ValueExpression, error) {

	tkn := p.t.Next()

	switch tkn.Type {
	case TokenNull:
		return NullLiteral{}, nil
	case TokenStringRaw:
		str, err := p.t.processRawStringToken(tkn)
		if err != nil {
			return StringLiteral{}, err
		}
		return StringLiteral{
			Value: str,
		}, nil
	case TokenNumberRaw:
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
				return IntLiteral{}, internalErr("string conversion failed: %v", err)
			}
			return IntLiteral{
				Value: i,
			}, nil
		case 1:
			f, err := strconv.ParseFloat(tkn.Value, 64)
			if err != nil {
				return FloatLiteral{}, internalErr("string conversion failed: %v", err)
			}
			return FloatLiteral{
				Value: f,
			}, nil
		default:
			return IntLiteral{}, p.t.TknErr(tkn, "invalid number value: %s", tkn.Value)
		}
	case TokenParenL:
		return parseList(p)

	default:
		return nil, p.t.TknErr(tkn, "unexpected value %s", tkn.Value)
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

func parseList(p *parser) (ValueExpression, error) {
	first, err := parseValue(p)
	if err != nil {
		return nil, err
	}

	if first == nil {
		return nil, syntaxErr("a list literal must have at least 1 element")
	}

	switch first := first.(type) {
	case StringLiteral:
		vs, err := parseListElements(
			p,
			first.GetTypeName(),
			func(v StringLiteral) string {
				return v.Value
			},
		)
		if err != nil {
			return nil, err
		}
		return &StringListLiteral{
			Values: slices.Concat([]string{first.Value}, vs),
		}, nil
	case IntLiteral:
		vs, err := parseListElements(
			p,
			first.GetTypeName(),
			func(v IntLiteral) int {
				return v.Value
			},
		)
		if err != nil {
			return nil, err
		}
		return &IntListLiteral{
			Values: slices.Concat([]int{first.Value}, vs),
		}, nil
	case FloatLiteral:
		vs, err := parseListElements(
			p,
			first.GetTypeName(),
			func(v FloatLiteral) float64 {
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
	p *parser,
	listType string,
	unwrap func(WT) RT,
) ([]RT, error) {
	o := []RT{}
	for {
		if p.t.Peek().Type == TokenParenR {
			_ = p.t.Next()
			return o, nil
		}

		if p.t.Peek().Type == TokenComma {
			_ = p.t.Next()
			continue
		}

		v, err := parseValue(p)
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
