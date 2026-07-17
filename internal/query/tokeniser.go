package query

import (
	"fmt"
)

type tokenType uint64

const (
	TokenEOF   tokenType = iota
	TokenEmpty tokenType = iota
	TokenIdentifier
	TokenOperator
	TokenNumber
	TokenString
	TokenParenL
	TokenParenR
	TokenAmper
	TokenSemiColon
	TokenComma
	TokenEquals
)

var singleCharTokenMap = map[byte]tokenType{
	'(': TokenParenL,
	')': TokenParenR,
	'&': TokenAmper,
	';': TokenSemiColon,
	',': TokenComma,
	'=': TokenEquals,
}

type token struct {
	Type  tokenType
	Value string
}

type newTokeniser struct {
	input   string
	idx     int
	current token
}

func NewNewTokeniser(input string) *newTokeniser {
	return &newTokeniser{
		input:   input,
		current: token{Type: TokenEmpty},
	}
}

func (t *newTokeniser) CurrentToken() token {
	return t.current
}

func (t *newTokeniser) NextToken() (token, error) {
	for t.idx < len(t.input) && t.input[t.idx] == ' ' {
		t.idx++
	}

	if t.idx >= (len(t.input)) {
		t.current = token{Type: TokenEOF}
		return t.current, nil
	}

	var err error = nil
	switch b := t.input[t.idx]; {
	case b == '(':
		t.idx++
		t.current, err = token{Type: TokenParenL, Value: "("}, nil
	case b == ')':
		t.idx++
		t.current, err = token{Type: TokenParenR, Value: ")"}, nil
	case b == '&':
		t.idx++
		t.current, err = token{Type: TokenAmper, Value: "&"}, nil
	case b == ';':
		t.idx++
		t.current, err = token{Type: TokenSemiColon, Value: ";"}, nil
	case b == ',':
		t.idx++
		t.current, err = token{Type: TokenComma, Value: ","}, nil
	case b == '=':
		t.idx++
		t.current, err = token{Type: TokenEquals, Value: "="}, nil
	case b == '\'':
		t.current, err = t.readString()
	case b >= '0' && b <= '9':
		t.current, err = t.readNumber()
	default:
		t.current, err = t.readWord()
	}
	return t.current, err
}

func (t newTokeniser) readString() (token, error) {
	t.idx++
	start := t.idx
	for t.idx < len(t.input) && t.input[t.idx] != '\'' {
		t.idx++
	}

	value := t.input[start:t.idx]
	if t.input[t.idx-1] != '\'' {
		return token{}, fmt.Errorf("unterminated string value: %s", value)
	}

	return token{Type: TokenString, Value: value}, nil
}

func (t newTokeniser) readNumber() (token, error) {
	start := t.idx
	for t.idx < len(t.input) {
		b := t.input[t.idx]
		if (b < '0' || b > '9') && b != '.' {
			break
		}
		t.idx++
	}

	return token{
		Type:  TokenNumber,
		Value: t.input[start:t.idx],
	}, nil
}

func (t *newTokeniser) readWord() (token, error) {
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

	value := t.input[start:t.idx]

	switch value {
	case "eq", "ne", "gt", "lt", "and", "or":
		return token{
			Type:  TokenOperator,
			Value: value,
		}, nil

	default:
		return token{
			Type:  TokenIdentifier,
			Value: value,
		}, nil
	}
}
