package queryPlanner

import (
	"github.com/turnerbenjamin/heterogen_portal/internal/query/paginationTokens"
	tkns "github.com/turnerbenjamin/heterogen_portal/internal/query/paginationTokens"
	qstore "github.com/turnerbenjamin/heterogen_portal/internal/query/queryDataStore"
	qerr "github.com/turnerbenjamin/heterogen_portal/internal/query/queryError"
	mdl "github.com/turnerbenjamin/heterogen_portal/internal/query/queryModel"
)

// TODO: SORT OUT SOME CONFIG
const MAX_LIMIT = 5000

type PagingTokenBuilder interface {
	BuildToken(
		queryDataStore qstore.QueryDataStore,
		lastRecord mdl.TableModel,
	) (string, error)

	ParseToken(token string, valueBuilder mdl.ValueBuilder) (*tkns.PagingToken, error)
}

type QueryParser interface {
	Parse(
		queryString string,
		queryDataStore qstore.QueryDataStore,
	) error
}

func ConfigureQuery(
	queryString string,
	queryParser QueryParser,
	pagingTokenBuilder PagingTokenBuilder,
	valueBuilder mdl.ValueBuilder,
	rootResource mdl.TableMetadata,
	accessPolicy mdl.AccessPolicy,
) (qstore.QueryDataStore, error) {
	// Initialise query data store
	s, err := initStore(
		queryString,
		queryParser,
		valueBuilder,
		rootResource,
		accessPolicy,
	)
	if err != nil {
		return nil, err
	}

	// Parse Token
	var pagingToken *tkns.PagingToken
	if tknStr, exists := s.PagingToken(); exists {
		pagingToken, err = pagingTokenBuilder.ParseToken(
			tknStr,
			valueBuilder,
		)
		if err != nil {
			return nil, err
		}

		// TODO implement token validation logic
		s, err = initStore(
			pagingToken.QueryString,
			queryParser,
			valueBuilder,
			rootResource,
			accessPolicy,
		)
		if err != nil {
			return nil, err
		}
	}

	// Configure the query
	if err := configureQuery(s, pagingToken); err != nil {
		return nil, err
	}

	return s, err
}

func initStore(
	queryString string,
	queryParser QueryParser,
	valueBuilder mdl.ValueBuilder,
	rootResource mdl.TableMetadata,
	accessPolicy mdl.AccessPolicy,
) (qstore.QueryDataStore, error) {
	s, err := qstore.NewQueryDataStore(
		queryString,
		rootResource,
		accessPolicy,
		valueBuilder,
	)
	if err != nil {
		return nil, err
	}

	// Parse query string into data store
	if err := queryParser.Parse(queryString, s); err != nil {
		return nil, err
	}

	return s, nil
}

func configureQuery(
	s qstore.QueryDataStore,
	pagingToken *paginationTokens.PagingToken,
) error {
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

	// Set system limit for pagination record
	if err := setSystemLimit(s); err != nil {
		return err
	}

	// Configure expanded queries
	for _, expansion := range s.Expands() {
		if err := configureQuery(expansion.QueryData, nil); err != nil {
			return err
		}
	}

	// Add cursor filter where required
	addCursorFilter(s, pagingToken)

	return nil
}

func validateUserQuery(s qstore.QueryDataStore) error {
	if s.Limit() > MAX_LIMIT {
		return qerr.SyntaxErr("limit must be between 1 and %d", MAX_LIMIT)
	}
	return nil
}

func addSystemDefaults(s qstore.QueryDataStore) error {
	p := s.Projection()
	if p.IsEmpty() {
		if err := addDefaultSelects(s); err != nil {
			return err
		}
	}

	if s.Limit() == 0 {
		s.SetLimit(MAX_LIMIT)
	}

	if s.OrderByLen() == 0 {
		s.AddOrderBy(s.RootResource().PrimaryKeyField().Name(), mdl.SortDirectionAsc)

	}

	return nil
}

