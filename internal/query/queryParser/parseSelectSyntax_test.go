package queryParser

// import (
// 	"fmt"
// 	"strings"
// 	"testing"

// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/require"
// )

// func TestParseSelect_Success(t *testing.T) {
// 	testData := []struct {
// 		name string
// 		cols []string
// 	}{
// 		{
// 			name: "single column",
// 			cols: []string{"firstname"},
// 		},
// 		{
// 			name: "multiple columns",
// 			cols: []string{"firstname", "lastname", "username", "id"},
// 		},
// 		{
// 			name: "column names with underscores",
// 			cols: []string{"first_name", "last_name", "user_name"},
// 		},
// 		{
// 			name: "column names with numbers",
// 			cols: []string{"firstname1", "la5tname", "u53rnam3"},
// 		},
// 		{
// 			name: "mixed identifier styles",
// 			cols: []string{"firstname", "last_name", "u53rnam3", "id_1"},
// 		},
// 	}

// 	for _, d := range testData {
// 		t.Run(d.name, func(t *testing.T) {
// 			tokeniser := NewTokeniser(strings.Join(d.cols, ","))
// 			op, err := parseSelectOperation(tokeniser, TokenAmpersand, TokenEOF)

// 			selectOperation, ok := op.(*SelectOperation)
// 			require.True(t, ok)

// 			require.NoError(t, err)
// 			require.NotNil(t, op)
// 			require.Len(t, selectOperation.Columns, len(d.cols))

// 			for i, col := range selectOperation.Columns {
// 				expected := d.cols[i]
// 				assert.Equal(t, expected, col.ColumnName)
// 			}
// 		})
// 	}
// }

// func TestParseSelect_TerminatesCorrectly(t *testing.T) {

// 	testData := []struct {
// 		name                  string
// 		query                 string
// 		operatorSeparator     tokenType
// 		operatorTerminator    tokenType
// 		expectedNextTokenType tokenType
// 	}{
// 		{
// 			name:                  "top level select terminates correctly with operator separator",
// 			query:                 "col1,col2,col3,col4&next",
// 			operatorSeparator:     TokenAmpersand,
// 			operatorTerminator:    TokenEOF,
// 			expectedNextTokenType: TokenAmpersand,
// 		},
// 		{
// 			name:                  "top level select terminates correctly with operator terminator",
// 			query:                 "col1,col2,col3,col4",
// 			operatorSeparator:     TokenAmpersand,
// 			operatorTerminator:    TokenEOF,
// 			expectedNextTokenType: TokenEOF,
// 		},
// 		{
// 			name:                  "nested select terminates correctly with operator separator",
// 			query:                 "col1,col2,col3,col4;next",
// 			operatorSeparator:     TokenSemiColon,
// 			operatorTerminator:    TokenParenR,
// 			expectedNextTokenType: TokenSemiColon,
// 		},
// 		{
// 			name:                  "nested select terminates correctly with operator terminator",
// 			query:                 "col1,col2,col3,col4)",
// 			operatorSeparator:     TokenSemiColon,
// 			operatorTerminator:    TokenParenR,
// 			expectedNextTokenType: TokenParenR,
// 		},
// 	}

// 	for _, d := range testData {
// 		t.Run(d.name, func(t *testing.T) {
// 			tokeniser := NewTokeniser(d.query)
// 			op, err := parseSelectOperation(tokeniser, d.operatorSeparator, d.operatorTerminator)

// 			_, ok := op.(*SelectOperation)
// 			require.True(t, ok)

// 			nextToken := tokeniser.Peek()

// 			assert.NotNil(t, op)
// 			assert.NoError(t, err)
// 			assert.Equal(t, d.expectedNextTokenType, nextToken.Type)
// 		})
// 	}
// }

// func TestParseSelect_InvalidEOFToken(t *testing.T) {
// 	t.Parallel()

