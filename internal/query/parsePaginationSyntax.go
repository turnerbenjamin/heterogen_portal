package query

import (
	"strconv"
	"strings"
)

// LimitOperation is used to set the maximim number of rows to return
type LimitOperation struct {
	Limit int `json:"limit"`
}

// IsQueryOperation indicates that LimitOperation is a QueryOperation
func (o *LimitOperation) IsQueryOperation() {}

// CountOperation is a flag used to indicate whether Count should be included
// with the results
type CountOperation struct {
	DoCount bool `json:"doCount"`
}

// IsQueryOperation indicates that CountOperation is a QueryOperation
func (o *CountOperation) IsQueryOperation() {}

// Paging token operations is used to access the next page of results
type PagingTokenOperation struct {
	token string
}

// IsQueryOperation indicates that PagingTokenOperation is a QueryOperation
func (o *PagingTokenOperation) IsQueryOperation() {}

// parseSelectOperation is responsible for parsing limit operations, it expects
// a single number value will be provided
func parseLimitOperation(t *Tokeniser) (*LimitOperation, error) {
	limitValue := t.Next()
	if limitValue.Type != TokenNumberRaw {
		return nil, t.TknErr(limitValue, "expected an integer but received '%s'", limitValue.Value)
	}

	if strings.Contains(limitValue.Value, ".") {
		return nil, t.TknErr(limitValue, "expected an integer but received '%s'", limitValue.Value)
	}

	limitInt, err := strconv.Atoi(limitValue.Value)
	if err != nil {
		return nil, internalErr("unable to parse string as int: %v", err)
	}

	return &LimitOperation{
		Limit: limitInt,
	}, nil
}

// parseCountOperation is responsible for parsing count operations. It expects a
// single boolean value will be provided
func parseCountOperation(t *Tokeniser) (*CountOperation, error) {

	countValue := t.Next()
	if countValue.Type != TokenBool {
		return nil, t.TknErr(countValue, "expected true/false but received '%s'", countValue.Value)
	}

	return &CountOperation{
		DoCount: countValue.Value == "true",
	}, nil
}

// parsePagingTokenOperation is responsible for parsing paging tokens, it
// expects a single string value - It will treat all token types, other than
// operation separator, endOfOperationsSentinal and EOF as part of the token and
// leave validation for the token parser
func parsePagingTokenOperation(
	t *Tokeniser,
	operationSeparator tokenType,
	endOfOperationsSentinal tokenType,
) (*PagingTokenOperation, error) {
	o := &PagingTokenOperation{}

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
			return nil, internalErr("unable to add token to token builder: %v", err)
		}

	}
	o.token = tknBuilder.String()

	if o.token == "" {
		return nil, syntaxErr("pagingToken must be provided")
	}

	return o, nil
}
