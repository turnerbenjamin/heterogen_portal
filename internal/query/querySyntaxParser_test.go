package query

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const validSelect = "select=col1,col2"
const validExpand = "expand=col1,col2(select=col1),col3"
const validFilter = "filter=name eq 'name'"

func TestParse_Success(t *testing.T) {
	t.Parallel()

	testData := []struct {
		name          string
		operations    []string
		expectedTypes []QueryOperation
	}{
		{
			name:          "empty query",
			operations:    []string{},
			expectedTypes: []QueryOperation{},
		},
		{
			name:          "valid select",
			operations:    []string{validSelect},
			expectedTypes: []QueryOperation{&SelectOperation{}},
		},
		{
			name:          "valid expand",
			operations:    []string{validExpand},
			expectedTypes: []QueryOperation{&ExpandOperation{}},
		},
		{
			name:          "valid filter",
			operations:    []string{validFilter},
			expectedTypes: []QueryOperation{&FilterOperation{}}},
		{
			name:       "multiple valid operations",
			operations: []string{validSelect, validExpand, validFilter},
			expectedTypes: []QueryOperation{
				&SelectOperation{},
				&ExpandOperation{},
				&FilterOperation{},
			},
		},
	}

	testParser := QuerySyntaxParser{}
	for _, d := range testData {
		t.Run(d.name, func(t *testing.T) {
			query := strings.Join(d.operations, "&")
			ops, err := testParser.Parse(query)

			require.NoError(t, err)
			require.NotNil(t, ops)
			require.Len(t, ops, len(d.operations))

			for i, op := range ops {
				expectedType := d.expectedTypes[i]
				assert.IsType(t, expectedType, op)
			}
		})
	}
}

func TestParseOperations_Success(t *testing.T) {
	t.Parallel()

	validOperationTestData := []struct {
		name          string
		operations    []string
		expectedTypes []QueryOperation
	}{
		{
			name:          "valid select",
			operations:    []string{validSelect},
			expectedTypes: []QueryOperation{&SelectOperation{}},
		},
		{
			name:          "valid expand",
			operations:    []string{validExpand},
			expectedTypes: []QueryOperation{&ExpandOperation{}},
		},
		{
			name:          "valid filter",
			operations:    []string{validFilter},
			expectedTypes: []QueryOperation{&FilterOperation{}}},
		{
			name:       "multiple valid operations",
			operations: []string{validSelect, validExpand, validFilter},
			expectedTypes: []QueryOperation{
				&SelectOperation{},
				&ExpandOperation{},
				&FilterOperation{},
			},
		},
	}

	operationDelimiterSets := []struct {
		operationSeparator  byte
		operationTerminator byte
	}{
		{
			operationSeparator:  '&',
			operationTerminator: ' ',
		},
		{
			operationSeparator:  ';',
			operationTerminator: ')',
		},
	}

	for _, operationData := range validOperationTestData {
		for _, testDelimiterSet := range operationDelimiterSets {
			testName := fmt.Sprintf(
				"%s - separator: '%c', terminator '%c'",
				operationData.name,
				testDelimiterSet.operationSeparator,
				testDelimiterSet.operationTerminator,
			)

			t.Run(testName, func(t *testing.T) {
				query := strings.Join(operationData.operations, string(testDelimiterSet.operationSeparator))
				query = query + string(testDelimiterSet.operationTerminator)

				separatorToken, ok := singleCharTokenMap[testDelimiterSet.operationSeparator]
				require.True(t, ok)

				terminatorToken, ok := singleCharTokenMap[testDelimiterSet.operationTerminator]
				if !ok {
					terminatorToken = TokenEOF
				}

				tokeniser := NewTokeniser(query)
				ops, err := parseOperations(tokeniser, separatorToken, terminatorToken)

				require.NoError(t, err)
				require.NotNil(t, ops)
				require.Len(t, ops, len(operationData.operations))

				for i, op := range ops {
					expectedType := operationData.expectedTypes[i]
					assert.IsType(t, expectedType, op)
				}
			})
		}
	}
}

