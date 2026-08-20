package query

import (
	"fmt"
	"strings"
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
func (q *queryWriter) arg(v any) string {
	q.args = append(q.args, v)
	return fmt.Sprintf("@p%d", len(q.args))
}

// aliasedResource binds a resourse to a table alias in an sql query
type aliasedResource struct {
	resource TableMetadata
	alias    string
}

// nestedQuery represents an expansion query. It binds a queryBuilder, for the
// nested query, to a traversal step from the parent table. This enables the
// query response to be stiched to the parent response server side
type nestedQuery struct {
	queryBuilder *sqlQueryBuilder
	link         *TraversalStep
}

// sqlQueryBuilder is used to build and execute sql queries
type sqlQueryBuilder struct {
	depth                int
	pagingTokenParser    PagingTokenBuilder
	metadataBinder       *metadataBinder
	relationshipPlan     *RelationshipPlanner
	rootResource         TableMetadata
	accessPolicy         AccessPolicy
	resourceAccessPolicy TableAccessPolicy
	operations           *Operations
	doIncludeCount       bool
	nestedQueries        map[string]*nestedQuery
}

// newSqlQueryBuilder constructs a new sqlQueryBuilderInstance from
// QueryOperations. It will return an error if the operations cannot be bound to
// the schema metadata
func newSqlQueryBuilder(
	rootResource TableMetadata,
	accessPolicy AccessPolicy,
	queryOperations *Operations,
	pagingTokenParser PagingTokenBuilder,
) (*sqlQueryBuilder, error) {

	if queryOperations.state != operationsStateReadyForBuild {
		return nil, internalErr("query operations must be ready for build")
	}

	b := &sqlQueryBuilder{
		operations:        queryOperations,
		pagingTokenParser: pagingTokenParser,
		metadataBinder:    &metadataBinder{maximumDepth: 5, accessPolicy: accessPolicy},
		rootResource:      rootResource,
		accessPolicy:      accessPolicy,
		nestedQueries:     map[string]*nestedQuery{},
		depth:             0,
	}

	// Build relationship plan for exists and joins
	relationshipPlan, err := NewRelationshipPlan(queryOperations)
	if err != nil {
		return nil, err
	}
	b.relationshipPlan = relationshipPlan

	if err := b.createBuildersforNestedOperations(); err != nil {
		return nil, err
	}

	return b, nil
}

func (b *sqlQueryBuilder) newNestedSqlQueryBuilder(
	rootResource TableMetadata,
	accessPolicy AccessPolicy,
	rawOperations *Operations,
) (*sqlQueryBuilder, error) {
	nestedBuilder, err := newSqlQueryBuilder(
		rootResource,
		accessPolicy,
		rawOperations,
		b.pagingTokenParser,
	)
	if err != nil {
		return nil, err
	}
	nestedBuilder.depth = b.depth + 1
	return nestedBuilder, nil
}

func (b *sqlQueryBuilder) createBuildersforNestedOperations() error {
	op := b.operations.ExpandOperation
	if op == nil {
		return nil
	}

	for _, expand := range op.Expands {
		nestedQueryBuilder, err := b.newNestedSqlQueryBuilder(
			expand.Link.To,
			b.accessPolicy,
			expand.Operations,
		)
		if err != nil {
			return err
		}

		b.nestedQueries[string(expand.Link.Relationship.Id())] = &nestedQuery{
			queryBuilder: nestedQueryBuilder,
			link:         expand.Link,
		}
	}
	return nil
}

func (b *sqlQueryBuilder) addAssociatedWithParentFilter(
	linkFromParent *TraversalStep,
	joinParentOnValues []string,
) error {

	associationFilter, err := b.operations.addAssociatedWithParentFilter(
		linkFromParent,
		joinParentOnValues,
	)
	if err != nil {
		return err
	}

	return b.relationshipPlan.processFilterExpression(
		b.relationshipPlan.rootAlias,
		associationFilter,
	)
}

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

	// Increment limit by 1 so that we can identify if a next page of records
	// exists
	b.operations.LimitOperation.Limit = b.operations.LimitOperation.Limit + 1

	// Write the main query
	err := b.writeQuery(w)
	if err != nil {
		return nil, "", err
	}
	w.statement = w.sb.String()

	// Restore the original limit value
	b.operations.LimitOperation.Limit = b.operations.LimitOperation.Limit - 1

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

	b.writeJoins(w, b.relationshipPlan.Joins)

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

	w.sb.WriteString("FROM ")
	w.sb.WriteString(b.rootResource.FullyQualifiedName())
	w.sb.WriteRune(' ')
	w.sb.WriteString(b.relationshipPlan.rootAlias)

	w.fromLocation.right = w.sb.Len()

	w.sb.WriteRune(' ')
}

