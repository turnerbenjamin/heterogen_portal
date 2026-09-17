package queryParser

import (
	"fmt"
	"strings"

	qerr "github.com/turnerbenjamin/heterogen_portal/internal/query/queryError"
	mdl "github.com/turnerbenjamin/heterogen_portal/internal/query/queryModel"
)

// tokenType identifies the kinds of tokens recognised by the tokeniser.
type tokenType uint64

const (
	// TokenEmpty is an invalid token type returned with errors
	TokenEmpty tokenType = iota

	// TokenEOF represents the end of the input
	TokenEOF

	// TokenSpace represents a space
	TokenSpace

	// TokenLogicalOperator is an identifier matching a supported logical
	// operator
	TokenLogicalOperator

	// TokenComparisonOperator is an identifier matching a supported comparison
	// operator
	TokenComparisonOperator

	// TokenCollectionOperator is an identifier matching a supported collection
	// operator
	TokenCollectionOperator

	// TokenSortDirectionOperator is an identifier matching a supported Sort
	// direction operator
	TokenSortDirectionOperator

	// TokenNull is an identifier equal to null
	TokenNull

	// TokenNullByte is an identifier representing a null byte value
	TokenNullByte

	// TokenIdentifier represents a generic identifier
	TokenIdentifier

	// TokenBool represents a boolean literal
	TokenBool

	// TokenNumberRaw represents a raw number value before validation
	TokenNumberRaw

	// TokenStringRaw represents a raw string value before validation
	TokenStringRaw

	// TokenParenL represents an opening parenthesis '('
	TokenParenL

	// TokenParenR represents a closing parenthesis ')'
	TokenParenR

	// TokenAmpersand represents an ampersand '&'
	TokenAmpersand

	// TokenSemiColon represents a semi-colon ';'
	TokenSemiColon

	// TokenComma represents a comma ','
	TokenComma

	// TokenEquals represents an equals symbol '='
	TokenEquals

	// TokenSlash represents a forward slash '/'
	TokenSlash

	// TokenHyphen represents a hyphen '-'
	TokenHyphen
)

// singleCharTokenMap is used to identify characters that signal the end of an
// identifier
var singleCharTokenMap = map[byte]tokenType{
	'(':    TokenParenL,
	')':    TokenParenR,
	'&':    TokenAmpersand,
	';':    TokenSemiColon,
	',':    TokenComma,
	'=':    TokenEquals,
	'/':    TokenSlash,
	'-':    TokenHyphen,
	'\x00': TokenNullByte,
}

// token represents a single token value
type token struct {
	endIdx int
	Type   tokenType
	Value  string
}

// Tokeniser is used to tokenise a string input
type Tokeniser struct {
	input   string
	idx     int
	buffIdx int
	buffer  []token
}

// NewTokeniser returns a new tokeniser for a given string input
func NewTokeniser(input string) *Tokeniser {
	return &Tokeniser{
		input:   input,
		buffer:  make([]token, 0, 64),
		buffIdx: -1,
	}
}

// newTkn is a helper for creating tokens concisely
func (t *Tokeniser) newTkn(tType tokenType, v string) token {
	return token{
		endIdx: t.idx,
		Type:   tType,
		Value:  v,
	}
}

// Peek returns the next token without moving the tokeniser forward
func (t *Tokeniser) Peek() token {
	return t.PeekN(1)
}

// PeekN returns the nth token from the current position without advancing the
// tokeniser.
func (t *Tokeniser) PeekN(n int) token {
	o := t.newTkn(TokenEmpty, "")
	savedBuffIdx := t.buffIdx
	for range n {
		o = t.Next()
	}
	t.buffIdx = savedBuffIdx
	return o
}

// // Current returns the current token
// func (t *Tokeniser) Current() token {
// 	if t.buffIdx < 0 {
// 		return t.newTkn(TokenEmpty, "")
// 	}
// 	return t.buffer[t.buffIdx]
// }

// Next moves the tokeniser one token forward and returns the token - Space
// tokens are skipped
func (t *Tokeniser) Next() token {
	return t.next(true)
}

// NextIncWhitespace returns the next token, Space tokens are not skipped
func (t *Tokeniser) NextIncSpace() token {
	return t.next(false)
}