func TestParseOperations_DuplicateOperators(t *testing.T) {
	t.Parallel()

	testData := []struct {
		n   string
		q   string
		dup string
	}{
		{
			n:   "reject duplicate select operators",
			q:   "select=trading_name&filter=trading_name eq 'b1'&select=reference&expand=created_by_id",
			dup: "select",
		},
		{
			n:   "reject duplicate filter operators",
			q:   "filter=price gt 2&expand=address(select=line_1)&filter=name",
			dup: "filter",
		},
		{
			n:   "reject duplicate expand operators",
			q:   "filter=name contains 'a'&expand=created_by_id&select=name&expand=address",
			dup: "expand",
		},
		{
			n:   "reject duplicate nested select operators",
			q:   "expand=business_id(select=trading_name;filter=trading_name eq 'b1';select=reference&expand=created_by_id)",
			dup: "select",
		},
		{
			n:   "reject duplicate nested filter operators",
			q:   "expand=business_id(filter=price gt 2;expand=address(select=line_1);filter=name)",
			dup: "filter",
		},
		{
			n:   "reject duplicate nested expand operators",
			q:   "expand=business_id(filter=trading_name contains 'a';expand=created_by_id;select=name;expand=address)",
			dup: "expand",
		},
		{
			n: "allow multiple select operations in seperate scopes",
			q: "select=trading_name&expand=owner_id(select=user_name)",
		},
		{
			n: "allow multiple filter operations in seperate scopes",
			q: "filter=trading_name startswith 'a'&expand=owner_id(filter=user_name startsWith 'b')",
		},
		{
			n: "allow multiple expand operations in seperate scopes",
			q: "expand=business_id(expand=owner_id(expand=primary_business_id))",
		},
	}

	for _, d := range testData {
		t.Run(d.n, func(t *testing.T) {
			tokensier := NewTokeniser(d.q)
			ops, err := parseOperations(tokensier, TokenAmpersand, TokenEOF)

			var expectedErr error
			if d.dup != "" {
				lastToken := tokensier.Current()

				expectedErr = tokensier.TknErr(
					lastToken,
					"invalid re-declaration of the '%s' operator",
					d.dup,
				)
			}

			if expectedErr == nil {
				assert.NotNil(t, ops)
				assert.NoError(t, err)
			} else {
				assert.Nil(t, ops)
				assert.EqualError(t, err, expectedErr.Error())
			}
		})
	}
}

func TestParseOperations_InvalidOperatorValues(t *testing.T) {
	testData := []struct {
		name         string
		input        string
		invalidValue string
		errPattern   string
	}{
		{
			name:         "invalid operator: =",
			input:        "select=trading_name&=",
			invalidValue: "=",
			errPattern:   "expected an operator but received '%s'",
		},
		{
			name:         "invalid operator: ,",
			input:        ",=trading_name&filter=type eq 3",
			invalidValue: ",",
			errPattern:   "expected an operator but received '%s'",
		},
		{
			name:         "invalid operator: (",
			input:        "expand=modifiedby((select=name)",
			invalidValue: "(",
			errPattern:   "expected an operator but received '%s'",
		},
		{
			name:         "invalid operator: )",
			input:        "expand=modifiedby(select=name;))",
			invalidValue: ")",
			errPattern:   "expected an operator but received '%s'",
		},
		{
			name:         "invalid operator: ;",
			input:        "expand=modified_by&;",
			invalidValue: ";",
			errPattern:   "expected an operator but received '%s'",
		},
		{
			name:         "invalid operator: &",
			input:        "expand=modified_by&&",
			invalidValue: "&",
			errPattern:   "expected an operator but received '%s'",
		},
	}

	for _, d := range testData {
		t.Run(d.name, func(t *testing.T) {
			tokensier := NewTokeniser(d.input)
			ops, err := parseOperations(tokensier, TokenAmpersand, TokenEOF)

			expectedErr := tokensier.TknErr(
				tokensier.Current(),
				d.errPattern,
				d.invalidValue,
			)

			assert.Nil(t, ops)
			assert.EqualError(t, err, expectedErr.Error())
		})
	}
}

