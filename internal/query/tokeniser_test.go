package query

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTokeniser_ReadsSingleStringTokens(t *testing.T) {
	t.Parallel()

	testData := []struct {
		name string
		str  string
	}{
		{"Reads single quoted string", "'test'"},
		{"Reads double quoted string", "\"test\""},
		{"Reads empty single quoted string", "''"},
		{"Reads empty double quoted string", "\"\""},
		{"Reads string containing whitespace", "\"a test\t strin \n g\""},
		{"Reads string containing double quotes when single quoted", "'\"a test\"'"},
		{"Reads string containing single quotes when double quoted", "\"'a test'\""},
		{"Reads unterminated single quote string", "'a test"},
		{"Reads unterminated double quote string", "\"a test"},
		{"Reads single quote strings containing other tokens", "'test ()&;,=/ test'"},
	}

	for _, td := range testData {
		t.Run(t.Name(), func(t *testing.T) {
			tokeniser := NewTokeniser(td.str)
			tkn := tokeniser.Next()

			assert.Equal(t, td.str, tkn.Value)
			assert.Equal(t, TokenStringRaw, tkn.Type)
		})
	}
}

func TestTokeniser_CaseHandling(t *testing.T) {
	t.Parallel()

	testData := []struct {
		input           string
		expectedTknType tokenType
	}{
		{"and", TokenLogicalOperator},
		{"AND", TokenLogicalOperator},
		{"or", TokenLogicalOperator},
		{"OR", TokenLogicalOperator},
		{"eq", TokenComparisonOperator},
		{"EQ", TokenComparisonOperator},
		{"ne", TokenComparisonOperator},
		{"NE", TokenComparisonOperator},
		{"gt", TokenComparisonOperator},
		{"GT", TokenComparisonOperator},
		{"ge", TokenComparisonOperator},
		{"gE", TokenComparisonOperator},
		{"Le", TokenComparisonOperator},
		{"LE", TokenComparisonOperator},
		{"lt", TokenComparisonOperator},
		{"lT", TokenComparisonOperator},
		{"in", TokenComparisonOperator},
		{"In", TokenComparisonOperator},
		{"CONTAINS", TokenComparisonOperator},
		{"contains", TokenComparisonOperator},
		{"startswith", TokenComparisonOperator},
		{"startsWith", TokenComparisonOperator},
		{"STARTSWITH", TokenComparisonOperator},
		{"endswith", TokenComparisonOperator},
		{"ENDSWITH", TokenComparisonOperator},
		{"endsWith", TokenComparisonOperator},
		{"any", TokenCollectionOperator},
		{"ANY", TokenCollectionOperator},
		{"Any", TokenCollectionOperator},
		{"all", TokenCollectionOperator},
		{"ALL", TokenCollectionOperator},
		{"All", TokenCollectionOperator},
		{"null", TokenNull},
		{"Null", TokenNull},
		{"NULL", TokenNull},
		{"identifier", TokenIdentifier},
		{"idenTIfier", TokenIdentifier},
		{"IDENTIFIER", TokenIdentifier},
	}

	for _, td := range testData {
		t.Run("test operator type case insensitive but casing retained for identifier values: "+td.input, func(t *testing.T) {
			tokeniser := NewTokeniser(td.input)
			tkn := tokeniser.Next()

			expectedValue := td.input
			if td.expectedTknType != TokenIdentifier {
				expectedValue = strings.ToLower(expectedValue)
			}

			assert.Equal(t, td.expectedTknType, tkn.Type)
			assert.Equal(t, expectedValue, tkn.Value)
		})
	}
}

