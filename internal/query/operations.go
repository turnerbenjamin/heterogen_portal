package query

type operationsState uint8

const (
	operationsStateEmpty operationsState = iota
	operationsStateRawQueriesParsed
	operationsStateEnrichmentDepsSet
	operationsStateDefaultOperationsSet
	operationsStateQueryShapeFinalised
	operationsStateMetadataBound
	operationsStateProjectionsSet
	operationsStateReadyForBuild
)

type Operations struct {
	state operationsState

	depth            int
	rootResource     TableMetadata
	rootAccessPolicy TableAccessPolicy
	accessPolicy     AccessPolicy
	metadataBinder   *metadataBinder

	projection []string

	SelectOperation       *SelectOperation
	SystemSelectOperation *SelectOperation
	ExpandOperation       *ExpandOperation
	FilterOperation       *FilterOperation
	OrderByOperation      *OrderByOperation
	CountOperation        *CountOperation
	LimitOperation        *LimitOperation
	PagingTokenOperation  *PagingTokenOperation

	CursorValues []ValueExpression

	nestedOperations map[string]*Operations
}

func (o *Operations) moveState(newState operationsState) error {
	if newState != o.state+1 {
		return internalErr("invalid state transition from %d to %d", o.state, newState)
	}
	o.state = newState
	return nil
}

func buildQueryOperations(
	rootResource TableMetadata,
	accessPolicy AccessPolicy,
	queryString string,
	queryParser QueryParser,
	pagingTokenBuilder PagingTokenBuilder,
) (*Operations, error) {
	// get raw operations from query string
	queryOperations, err := queryParser.Parse(queryString)
	if err != nil {
		return nil, err
	}

	if queryOperations.PagingTokenOperation != nil &&
		queryOperations.PagingTokenOperation.token != "" {
		token, err := pagingTokenBuilder.ParseToken(queryOperations.PagingTokenOperation.token)

		tokenOperations, err := queryParser.Parse(token.QueryString)
		if err != nil {
			return nil, err
		}
		queryOperations = tokenOperations
		queryOperations.CursorValues = token.CursorValues
	}

	if err := queryOperations.moveState(operationsStateRawQueriesParsed); err != nil {
		return nil, err
	}

	if err := queryOperations.enrichOperations(
		rootResource,
		accessPolicy,
		0,
	); err != nil {
		return nil, err
	}

	if queryOperations.state != operationsStateReadyForBuild {
		return nil, internalErr("expected query operations to be ready for build post enrichment")
	}

	return queryOperations, nil
}

func (ops *Operations) enrichOperations(
	rootResource TableMetadata,
	accessPolicy AccessPolicy,
	depth int,
) error {
	// Set dependencies required for enrichment such as resource metadata and
	// access policies
	if err := ops.setEnrichmentDependencies(depth, rootResource, accessPolicy); err != nil {
		return err
	}

	// add system defaults where operations are missing
	if err := ops.addSystemDefaults(); err != nil {
		return err
	}

	if len(ops.CursorValues) > 0 {
		if err := addCursorFilter(ops, ops.CursorValues); err != nil {
			return err
		}
	}

	// Bind metadata to the operations
	if err := ops.bindMetadata(); err != nil {
		return err
	}

	// Perform validations
	if err := ops.performValidations(); err != nil {
		return err
	}

	// Add select and expand relationship fields to the projection
	if err := ops.buildProjection(); err != nil {

		return err
	}

	// Handle nested operations
	if err := ops.addRequiredOperationsForExpands(); err != nil {
		return err
	}

	// Ensure that columns required for cursor pagination retrieved
	if err := ops.addRequiredOperationsForCursorPagination(); err != nil {
		return err
	}

	return ops.moveState(operationsStateReadyForBuild)
}

func (ops *Operations) setEnrichmentDependencies(
	depth int,
	rootResource TableMetadata,
	accessPolicy AccessPolicy,
) error {
	if ops.state < operationsStateRawQueriesParsed {
		return internalErr("parse raw user queries before enriching operations")
	}

	ops.depth = depth
	ops.rootResource = rootResource
	ops.accessPolicy = accessPolicy

	rootAccessPolicy, err := getTableAccessPolicy(accessPolicy, rootResource)
	if err != nil {
		return err
	}
	ops.rootAccessPolicy = rootAccessPolicy
	if !ops.rootAccessPolicy.CanAccess() {
		return accessErr("access to the %s table is denied", rootResource.Name())
	}

	ops.metadataBinder = &metadataBinder{
		maximumDepth: 5,
		accessPolicy: accessPolicy,
	}
	return ops.moveState(operationsStateEnrichmentDepsSet)
}

