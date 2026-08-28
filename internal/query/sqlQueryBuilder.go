package query

import (
	"fmt"
	"strings"

	qstore "github.com/turnerbenjamin/heterogen_portal/internal/query/queryDataStore"
	qerr "github.com/turnerbenjamin/heterogen_portal/internal/query/queryError"
	mdl "github.com/turnerbenjamin/heterogen_portal/internal/query/queryModel"
	"github.com/turnerbenjamin/heterogen_portal/internal/query/relationships"
)

// sqlQuery is used to build sql query data, it includes a strings builder for
// constructing a statement and an args slice containing placeholder values.
// statement is a materialisation of the strings builder post build
type sqlQuery struct {
	args      []any
	statement string
}

type stringCoords struct {
	left  int
	right int
}

type queryWriter struct {
	sb              *strings.Builder
	args            []any
	statement       string
	selectLocation  stringCoords
	fromLocation    stringCoords
	filterLocation  stringCoords
	orderByLocation stringCoords
}

// arg adds a new argument to the args list and returns a unique placeholder for
// use in the sql statement
func (q *queryWriter) Placeholder(v any) string {
	q.args = append(q.args, v)
	return fmt.Sprintf("@p%d", len(q.args))
}

func (q *queryWriter) Write(statement string, args ...any) {
	fmt.Fprintf(q.sb, statement, args...)
}

// aliasedResource binds a resourse to a table alias in an sql query
type aliasedResource struct {
	resource mdl.TableMetadata
	alias    string
}

// nestedQuery represents an expansion query. It binds a queryBuilder, for the
// nested query, to a traversal step from the parent table. This enables the
// query response to be stiched to the parent response server side
// type nestedQuery struct {
// 	queryBuilder *sqlQueryBuilder
// 	link         *TraversalStep
// }

// sqlQueryBuilder is used to build and execute sql queries
type sqlQueryBuilder struct {
	queryDataStore qstore.QueryDataStore
	aliasStore     relationships.AliasStore
}

// newSqlQueryBuilder constructs a new sqlQueryBuilderInstance from
// QueryOperations. It will return an error if the operations cannot be bound to
// the schema metadata
func newSqlQueryBuilder(queryDataStore qstore.QueryDataStore) *sqlQueryBuilder {
	return &sqlQueryBuilder{
		queryDataStore: queryDataStore,
		aliasStore:     queryDataStore.GetAliasStore(),
	}
}

// func (b *sqlQueryBuilder) newNestedSqlQueryBuilder(
// 	rootResource TableMetadata,
// 	accessPolicy AccessPolicy,
// 	rawOperations *Operations,
// ) (*sqlQueryBuilder, error) {
// 	nestedBuilder, err := newSqlQueryBuilder(
// 		rootResource,
// 		accessPolicy,
// 		rawOperations,
// 		b.pagingTokenParser,
// 	)
// 	if err != nil {
// 		return nil, err
// 	}
// 	nestedBuilder.depth = b.depth + 1
// 	return nestedBuilder, nil
// }

// func (b *sqlQueryBuilder) createBuildersforNestedOperations() error {
// 	op := b.operations.ExpandOperation
// 	if op == nil {
// 		return nil
// 	}

// 	for _, expand := range op.Expands {
// 		nestedQueryBuilder, err := b.newNestedSqlQueryBuilder(
// 			expand.Link.To,
// 			b.accessPolicy,
// 			expand.Operations,
// 		)
// 		if err != nil {
// 			return err
// 		}

// 		b.nestedQueries[string(expand.Link.Relationship.Id())] = &nestedQuery{
// 			queryBuilder: nestedQueryBuilder,
// 			link:         expand.Link,
// 		}
// 	}
// 	return nil
// }

// func (b *sqlQueryBuilder) addAssociatedWithParentFilter(
// 	linkFromParent *TraversalStep,
// 	joinParentOnValues []string,
// ) error {

// 	associationFilter, err := b.operations.addAssociatedWithParentFilter(
// 		linkFromParent,
// 		joinParentOnValues,
// 	)
// 	if err != nil {
// 		return err
// 	}

// 	return b.relationshipPlan.processFilterExpression(
// 		b.relationshipPlan.rootAlias,
// 		associationFilter,
// 	)
// }

func (b *sqlQueryBuilder) build() (*sqlQuery, error) {
	w := &queryWriter{
		sb:   &strings.Builder{},
		args: []any{},
	}

	err := b.writeQuery(w)
	if err != nil {
		return nil, err
	}
	w.statement = w.sb.String()

	return &sqlQuery{
		statement: w.sb.String(),
		args:      w.args,
	}, nil
}