func (b *sqlQueryBuilder) writeSelectStatement(w *queryWriter) error {
	// Expect either a user or system defined select statemet
	selectOp := b.operations.SelectOperation
	systemSelectOp := b.operations.SystemSelectOperation

	if selectOp == nil || len(selectOp.Columns) == 0 {
		return internalErr("expected a select operation with at least one column specified")
	}

	columns := make([]*ColumnValue, 0, len(selectOp.Columns)+len(systemSelectOp.Columns))

	// Add all columns to select
	for _, v := range selectOp.Columns {
		columns = append(columns, v)
	}

	if systemSelectOp != nil && systemSelectOp.Columns != nil {
		for _, v := range systemSelectOp.Columns {
			columns = append(columns, v)
		}
	}

	w.selectLocation.left = w.sb.Len()
	w.sb.WriteString("SELECT ")

	i := 0
	for _, columnValue := range columns {
		if i > 0 {
			w.sb.WriteString(",")
		}

		fmt.Fprintf(
			w.sb,
			"%s.%s",
			b.relationshipPlan.rootAlias,
			b.formatSelectValue(columnValue.ColumnData),
		)
		i++
	}

	w.selectLocation.right = w.sb.Len()

	w.sb.WriteRune(' ')
	return nil
}

func (b *sqlQueryBuilder) formatSelectValue(columnData ColumnMetadata) string {
	switch columnData.Type() {
	case DbTypeGeography:
		return fmt.Sprintf("%s.STAsText() AS %s", columnData.Name(), columnData.Name())
	default:
		return columnData.Name()
	}
}

func (b *sqlQueryBuilder) writeJoins(w *queryWriter, joins map[string]*Join) {
	for _, join := range joins {
		fmt.Fprintf(w.sb,
			"LEFT JOIN %s %s on %s.%s = %s.%s",
			join.step.To.FullyQualifiedName(),
			join.alias,
			join.ParentAlias,
			join.step.Relationship.FromColumn().Name(),
			join.alias,
			join.step.Relationship.ToColumn().Name(),
		)

		if len(join.Joins) > 0 {
			w.sb.WriteRune(' ')
			b.writeJoins(w, join.Joins)
		}
		w.sb.WriteRune(' ')
	}
}

func (b *sqlQueryBuilder) writeOrderByStatement(w *queryWriter) error {
	if b.operations.OrderByOperation == nil || len(b.operations.OrderByOperation.Rules) == 0 {
		return internalErr("expected an orderby operation with at least the primary column specified")
	}

	w.orderByLocation.left = w.sb.Len()
	w.sb.WriteString("ORDER BY ")

	orderByRules := b.operations.OrderByOperation.Rules
	for i, r := range orderByRules {
		if r.ResolvedPath.Id == "" {
			return internalErr("resolved path id has not been populated")
		}

		tableAlias, exists := b.relationshipPlan.TableAliases[r.ResolvedPath.Id]
		if !exists {
			return internalErr("unable to access join alias for path %s", r.ResolvedPath.Id)
		}

		if i > 0 {
			w.sb.WriteString(", ")
		}

		fmt.Fprintf(
			w.sb,
			"%s.%s %s",
			tableAlias,
			r.ResolvedColumn.ColumnName,
			string(r.Direction),
		)
	}

	w.orderByLocation.right = w.sb.Len()
	w.sb.WriteRune(' ')

	return nil
}

func (b *sqlQueryBuilder) writeLimitStatement(w *queryWriter) error {
	limitOperation := b.operations.LimitOperation
	if limitOperation == nil {
		return internalErr("expected either a user or system defined limit operation")
	}

	w.sb.WriteString(fmt.Sprintf("OFFSET 0 ROWS FETCH NEXT %d ROWS ONLY", limitOperation.Limit))
	w.sb.WriteRune(' ')
	return nil
}

func (b *sqlQueryBuilder) writeFilterStatement(w *queryWriter) error {
	if b.operations.FilterOperation == nil {
		return nil
	}

	filterExpression := b.operations.FilterOperation.FilterExpression
	if filterExpression == nil {
		return nil
	}

	w.filterLocation.left = w.sb.Len()

	w.sb.WriteString("WHERE ")

	rootResource := &aliasedResource{resource: b.rootResource, alias: b.relationshipPlan.rootAlias}
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
	expression FilterExpression,
	o *queryWriter,
) error {
	switch expression := expression.(type) {
	case *LogicalExpression:
		return b.buildLogicalExpression(rootResource, expression, o)
	case *ComparisonExpression:
		return b.buildComparisonExpression(expression, rootResource, o)
	case *CollectionExpression:
		return b.buildCollectionExpression(expression, rootResource, o)
	default:
		return internalErr("unexpected filter expression received")
	}
}