func (ops *Operations) addRequiredOperationsForExpands() error {
	if ops.state < operationsStateProjectionsSet {
		return internalErr("set projection before adding system operations for expands")
	}

	if ops.ExpandOperation != nil && ops.ExpandOperation.Expands != nil {
		for _, e := range ops.ExpandOperation.Expands {
			if err := ops.addRequiredOperationsForExpand(e); err != nil {
				return err
			}
		}
	}
	return nil
}

func (ops *Operations) addRequiredOperationsForExpand(e *Expand) error {
	if ops.state < operationsStateProjectionsSet {
		return internalErr("set projection before adding system operations for expand")
	}

	if e.Operations == nil {
		e.Operations = &Operations{}
	}

	// initialise nested operations
	nestedOps := e.Operations
	if err := nestedOps.moveState(operationsStateRawQueriesParsed); err != nil {
		return err
	}

	// enrich the nested operations with metadata
	if err := nestedOps.enrichOperations(
		e.Link.Relationship.To(),
		ops.accessPolicy,
		ops.depth+1,
	); err != nil {
		return err
	}

	// Add nested operations to mapping
	if ops.nestedOperations == nil {
		ops.nestedOperations = make(map[string]*Operations)
	}
	ops.nestedOperations[e.Link.Relationship.Id()] = nestedOps

	// Add system selects required for join
	if err := ops.addSystemSelect(e.Link.Relationship.FromColumn().Name()); err != nil {
		return err
	}
	if err := nestedOps.addSystemSelect(e.Link.Relationship.ToColumn().Name()); err != nil {
		return err
	}

	return nil
}

func (ops *Operations) addSystemDefaults() error {
	if ops.state < operationsStateEnrichmentDepsSet {
		return internalErr("set enrichment dependencies before setting system defaults")
	}

	// if no selects is given, default to selecting all fields
	if ops.SelectOperation == nil {
		if err := ops.addDefaultSelect(); err != nil {
			return err
		}
	}

	// if there is no limit, set a default limit of 5,000
	if ops.LimitOperation == nil {
		if err := ops.addDefaultLimit(); err != nil {
			return err
		}
	}

	// if there is no order by, order by the primary column
	if ops.OrderByOperation == nil {
		if err := ops.addDefaultOrderBy(); err != nil {
			return err
		}
	}

	// move state forward as all system operations have been set
	if err := ops.moveState(operationsStateDefaultOperationsSet); err != nil {
		return err
	}

	// Ensure that the primary key of the table is always contained in a sorting
	// operation
	if err := ops.ensureDeterministicSorting(); err != nil {
		return err
	}

	return ops.moveState(operationsStateQueryShapeFinalised)
}

func (ops *Operations) addDefaultSelect() error {
	if ops.state < operationsStateEnrichmentDepsSet {
		return internalErr("set enrichment dependencies before adding default select operation")
	}

	ops.SelectOperation = &SelectOperation{
		Columns: make(map[string]*ColumnValue, ops.rootResource.ColumnCount()),
	}

	i := 0
	for col := range ops.rootResource.Columns() {
		accessPolicy := ops.rootAccessPolicy.GetColumnAccessPolicy(col.Name())
		if accessPolicy == nil {
			return internalErr("unable to find access policy for %s.%s", ops.rootResource.Name(), col.Name())
		}

		if !accessPolicy.CanAccess() {
			continue
		}

		ops.SelectOperation.Columns[col.Name()] = &ColumnValue{
			ColumnName: col.Name(),
		}
		i++
	}
	return nil
}

func (ops *Operations) addDefaultOrderBy() error {
	if ops.state < operationsStateEnrichmentDepsSet {
		return internalErr("set enrichment dependencies before adding default orderby operation")
	}

	ops.OrderByOperation = &OrderByOperation{
		Rules: make([]*SortingRule, 0, 1),
	}

	defaultOrderBy := ops.OrderByOperation
	defaultOrderBy.Rules = append(defaultOrderBy.Rules, &SortingRule{
		Path: PropertyPath{
			Segments: []string{ops.rootResource.PrimaryKeyField().Name()},
		},
		Direction: "ASC",
	})
	return nil
}

func (ops *Operations) addDefaultLimit() error {
	if ops.state < operationsStateEnrichmentDepsSet {
		return internalErr("set enrichment dependencies before adding default limit operation")
	}

	ops.LimitOperation = &LimitOperation{
		Limit: 5000,
	}
	return nil
}

