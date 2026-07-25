package query

import (
	"fmt"
	"slices"
	"strings"
)

// sqlQuery is used to build sql query data, it includes a strings builder for
// constructing a statement and an args slice containing placeholder values.
// statement is a materialisation of the strings builder post build
type sqlQuery struct {
	sb        *strings.Builder
	args      []any
	statement string
}

// arg adds a new argument to the args list and returns a unique placeholder for
// use in the sql statement
func (q *sqlQuery) arg(v any) string {
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
	metadataBinder        *metadataBinder
	rootResource          TableMetadata
	selectOperation       *SelectOperation
	systemSelectOperation *SelectOperation
	filterExpression      FilterExpression
	nestedQueries         map[string]*nestedQuery
	aliasCount            int
	rootAlias             string
}

// NewSqlQueryBuilder constructs a new sqlQueryBuilderInstance from
// QueryOperations. It will return an error if the operations cannot be bound to
// the schema metadata
func NewSqlQueryBuilder(
	rootResource TableMetadata,
	operations []QueryOperation,
) (*sqlQueryBuilder, error) {
	b := &sqlQueryBuilder{
		metadataBinder: &metadataBinder{maximumDepth: 5},
		rootResource:   rootResource,
		nestedQueries:  map[string]*nestedQuery{},
		rootAlias:      "ra",
	}

	for _, op := range operations {
		var err error
		switch op := op.(type) {
		case *SelectOperation:
			err = b.metadataBinder.bindSelectOperation(rootResource, op)
			if err != nil {
				return nil, err
			}
			b.selectOperation = op
		case FilterOperation:
			err = b.metadataBinder.bindFilterOperation(rootResource, 0, op.FilterExpression)
			b.filterExpression = op.FilterExpression
		case *ExpandOperation:
			err = b.metadataBinder.bindExpandOperation(rootResource, 0, op)
			if err != nil {
				return nil, err
			}
			err = b.processExpandOperation(op)

		default:
			return nil, fmt.Errorf("unexpected query operation received: %v", op)
		}

		if err != nil {
			return nil, err
		}
	}

	return b, nil
}

func (b *sqlQueryBuilder) nextTableAlias() string {
	b.aliasCount++
	return fmt.Sprintf("a%d", b.aliasCount)
}

func (b *sqlQueryBuilder) processExpandOperation(op *ExpandOperation) error {
	for _, expand := range op.Expands {
		nestedQueryBuilder, err := NewSqlQueryBuilder(
			expand.Link.To,
			expand.Operations,
		)
		if err != nil {
			return err
		}

		// Select join on columns as they are required for server-side join
		b.addSystemSelect(expand.Link.Relationship.FromColumn().Name())
		nestedQueryBuilder.addSystemSelect(expand.Link.Relationship.ToColumn().Name())

		b.nestedQueries[string(expand.Link.Relationship.Id())] = &nestedQuery{
			queryBuilder: nestedQueryBuilder,
			link:         expand.Link,
		}
	}
	return nil
}

func (b *sqlQueryBuilder) addSystemSelect(columnName string) error {
	if b.systemSelectOperation == nil {
		b.systemSelectOperation = &SelectOperation{
			Columns: []*ColumnValue{},
		}
	}

	b.systemSelectOperation.Columns = append(
		b.systemSelectOperation.Columns,
		&ColumnValue{columnName: columnName},
	)
	return b.metadataBinder.bindSelectOperation(b.rootResource, b.systemSelectOperation)
}

func (b *sqlQueryBuilder) addAssociatedWithParentFilter(
	linkFromParent *TraversalStep,
	joinParentOnValues []string,
) error {
	associationFilter := &ComparisonExpression{
		Path: PropertyPath{
			Segments: []string{linkFromParent.Relationship.ToColumn().Name()},
		},
		Operator: ComparisonIn,
		Value: &StringListLiteral{
			Values: joinParentOnValues,
		},
	}
	err := b.metadataBinder.bindFilterOperation(b.rootResource, 0, associationFilter)
	if err != nil {
		return err
	}

	if b.filterExpression == nil {
		b.filterExpression = associationFilter
	} else {
		b.filterExpression = &LogicalExpression{
			Left:     associationFilter,
			Operator: LogicalAnd,
			Right:    b.filterExpression,
		}
	}
	return nil
}

func (b *sqlQueryBuilder) build() (*sqlQuery, error) {
	o := &sqlQuery{
		sb:   &strings.Builder{},
		args: []any{},
	}

	b.buildSelectStatement(o)
	o.sb.WriteRune(' ')
	b.buildFromStatement(o)
	o.sb.WriteRune(' ')
	err := b.buildFilterStatement(o)
	if err != nil {
		return nil, err
	}

	o.sb.WriteRune(' ')

	o.sb.WriteString("FOR JSON PATH;")

	o.statement = o.sb.String()

	return o, nil
}

func (b *sqlQueryBuilder) buildFromStatement(o *sqlQuery) {
	o.sb.WriteString("FROM ")
	o.sb.WriteString(b.rootResource.FullyQualifiedName())
	o.sb.WriteRune(' ')
	o.sb.WriteString(b.rootAlias)
}