func (b *sqlQueryBuilder) buildTopLevelQuery() (*sqlQuery, string, error) {
	// Initialise a streaming query writer
	w := &queryWriter{
		sb:   &strings.Builder{},
		args: []any{},
	}

	// Write the main query
	err := b.writeQuery(w)
	if err != nil {
		return nil, "", err
	}
	w.statement = w.sb.String()

	// Construct the count statement
	countStatement := b.buildCountStatement(w)

	return &sqlQuery{
		statement: w.sb.String(),
		args:      w.args,
	}, countStatement, nil

}

func (b *sqlQueryBuilder) writeQuery(w *queryWriter) error {
	err := b.writeSelectStatement(w)
	if err != nil {
		return err
	}

	b.writeFromStatement(w)

	b.writeJoins(w, b.queryDataStore.JoinCollection())

	err = b.writeFilterStatement(w)
	if err != nil {
		return err
	}

	err = b.writeOrderByStatement(w)
	if err != nil {
		return err
	}

	err = b.writeLimitStatement(w)
	if err != nil {
		return err
	}

	// request json response
	w.sb.WriteString("FOR JSON PATH;")

	return nil
}

func (b *sqlQueryBuilder) buildCountStatement(w *queryWriter) string {
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

	return countSb.String()
}

func (b *sqlQueryBuilder) writeFromStatement(w *queryWriter) {
	w.fromLocation.left = w.sb.Len()

	rootResource := b.queryDataStore.RootResource()
	rootAlias := b.aliasStore.GetRootAlias()

	w.sb.WriteString("FROM ")
	w.sb.WriteString(rootResource.FullyQualifiedName())
	w.sb.WriteRune(' ')
	w.sb.WriteString(rootAlias)

	w.fromLocation.right = w.sb.Len()

	w.sb.WriteRune(' ')
}

func (b *sqlQueryBuilder) writeSelectStatement(w *queryWriter) error {

	if b.queryDataStore.SelectsLen() == 0 {
		return qerr.InternalErr("expected a select operation with at least one column specified")
	}

	w.selectLocation.left = w.sb.Len()
	w.sb.WriteString("SELECT ")

	i := 0
	for _, columnValue := range b.queryDataStore.Selects() {
		if i > 0 {
			w.sb.WriteString(",")
		}

		fmt.Fprintf(
			w.sb,
			"%s.%s",
			b.aliasStore.GetRootAlias(),
			b.formatSelectValue(columnValue.Metadata),
		)
		i++
	}

	w.selectLocation.right = w.sb.Len()

	w.sb.WriteRune(' ')
	return nil
}

func (b *sqlQueryBuilder) formatSelectValue(columnData mdl.ColumnMetadata) string {
	switch columnData.Type() {
	case mdl.DbTypeGeography:
		return fmt.Sprintf("%s.STAsText() AS %s", columnData.Name(), columnData.Name())
	default:
		return columnData.Name()
	}
}

func (b *sqlQueryBuilder) writeJoins(w *queryWriter, joinCollection relationships.JoinCollection) {
	for _, join := range joinCollection.Joins() {
		relationship := join.Step.Relationship

		fmt.Fprintf(w.sb,
			"LEFT JOIN %s %s on %s.%s = %s.%s",
			relationship.To().FullyQualifiedName(),
			join.Alias,
			join.ParentAlias,
			relationship.FromColumn().Name(),
			join.Alias,
			relationship.ToColumn().Name(),
		)

		if join.SubJoins.JoinsLen() > 0 {
			w.sb.WriteRune(' ')
			b.writeJoins(w, join.SubJoins)
		}
		w.sb.WriteRune(' ')
	}
}

func (b *sqlQueryBuilder) writeOrderByStatement(w *queryWriter) error {
	if b.queryDataStore.OrderByLen() == 0 {
		return qerr.InternalErr("expected an orderby operation with at least the primary column specified")
	}

	w.orderByLocation.left = w.sb.Len()
	w.sb.WriteString("ORDER BY ")

	i := 0
	for r := range b.queryDataStore.OrderBy() {
		tableAlias, exists := b.aliasStore.GetJoinAlias(r.ResolvedColumn.ResolvedPath)
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
			r.ResolvedColumn.Metadata.Name(),
			string(r.Direction),
		)
		i++
	}

	w.orderByLocation.right = w.sb.Len()
	w.sb.WriteRune(' ')

	return nil
}

func (b *sqlQueryBuilder) writeLimitStatement(w *queryWriter) error {
	limit := b.queryDataStore.Limit()
	if limit == 0 {
		return qerr.InternalErr("expected either a user or system defined limit operation")
	}

	w.sb.WriteString(fmt.Sprintf("OFFSET 0 ROWS FETCH NEXT %d ROWS ONLY", limit))
	w.sb.WriteRune(' ')
	return nil
}