func (ops *Operations) ensureDeterministicSorting() error {
	if ops.state < operationsStateDefaultOperationsSet {
		return internalErr("set default operations before ensuring deterministic sorting")
	}

	if ops.OrderByOperation == nil || len(ops.OrderByOperation.Rules) == 0 {
		return internalErr(
			"expected an orderby operations with at least one rule specified",
		)
	}

	rules := ops.OrderByOperation.Rules

	// If primary key field present in order by rule return
	primaryKeyField := ops.rootResource.PrimaryKeyField()
	for _, rule := range rules {
		if len(rule.Path.Segments) == 1 && rule.Path.Segments[0] == primaryKeyField.Name() {
			return nil
		}
	}

	// else add to the rule
	ops.OrderByOperation.Rules = append(rules, &SortingRule{
		Path: PropertyPath{
			Segments: []string{ops.rootResource.PrimaryKeyField().Name()},
		},
		Direction: "ASC",
	})
	return nil
}

func (ops *Operations) performValidations() error {
	if ops.state < operationsStateQueryShapeFinalised {
		return internalErr("finalise query shape before performing validations")
	}

	if ops.LimitOperation == nil || ops.LimitOperation.Limit <= 0 {
		return syntaxErr("limit must be greater than 0")
	}
	return nil
}

func (ops *Operations) bindMetadata() error {
	if ops.state < operationsStateQueryShapeFinalised {
		return internalErr("finalise query shape before binding metadata")
	}

	if err := ops.metadataBinder.bindMetadata(
		ops.rootResource,
		ops,
	); err != nil {
		return err
	}
	return ops.moveState(operationsStateMetadataBound)
}

func (ops *Operations) addSystemSelect(columnName string) error {
	// create system selects mapping if nil
	if ops.SystemSelectOperation == nil {
		ops.SystemSelectOperation = &SelectOperation{
			Columns: map[string]*ColumnValue{},
		}
	}

	// return if user selects already contains the column
	if _, exists := ops.SelectOperation.Columns[columnName]; exists {
		return nil
	}

	// add the system select field and rebind the select operation
	ops.SystemSelectOperation.Columns[columnName] = &ColumnValue{ColumnName: columnName}
	return ops.metadataBinder.bindSelectOperation(
		ops.rootResource,
		ops.SystemSelectOperation,
	)
}

func (ops *Operations) buildProjection() error {
	if ops.state < operationsStateMetadataBound {
		return internalErr("bind metadata before building the query projection")
	}

	ops.projection = make([]string, 0, len(ops.SelectOperation.Columns))

	// add selects to projection
	for _, col := range ops.SelectOperation.Columns {
		ops.projection = append(ops.projection, col.ColumnName)
	}

	// add expand relationship columns to projection
	if ops.ExpandOperation != nil && len(ops.ExpandOperation.Expands) > 0 {
		for _, e := range ops.ExpandOperation.Expands {
			if e == nil {
				continue
			}
			ops.projection = append(
				ops.projection,
				e.Link.Relationship.RelationshipColumn(),
			)
		}
	}

	return ops.moveState(operationsStateProjectionsSet)
}

func (ops *Operations) addRequiredOperationsForCursorPagination() error {
	if ops.state < operationsStateProjectionsSet {
		return internalErr("please set projection before adding system operations")
	}

	// cursor pagination only needed at the top level query
	if ops.depth != 0 {
		return nil
	}

	rules := ops.OrderByOperation.Rules
	for _, rule := range rules {
		if rule.ResolvedPath == nil {
			return internalErr("resolved path has not been populated for orderby rule")
		}

		if rule.ResolvedPath.Id == "" {
			return internalErr("resolved path id has not been populated")
		}

		// if order by field is on the top level query just add the field to the
		// system select so that it is available
		if len(rule.ResolvedPath.Steps) == 0 {
			ops.addSystemSelect(rule.ResolvedColumn.ColumnName)
		} else {
			ops.addNestedSystemSelect(rule.ResolvedPath.Steps, rule.ResolvedColumn)
		}
	}

	return nil
}

func (ops *Operations) addNestedSystemSelect(pathSteps []TraversalStep, requiredColumn *ColumnValue) error {

	nextStep := pathSteps[0]
	nestedOperation, exists := ops.nestedOperations[nextStep.Relationship.Id()]

	if exists && nestedOperation != nil {
		// If an existing nested query is found for the orderby column, add a
		// system select to that query for the column and return
		remainingPath := pathSteps[1:]
		if len(remainingPath) == 0 {
			return nestedOperation.addSystemSelect(requiredColumn.ColumnName)
		} else {
			// If a nested query is found, but it is not the final resource, recall
			// the current function against the nested query with the remaining path
			return nestedOperation.addNestedSystemSelect(remainingPath, requiredColumn)
		}
	} else {
		// If a nested query is not found one, we need to construct a system
		// expand and select the required column
		return ops.addSystemExpand(pathSteps, requiredColumn)
	}
}

