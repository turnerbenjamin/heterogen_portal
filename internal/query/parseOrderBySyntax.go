package query

var supportedSortDirectionOperators = map[string]SortDirectionOperator{
	"asc":  SortDirectionAsc,
	"desc": SortDirectionDesc,
}

type SortingRule struct {
	Path           PropertyPath
	Direction      SortDirectionOperator
	ResolvedPath   *ResolvedPath
	ResolvedColumn *ColumnValue
}

type OrderByOperation struct {
	Rules []*SortingRule
}

// IsQueryOperation indicates that OrderByOperation is a QueryOperation
func (o *OrderByOperation) IsQueryOperation() {}

// parseSelectOperation is responsible for parsing select operations It expects
// a simple list of comma-separated values containing at least one column
// identifier
func parseOrderByOperation(
	t *Tokeniser,
	operationSeparator tokenType,
	operationTerminator tokenType,
) (QueryOperation, error) {
	o := &OrderByOperation{
		Rules: make([]*SortingRule, 0, 4),
	}

	for {
		tkn := t.Peek()
		if tkn.Type != TokenIdentifier {
			return o, t.TknErr(tkn, "expected a column path but received '%s'", tkn.Value)
		}

		path, err := parsePath(t)
		if err != nil {
			return nil, err
		}

		sortingRule := &SortingRule{
			Path:      path,
			Direction: SortDirectionAsc,
		}
		o.Rules = append(o.Rules, sortingRule)

		nxt := t.Peek()
		if nxt.Type == TokenSortDirectionOperator {
			sortDirectionStr := t.Next()
			sortDirection, ok := supportedSortDirectionOperators[sortDirectionStr.Value]
			if !ok {
				return nil, internalErr(
					"unable to match sort direction token %s to an operator",
					sortDirectionStr.Value,
				)
			}
			sortingRule.Direction = sortDirection
		}

		nxt = t.Peek()
		switch nxt.Type {
		case operationSeparator, operationTerminator:
			return o, nil
		case TokenComma:
			_ = t.Next()
		default:
			return o, t.TknErr(nxt, "unexpected token encountered '%s'", nxt.Value)
		}
	}
}
