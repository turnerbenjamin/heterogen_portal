package queryParser

import (
	"strconv"
	"strings"

	qstore "github.com/turnerbenjamin/heterogen_portal/internal/query/queryDataStore"
	qerr "github.com/turnerbenjamin/heterogen_portal/internal/query/queryError"
)

// parseSelectOperation is responsible for parsing limit operations, it expects
// a single number value will be provided
func parseLimitOperation(s qstore.QueryDataStore, t *Tokeniser) error {
	limitValue := t.Next()
	if limitValue.Type != TokenNumberRaw {
		return t.TknErr(limitValue, "expected an integer but received '%s'", limitValue.Value)
	}

	if strings.Contains(limitValue.Value, ".") {
		return t.TknErr(limitValue, "expected an integer but received '%s'", limitValue.Value)
	}

	limitInt, err := strconv.Atoi(limitValue.Value)
	if err != nil {
		return qerr.InternalErr("unable to parse string as int: %v", err)
	}

	return s.SetLimit(limitInt)
}

// parseCountOperation is responsible for parsing count operations. It expects a
// single boolean value will be provided
func parseCountOperation(s qstore.QueryDataStore, t *Tokeniser) error {

	countValue := t.Next()
	if countValue.Type != TokenBool {
		return t.TknErr(countValue, "expected true/false but received '%s'", countValue.Value)
	}

	s.SetDoCount(countValue.Value == "true")
	return nil
}

// parsePagingTokenOperation is responsible for parsing paging tokens, it
// expects a single string value - It will treat all token types, other than
// operation separator, endOfOperationsSentinal and EOF as part of the token and
// leave validation for the token parser
func parsePagingTokenOperation(
	s qstore.QueryDataStore,
	t *Tokeniser,
	operationSeparator tokenType,
	endOfOperationsSentinal tokenType,
) error {
	tknBuilder := strings.Builder{}
	for {
		nxt := t.Peek()
		if nxt.Type == operationSeparator ||
			nxt.Type == endOfOperationsSentinal ||
			nxt.Type == TokenEOF {
			break
		}
		tkn := t.Next()
		_, err := tknBuilder.WriteString(tkn.Value)
		if err != nil {
			return qerr.InternalErr("unable to add token to token builder: %v", err)
		}

	}

	s.SetPagingToken(tknBuilder.String())
	if _, exists := s.PagingToken(); !exists {
		return qerr.SyntaxErr("pagingToken must be provided")
	}
	return nil
}