func (b *sqlQueryBuilder) buildLogicalExpression(
	rootResource *aliasedResource,
	expression *LogicalExpression,
	w *queryWriter,
) error {
	operator := ""
	switch expression.Operator {
	case LogicalOr:
		operator = "or"
	case LogicalAnd:
		operator = "and"
	default:
		return internalErr("unexpected logical operator received '%v'", operator)
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
	ex *ComparisonExpression,
	rootResource *aliasedResource,
	w *queryWriter,
) error {
	if ex == nil {
		return internalErr("comparison expression is nil")
	}

	writeExpression := func(endResource *aliasedResource) error {
		if endResource == nil {
			endResource = rootResource
		}
		return ex.Value.WriteFilterExpression(
			w,
			fmt.Sprintf("%s.%s", endResource.alias, ex.ResolvedColumn.ColumnName),
			ex.Operator,
		)
	}

	if ex.ExistsPlan == nil {
		return internalErr("comparison exists plan has not been populated")
	}

	return b.writeExpressionWithPath(
		ex.ExistsPlan.FirstNode,
		writeExpression,
		false,
		w,
	)
}

func (b *sqlQueryBuilder) buildCollectionExpression(
	ex *CollectionExpression,
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

	doNegate := ex.Operator == CollectionAll
	if doNegate {
		negatedCondition, err := negate(ex.FilterExpression)
		if err != nil {
			return err
		}

		ex.FilterExpression = negatedCondition
	}

	return b.writeExpressionWithPath(
		ex.ExistsPlan.FirstNode,
		writeExpression,
		doNegate,
		w,
	)
}

func (b *sqlQueryBuilder) writeExpressionWithPath(
	existsNode *ExistsNode,
	writeExpression func(resource *aliasedResource) error,
	doNegate bool,
	w *queryWriter,
) error {
	if existsNode == nil {
		return writeExpression(nil)
	}

	if doNegate {
		w.sb.WriteString("NOT ")
	}

	relationship := existsNode.step.Relationship

	w.sb.WriteString("EXISTS (SELECT 1 FROM ")
	w.sb.WriteString(relationship.To().FullyQualifiedName())
	w.sb.WriteRune(' ')
	w.sb.WriteString(existsNode.alias)
	w.sb.WriteString(" WHERE ")
	w.sb.WriteString(existsNode.alias)
	w.sb.WriteRune('.')
	w.sb.WriteString(relationship.ToColumn().Name())
	w.sb.WriteString(" = ")
	w.sb.WriteString(existsNode.ParentAlias)
	w.sb.WriteRune('.')
	w.sb.WriteString(relationship.FromColumn().Name())
	w.sb.WriteString(" AND ")

	if existsNode.Next == nil {
		if err := writeExpression(&aliasedResource{
			alias:    existsNode.alias,
			resource: relationship.To(),
		}); err != nil {
			return err
		}
	} else {
		if err := b.writeExpressionWithPath(
			existsNode.Next,
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

func negate(expression FilterExpression) (FilterExpression, error) {
	switch ex := expression.(type) {
	case *LogicalExpression:
		l, err := negate(ex.Left)
		if err != nil {
			return nil, err
		}

		r, err := negate(ex.Right)
		if err != nil {
			return nil, err
		}

		switch ex.Operator {
		case LogicalAnd:
			return &LogicalExpression{
				Left:     l,
				Operator: LogicalOr,
				Right:    r,
			}, nil

		case LogicalOr:
			return &LogicalExpression{
				Left:     l,
				Operator: LogicalAnd,
				Right:    r,
			}, nil
		default:
			return nil, fmt.Errorf("unexpected logical operator received '%v'", ex.Operator)
		}

	case *ComparisonExpression:
		negatedOperator, err := negateComparisonOperator(ex.Operator)
		if err != nil {
			return nil, err
		}

		return &ComparisonExpression{
			Path:           ex.Path,
			ResolvedPath:   ex.ResolvedPath,
			ExistsPlan:     ex.ExistsPlan,
			ResolvedColumn: ex.ResolvedColumn,
			Value:          ex.Value,
			Operator:       negatedOperator,
		}, nil

	case *CollectionExpression:
		operator := ex.Operator
		switch ex.Operator {
		case CollectionAny:
			operator = CollectionAll
		case CollectionAll:
			operator = CollectionAny
		default:
			return nil, fmt.Errorf("unexpected collection operator received '%v'", ex.Operator)
		}

		expression, err := negate(ex.FilterExpression)
		if err != nil {
			return nil, err
		}

		return &CollectionExpression{
			Path:             ex.Path,
			ResolvedPath:     ex.ResolvedPath,
			ExistsPlan:       ex.ExistsPlan,
			Operator:         operator,
			FilterExpression: expression,
		}, nil
	default:
		return nil, fmt.Errorf("unknown filter expression received")
	}
}

func negateComparisonOperator(operator ComparisonOperator) (ComparisonOperator, error) {
	switch operator {
	case ComparisonEq:
		return ComparisonNe, nil

	case ComparisonNe:
		return ComparisonEq, nil

	case ComparisonStartsWith:
		return comparisonNotStartsWith, nil

	case ComparisonEndsWith:
		return comparisonNotEndsWith, nil

	case ComparisonContains:
		return comparisonNotContains, nil

	case ComparisonIn:
		return comparisonNotIn, nil

	case ComparisonGt:
		return ComparisonLe, nil

	case ComparisonGe:
		return ComparisonLt, nil

	case ComparisonLt:
		return ComparisonGe, nil

	case ComparisonLe:
		return ComparisonGt, nil

	default:
		return operator, fmt.Errorf("no negation defined for comparison operatior %v", operator)
	}
}
