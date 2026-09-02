package queryOrchestrator

import (
	"context"
	"sync"

	bldr "github.com/turnerbenjamin/heterogen_portal/internal/query/queryBuilder"
	qstore "github.com/turnerbenjamin/heterogen_portal/internal/query/queryDataStore"
	qerr "github.com/turnerbenjamin/heterogen_portal/internal/query/queryError"
	mdl "github.com/turnerbenjamin/heterogen_portal/internal/query/queryModel"
	"golang.org/x/sync/errgroup"
)

type nestedQueryResult struct {
	link    mdl.TraversalStep
	results []mdl.TableModel
}

var emptyResult = mdl.ExecuteResult{}

type QueryOrchestrator struct {
	newWriter          func(s qstore.QueryDataStore) (mdl.QueryWriter, error)
	repository         mdl.Repository
	pagingTokenBuilder bldr.PagingTokenBuilder
	valueBuilder       mdl.ValueBuilder
}

func NewQueryExecutor(
	repository mdl.Repository,
	newWriter func(s qstore.QueryDataStore) (mdl.QueryWriter, error),
	pagingTokeBuilder bldr.PagingTokenBuilder,
	valueBuilder mdl.ValueBuilder,
) *QueryOrchestrator {
	return &QueryOrchestrator{
		newWriter:          newWriter,
		repository:         repository,
		pagingTokenBuilder: pagingTokeBuilder,
		valueBuilder:       valueBuilder,
	}
}

func (e *QueryOrchestrator) ExecuteQuery(
	ctx context.Context,
	s qstore.QueryDataStore,
) (mdl.ExecuteResult, error) {
	rootResource := s.RootResource()
	resourceModel := rootResource.GetModel()
	if resourceModel == nil {
		return emptyResult, qerr.BindingErr(
			"unable to access model for %s",
			rootResource.Name(),
		)
	}

	w, err := e.newWriter(s)
	if err != nil {
		return emptyResult, err
	}

	queryStatement := w.WriteQueryStatement()
	countStatement := w.WriteCountStatement()
	args := w.Args()

	json, count, err := e.repository.ExecuteJsonRequestWithCount(
		ctx,
		queryStatement,
		countStatement,
		args,
	)
	if err != nil {
		return emptyResult, qerr.InternalErr("query executor failed: %v", err)
	}

	queryResults, stdErr := resourceModel.NewSlice(json, s.Projection())
	if stdErr != nil {
		return emptyResult, qerr.InternalErr("unable to create model slice: %v", stdErr)
	}

	err = e.populateNestedResults(ctx, s, queryResults)
	if err != nil {
		return emptyResult, err
	}

	nextPageToken, err := e.getNextPageToken(s, &queryResults)
	if err != nil {
		return emptyResult, err
	}

	// Top level queries set the limit to the requested limit + 1 so that it is
	// possible to determine if a next page of results exists
	return mdl.ExecuteResult{
		Count:         count,
		NextPageToken: nextPageToken,
		Data:          queryResults,
	}, nil
}

func (e *QueryOrchestrator) populateNestedResults(
	ctx context.Context,
	s qstore.QueryDataStore,
	queryResults []mdl.TableModel,
) error {
	nestedQueryResults, err := e.getNestedQueryResults(ctx, s, queryResults)
	if err != nil {
		return err
	}

	for _, result := range nestedQueryResults {
		link := result.link
		nestedResults := result.results

		switch link.Relationship.Type() {
		case mdl.RelationshipManyToOne:
			err := attachNestedResultsForManyToOneQuery(link.Relationship, queryResults, nestedResults)
			if err != nil {
				return err
			}
		case mdl.RelationshipOneToMany:
			err := attachNestedResultsForOneToManyQuery(link.Relationship, queryResults, nestedResults)
			if err != nil {
				return err
			}
		default:
			return qerr.InternalErr(
				"unsupported relationship type %v",
				link.Relationship.Type(),
			)
		}
	}
	return nil
}

func (e *QueryOrchestrator) getNestedQueryResults(
	ctx context.Context,
	s qstore.QueryDataStore,
	queryResults []mdl.TableModel,
) (map[string]*nestedQueryResult, error) {

	if s.ExpandsLen() == 0 || len(queryResults) == 0 {
		return nil, nil
	}

	// Execute nested queries asynchronously and collect the results in a map
	nestedQueryResults := make(map[string]*nestedQueryResult)
	var mu sync.Mutex
	g, _ := errgroup.WithContext(ctx)
	for _, nestedQuery := range s.Expands() {
		g.Go(func() error {
			nestedResults, err := e.executeNestedQuery(ctx, nestedQuery, queryResults)
			if err != nil {
				return err
			}

			mu.Lock()
			nestedQueryResults[nestedResults.link.Relationship.Id()] = nestedResults
			mu.Unlock()

			return nil
		})
	}

	// Wait for all nested queries to resolve
	err := g.Wait()
	if err != nil {
		return nil, err
	}

	// return results map
	return nestedQueryResults, nil
}

