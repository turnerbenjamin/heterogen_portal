package azSqlWriter

import (
	"fmt"
	"strings"
	"time"

	qstore "github.com/turnerbenjamin/heterogen_portal/internal/query/queryDataStore"
	qerr "github.com/turnerbenjamin/heterogen_portal/internal/query/queryError"
	mdl "github.com/turnerbenjamin/heterogen_portal/internal/query/queryModel"
	"github.com/turnerbenjamin/heterogen_portal/internal/query/relationships"
)

type stringCoords struct {
	left  int
	right int
}

// arg adds a new argument to the args list and returns a unique placeholder for
// use in the sql statement
func (w *queryWriter) placeholder(v any) string {
	w.args = append(w.args, v)
	return fmt.Sprintf("@p%d", len(w.args))
}

func (w *queryWriter) Write(statement string, args ...any) {
	fmt.Fprintf(w.sb, statement, args...)
}

// aliasedResource binds a resourse to a table alias in an sql query
type aliasedResource struct {
	resource mdl.TableData
	alias    string
}

type queryWriter struct {
	queryDataStore qstore.QueryDataStore
	aliasStore     relationships.AliasStore

	sb             *strings.Builder
	args           []any
	statement      string
	countStatement string
	fromLocation   stringCoords
	filterLocation stringCoords
}

func NewQueryWriter(s qstore.QueryDataStore) (mdl.QueryWriter, error) {
	w := queryWriter{
		queryDataStore: s,
		aliasStore:     s.GetAliasStore(),
		sb:             &strings.Builder{},
		args:           []any{},
	}

	err := w.writeQuery()
	if err != nil {
		return nil, err
	}
	w.statement = w.sb.String()

	return &w, nil
}

func (w queryWriter) WriteQueryStatement() string {
	return w.statement
}

func (w queryWriter) WriteCountStatement() string {
	if w.countStatement == "" {
		w.buildCountStatement()
	}
	return w.countStatement
}

func (w queryWriter) Args() []any {
	return w.args
}

func (w *queryWriter) writeQuery() error {
	err := w.writeSelectAndExpandStatement()
	if err != nil {
		return err
	}

	w.writeFromStatement()

	w.writeJoins(w.queryDataStore.JoinCollection())

	err = w.writeFilterStatement()
	if err != nil {
		return err
	}

	err = w.writeOrderByStatement()
	if err != nil {
		return err
	}

	// request json response
	w.sb.WriteString("FOR JSON PATH")
	if w.queryDataStore.IsTopLevelQuery() {
		w.sb.WriteRune(';')
	}

	return nil
}

func (w *queryWriter) buildCountStatement() {
	coreQueryString := w.sb.String()

	countSb := strings.Builder{}
	countSb.WriteString("SELECT COUNT(*) AS total_count")
	countSb.WriteRune(' ')

	fromString := coreQueryString[w.fromLocation.left:w.fromLocation.right]
	countSb.WriteString(fromString)
	countSb.WriteRune(' ')

	filterString := coreQueryString[w.filterLocation.left:w.filterLocation.right]
	countSb.WriteString(filterString)
	countSb.WriteRune(';')

	w.countStatement = countSb.String()
}

func (w *queryWriter) writeFromStatement() {
	w.fromLocation.left = w.sb.Len()
	rootResourceMetadata := w.queryDataStore.RootResourceMetadata()

	rootAlias := w.aliasStore.GetRootAlias()

	w.sb.WriteString("FROM ")
	w.sb.WriteString(rootResourceMetadata.FullyQualifiedName)
	w.sb.WriteRune(' ')
	w.sb.WriteString(rootAlias)

	w.fromLocation.right = w.sb.Len()

	w.sb.WriteRune(' ')
}

