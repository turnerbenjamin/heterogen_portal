package query

import (
	"strings"
)

type tokenType uint64

const (
	TokenEmpty tokenType = iota
	TokenEOF
	TokenIdentifier
	TokenLogicalOperator
	TokenComparisonOperator
	TokenCollectionOperator
	TokenNumberRaw
	TokenStringRaw
	TokenString
	TokenNull
	TokenParenL
	TokenParenR
	TokenAmper
	TokenSemiColon
	TokenComma
	TokenEquals
	TokenSlash
)

var comparisonOperators = map[string]ComparisonOperator{
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

var logicalOperators = map[string]LogicalOperator{
	"and": LogicalAnd,
	"or":  LogicalOr,
}

var collectionOperators = map[string]CollectionOperator{
	"any": CollectionAny,
	"all": CollectionAll,
}

var singleCharTokenMap = map[byte]tokenType{
	'(': TokenParenL,
	')': TokenParenR,
	'&': TokenAmper,
	';': TokenSemiColon,
	',': TokenComma,
	'=': TokenEquals,
	'/': TokenSlash,
}

type token struct {
	Type  tokenType
	Value string
}

type Tokeniser struct {
	input   string
	idx     int
	buffIdx int
	buffer  []token
}

func NewTokeniser(input string) *Tokeniser {
	return &Tokeniser{
		input:   input,
		buffer:  make([]token, 0, 64),
		buffIdx: -1,
	}
}

func (t *Tokeniser) Peek() token {
	return t.PeekN(1)
}

func (t *Tokeniser) PeekN(n int) token {
	o := token{Type: TokenEmpty}
	savedBuffIdx := t.buffIdx
	for range n {
		o = t.Next()
	}
	t.buffIdx = savedBuffIdx
	return o
}

func (t *Tokeniser) Current() token {
	if t.buffIdx < 0 {
		return token{Type: TokenEmpty}
	}
	return t.buffer[t.buffIdx]
}

func (t *Tokeniser) Next() token {
	if (t.buffIdx + 1) < len(t.buffer) {
		t.buffIdx++
		return t.buffer[t.buffIdx]
	}

	for t.idx < len(t.input) && t.input[t.idx] == ' ' {
		t.idx++
	}

	if t.idx >= (len(t.input)) {
		t.idx++

		t.buffer = append(t.buffer, token{Type: TokenEOF})
		t.buffIdx++

		return t.buffer[t.buffIdx]
	}

	var next = token{Type: TokenEmpty}
	switch b := t.input[t.idx]; {
	case b == '(':
		t.idx++
		next = token{Type: TokenParenL, Value: "("}
	case b == ')':
		t.idx++
		next = token{Type: TokenParenR, Value: ")"}
	case b == '&':
		t.idx++
		next = token{Type: TokenAmper, Value: "&"}
	case b == ';':
		t.idx++
		next = token{Type: TokenSemiColon, Value: ";"}
	case b == ',':
		t.idx++
		next = token{Type: TokenComma, Value: ","}
	case b == '=':
		t.idx++
		next = token{Type: TokenEquals, Value: "="}
	case b == '/':
		t.idx++
		next = token{Type: TokenSlash, Value: "/"}
	case b == '\'', b == '"':
		next = t.readString()
	case b >= '0' && b <= '9':
		next = t.readNumber()
	default:
		next = t.readIdentifier()
	}

	t.buffer = append(t.buffer, next)
	t.buffIdx++
	return t.buffer[t.buffIdx]
}

func (t *Tokeniser) readString() token {
	// opening and closing chars included in value so consumer can validate the
	// syntax; avoids handling errors for token types that will never return an
	// error
	openingQuotationChar := t.input[t.idx]
	start := t.idx
	t.idx++

	for t.idx < len(t.input) {
		b := t.input[t.idx]
		t.idx++

		if b == openingQuotationChar {
			break
		}
	}

	return token{
		Type:  TokenStringRaw,
		Value: t.input[start:t.idx],
	}
}

func (t *Tokeniser) readNumber() token {
	start := t.idx
	for t.idx < len(t.input) {
		b := t.input[t.idx]
		if (b < '0' || b > '9') && b != '.' {
			break
		}
		t.idx++
	}

	return token{
		Type:  TokenNumberRaw,
		Value: t.input[start:t.idx],
	}
}

func (t *Tokeniser) readIdentifier() token {
	start := t.idx

	for t.idx < len(t.input) {
		b := t.input[t.idx]

		if b == ' ' {
			break
		}

		if _, exists := singleCharTokenMap[b]; exists {
			break
		}

		t.idx++
	}

	value := strings.ToLower(t.input[start:t.idx])
	if _, isLogicalOperation := logicalOperators[value]; isLogicalOperation {
		return token{
			Type:  TokenLogicalOperator,
			Value: value,
		}
	}

	if _, isComparisonOperation := comparisonOperators[value]; isComparisonOperation {
		return token{
			Type:  TokenComparisonOperator,
			Value: value,
		}
	}

	if _, isCollectionOperation := collectionOperators[value]; isCollectionOperation {
		return token{
			Type:  TokenCollectionOperator,
			Value: value,
		}
	}

	if value == "null" {
		return token{
			Type:  TokenNull,
			Value: value,
		}
	}

	return token{
		Type:  TokenIdentifier,
		Value: value,
	}
}

func (t *Tokeniser) processRawStringToken(raw token) (token, error) {
	if raw.Type != TokenStringRaw {
		return token{}, syntaxErr("unexpected token type received")
	}

	if len(raw.Value) < 2 {
		return token{}, syntaxErr("raw string token should be at least 2 characters")
	}

	openingQuotationMark := raw.Value[0]
	closingQuotationMark := raw.Value[len(raw.Value)-1]

	if openingQuotationMark != '\'' && openingQuotationMark != '"' {
		return token{}, syntaxErr("raw string token should be prefixed with a ' or \"")
	}

	if closingQuotationMark != openingQuotationMark {
		return token{}, syntaxErr("unterminated string literal %s", raw.Value)
	}

	return token{
		Type:  TokenString,
		Value: raw.Value[1 : len(raw.Value)-1],
	}, nil
}
