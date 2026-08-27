package query

var supportedSortDirectionOperators = map[string]SortDirectionOperator{
	"asc":  SortDirectionAsc,
	"desc": SortDirectionDesc,
}

type SortingRule struct {
	ResolvedColumn ResolvedColumn
	Direction      SortDirectionOperator
}

// type OrderByOperation struct {
// 	Rules []*SortingRule `json:"rules"`
// }

// // IsQueryOperation indicates that OrderByOperation is a QueryOperation
// func (o *OrderByOperation) IsQueryOperation() {}

// parseSelectOperation is responsible for parsing select operations It expects
// a simple list of comma-separated values containing at least one column
// identifier
func parseOrderByOperation(
	s QueryDataStore,
	t *Tokeniser,
	operationSeparator tokenType,
	operationTerminator tokenType,
) error {
	for {
		tkn := t.Peek()
		if tkn.Type != TokenIdentifier {
			return t.TknErr(tkn, "expected a column path but received '%s'", tkn.Value)
		}

		path, err := parsePath(t)
		if err != nil {
			return err
		}
		dir := SortDirectionAsc

		nxt := t.Peek()
		if nxt.Type == TokenSortDirectionOperator {
			sortDirectionStr := t.Next()
			sortDirection, ok := supportedSortDirectionOperators[sortDirectionStr.Value]
			if !ok {
				return internalErr(
					"unable to match sort direction token %s to an operator",
					sortDirectionStr.Value,
				)
			}
			dir = sortDirection
		}

		s.AddOrderBy(path, dir)

		nxt = t.Peek()
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