func (w *queryWriter) writeSelectAndExpandStatement() error {
	w.sb.WriteString("SELECT ")

	// Add top statement to implement limit
	w.writeTopStatement()

	// Write all selected columns
	i := 0
	for _, columnValue := range w.queryDataStore.Selects() {
		if i > 0 {
			w.sb.WriteString(",")
		}

		fmt.Fprintf(
			w.sb,
			"%s.%s",
			w.aliasStore.GetRootAlias(),
			w.formatSelectValue(columnValue.Metadata),
		)
		i++
	}
	if i == 0 {
		return qerr.InternalErr("expected a select operation with at least one column specified")
	}

	// Write expansions
	for _, expansion := range w.queryDataStore.Expands() {
		w.sb.WriteString(",")
		if err := w.writeExpandColumn(expansion); err != nil {
			return err
		}
	}

	w.sb.WriteRune(' ')
	return nil
}

func (w *queryWriter) formatSelectValue(columnData mdl.ColumnData) string {
	switch columnData.Type {
	case mdl.DbTypePoint:
		return fmt.Sprintf("%s.STAsText() AS %s", columnData.Name, columnData.Name)
	default:
		return columnData.Name
	}
}

func (w *queryWriter) writeExpandColumn(expansion qstore.Expansion) error {
	nestedWriter, err := NewQueryWriter(expansion.QueryData)
	if err != nil {
		return err
	}

	w.sb.WriteString("JSON_QUERY((")
	w.sb.WriteString(nestedWriter.WriteQueryStatement())
	if expansion.TraversalStep.Relationship.Type == mdl.RelationshipManyToOne {
		w.sb.WriteString(", WITHOUT_ARRAY_WRAPPER")
	}
	fmt.Fprintf(w.sb,
		")) AS %s",
		expansion.TraversalStep.Relationship.ExpansionColumnName,
	)
	return nil
}

func (w *queryWriter) writeJoins(joinCollection relationships.JoinCollection) {
	for _, join := range joinCollection.Joins() {
		relationship := join.Step.Relationship
		toResourceMetadata := relationship.To.GetMetadata()

		fmt.Fprintf(w.sb,
			"LEFT JOIN %s %s on %s.%s = %s.%s",
			toResourceMetadata.FullyQualifiedName,
			join.Alias,
			join.ParentAlias,
			relationship.FromColumn.Name,
			join.Alias,
			relationship.ToColumn.Name,
		)

		if join.SubJoins.JoinsLen() > 0 {
			w.sb.WriteRune(' ')
			w.writeJoins(join.SubJoins)
		}
		w.sb.WriteRune(' ')
	}
}

func (w *queryWriter) writeOrderByStatement() error {
	if w.queryDataStore.OrderByLen() == 0 {
		return qerr.InternalErr("expected an orderby operation with at least the primary column specified")
	}

	w.sb.WriteString("ORDER BY ")

	i := 0
	for r := range w.queryDataStore.OrderBy() {
		tableAlias, exists := w.aliasStore.GetJoinAlias(r.ResolvedColumn.ResolvedPath)
		if !exists {
			return qerr.InternalErr(
				"unable to access join alias for path %s",
				r.ResolvedColumn.ResolvedPath.Id,
			)
		}

		if i > 0 {
			w.sb.WriteString(", ")
		}

		fmt.Fprintf(
			w.sb,
			"%s.%s %s",
			tableAlias,
			r.ResolvedColumn.Metadata.Name,
			string(r.Direction),
		)
		i++
	}

	w.sb.WriteRune(' ')

	return nil
}

func (w *queryWriter) writeTopStatement() error {
	limit := w.queryDataStore.SystemLimit()
	if limit == 0 {
		return qerr.InternalErr("expected either a user or system defined limit operation")
	}

	fmt.Fprintf(w.sb, "TOP (%d)", limit)
	w.sb.WriteRune(' ')
	return nil
}