func TestParseOperations_operatorParsing(t *testing.T) {
	testData := []struct {
		name                string
		input               string
		unsupportedOperator string
	}{
		{
			name:                "case insensitive parsing: SELECT",
			input:               "SELECT=trading_name",
			unsupportedOperator: "",
		},
		{
			name:                "case insensitive parsing: Select",
			input:               "Select=trading_name",
			unsupportedOperator: "",
		},
		{
			name:                "case insensitive parsing: EXPAND",
			input:               "EXPAND=trading_name",
			unsupportedOperator: "",
		},
		{
			name:                "case insensitive parsing: ExpaNd",
			input:               "ExpaNd=trading_name",
			unsupportedOperator: "",
		},
		{
			name:                "case insensitive parsing: FILTER",
			input:               "FILTER=trading_name eq 'b1'",
			unsupportedOperator: "",
		},
		{
			name:                "case insensitive parsing: filteR",
			input:               "filteR=trading_name contains 'a'",
			unsupportedOperator: "",
		},
		{
			name:                "unsupported operator: $select",
			input:               "$select=trading_name",
			unsupportedOperator: "$select",
		},
		{
			name:                "unsupported operator: selet",
			input:               "selet=trading_name",
			unsupportedOperator: "selet",
		},
		{
			name:                "unsupported operator: _expand",
			input:               "_expand=trading_name",
			unsupportedOperator: "_expand",
		},
		{
			name:                "unsupported operator: expan",
			input:               "expan=trading_name",
			unsupportedOperator: "expan",
		},
		{
			name:                "unsupported operator: $filter",
			input:               "$filter=type gt 1",
			unsupportedOperator: "$filter",
		},
		{
			name:                "unsupported operator: filterr",
			input:               "filterr=trading_name eq 'b1'",
			unsupportedOperator: "filterr",
		},
	}

	for _, d := range testData {
		t.Run(d.name, func(t *testing.T) {
			tokensier := NewTokeniser(d.input)
			ops, err := parseOperations(tokensier, TokenAmpersand, TokenEOF)

			if d.unsupportedOperator != "" {
				expectedErr := syntaxErr("unsupported operator: %s", d.unsupportedOperator)

				assert.Nil(t, ops)
				assert.EqualError(t, err, expectedErr.Error())
			} else {
				assert.NotNil(t, ops)
				assert.NoError(t, err)
			}
		})
	}
}

func TestParseOperations_missingEqualsAfterOperator(t *testing.T) {
	testData := []struct {
		name           string
		query          string
		incorrectValue string
	}{
		{
			name:           "EOF after operator",
			query:          "select",
			incorrectValue: "",
		},
		{
			name:           "& after operator",
			query:          "select&expand=created_by",
			incorrectValue: "&",
		},
		{
			name:           "( after operator",
			query:          "expand(created_by)",
			incorrectValue: "(",
		},
		{
			name:           ", after operator",
			query:          "select,expand",
			incorrectValue: ",",
		},
	}

	for _, d := range testData {
		t.Run(d.name, func(t *testing.T) {
			tokensier := NewTokeniser(d.query)
			ops, err := parseOperations(tokensier, TokenAmpersand, TokenEOF)

			expectedErr := tokensier.TknErr(
				tokensier.Current(),
				"expected '=' but received '%s'",
				d.incorrectValue,
			)

			assert.Nil(t, ops)
			assert.EqualError(t, err, expectedErr.Error())
		})
	}
}