func TestTokeniser_ReadsNumbers(t *testing.T) {
	t.Parallel()

	testData := []struct {
		input  string
		expect []token
	}{
		{"5", []token{numTkn("5")}},
		{"5.", []token{numTkn("5.0")}},
		{"5.2", []token{numTkn("5.2")}},
		{"5.268", []token{numTkn("5.268")}},
		{"5.795.", []token{numTkn("5.795"), numTkn("0.0")}},
		{"5.798.2", []token{numTkn("5.798"), numTkn("0.2")}},
		{".6", []token{numTkn("0.6")}},
	}

	for _, td := range testData {
		t.Run("reads numbers: "+td.input, func(t *testing.T) {
			tokeniser := NewTokeniser(td.input)

			for _, expected := range td.expect {
				tkn := tokeniser.Next()
				assert.Equal(t, expected.Type, tkn.Type)
				assert.Equal(t, expected.Value, tkn.Value)
			}

			sentinal := tokeniser.Next()
			assert.Equal(t, TokenEOF, sentinal.Type)
		})
	}
}

func TestTokeniser_ReadsComparisonExpressions(t *testing.T) {
	t.Parallel()

	testData := []struct {
		str      string
		expected []token
	}{
		{"given_name eq 'clippy'", []token{
			idntTkn("given_name"),
			compOpTkn("eq"),
			strTkn("'clippy'"),
		}},
		{"user_type ne 5", []token{
			idntTkn("user_type"),
			compOpTkn("ne"),
			numTkn("5"),
		}},
		{"price ge 7.2", []token{
			idntTkn("price"),
			compOpTkn("ge"),
			numTkn("7.2"),
		}},
		{"net gt 9.54", []token{
			idntTkn("net"),
			compOpTkn("gt"),
			numTkn("9.54"),
		}},
		{"net le 5", []token{
			idntTkn("net"),
			compOpTkn("le"),
			numTkn("5"),
		}},
		{"gross lt 2.24", []token{
			idntTkn("gross"),
			compOpTkn("lt"),
			numTkn("2.24"),
		}},
		{"trading_name in ('tn_1', \"tn_2\")", []token{
			idntTkn("trading_name"),
			compOpTkn("in"),
			parenLTkn(),
			strTkn("'tn_1'"),
			commaTkn(),
			strTkn("\"tn_2\""),
			parenRTkn(),
		}},
		{"trading_name startsWith \"Bob's\"", []token{
			idntTkn("trading_name"),
			compOpTkn("startswith"),
			strTkn("\"Bob's\""),
		}},
		{"trading_name contains \"s Bur\"", []token{
			idntTkn("trading_name"),
			compOpTkn("contains"),
			strTkn("\"s Bur\""),
		}},
		{"trading_name endsWith 'rgers'", []token{
			idntTkn("trading_name"),
			compOpTkn("endswith"),
			strTkn("'rgers'"),
		}},
	}

	for _, td := range testData {
		t.Run("Read: "+td.str, func(t *testing.T) {
			tokeniser := NewTokeniser(td.str)

			for _, expectedTkn := range td.expected {
				tkn := tokeniser.Next()
				assert.Equal(t, expectedTkn.Value, tkn.Value)
				assert.Equal(t, expectedTkn.Type, tkn.Type)
			}

			sentinal := tokeniser.Next()
			assert.Equal(t, TokenEOF, sentinal.Type)
		})
	}
}

func TestTokeniser_ReadsLogicalExpressions(t *testing.T) {
	t.Parallel()

	testData := []struct {
		str      string
		expected []token
	}{
		{"e eq 'mc2' and (n gt 3 or (s le 5))", []token{
			idntTkn("e"),
			compOpTkn("eq"),
			strTkn("'mc2'"),
			logicOpTkn("and"),
			parenLTkn(),
			idntTkn("n"),
			compOpTkn("gt"),
			numTkn("3"),
			logicOpTkn("or"),
			parenLTkn(),
			idntTkn("s"),
			compOpTkn("le"),
			numTkn("5"),
			parenRTkn(),
			parenRTkn(),
		}},
		{"trading_name in ('tn1', 'tn2') or (business_type eq 2 and reference startsWith 'a')", []token{
			idntTkn("trading_name"),
			compOpTkn("in"),
			parenLTkn(),
			strTkn("'tn1'"),
			commaTkn(),
			strTkn("'tn2'"),
			parenRTkn(),
			logicOpTkn("or"),
			parenLTkn(),
			idntTkn("business_type"),
			compOpTkn("eq"),
			numTkn("2"),
			logicOpTkn("and"),
			idntTkn("reference"),
			compOpTkn("startswith"),
			strTkn("'a'"),
			parenRTkn(),
		}},
	}

	for _, td := range testData {
		t.Run("Read: "+td.str, func(t *testing.T) {
			tokeniser := NewTokeniser(td.str)

			for _, expectedTkn := range td.expected {
				tkn := tokeniser.Next()
				assert.Equal(t, expectedTkn.Value, tkn.Value)
				assert.Equal(t, expectedTkn.Type, tkn.Type)
			}

			sentinal := tokeniser.Next()
			assert.Equal(t, TokenEOF, sentinal.Type)
		})
	}
}