func (b *sqlQueryBuilder) writeFilterStatement(w *queryWriter) error {
	filterExpression := b.queryDataStore.FilterExpression()
	if filterExpression == nil {
		return nil
	}

	w.filterLocation.left = w.sb.Len()

	w.sb.WriteString("WHERE ")

	rootResource := &aliasedResource{
		resource: b.queryDataStore.RootResource(),
		alias:    b.aliasStore.GetRootAlias(),
	}

	err := b.buildFilterExpression(rootResource, filterExpression, w)
	if err != nil {
		return err
	}

	w.filterLocation.right = w.sb.Len()

	w.sb.WriteRune(' ')
	return nil
}

func (b *sqlQueryBuilder) buildFilterExpression(
	rootResource *aliasedResource,
	expression mdl.FilterExpression,
	o *queryWriter,
) error {
	switch expression := expression.(type) {
	case *mdl.LogicalExpression:
		return b.buildLogicalExpression(rootResource, expression, o)
	case *mdl.ComparisonExpression:
		return b.buildComparisonExpression(expression, rootResource, o)
	case *mdl.CollectionExpression:
		return b.buildCollectionExpression(expression, rootResource, o)
	default:
		return qerr.InternalErr("unexpected filter expression received")
	}
}

func (b *sqlQueryBuilder) buildLogicalExpression(
	rootResource *aliasedResource,
	expression *mdl.LogicalExpression,
	w *queryWriter,
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
	b.buildFilterExpression(rootResource, expression.Left, w)
	w.sb.WriteRune(' ')
	w.sb.WriteString(operator)
	w.sb.WriteRune(' ')
	b.buildFilterExpression(rootResource, expression.Right, w)
	w.sb.WriteRune(')')

	return nil
}

func (b *sqlQueryBuilder) buildComparisonExpression(
	ex *mdl.ComparisonExpression,
	rootResource *aliasedResource,
	w *queryWriter,
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
			fmt.Sprintf("%s.%s", endResource.alias, ex.ResolvedColumn.Metadata.Name()),
			ex.Operator,
		)
	}

	return b.writeExpressionWithPath(
		ex.ExistsNodes,
		writeExpression,
		false,
		w,
	)
}

func (b *sqlQueryBuilder) buildCollectionExpression(
	ex *mdl.CollectionExpression,
	rootResource *aliasedResource,
	w *queryWriter,
) error {
	writeExpression := func(endResource *aliasedResource) error {
		if endResource == nil {
			endResource = rootResource
		}
		return b.buildFilterExpression(
			endResource,
			ex.FilterExpression,
			w,
		)
	}

	doNegate := ex.Operator == mdl.CollectionAll
	if doNegate {
		negatedCondition, err := b.negate(ex.FilterExpression)
		if err != nil {
			return err
		}

		ex.FilterExpression = negatedCondition
	}

	return b.writeExpressionWithPath(
		ex.ExistsNodes,
		writeExpression,
		doNegate,
		w,
	)
}

func (b *sqlQueryBuilder) writeExpressionWithPath(
	existsNodes []mdl.ExistsNodeNew,
	writeExpression func(resource *aliasedResource) error,
	doNegate bool,
	w *queryWriter,
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

	w.sb.WriteString("EXISTS (SELECT 1 FROM ")
	w.sb.WriteString(relationship.To().FullyQualifiedName())
	w.sb.WriteRune(' ')
	w.sb.WriteString(currentNode.Alias)
	w.sb.WriteString(" WHERE ")
	w.sb.WriteString(currentNode.Alias)
	w.sb.WriteRune('.')
	w.sb.WriteString(relationship.ToColumn().Name())
	w.sb.WriteString(" = ")
	w.sb.WriteString(currentNode.ParentAlias)
	w.sb.WriteRune('.')
	w.sb.WriteString(relationship.FromColumn().Name())
	w.sb.WriteString(" AND ")

	// If last node write expression and return
	if len(existsNodes) == 1 {
		if err := writeExpression(&aliasedResource{
			alias:    currentNode.Alias,
			resource: relationship.To(),
		}); err != nil {
			return err
		}
	} else {
		if err := b.writeExpressionWithPath(
			existsNodes[1:],
			writeExpression,
			doNegate,
			w,
		); err != nil {
			return err
		}
	}

	w.sb.WriteRune(')')
	return nil
}

func (b *sqlQueryBuilder) negate(expression mdl.FilterExpression) (mdl.FilterExpression, error) {
	filterExpressionBuilder := b.queryDataStore.FilterExpressionBuilder()

	switch ex := expression.(type) {
	case *mdl.LogicalExpression:
		l, err := b.negate(ex.Left)
		if err != nil {
			return nil, err
		}

		r, err := b.negate(ex.Right)
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

		return ex.WithValues(negatedOperator, ex.Value), nil

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

		expression, err := b.negate(ex.FilterExpression)
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