func TestParseOperations_Termination(t *testing.T) {

	t.Run("EOF terminates correctly", func(t *testing.T) {
		tokeniser := NewTokeniser("select=trading_name")

		ops, err := parseOperations(tokeniser, TokenAmpersand, TokenEOF)
		currentTkn := tokeniser.Current()

		assert.NotNil(t, ops)
		assert.NoError(t, err)
		assert.Equal(t, TokenEOF, currentTkn.Type)
	})

	t.Run("Right parenthesis terminates correctly", func(t *testing.T) {
		tokeniser := NewTokeniser("select=trading_name)&select=type")

		ops, err := parseOperations(tokeniser, TokenSemiColon, TokenParenR)

		currentTkn := tokeniser.Current()
		nextTkn := tokeniser.Peek()

		assert.NotNil(t, ops)
		assert.NoError(t, err)
		assert.Equal(t, TokenParenR, currentTkn.Type)
		assert.Equal(t, TokenAmpersand, nextTkn.Type)
	})
}

func TestParseSelect_Success(t *testing.T) {
	testData := []struct {
		name string
		cols []string
	}{
		{
			name: "single column",
			cols: []string{"firstname"},
		},
		{
			name: "multiple columns",
			cols: []string{"firstname", "lastname", "username", "id"},
		},
		{
			name: "column names with underscores",
			cols: []string{"first_name", "last_name", "user_name"},
		},
		{
			name: "column names with numbers",
			cols: []string{"firstname1", "la5tname", "u53rnam3"},
		},
		{
			name: "mixed identifier styles",
			cols: []string{"firstname", "last_name", "u53rnam3", "id_1"},
		},
	}

	for _, d := range testData {
		t.Run(d.name, func(t *testing.T) {
			tokeniser := NewTokeniser(strings.Join(d.cols, ","))
			op, err := parseSelectOperation(tokeniser, TokenAmpersand, TokenEOF)

			selectOperation, ok := op.(*SelectOperation)
			require.True(t, ok)

			require.NoError(t, err)
			require.NotNil(t, op)
			require.Len(t, selectOperation.Columns, len(d.cols))

			for i, col := range selectOperation.Columns {
				expected := d.cols[i]
				assert.Equal(t, expected, col.ColumnName)
			}
		})
	}
}

func TestParseSelect_TerminatesCorrectly(t *testing.T) {

	testData := []struct {
		name                  string
		query                 string
		operatorSeparator     tokenType
		operatorTerminator    tokenType
		expectedNextTokenType tokenType
	}{
		{
			name:                  "top level select terminates correctly with operator separator",
			query:                 "col1,col2,col3,col4&next",
			operatorSeparator:     TokenAmpersand,
			operatorTerminator:    TokenEOF,
			expectedNextTokenType: TokenAmpersand,
		},
		{
			name:                  "top level select terminates correctly with operator terminator",
			query:                 "col1,col2,col3,col4",
			operatorSeparator:     TokenAmpersand,
			operatorTerminator:    TokenEOF,
			expectedNextTokenType: TokenEOF,
		},
		{
			name:                  "nested select terminates correctly with operator separator",
			query:                 "col1,col2,col3,col4;next",
			operatorSeparator:     TokenSemiColon,
			operatorTerminator:    TokenParenR,
			expectedNextTokenType: TokenSemiColon,
		},
		{
			name:                  "nested select terminates correctly with operator terminator",
			query:                 "col1,col2,col3,col4)",
			operatorSeparator:     TokenSemiColon,
			operatorTerminator:    TokenParenR,
			expectedNextTokenType: TokenParenR,
		},
	}

	for _, d := range testData {
		t.Run(d.name, func(t *testing.T) {
			tokeniser := NewTokeniser(d.query)
			op, err := parseSelectOperation(tokeniser, d.operatorSeparator, d.operatorTerminator)

			_, ok := op.(*SelectOperation)
			require.True(t, ok)

			nextToken := tokeniser.Peek()

			assert.NotNil(t, op)
			assert.NoError(t, err)
			assert.Equal(t, d.expectedNextTokenType, nextToken.Type)
		})
	}
}