func TestTokeniser_ReadsCollectionExpressions(t *testing.T) {
	t.Parallel()

	testData := []struct {
		str      string
		expected []token
	}{
		{"business_id/all(reference startswith 'a')", []token{
			idntTkn("business_id"),
			slashTkn(),
			collectionOpTkn("all"),
			parenLTkn(),
			idntTkn("reference"),
			compOpTkn("startswith"),
			strTkn("'a'"),
			parenRTkn(),
		}},
		{"user_id/any(height gt 165.34)", []token{
			idntTkn("user_id"),
			slashTkn(),
			collectionOpTkn("any"),
			parenLTkn(),
			idntTkn("height"),
			compOpTkn("gt"),
			numTkn("165.34"),
			parenRTkn(),
		}},
	}

	for _, td := range testData {
		t.Run("Read: "+td.str, func(t *testing.T) {
			tokeniser := NewTokeniser(td.str)

			for _, expectedTkn := range td.expected {
				tkn := tokeniser.Next()
				assert.Equal(t, expectedTkn.Value, tkn.Value)
				assert.Equal(t, expectedTkn.Type, tkn.Type)
			}

			sentinal := tokeniser.Next()
			assert.Equal(t, TokenEOF, sentinal.Type)
		})
	}
}

func TestTokeniser_Peek(t *testing.T) {
	t.Parallel()

	t.Run("peek does not advance", func(t *testing.T) {
		ids := []string{"id1", "id2", "id3"}
		tokeniser := NewTokeniser(strings.Join(ids, " "))

		for _, id := range ids {
			peekedFirst := tokeniser.Peek()
			peekedSecond := tokeniser.Peek()
			tkn := tokeniser.Next()

			assert.Equal(t, TokenIdentifier, peekedFirst.Type)
			assert.Equal(t, id, peekedFirst.Value)

			assert.Equal(t, TokenIdentifier, peekedSecond.Type)
			assert.Equal(t, id, peekedSecond.Value)

			assert.Equal(t, TokenIdentifier, tkn.Type)
			assert.Equal(t, id, tkn.Value)
		}
	})

	t.Run("peek after end of input returns EOF", func(t *testing.T) {
		ids := []string{""}
		tokeniser := NewTokeniser(strings.Join(ids, " "))

		sentinal := tokeniser.Next()
		peekedFirst := tokeniser.Peek()
		peekedSecond := tokeniser.Peek()

		assert.Equal(t, TokenEOF, sentinal.Type)
		assert.Equal(t, TokenEOF, peekedFirst.Type)
		assert.Equal(t, TokenEOF, peekedSecond.Type)
	})
}