func (w *queryWriter) writeFilterStatement() error {
	filterExpression := w.queryDataStore.FilterExpression()

	rootResource := &aliasedResource{
		resource: w.queryDataStore.RootResourceMetadata(),
		alias:    w.aliasStore.GetRootAlias(),
	}

	parentQuery := w.queryDataStore.Parent()
	linkToParent := w.queryDataStore.LinkFromParent()

	if parentQuery != nil && linkToParent != nil {
		w.filterLocation.left = w.sb.Len()
		w.sb.WriteString("WHERE ")

		if err := w.writeFilterExpressionWithLink(
			parentQuery,
			linkToParent,
			rootResource,
			filterExpression,
		); err != nil {
			return err
		}
		w.filterLocation.right = w.sb.Len()
	} else {
		if filterExpression == nil {
			return nil
		}

		w.filterLocation.left = w.sb.Len()
		w.sb.WriteString("WHERE ")
		err := w.writeFilterExpression(rootResource, filterExpression)
		if err != nil {
			return err
		}
		w.filterLocation.right = w.sb.Len()
	}

	w.sb.WriteRune(' ')
	return nil
}

func (w *queryWriter) writeFilterExpressionWithLink(
	parentQuery qstore.QueryDataStore,
	linkToParent *mdl.TraversalStep,
	rootResource *aliasedResource,
	filterExpression mdl.FilterExpression,
) error {
	w.sb.WriteString("(")

	fmt.Fprintf(
		w.sb,
		"%s.%s = %s.%s",
		w.queryDataStore.Alias(),
		linkToParent.Relationship.ToColumn.Name,
		parentQuery.Alias(),
		linkToParent.Relationship.FromColumn.Name,
	)

	if filterExpression != nil {
		w.sb.WriteString(" and ")
		w.writeFilterExpression(rootResource, filterExpression)
	}

	w.sb.WriteString(") ")

	return nil
}

func (w *queryWriter) writeFilterExpression(
	rootResource *aliasedResource,
	expression mdl.FilterExpression,
) error {
	switch expression := expression.(type) {
	case *mdl.LogicalExpression:
		return w.writeLogicalExpression(rootResource, expression)
	case *mdl.ComparisonExpression:
		return w.writeComparisonExpression(expression, rootResource)
	case *mdl.CollectionExpression:
		return w.writeCollectionExpression(expression, rootResource)
	default:
		return qerr.InternalErr("unexpected filter expression received")
	}
}

func (w *queryWriter) writeLogicalExpression(
	rootResource *aliasedResource,
	expression *mdl.LogicalExpression,
) error {
	operator := ""
	switch expression.Operator {
	case mdl.LogicalOr:
		operator = "or"
	case mdl.LogicalAnd:
		operator = "and"
	default:
		return qerr.InternalErr("unexpected logical operator received '%v'", operator)
	}
	w.sb.WriteRune('(')
	if err := w.writeFilterExpression(rootResource, expression.Left); err != nil {
		return err
	}
	w.sb.WriteRune(' ')
	w.sb.WriteString(operator)
	w.sb.WriteRune(' ')
	if err := w.writeFilterExpression(rootResource, expression.Right); err != nil {
		return err
	}
	w.sb.WriteRune(')')

	return nil
}

func (w *queryWriter) writeComparisonExpression(
	ex *mdl.ComparisonExpression,
	rootResource *aliasedResource,
) error {
	if ex == nil {
		return qerr.InternalErr("comparison expression is nil")
	}

	writeExpression := func(endResource *aliasedResource) error {
		if endResource == nil {
			endResource = rootResource
		}
		return ex.Value.WriteFilterExpression(
			w,
			fmt.Sprintf("%s.%s", endResource.alias, ex.ResolvedColumn.Metadata.Name),
			ex.Operator,
		)
	}

	return w.writeExpressionWithPath(
		ex.ExistsNodes,
		writeExpression,
		false,
	)
}

func (w *queryWriter) writeCollectionExpression(
	ex *mdl.CollectionExpression,
	rootResource *aliasedResource,
) error {
	filterExpression := ex.FilterExpression
	doNegate := ex.Operator == mdl.CollectionAll
	if doNegate {
		negatedCondition, err := w.negate(ex.FilterExpression)
		if err != nil {
			return err
		}

		filterExpression = negatedCondition
	}

	writeExpression := func(endResource *aliasedResource) error {
		if endResource == nil {
			endResource = rootResource
		}
		return w.writeFilterExpression(
			endResource,
			filterExpression,
		)
	}

	return w.writeExpressionWithPath(
		ex.ExistsNodes,
		writeExpression,
		doNegate,
	)
}

