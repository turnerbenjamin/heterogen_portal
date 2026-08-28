package queryParser

import qstore "github.com/turnerbenjamin/heterogen_portal/internal/query/queryDataStore"

// parseSelectOperation is responsible for parsing select operations It expects
// a simple list of comma-separated values containing at least one column
// identifier
func parseSelectOperation(
	queryDataStore qstore.QueryDataStore,
	t *Tokeniser,
	operationSeparator tokenType,
	operationTerminator tokenType,
) error {
	for {
		tkn := t.Next()
		if tkn.Type != TokenIdentifier {
			return t.TknErr(tkn, "expected a column identifier but received '%s'", tkn.Value)
		}

		if err := queryDataStore.AddSelect(tkn.Value); err != nil {
			return err
		}

		nxt := t.Peek()
		switch nxt.Type {
		case operationSeparator, operationTerminator:
			return nil
		case TokenComma:
			_ = t.Next()
		default:
			return t.TknErr(nxt, "unexpected token encountered '%s'", nxt.Value)
		}
	}
}