func (t *Tokeniser) next(doSkipWhitespace bool) token {
	for (t.buffIdx + 1) < len(t.buffer) {
		t.buffIdx++
		if t.buffer[t.buffIdx].Type != TokenSpace || !doSkipWhitespace {
			return t.buffer[t.buffIdx]
		}
	}

	// collapse whitespace into single space character
	if t.idx < len(t.input) && isWhiteSpace(t.input[t.idx]) {
		for t.idx < len(t.input) && isWhiteSpace(t.input[t.idx]) {
			t.idx++
		}
		t.buffer = append(t.buffer, t.newTkn(TokenSpace, ""))
		t.buffIdx++

		if !doSkipWhitespace {
			return t.buffer[t.buffIdx]
		}
	}

	if t.idx >= len(t.input) {
		t.idx++

		t.buffer = append(t.buffer, t.newTkn(TokenEOF, ""))
		t.buffIdx++

		return t.buffer[t.buffIdx]
	}

	var next = token{Type: TokenEmpty}
	b := t.input[t.idx]
	if tt, exists := singleCharTokenMap[b]; exists {
		t.idx++
		next = t.newTkn(tt, string(b))
	} else {
		switch {
		case b == '\'', b == '"':
			next = t.readString()
		case b >= '0' && b <= '9':
			next = t.readNumber()
		default:
			next = t.readIdentifier()
		}
	}

	t.buffer = append(t.buffer, next)
	t.buffIdx++
	return t.buffer[t.buffIdx]
}

// readString parses string literals. It does not validate that the string is
// terminated. The opening, and if present the closing, quotation marks are
// included to allow validation by the consumer.
func (t *Tokeniser) readString() token {
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

// readNumber reads a number, treating both digits and decimal places to be
// valid - The tokeniser does not validate the number
func (t *Tokeniser) readNumber() token {
	start := t.idx

	dpSeen := false
	for t.idx < len(t.input) {
		b := t.input[t.idx]
		if (b < '0' || b > '9') && b != '.' {
			break
		}

		if b == '.' {
			if dpSeen {
				break
			}
			dpSeen = true
		}

		t.idx++
	}

	numString := t.input[start:t.idx]

	if strings.HasSuffix(numString, ".") {
		numString = numString + "0"
	}

	return t.newTkn(TokenNumberRaw, numString)
}

// readIdentifier parses words separated by whitespace or single character
// tokens.
func (t *Tokeniser) readIdentifier() token {
	start := t.idx

	for t.idx < len(t.input) {
		b := t.input[t.idx]

		if isWhiteSpace(b) {
			break
		}

		if _, exists := singleCharTokenMap[b]; exists {
			break
		}

		if b == '\'' || b == '"' {
			break
		}

		t.idx++
	}

	value := t.input[start:t.idx]
	tokenType, formattedValue := classifyIdentifier(value)

	return t.newTkn(tokenType, formattedValue)
}

// classifyIdentifier tries to match the identifier with a specific identifier
// type, e.g. TokenLogicalOperator. If a specific match cannot be found it falls
// back to the generic TokenIdentifier
func classifyIdentifier(identifier string) (tokenType, string) {
	lIdentifier := strings.ToLower(identifier)

	if _, isLogicalOperation := mdl.SupportedLogicalOperators[lIdentifier]; isLogicalOperation {
		return TokenLogicalOperator, lIdentifier
	}

	if _, isComparisonOperation := mdl.SupportedComparisonOperators[lIdentifier]; isComparisonOperation {
		return TokenComparisonOperator, lIdentifier
	}

	if _, isCollectionOperation := mdl.SupportedCollectionOperators[lIdentifier]; isCollectionOperation {
		return TokenCollectionOperator, lIdentifier
	}

	if _, isSortDirectionOperator := supportedSortDirectionOperators[lIdentifier]; isSortDirectionOperator {
		return TokenSortDirectionOperator, lIdentifier
	}

	if lIdentifier == "null" {
		return TokenNull, lIdentifier
	}

	if lIdentifier == "true" || lIdentifier == "false" {
		return TokenBool, lIdentifier
	}

	return TokenIdentifier, identifier
}

func isWhiteSpace(b byte) bool {
	switch b {
	case '\t', '\n', '\v', '\f', '\r', ' ':
		return true
	default:
		return false
	}
}

// TknErr builds a syntax err with an input substring which terminates at the
// end of the offending tkn to provide context
func (t Tokeniser) TknErr(tkn token, m string, a ...any) error {
	if tkn.Type == TokenEmpty || tkn.Type == TokenEOF {
		return qerr.SyntaxErr(m, a...)
	}

	em := fmt.Sprintf(m, a...)

	maxCtxLen := 50
	ctx := t.input[max(0, tkn.endIdx-maxCtxLen):min(len(t.input), tkn.endIdx)]

	return qerr.SyntaxErr("%s: __%s <--", em, ctx)
}