func (w *queryWriter) writeExpressionWithPath(
	existsNodes []mdl.ExistsNode,
	writeExpression func(resource *aliasedResource) error,
	doNegate bool,
) error {
	if existsNodes == nil {
		return qerr.InternalErr("exists nodes should not be nil")
	}

	if len(existsNodes) == 0 {
		return writeExpression(nil)
	}

	currentNode := existsNodes[0]

	if doNegate {
		w.sb.WriteString("NOT ")
	}

	relationship := currentNode.Step.Relationship
	toResourceMetadata := relationship.To.GetMetadata()

	w.sb.WriteString("EXISTS (SELECT 1 FROM ")
	w.sb.WriteString(toResourceMetadata.FullyQualifiedName)
	w.sb.WriteRune(' ')
	w.sb.WriteString(currentNode.Alias)
	w.sb.WriteString(" WHERE ")
	w.sb.WriteString(currentNode.Alias)
	w.sb.WriteRune('.')
	w.sb.WriteString(relationship.ToColumn.Name)
	w.sb.WriteString(" = ")
	w.sb.WriteString(currentNode.ParentAlias)
	w.sb.WriteRune('.')
	w.sb.WriteString(relationship.FromColumn.Name)
	w.sb.WriteString(" AND ")

	// If last node write expression and return
	if len(existsNodes) == 1 {
		if err := writeExpression(&aliasedResource{
			alias:    currentNode.Alias,
			resource: relationship.To.GetMetadata(),
		}); err != nil {
			return err
		}
	} else {
		if err := w.writeExpressionWithPath(
			existsNodes[1:],
			writeExpression,
			doNegate,
		); err != nil {
			return err
		}
	}

	w.sb.WriteRune(')')
	return nil
}

func (w *queryWriter) negate(expression mdl.FilterExpression) (mdl.FilterExpression, error) {
	filterExpressionBuilder := w.queryDataStore.FilterExpressionBuilder()

	switch ex := expression.(type) {
	case *mdl.LogicalExpression:
		l, err := w.negate(ex.Left)
		if err != nil {
			return nil, err
		}

		r, err := w.negate(ex.Right)
		if err != nil {
			return nil, err
		}

		switch ex.Operator {
		case mdl.LogicalAnd:
			return filterExpressionBuilder.NewLogicalExpression(l, mdl.LogicalOr, r)
		case mdl.LogicalOr:
			return filterExpressionBuilder.NewLogicalExpression(l, mdl.LogicalAnd, r)
		default:
			return nil, fmt.Errorf("unexpected logical operator received '%v'", ex.Operator)
		}

	case *mdl.ComparisonExpression:
		negatedOperator, err := negateComparisonOperator(ex.Operator)
		if err != nil {
			return nil, err
		}

		return ex.WithValue(negatedOperator, ex.Value), nil

	case *mdl.CollectionExpression:
		operator := ex.Operator
		switch ex.Operator {
		case mdl.CollectionAny:
			operator = mdl.CollectionAll
		case mdl.CollectionAll:
			operator = mdl.CollectionAny
		default:
			return nil, fmt.Errorf("unexpected collection operator received '%v'", ex.Operator)
		}

		expression, err := w.negate(ex.FilterExpression)
		if err != nil {
			return nil, err
		}

		return ex.WithValues(operator, expression), nil
	default:
		return nil, fmt.Errorf("unknown filter expression received")
	}
}