func TestTokeniser_PeekN(t *testing.T) {
	t.Parallel()

	t.Run("peek n returns nth", func(t *testing.T) {
		ids := []string{"id1", "id2", "id3"}
		tokeniser := NewTokeniser(strings.Join(ids, " "))

		for i := range ids {
			tknI := tokeniser.PeekN(i + 1)
			assert.Equal(t, ids[i], tknI.Value)
		}
	})

	t.Run("peek n does not advance tokeniser", func(t *testing.T) {
		ids := []string{"id1", "id2", "id3"}
		tokeniser := NewTokeniser(strings.Join(ids, " "))

		peekedFirst := make([]string, len(ids))
		peekedSecond := make([]string, len(ids))
		for i := range ids {
			peekedFirst[i] = tokeniser.PeekN(i + 1).Value
			peekedSecond[i] = tokeniser.PeekN(i + 1).Value
		}

		for i, id := range ids {
			tkn := tokeniser.Next()

			assert.Equal(t, id, tkn.Value)
			assert.Equal(t, peekedFirst[i], tkn.Value)
			assert.Equal(t, peekedSecond[i], tkn.Value)
		}
	})

	t.Run("peek n beyond end of input returns EOF", func(t *testing.T) {
		ids := []string{""}
		tokeniser := NewTokeniser(strings.Join(ids, " "))

		sentinal := tokeniser.Next()
		peek1 := tokeniser.Peek()
		peek5 := tokeniser.Peek()
		peek25 := tokeniser.Peek()

		assert.Equal(t, TokenEOF, sentinal.Type)
		assert.Equal(t, TokenEOF, peek1.Type)
		assert.Equal(t, TokenEOF, peek5.Type)
		assert.Equal(t, TokenEOF, peek25.Type)
	})

	t.Run("peek n 1 matches peek", func(t *testing.T) {
		ids := []string{"id1", "id2", "id3"}
		tokeniser := NewTokeniser(strings.Join(ids, " "))

		for _, id := range ids {
			peeked := tokeniser.Peek()
			peekedN1 := tokeniser.PeekN(1)

			assert.Equal(t, id, peeked.Value)
			assert.Equal(t, id, peekedN1.Value)

			_ = tokeniser.Next()
		}
	})
}

func TestTokeniser_Current(t *testing.T) {
	t.Parallel()

	t.Run("current after next returns the current token", func(t *testing.T) {
		vals := []string{"val1", "val2"}
		tokeniser := NewTokeniser(strings.Join(vals, " "))

		for _, val := range vals {
			tkn := tokeniser.Next()
			currTkn := tokeniser.Current()

			assert.Equal(t, val, currTkn.Value)
			assert.Equal(t, val, tkn.Value)
		}
	})

	t.Run("current after peek remains unchanged", func(t *testing.T) {
		tokeniser := NewTokeniser("val1 val2")
		_ = tokeniser.Next()

		nxtTkn := tokeniser.Peek()
		currTkn := tokeniser.Current()

		assert.Equal(t, "val1", currTkn.Value)
		assert.Equal(t, "val2", nxtTkn.Value)
	})

	t.Run("current after EOF returns EOF", func(t *testing.T) {
		tokeniser := NewTokeniser("")
		_ = tokeniser.Next()

		currTkn := tokeniser.Current()
		assert.Equal(t, TokenEOF, currTkn.Type)
	})

	t.Run("current before next returns an empty token", func(t *testing.T) {
		tokeniser := NewTokeniser("some string")

		currTkn := tokeniser.Current()
		assert.Equal(t, TokenEmpty, currTkn.Type)
	})
}

func TestTokeniser_TknErr(t *testing.T) {
	t.Parallel()

	const customErrPattern = "custom err"

	testData := []struct {
		name   string
		input  string
		expect string
	}{
		{
			name:   "err message includes context",
			input:  "This is a test string which is going to contain an error right here",
			expect: "custom err: __ring which is going to contain an error right here <--",
		},
		{
			name:   "handles errs with short context",
			input:  "an err",
			expect: "custom err: __an err <--",
		},
		{
			name:   "err",
			input:  "err",
			expect: "custom err: __err <--",
		},
	}

	for _, td := range testData {
		t.Run(td.name, func(t *testing.T) {
			tokeniser := NewTokeniser(td.input)

			tkn := token{}
			for {
				nxt := tokeniser.Peek()
				if nxt.Type == TokenEOF {
					break
				}
				tkn = tokeniser.Next()
			}

			actualErr, ok := tokeniser.TknErr(tkn, customErrPattern).(*queryError)

			assert.True(t, ok)
			assert.Equal(t, QueryErrSyntaxErr, actualErr.category)
			assert.Equal(t, td.expect, actualErr.Error())
		})
	}

	t.Run("omits context when token is EOF", func(t *testing.T) {
		tokeniser := NewTokeniser("")
		tkn := token{Type: TokenEOF}

		actualErr, ok := tokeniser.TknErr(tkn, customErrPattern).(*queryError)

		assert.True(t, ok)
		assert.Equal(t, QueryErrSyntaxErr, actualErr.category)
		assert.Equal(t, customErrPattern, actualErr.Error())
	})

	t.Run("omits context when token is empty", func(t *testing.T) {
		tokeniser := NewTokeniser("")
		tkn := token{Type: TokenEmpty}

		actualErr, ok := tokeniser.TknErr(tkn, customErrPattern).(*queryError)

		assert.True(t, ok)
		assert.Equal(t, QueryErrSyntaxErr, actualErr.category)
		assert.Equal(t, customErrPattern, actualErr.Error())
	})
}