func (b *sqlQueryBuilder) buildSelectStatement(o *sqlQuery) {
	// if no columns selected add all columns
	if b.selectOperation == nil || len(b.selectOperation.Columns) == 0 {
		b.selectOperation = &SelectOperation{
			Columns: make([]*ColumnValue, b.rootResource.ColumnCount()),
		}

		i := 0
		for col := range b.rootResource.Columns() {
			b.selectOperation.Columns[i] = &ColumnValue{
				columnName: col.Name(),
				columnData: col,
			}
			i++
		}
	} else if b.systemSelectOperation != nil && b.systemSelectOperation.Columns != nil {
		b.selectOperation.Columns = slices.Concat(
			b.selectOperation.Columns,
			b.systemSelectOperation.Columns,
		)
	}

	o.sb.WriteString("SELECT ")

	seen := map[string]struct{}{}
	i := 0
	for _, columnValue := range b.selectOperation.Columns {
		if _, ok := seen[columnValue.columnName]; ok {
			continue
		}
		seen[columnValue.columnName] = struct{}{}

		if i > 0 {
			o.sb.WriteString(",")
		}
		o.sb.WriteString(b.formatSelectValue(columnValue.columnData))
		i++
	}
}

func (b *sqlQueryBuilder) formatSelectValue(columnData ColumnMetadata) string {
	switch columnData.Type() {
	case DbTypeGeography:
		return fmt.Sprintf("%s.STAsText() AS %s", columnData.Name(), columnData.Name())
	default:
		return columnData.Name()
	}
}

func (b *sqlQueryBuilder) buildFilterStatement(o *sqlQuery) error {
	if b.filterExpression == nil {
		return nil
	}

	o.sb.WriteString("WHERE ")

	rootResource := &aliasedResource{resource: b.rootResource, alias: b.rootAlias}
	return b.buildFilterExpression(rootResource, b.filterExpression, o)
}

func (b *sqlQueryBuilder) buildFilterExpression(
	rootResource *aliasedResource,
	expression FilterExpression,
	o *sqlQuery,
) error {
	switch expression := expression.(type) {
	case *LogicalExpression:
		return b.buildLogicalExpression(rootResource, expression, o)
	case *ComparisonExpression:
		return b.buildComparisonExpression(rootResource, expression, o)
	case *CollectionExpression:
		return b.buildCollectionExpression(rootResource, expression, o)
	default:
		return internalErr("unexpected filter expression received")
	}
}

func (b *sqlQueryBuilder) buildLogicalExpression(
	rootResource *aliasedResource,
	expression *LogicalExpression,
	o *sqlQuery,
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
	o.sb.WriteRune('(')
	b.buildFilterExpression(rootResource, expression.Left, o)
	o.sb.WriteRune(' ')
	o.sb.WriteString(operator)
	o.sb.WriteRune(' ')
	b.buildFilterExpression(rootResource, expression.Right, o)
	o.sb.WriteRune(')')

	return nil
}

func (b *sqlQueryBuilder) buildComparisonExpression(
	rootResource *aliasedResource,
	ex *ComparisonExpression,
	o *sqlQuery,
) error {
	writeExpression := func(_ *aliasedResource) error {
		return ex.Value.WriteFilterExpression(o, ex.ResolvedColumn.columnName, ex.Operator)
	}

	b.writeExpressionWithPath(
		rootResource,
		ex.ResolvedPath,
		0,
		writeExpression,
		false,
		o,
	)
	return nil
}

func (b *sqlQueryBuilder) buildCollectionExpression(
	rootResource *aliasedResource,
	ex *CollectionExpression,
	o *sqlQuery,
) error {

	writeExpression := func(endResource *aliasedResource) error {
		return b.buildFilterExpression(
			endResource,
			ex.FilterExpression,
			o,
		)
	}

	// Invert if AND
	doNegate := ex.Operator == CollectionAll
	if doNegate {
		negatedCondition, err := negate(ex.FilterExpression)
		if err != nil {
			return err
		}

		ex.FilterExpression = negatedCondition
	}

	b.writeExpressionWithPath(
		rootResource,
		ex.ResolvedPath,
		0,
		writeExpression,
		doNegate,
		o,
	)
	return nil
}

func (b *sqlQueryBuilder) writeExpressionWithPath(
	parentResource *aliasedResource,
	path ResolvedPath,
	depth int,
	writeExpression func(rootResource *aliasedResource) error,
	doNegate bool,
	o *sqlQuery,
) error {
	if depth == len(path.Steps) {
		return writeExpression(parentResource)
	}

	node := path.Steps[depth]
	childResource := &aliasedResource{
		resource: node.To,
		alias:    b.nextTableAlias(),
	}

	if doNegate {
		o.sb.WriteString("NOT ")
	}

	o.sb.WriteString("EXISTS (SELECT 1 FROM ")
	o.sb.WriteString(childResource.resource.FullyQualifiedName())
	o.sb.WriteRune(' ')
	o.sb.WriteString(childResource.alias)
	o.sb.WriteString(" WHERE ")
	o.sb.WriteString(childResource.alias)
	o.sb.WriteRune('.')
	o.sb.WriteString(node.Relationship.ToColumn().Name())
	o.sb.WriteString(" = ")
	o.sb.WriteString(parentResource.alias)
	o.sb.WriteRune('.')
	o.sb.WriteString(node.Relationship.FromColumn().Name())
	o.sb.WriteString(" AND ")
	b.writeExpressionWithPath(
		childResource,
		path,
		depth+1,
		writeExpression,
		false,
		o,
	)

	o.sb.WriteRune(')')
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
		return ComparisonNotStartsWith, nil

	case ComparisonEndsWith:
		return ComparisonNotEndsWith, nil

	case ComparisonContains:
		return ComparisonNotContains, nil

	case ComparisonIn:
		return ComparisonNotIn, nil

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
