package query

import (
	"fmt"
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
	endIdx int
	Type   tokenType
	Value  string
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

func (t *Tokeniser) newTkn(tType tokenType, v string) token {
	return token{
		endIdx: t.idx,
		Type:   tType,
		Value:  v,
	}
}

func (t *Tokeniser) Peek() token {
	return t.PeekN(1)
}

func (t *Tokeniser) PeekN(n int) token {
	o := t.newTkn(TokenEmpty, "")
	savedBuffIdx := t.buffIdx
	for range n {
		o = t.Next()
	}
	t.buffIdx = savedBuffIdx
	return o
}

func (t *Tokeniser) Current() token {
	if t.buffIdx < 0 {
		return t.newTkn(TokenEmpty, "")
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

		t.buffer = append(t.buffer, t.newTkn(TokenEOF, ""))
		t.buffIdx++

		return t.buffer[t.buffIdx]
	}

	var next = token{Type: TokenEmpty}
	switch b := t.input[t.idx]; {
	case b == '(':
		t.idx++
		next = t.newTkn(TokenParenL, "(")
	case b == ')':
		t.idx++
		next = t.newTkn(TokenParenR, ")")
	case b == '&':
		t.idx++
		next = t.newTkn(TokenAmper, "&")
	case b == ';':
		t.idx++
		next = t.newTkn(TokenSemiColon, ";")
	case b == ',':
		t.idx++
		next = t.newTkn(TokenComma, ",")
	case b == '=':
		t.idx++
		next = t.newTkn(TokenEquals, "=")
	case b == '/':
		t.idx++
		next = t.newTkn(TokenSlash, "/")
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

	return t.newTkn(TokenStringRaw, t.input[start:t.idx])
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

	return t.newTkn(TokenNumberRaw, t.input[start:t.idx])
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
		return t.newTkn(TokenLogicalOperator, value)
	}

	if _, isComparisonOperation := comparisonOperators[value]; isComparisonOperation {
		return t.newTkn(TokenComparisonOperator, value)
	}

	if _, isCollectionOperation := collectionOperators[value]; isCollectionOperation {
		return t.newTkn(TokenCollectionOperator, value)
	}

	if value == "null" {
		return t.newTkn(TokenNull, value)
	}

	return t.newTkn(TokenIdentifier, value)
}

func (t *Tokeniser) processRawStringToken(raw token) (token, error) {
	if raw.Type != TokenStringRaw {
		return t.newTkn(TokenEmpty, ""), t.tknErr(raw, "unexpected token type received: '%s'", raw.Value)
	}

	if len(raw.Value) < 2 {
		return t.newTkn(TokenEmpty, ""), syntaxErr("raw string token should be at least 2 characters")
	}

	openingQuotationMark := raw.Value[0]
	closingQuotationMark := raw.Value[len(raw.Value)-1]

	if openingQuotationMark != '\'' && openingQuotationMark != '"' {
		return t.newTkn(TokenEmpty, ""), syntaxErr("raw string token should be prefixed with a ' or \"")
	}

	if closingQuotationMark != openingQuotationMark {
		return t.newTkn(TokenEmpty, ""), syntaxErr("unterminated string literal %s", raw.Value)
	}

	return t.newTkn(TokenString, raw.Value[1:len(raw.Value)-1]), nil
}

func (t Tokeniser) tknErr(tkn token, m string, a ...any) error {
	em := fmt.Sprintf(m, a...)

	maxCtxLen := 50
	ctx := t.input[max(0, tkn.endIdx-maxCtxLen):min(len(t.input), tkn.endIdx)]

	return syntaxErr("%s: __%s <--", em, ctx)

}