func negateComparisonOperator(operator mdl.ComparisonOperator) (mdl.ComparisonOperator, error) {
	switch operator {
	case mdl.ComparisonEq:
		return mdl.ComparisonNe, nil

	case mdl.ComparisonNe:
		return mdl.ComparisonEq, nil

	case mdl.ComparisonStartsWith:
		return mdl.ComparisonNotStartsWith, nil

	case mdl.ComparisonEndsWith:
		return mdl.ComparisonNotEndsWith, nil

	case mdl.ComparisonContains:
		return mdl.ComparisonNotContains, nil

	case mdl.ComparisonIn:
		return mdl.ComparisonNotIn, nil

	case mdl.ComparisonGt:
		return mdl.ComparisonLe, nil

	case mdl.ComparisonGe:
		return mdl.ComparisonLt, nil

	case mdl.ComparisonLt:
		return mdl.ComparisonGe, nil

	case mdl.ComparisonLe:
		return mdl.ComparisonGt, nil

	default:
		return operator, fmt.Errorf("no negation defined for comparison operatior %v", operator)
	}
}

func (w *queryWriter) WriteFilterExpressionNull(
	fieldName string,
	op mdl.ComparisonOperator,
) error {
	switch op {
	case mdl.ComparisonEq:
		w.Write("%s IS NULL", fieldName)
	case mdl.ComparisonNe:
		w.Write("%s IS NOT NULL", fieldName)
	default:
		return fmt.Errorf("unsupported operation: %s", string(op))
	}
	return nil
}

func (w *queryWriter) WriteFilterExpressionString(
	fieldName string,
	op mdl.ComparisonOperator,
	value string,
) error {
	switch op {
	case mdl.ComparisonEq:
		w.Write("%s = %s", fieldName, w.placeholder(value))
	case mdl.ComparisonNe:
		w.Write("%s != %s", fieldName, w.placeholder(value))
	case mdl.ComparisonGt:
		w.Write("%s > %s", fieldName, w.placeholder(value))
	case mdl.ComparisonGe:
		w.Write("%s >= %s", fieldName, w.placeholder(value))
	case mdl.ComparisonLt:
		w.Write("%s < %s", fieldName, w.placeholder(value))
	case mdl.ComparisonLe:
		w.Write("%s <= %s", fieldName, w.placeholder(value))
	case mdl.ComparisonContains, mdl.ComparisonNotContains:
		modifier := ""
		if op == mdl.ComparisonNotContains {
			modifier = "NOT "
		}

		pattern := fmt.Sprintf("%%%s%%", value)
		w.Write("%s %sLIKE %s", fieldName, modifier, w.placeholder(pattern))

	case mdl.ComparisonStartsWith, mdl.ComparisonNotStartsWith:
		modifier := ""
		if op == mdl.ComparisonNotStartsWith {
			modifier = "NOT "
		}

		pattern := fmt.Sprintf("%s%%", value)
		w.Write("%s %sLIKE %s", fieldName, modifier, w.placeholder(pattern))
	case mdl.ComparisonEndsWith, mdl.ComparisonNotEndsWith:
		modifier := ""
		if op == mdl.ComparisonNotEndsWith {
			modifier = "NOT "
		}

		pattern := fmt.Sprintf("%%%s", value)
		w.Write("%s %sLIKE %s", fieldName, modifier, w.placeholder(pattern))
	default:
		return fmt.Errorf("unsupported string operation: %s", string(op))
	}
	return nil
}

func (w *queryWriter) WriteFilterExpressionInt(
	fieldName string,
	op mdl.ComparisonOperator,
	value int64,
) error {
	switch op {
	case mdl.ComparisonEq:
		w.Write("%s = %s", fieldName, w.placeholder(value))
	case mdl.ComparisonNe:
		w.Write("%s != %s", fieldName, w.placeholder(value))
	case mdl.ComparisonGt:
		w.Write("%s > %s", fieldName, w.placeholder(value))
	case mdl.ComparisonGe:
		w.Write("%s >= %s", fieldName, w.placeholder(value))
	case mdl.ComparisonLt:
		w.Write("%s < %s", fieldName, w.placeholder(value))
	case mdl.ComparisonLe:
		w.Write("%s <= %s", fieldName, w.placeholder(value))
	default:
		return fmt.Errorf("unsupported int operation: %s", string(op))
	}
	return nil
}