// 	testData := []struct {
// 		name               string
// 		operatorSeparator  tokenType
// 		operatorTerminator tokenType
// 		query              string
// 		expectedError      string
// 	}{
// 		{
// 			name:               "EOF at start of string",
// 			operatorSeparator:  TokenAmpersand,
// 			operatorTerminator: TokenEOF,
// 			query:              "",
// 			expectedError:      "expected a column identifier but received ''",
// 		},
// 		{
// 			name:               "EOF after a column separator",
// 			operatorSeparator:  TokenAmpersand,
// 			operatorTerminator: TokenEOF,
// 			query:              "col1,",
// 			expectedError:      "expected a column identifier but received ''",
// 		},
// 		{
// 			name:               "EOF at the end of string when not a terminator",
// 			operatorSeparator:  TokenSemiColon,
// 			operatorTerminator: TokenParenR,
// 			query:              "col1",
// 			expectedError:      "unexpected token encountered ''",
// 		},
// 	}

// 	for _, d := range testData {
// 		t.Run(d.name, func(t *testing.T) {
// 			tokeniser := NewTokeniser(d.query)
// 			op, err := parseSelectOperation(tokeniser, d.operatorSeparator, d.operatorTerminator)

// 			assert.NotNil(t, op)
// 			assert.EqualError(t, err, d.expectedError)
// 		})
// 	}
// }

// func TestParseSelect_InvalidIdentifiers(t *testing.T) {
// 	t.Parallel()

// 	invalidTokens := []token{
// 		commaTkn(),
// 		parenLTkn(),
// 		parenRTkn(),
// 		equalsTkn(),
// 		semicolonTkn(),
// 		andTkn(),
// 	}

// 	testPatterns := []struct {
// 		testName string
// 		ptn      string
// 	}{
// 		{
// 			testName: "invalid token %s at start of string",
// 			ptn:      "%s,col2,col3",
// 		},
// 		{
// 			testName: "invalid token '%s' in middle of string",
// 			ptn:      "col1,%s,col3",
// 		},
// 		{
// 			testName: "invalid token '%s' at end of string",
// 			ptn:      "col1,col2,%s",
// 		},
// 	}

// 	for _, patternData := range testPatterns {
// 		for _, tkn := range invalidTokens {
// 			testName := fmt.Sprintf(patternData.testName, tkn.Value)
// 			t.Run(testName, func(t *testing.T) {
// 				q := fmt.Sprintf(patternData.ptn, tkn.Value)
// 				tokeniser := NewTokeniser(q)

// 				opSeparator := TokenAmpersand
// 				if tkn.Type == TokenAmpersand {
// 					opSeparator = TokenSemiColon
// 				}

// 				op, err := parseSelectOperation(tokeniser, opSeparator, TokenEOF)

// 				expectedError := tokeniser.TknErr(
// 					tokeniser.Current(),
// 					"expected a column identifier but received '%s'",
// 					tkn.Value,
// 				)

// 				assert.NotNil(t, op)
// 				assert.EqualError(t, err, expectedError.Error())
// 			})
// 		}
// 	}
// }

// func TestParseSelect_InvalidSeparators(t *testing.T) {
// 	t.Parallel()

// 	invalidTokens := []token{
// 		parenLTkn(),
// 		parenRTkn(),
// 		equalsTkn(),
// 		semicolonTkn(),
// 		andTkn(),
// 	}

// 	testPatterns := []struct {
// 		testName string
// 		ptn      string
// 	}{
// 		{
// 			testName: "invalid token '%s' as separator",
// 			ptn:      "col1%scol2",
// 		},
// 		{
// 			testName: "invalid token '%s' as trailing separator",
// 			ptn:      "col1,col2%s",
// 		},
// 	}

// 	for _, patternData := range testPatterns {
// 		for _, tkn := range invalidTokens {
// 			testName := fmt.Sprintf(patternData.testName, tkn.Value)
// 			t.Run(testName, func(t *testing.T) {
// 				q := fmt.Sprintf(patternData.ptn, tkn.Value)
// 				tokeniser := NewTokeniser(q)

// 				opSeparator := TokenAmpersand
// 				if tkn.Type == TokenAmpersand {
// 					opSeparator = TokenSemiColon
// 				}

// 				op, err := parseSelectOperation(tokeniser, opSeparator, TokenEOF)

// 				expectedError := tokeniser.TknErr(
// 					tokeniser.Peek(),
// 					"unexpected token encountered '%s'",
// 					tkn.Value,
// 				)

// 				assert.NotNil(t, op)
// 				assert.EqualError(t, err, expectedError.Error())
// 			})
// 		}
// 	}
// }
