package query

const MAX_LIMIT = 5000

func BuildQuery(
	queryString string,
	queryParser QueryParser,
	rootResource TableMetadata,
	accessPolicy AccessPolicy,
) (QueryDataStore, error) {
	// Initialise query data store
	s, err := NewQueryDataStore(
		queryString,
		rootResource,
		accessPolicy,
	)
	if err != nil {
		return nil, err
	}

	// Parse query string into data store
	if err := queryParser.Parse(queryString, s); err != nil {
		return nil, err
	}

	// Configure the query
	if err := configureQuery(s); err != nil {
		return nil, err
	}

	return s, err
}

func configureQuery(s QueryDataStore) error {
	// Validate user query as parsed before mutating
	if err := validateUserQuery(s); err != nil {
		return err
	}

	// Add default operations as required
	if err := addSystemDefaults(s); err != nil {
		return err
	}

	// Ensure deterministic order
	if err := ensureDeterministicOrdering(s); err != nil {
		return err
	}

	// Add any required systems selects/expands required for cursor pagination
	if err := addRequiredOperationsForCursorPagination(s); err != nil {
		return err
	}

	// Configure expanded queries
	for _, expansion := range s.Expands() {
		if err := configureQuery(expansion.queryData); err != nil {
			return err
		}
	}

	return nil
}

func validateUserQuery(s QueryDataStore) error {
	if s.Limit() > MAX_LIMIT {
		return syntaxErr("limit must be between 1 and %d", MAX_LIMIT)
	}
	return nil
}

func addSystemDefaults(s QueryDataStore) error {
	if len(s.Projection()) == 0 {
		if err := addDefaultSelects(s); err != nil {
			return err
		}
	}

	if s.Limit() == 0 {
		s.SetLimit(MAX_LIMIT)
	}

	if s.OrderByLen() == 0 {
		s.AddOrderBy(s.RootResource().PrimaryKeyField().Name(), SortDirectionAsc)
	}

	return nil
}

func addDefaultSelects(s QueryDataStore) error {
	// Add all columns the user can access to the table
	tableAccessPolicy := s.RootResourceAccessPolicy()
	for col := range s.RootResource().Columns() {
		colAccessPolicy := tableAccessPolicy.GetColumnAccessPolicy(col.Name())
		if colAccessPolicy == nil {
			return internalErr(
				"unable to add default select: access policy for %s is nil",
				col.Name(),
			)
		}

		if !colAccessPolicy.CanAccess() {
			continue
		}

		if err := s.AddSelect(col.Name()); err != nil {
			return err
		}
	}

	if s.SelectsLen() == 0 {
		return internalErr("invalid access policy, the user does not have access to any columns")
	}
	return nil
}

func ensureDeterministicOrdering(s QueryDataStore) error {
	primaryKeyField := s.RootResource().PrimaryKeyField().Name()

	// exit early if query is already sorted by the primary key field
	for rule := range s.OrderBy() {
		if len(rule.ResolvedColumn.ResolvedPath.Steps) == 0 &&
			rule.ResolvedColumn.Metadata.Name() == primaryKeyField {
			return nil
		}
	}
	return s.AddOrderBy(primaryKeyField, SortDirectionAsc)
}

func addRequiredOperationsForCursorPagination(s QueryDataStore) error {
	if !s.IsTopLevelQuery() {
		return nil
	}

	for rule := range s.OrderBy() {
		col := rule.ResolvedColumn
		// If column is on the root resource just add a system select
		if len(col.ResolvedPath.Steps) == 0 {
			if err := s.AddSystemSelect(col.Metadata.Name()); err != nil {
				return err
			}
		} else {
			// If the column is nested add a nested select
			if err := addNestedSystemSelect(s, col); err != nil {
				return err
			}
		}

	}
	return nil
}

func addNestedSystemSelect(s QueryDataStore, resolvedColumn ResolvedColumn) error {
	// Shift first step from the array
	nextStep := resolvedColumn.ResolvedPath.Steps[0]

	// Check for existing nested operation
	nestedOperation, exists := s.GetExpansionByRelationshipId(nextStep.Relationship.Id())
	if exists {
		// bring resolved column path forward to root from expanded entity
		remainingSteps := resolvedColumn.ResolvedPath.Steps[1:]
		resolvedColumn.ResolvedPath.Steps = remainingSteps

		// If an existing nested query is found for the orderby column, add a
		// system select to that query for the column and return
		if len(remainingSteps) == 0 {
			return nestedOperation.queryData.AddSystemSelect(resolvedColumn.Metadata.Name())
		} else {
			// If a nested query is found, but it is not the final resource, recall
			// the current function against the nested query with the remaining path
			return addNestedSystemSelect(s, resolvedColumn)
		}
	} else {
		// If a nested query is not found one, we need to construct a system
		// expand and select the required column
		return addSystemExpand(s, resolvedColumn)
	}
}

func addSystemExpand(s QueryDataStore, resolvedColumn ResolvedColumn) error {
	pathSteps := resolvedColumn.ResolvedPath.Steps
	totalSteps := len(pathSteps)

	currentStore := s
	for i := totalSteps - 1; i >= 0; i-- {
		step := pathSteps[i]

		nestedOperations, err := currentStore.AddSystemExpand(step.Relationship.Id())
		if err != nil {
			return err
		}

		nestedResource := nestedOperations.RootResource()

		if i == totalSteps-1 {
			if err := nestedOperations.AddSelect(resolvedColumn.Metadata.Name()); err != nil {
				return err
			}
		} else {
			if err := nestedOperations.AddSelect(nestedResource.Name()); err != nil {
				return err
			}
		}
		currentStore = nestedOperations
	}
	return nil
}