func TestTokeniser_MiscAndEdgeCases(t *testing.T) {
	t.Parallel()

	testData := []struct {
		name     string
		input    string
		expected []token
	}{
		{
			name:     "handles identifiers containing numbers",
			input:    "ident123ifier",
			expected: []token{idntTkn("ident123ifier")},
		},
		{
			name:     "handles identifiers ending in numbers",
			input:    "identifier123",
			expected: []token{idntTkn("identifier123")},
		},
		{
			name:     "handles number followed by an identifier",
			input:    "123.2some_col",
			expected: []token{numTkn("123.2"), idntTkn("some_col")},
		},
		{
			name:     "handles single quote string followed by an identifier",
			input:    "'a string'some_col",
			expected: []token{strTkn("'a string'"), idntTkn("some_col")},
		},
		{
			name:     "handles double quote string followed by an identifier",
			input:    "\"a string\"some_col",
			expected: []token{strTkn("\"a string\""), idntTkn("some_col")},
		},
		{
			name:     "handles identifier followed immedietely by single quote string",
			input:    "some_col'a string'",
			expected: []token{idntTkn("some_col"), strTkn("'a string'")},
		},
		{
			name:     "handles identifier followed immedietely by double quote string",
			input:    "some_col\"a string\"",
			expected: []token{idntTkn("some_col"), strTkn("\"a string\"")},
		},
		{
			name:     "handles adjacent single char tokens",
			input:    ")(;,&=/",
			expected: []token{parenRTkn(), parenLTkn(), semicolonTkn(), commaTkn(), andTkn(), equalsTkn(), slashTkn()},
		},
	}

	for _, td := range testData {
		t.Run(td.name, func(t *testing.T) {
			tokeniser := NewTokeniser(td.input)

			for _, expectedTkn := range td.expected {
				tkn := tokeniser.Next()
				assert.Equal(t, expectedTkn.Value, tkn.Value)
				assert.Equal(t, expectedTkn.Type, tkn.Type)
			}

			sentinal := tokeniser.Next()
			assert.Equal(t, TokenEOF, sentinal.Type)
		})
	}
}

func strTkn(v string) token {
	return token{Type: TokenStringRaw, Value: v}
}

func idntTkn(v string) token {
	return token{Type: TokenIdentifier, Value: v}
}

func numTkn(v string) token {
	return token{Type: TokenNumberRaw, Value: v}
}

func compOpTkn(v string) token {
	return token{Type: TokenComparisonOperator, Value: v}
}

func logicOpTkn(v string) token {
	return token{Type: TokenLogicalOperator, Value: v}
}

func collectionOpTkn(v string) token {
	return token{Type: TokenCollectionOperator, Value: v}
}

func parenLTkn() token    { return token{Type: TokenParenL, Value: "("} }
func parenRTkn() token    { return token{Type: TokenParenR, Value: ")"} }
func andTkn() token       { return token{Type: TokenAmpersand, Value: "&"} }
func semicolonTkn() token { return token{Type: TokenSemiColon, Value: ";"} }
func commaTkn() token     { return token{Type: TokenComma, Value: ","} }
func equalsTkn() token    { return token{Type: TokenEquals, Value: "="} }
func slashTkn() token     { return token{Type: TokenSlash, Value: "/"} }