func addDefaultSelects(s qstore.QueryDataStore) error {
	// Add all columns the user can access to the table
	tableAccessPolicy := s.RootResourceAccessPolicy()
	for col := range s.RootResource().Columns() {
		colAccessPolicy := tableAccessPolicy.GetColumnAccessPolicy(col.Name())
		if colAccessPolicy == nil {
			return qerr.InternalErr(
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
		return qerr.InternalErr("invalid access policy, the user does not have access to any columns")
	}
	return nil
}

func ensureDeterministicOrdering(s qstore.QueryDataStore) error {

	primaryKeyField := s.RootResource().PrimaryKeyField().Name()

	// exit early if query is already sorted by the primary key field
	for rule := range s.OrderBy() {
		if len(rule.ResolvedColumn.ResolvedPath.Steps) == 0 &&
			rule.ResolvedColumn.Metadata.Name() == primaryKeyField {
			return nil
		}
	}
	return s.AddOrderBy(primaryKeyField, mdl.SortDirectionAsc)
}

func addRequiredOperationsForCursorPagination(s qstore.QueryDataStore) error {
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

func addNestedSystemSelect(s qstore.QueryDataStore, resolvedColumn mdl.ResolvedColumn) error {
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
			return nestedOperation.QueryData.AddSystemSelect(resolvedColumn.Metadata.Name())
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

func addSystemExpand(s qstore.QueryDataStore, resolvedColumn mdl.ResolvedColumn) error {
	pathSteps := resolvedColumn.ResolvedPath.Steps
	totalSteps := len(pathSteps)

	currentStore := s
	for i := totalSteps - 1; i >= 0; i-- {
		step := pathSteps[i]

		nestedOperations, err := currentStore.AddSystemExpand(step.Relationship.FromColumn().Name())
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

func setSystemLimit(s qstore.QueryDataStore) error {
	l := int(s.Limit())

	// Set system limit as limit + 1 for top-level queries so that we can
	// check if there are additional records
	if s.IsTopLevelQuery() {
		l++
	}

	return s.SetSystemLimit(l)
}

func addCursorFilter(
	s qstore.QueryDataStore,
	paginationToken *paginationTokens.PagingToken,
) error {
	if paginationToken == nil {
		return nil
	}

	cursorValues := paginationToken.CursorValues

	// validate that order by set and order by can be zipped to cursor values
	if s.OrderByLen() == 0 || s.OrderByLen() != len(cursorValues) {
		return qerr.InternalErr(
			"expected at least one orderby rule with one value expression" +
				"for each rule",
		)
	}

	b := s.FilterExpressionBuilder()
	// Build the cursor filter
	var cursorFilter mdl.FilterExpression
	i := 0
	for rule := range s.OrderBy() {
		cursorValue := cursorValues[i]

		if rule.Direction == mdl.SortDirectionDesc &&
			cursorValue.Type() == mdl.LiteralTypeNull {
			continue
		}

		ruleExpression, err := getCursorFilterComparisonOperator(
			rule,
			cursorValue,
			b,
		)
		if err != nil {
			return err
		}

		if ruleExpression == nil {
			return qerr.InternalErr("unable to generate cursor filter")
		}

		j := 0
		for previousRule := range s.OrderBy() {
			if j == i {
				break
			}

			previousValue := cursorValues[j]
			previousRuleFilter, err := b.NewComparisonExpressionFromResolvedColumn(
				previousRule.ResolvedColumn,
				mdl.ComparisonEq,
				previousValue,
			)
			if err != nil {
				return err
			}

			ruleExpression, err = b.NewLogicalExpression(
				ruleExpression,
				mdl.LogicalAnd,
				previousRuleFilter,
			)
			if err != nil {
				return err
			}
			j++
		}

		if cursorFilter == nil {
			cursorFilter = ruleExpression
		} else {
			cursorFilter, err = b.NewLogicalExpression(
				cursorFilter,
				mdl.LogicalOr,
				ruleExpression,
			)
			if err != nil {
				return err
			}
		}
		i++
	}

	if s.FilterExpression() == nil {
		s.SetFilterExpression(cursorFilter)
		return nil
	}

	finalFilter, err := b.NewLogicalExpression(
		s.FilterExpression(),
		mdl.LogicalAnd,
		cursorFilter,
	)
	if err != nil {
		return err
	}
	s.SetFilterExpression(finalFilter)

	return nil
}

func getCursorFilterComparisonOperator(
	rule mdl.SortingRule,
	value mdl.ValueExpression,
	b qstore.FilterExpressionBuilder,
) (mdl.FilterExpression, error) {
	if rule.Direction == mdl.SortDirectionAsc {
		if value.Type() == mdl.LiteralTypeNull {
			// Ascending logic for null value
			return b.NewComparisonExpressionFromResolvedColumn(
				rule.ResolvedColumn,
				mdl.ComparisonNe,
				value,
			)
		} else {
			// Ascending logic for non-null value
			return b.NewComparisonExpressionFromResolvedColumn(
				rule.ResolvedColumn,
				mdl.ComparisonGt,
				value,
			)
		}
	} else {
		if value.Type() == mdl.LiteralTypeNull {
			// Descending logic for null value, null is already the last value
			// so do not add a filter
			return nil, nil
		}
		l, err := b.NewComparisonExpressionFromResolvedColumn(
			rule.ResolvedColumn,
			mdl.ComparisonLt,
			value,
		)
		if err != nil {
			return nil, err
		}

		r, err := b.NewComparisonExpressionFromResolvedColumn(
			rule.ResolvedColumn,
			mdl.ComparisonEq,
			b.ValueBuilder().Null(),
		)
		if err != nil {
			return nil, err
		}
		return b.NewLogicalExpression(l, mdl.LogicalOr, r)
	}
}