func (ops *Operations) addSystemExpand(pathSteps []TraversalStep, requiredColumn *ColumnValue) error {
	totalSteps := len(pathSteps)
	var lastExpand *Expand
	for i := totalSteps - 1; i >= 0; i-- {
		step := pathSteps[i]

		// Build Expand
		expand := &Expand{
			RelationshipName: step.Relationship.FromColumn().Name(),
			Operations:       &Operations{},
		}

		// Build Expand.select with primary key field only
		toPrimaryKeyField := step.To.PrimaryKeyField()
		selectColumns := map[string]*ColumnValue{
			toPrimaryKeyField.Name(): {
				ColumnName: toPrimaryKeyField.Name(),
				ColumnData: toPrimaryKeyField,
			},
		}
		expand.Operations.SelectOperation = &SelectOperation{Columns: selectColumns}

		if i == totalSteps-1 {
			// If we are on the final step, add the required column to the
			// select
			selectColumns[requiredColumn.ColumnName] = requiredColumn
		} else {
			// if we are on an intermediate step, add the last expands to the
			// current expand operations
			expand.Operations.ExpandOperation = &ExpandOperation{
				Expands: map[string]*Expand{
					expand.RelationshipName: expand,
				},
			}
		}
		lastExpand = expand
	}

	if ops.ExpandOperation == nil {
		ops.ExpandOperation = &ExpandOperation{
			Expands: map[string]*Expand{},
		}
	}
	ops.ExpandOperation.Expands[lastExpand.RelationshipName] = lastExpand

	err := ops.metadataBinder.bindExpand(ops.rootResource, lastExpand)
	if err != nil {
		return err
	}

	return ops.addRequiredOperationsForExpand(lastExpand)
}

func getTableAccessPolicy(accessPolicy AccessPolicy, tableData TableMetadata) (TableAccessPolicy, error) {
	if accessPolicy == nil {
		return nil, internalErr("access policy cannot be nil")
	}

	tablePolicy := accessPolicy.GetTableAccessPolicy(tableData.Name())
	if tablePolicy == nil {
		return nil, internalErr("unable to find access policy for the %s table", tableData.Name())
	}

	if !tablePolicy.CanAccess() {
		return nil, accessErr("you do not have permission to access the %s table", tableData.Name())
	}
	return tablePolicy, nil
}

func addCursorFilter(queryOperations *Operations, cursorValues []ValueExpression) error {
	// Validate that rules and cursor values can be zipped together
	if queryOperations.OrderByOperation == nil ||
		cursorValues == nil ||
		len(queryOperations.OrderByOperation.Rules) == 0 ||
		len(queryOperations.OrderByOperation.Rules) != len(cursorValues) {
		return internalErr(
			"expected at least one orderby rule with one value expression" +
				"for each rule",
		)
	}

	// Validate that the last value is non-null
	lastValue := cursorValues[len(cursorValues)-1]
	if lastValue.GetType() == literalTypeNull {
		return internalErr(
			"expected the final cursor value to be a non-null value",
		)
	}

	// Build the cursor filter
	var cursorFilter FilterExpression
	for i, rule := range queryOperations.OrderByOperation.Rules {
		cursorValue := cursorValues[i]

		if rule.Direction == SortDirectionDesc && cursorValue.GetType() == literalTypeNull {
			continue
		}

		ruleExpression := getCursorFilterComparisonOperator(
			rule,
			cursorValue,
		)
		if ruleExpression == nil {
			return internalErr("unable to generate cursor filter")
		}

		for j := range i {
			previousValue := cursorValues[j]
			previousRule := queryOperations.OrderByOperation.Rules[j]

			ruleExpression = &LogicalExpression{
				Left:     ruleExpression,
				Operator: LogicalAnd,
				Right: &ComparisonExpression{
					Path:     previousRule.Path,
					Operator: ComparisonEq,
					Value:    previousValue,
				},
			}
		}

		if cursorFilter == nil {
			cursorFilter = ruleExpression
		} else {
			cursorFilter = &LogicalExpression{
				Left:     cursorFilter,
				Operator: LogicalOr,
				Right:    ruleExpression,
			}
		}
	}

	// Add the cursor filter to the query operations
	if queryOperations.FilterOperation == nil {
		queryOperations.FilterOperation = &FilterOperation{
			FilterExpression: cursorFilter,
		}
	} else {
		queryOperations.FilterOperation.FilterExpression = &LogicalExpression{
			Left:     queryOperations.FilterOperation.FilterExpression,
			Operator: LogicalAnd,
			Right:    cursorFilter,
		}
	}

	return nil
}