func (w *queryWriter) WriteFilterExpressionFloat(
	fieldName string,
	op mdl.ComparisonOperator,
	value float64,
) error {
	switch op {
	case mdl.ComparisonEq:
		w.Write("%s = %s", fieldName, w.placeholder(value))
	case mdl.ComparisonNe:
		w.Write("%s != %s", fieldName, w.placeholder(value))
	case mdl.ComparisonGt:
		w.Write("%s > %s", fieldName, w.placeholder(value))
	case mdl.ComparisonGe:
		w.Write("%s >= %s", fieldName, w.placeholder(value))
	case mdl.ComparisonLt:
		w.Write("%s < %s", fieldName, w.placeholder(value))
	case mdl.ComparisonLe:
		w.Write("%s <= %s", fieldName, w.placeholder(value))
	default:
		return fmt.Errorf("unsupported float operation: %s", string(op))
	}
	return nil
}

func (w *queryWriter) WriteFilterExpressionPoint(
	fieldName string,
	op mdl.ComparisonOperator,
	value mdl.Point,
) error {
	return fmt.Errorf("no operators are currently supported for point")
}

func (w *queryWriter) WriteFilterExpressionDateTime(
	fieldName string,
	op mdl.ComparisonOperator,
	value time.Time,
) error {
	switch op {
	case mdl.ComparisonEq:
		w.Write("%s = %s", fieldName, w.placeholder(value))
	case mdl.ComparisonNe:
		w.Write("%s != %s", fieldName, w.placeholder(value))
	case mdl.ComparisonGt:
		w.Write("%s > %s", fieldName, w.placeholder(value))
	case mdl.ComparisonGe:
		w.Write("%s >= %s", fieldName, w.placeholder(value))
	case mdl.ComparisonLt:
		w.Write("%s < %s", fieldName, w.placeholder(value))
	case mdl.ComparisonLe:
		w.Write("%s <= %s", fieldName, w.placeholder(value))
	default:
		return fmt.Errorf("unsupported date/time operation: %s", string(op))
	}
	return nil
}

func (w *queryWriter) WriteFilterExpressionStringList(
	fieldName string,
	op mdl.ComparisonOperator,
	value []string,
) error {
	listLen := len(value)

	return w.writeList(
		fieldName,
		listLen,
		op,
		func(i int) error {
			if i < 0 || i > listLen-1 {
				return qerr.InternalErr("unable to write list value: index out of range")
			}
			w.Write("%s", value[i])
			return nil
		},
	)
}

func (w *queryWriter) WriteFilterExpressionIntList(
	fieldName string,
	op mdl.ComparisonOperator,
	value []int64,
) error {
	listLen := len(value)

	return w.writeList(
		fieldName,
		listLen,
		op,
		func(i int) error {
			if i < 0 || i > listLen-1 {
				return qerr.InternalErr("unable to write list value: index out of range")
			}
			w.Write("%d", value[i])
			return nil
		},
	)
}

func (w *queryWriter) WriteFilterExpressionFloatList(
	fieldName string,
	op mdl.ComparisonOperator,
	value []float64,
) error {
	listLen := len(value)

	return w.writeList(
		fieldName,
		listLen,
		op,
		func(i int) error {
			if i < 0 || i > listLen-1 {
				return qerr.InternalErr("unable to write list value: index out of range")
			}
			w.Write("%f", value[i])
			return nil
		},
	)
}

func (w *queryWriter) writeList(
	fieldName string,
	listLength int,
	op mdl.ComparisonOperator,
	writeValue func(i int) error,
) error {
	if op != mdl.ComparisonIn && op != mdl.ComparisonNotIn {
		return fmt.Errorf("unsupported list operation: %s", string(op))
	}

	negationModifier := ""
	if op == mdl.ComparisonNotIn {
		negationModifier = "NOT"
	}

	w.Write("%s %sIN (", fieldName, negationModifier)
	for i := range listLength {
		if i != 0 {
			w.Write(",")
		}
		writeValue(i)
	}
	w.Write(")")
	return nil

}