func TestParseSelect_InvalidEOFToken(t *testing.T) {
	t.Parallel()

	testData := []struct {
		name               string
		operatorSeparator  tokenType
		operatorTerminator tokenType
		query              string
		expectedError      string
	}{
		{
			name:               "EOF at start of string",
			operatorSeparator:  TokenAmpersand,
			operatorTerminator: TokenEOF,
			query:              "",
			expectedError:      "expected a column identifier but received ''",
		},
		{
			name:               "EOF after a column separator",
			operatorSeparator:  TokenAmpersand,
			operatorTerminator: TokenEOF,
			query:              "col1,",
			expectedError:      "expected a column identifier but received ''",
		},
		{
			name:               "EOF at the end of string when not a terminator",
			operatorSeparator:  TokenSemiColon,
			operatorTerminator: TokenParenR,
			query:              "col1",
			expectedError:      "unexpected token encountered ''",
		},
	}

	for _, d := range testData {
		t.Run(d.name, func(t *testing.T) {
			tokeniser := NewTokeniser(d.query)
			op, err := parseSelectOperation(tokeniser, d.operatorSeparator, d.operatorTerminator)

			assert.NotNil(t, op)
			assert.EqualError(t, err, d.expectedError)
		})
	}
}

func TestParseSelect_InvalidIdentifiers(t *testing.T) {
	t.Parallel()

	invalidTokens := []token{
		commaTkn(),
		parenLTkn(),
		parenRTkn(),
		equalsTkn(),
		semicolonTkn(),
		andTkn(),
	}

	testPatterns := []struct {
		testName string
		ptn      string
	}{
		{
			testName: "invalid token %s at start of string",
			ptn:      "%s,col2,col3",
		},
		{
			testName: "invalid token '%s' in middle of string",
			ptn:      "col1,%s,col3",
		},
		{
			testName: "invalid token '%s' at end of string",
			ptn:      "col1,col2,%s",
		},
	}

	for _, patternData := range testPatterns {
		for _, tkn := range invalidTokens {
			testName := fmt.Sprintf(patternData.testName, tkn.Value)
			t.Run(testName, func(t *testing.T) {
				q := fmt.Sprintf(patternData.ptn, tkn.Value)
				tokeniser := NewTokeniser(q)

				opSeparator := TokenAmpersand
				if tkn.Type == TokenAmpersand {
					opSeparator = TokenSemiColon
				}

				op, err := parseSelectOperation(tokeniser, opSeparator, TokenEOF)

				expectedError := tokeniser.TknErr(
					tokeniser.Current(),
					"expected a column identifier but received '%s'",
					tkn.Value,
				)

				assert.NotNil(t, op)
				assert.EqualError(t, err, expectedError.Error())
			})
		}
	}
}

func TestParseSelect_InvalidSeparators(t *testing.T) {
	t.Parallel()

	invalidTokens := []token{
		parenLTkn(),
		parenRTkn(),
		equalsTkn(),
		semicolonTkn(),
		andTkn(),
	}

	testPatterns := []struct {
		testName string
		ptn      string
	}{
		{
			testName: "invalid token '%s' as separator",
			ptn:      "col1%scol2",
		},
		{
			testName: "invalid token '%s' as trailing separator",
			ptn:      "col1,col2%s",
		},
	}

	for _, patternData := range testPatterns {
		for _, tkn := range invalidTokens {
			testName := fmt.Sprintf(patternData.testName, tkn.Value)
			t.Run(testName, func(t *testing.T) {
				q := fmt.Sprintf(patternData.ptn, tkn.Value)
				tokeniser := NewTokeniser(q)

				opSeparator := TokenAmpersand
				if tkn.Type == TokenAmpersand {
					opSeparator = TokenSemiColon
				}

				op, err := parseSelectOperation(tokeniser, opSeparator, TokenEOF)

				expectedError := tokeniser.TknErr(
					tokeniser.Peek(),
					"unexpected token encountered '%s'",
					tkn.Value,
				)

				assert.NotNil(t, op)
				assert.EqualError(t, err, expectedError.Error())
			})
		}
	}
}