func (e *QueryOrchestrator) executeNestedQuery(
	ctx context.Context,
	nestedQuery qstore.Expansion,
	queryResults []mdl.TableModel,
) (*nestedQueryResult, error) {
	link := nestedQuery.TraversalStep
	queryData := nestedQuery.QueryData

	joinOnValues, err := e.getJoinOnValues(link, queryResults)
	if err != nil {
		return nil, err
	}

	err = queryData.AddAssociatedWithParentFilter(
		link,
		joinOnValues,
	)
	if err != nil {
		return nil, err
	}

	results, err := e.executeNestedQueries(ctx, queryData)
	if err != nil {
		return nil, err
	}

	return &nestedQueryResult{
		link:    link,
		results: results,
	}, nil
}

func (e *QueryOrchestrator) executeNestedQueries(
	ctx context.Context,
	s qstore.QueryDataStore,
) ([]mdl.TableModel, error) {
	rootResource := s.RootResource()
	resourceModel := rootResource.GetModel()
	if resourceModel == nil {
		return nil, qerr.BindingErr(
			"unable to access model for %s",
			rootResource.Name(),
		)
	}

	w, err := e.newWriter(s)
	if err != nil {
		return nil, err
	}

	queryStatement := w.WriteQueryStatement()
	args := w.Args()

	json, err := e.repository.ExecuteJsonRequest(ctx, queryStatement, args)
	if err != nil {
		return nil, qerr.InternalErr("query executor failed: %v", err)
	}

	queryResults, stdErr := resourceModel.NewSlice(
		json,
		s.Projection(),
	)
	if stdErr != nil {
		return nil, qerr.InternalErr("unable to create model slice: %v", stdErr)
	}

	err = e.populateNestedResults(ctx, s, queryResults)
	if err != nil {
		return nil, err
	}

	return queryResults, nil
}

func (e *QueryOrchestrator) getJoinOnValues(
	link mdl.TraversalStep,
	fromResults []mdl.TableModel,
) ([]mdl.ValueExpression, error) {
	seen := map[string]struct{}{}
	joinValues := make([]mdl.ValueExpression, 0, len(fromResults))

	for _, fromResult := range fromResults {
		value, err := fromResult.GetJoinOnValue(link.Relationship.Id())
		if err != nil {
			return nil, err
		}

		if _, exists := seen[value]; !exists {
			joinValues = append(joinValues, e.valueBuilder.String(value))
		}
		seen[value] = struct{}{}
	}
	return joinValues, nil
}

func attachNestedResultsForManyToOneQuery(
	relationship mdl.RelationshipMetadata,
	parentResults,
	nestedResults []mdl.TableModel,
) error {
	nestedResultsMap := map[string]mdl.TableModel{}
	for _, nestedResult := range nestedResults {
		joinOnValue, err := nestedResult.GetJoinOnValue(relationship.Id())
		if err != nil {
			return qerr.InternalErr("unable to get join on value: %v", err)
		}
		nestedResultsMap[joinOnValue] = nestedResult
	}

	for _, result := range parentResults {
		joinOnValue, err := result.GetJoinOnValue(relationship.Id())
		if err != nil {
			return qerr.InternalErr("unable to get join on value: %v", err)
		}
		related, ok := nestedResultsMap[joinOnValue]
		if !ok || related == nil {
			continue
		}
		result.SetRelationshipField(relationship.Id(), related)
	}
	return nil
}

func attachNestedResultsForOneToManyQuery(
	relationship mdl.RelationshipMetadata,
	parentResults,
	nestedResults []mdl.TableModel,
) error {
	parentResultsMap := map[string]mdl.TableModel{}
	for _, parentResult := range parentResults {
		joinOnValue, err := parentResult.GetJoinOnValue(relationship.Id())
		if err != nil {
			return qerr.InternalErr("unable to get join on value: %v", err)
		}
		parentResultsMap[joinOnValue] = parentResult
		if err := parentResult.InitRelationshipField(relationship.Id()); err != nil {
			return err
		}
	}

	for _, nestedResult := range nestedResults {
		joinOnValue, err := nestedResult.GetJoinOnValue(relationship.Id())
		if err != nil || joinOnValue == "" {
			return qerr.InternalErr("unable to get join on value: %v", err)
		}
		parent, ok := parentResultsMap[joinOnValue]
		if !ok || parent == nil {
			return qerr.InternalErr("unable to join query results")
		}
		if err := parent.SetRelationshipField(relationship.Id(), nestedResult); err != nil {
			return err
		}
	}
	return nil
}

func (e *QueryOrchestrator) getNextPageToken(
	s qstore.QueryDataStore,
	queryResults *[]mdl.TableModel,
) (string, error) {
	results := *queryResults
	limit := s.Limit()
	isNextRecord := len(results) > int(limit)

	if !isNextRecord {
		return "", nil
	}
	lastRecord := results[limit-1]
	*queryResults = results[0:limit]

	return e.pagingTokenBuilder.BuildToken(
		s,
		lastRecord,
	)
}
