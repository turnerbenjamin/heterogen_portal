package queryOrchestrator

import (
	"context"

	qstore "github.com/turnerbenjamin/heterogen_portal/internal/query/queryDataStore"
	qerr "github.com/turnerbenjamin/heterogen_portal/internal/query/queryError"
	mdl "github.com/turnerbenjamin/heterogen_portal/internal/query/queryModel"
	qplan "github.com/turnerbenjamin/heterogen_portal/internal/query/queryPlanner"
)

var emptyResult = mdl.ExecuteResult{}

type QueryOrchestrator struct {
	newWriter          func(s qstore.QueryDataStore) (mdl.QueryWriter, error)
	repository         mdl.Repository
	pagingTokenBuilder qplan.PagingTokenBuilder
}

func NewQueryExecutor(
	repository mdl.Repository,
	newWriter func(s qstore.QueryDataStore) (mdl.QueryWriter, error),
	pagingTokeBuilder qplan.PagingTokenBuilder,
) *QueryOrchestrator {
	return &QueryOrchestrator{
		newWriter:          newWriter,
		repository:         repository,
		pagingTokenBuilder: pagingTokeBuilder,
	}
}

func (e *QueryOrchestrator) ExecuteQuery(
	ctx context.Context,
	s qstore.QueryDataStore,
) (mdl.ExecuteResult, error) {
	rootResource := s.RootResource()
	resourceModel := rootResource.InitModel()
	resourceMetadata := rootResource.GetMetadata()

	if resourceModel == nil {
		return emptyResult, qerr.BindingErr(
			"unable to access model for %s",
			resourceMetadata.Name,
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

	queryResults, stdErr := resourceModel.NewSlice(json, s.ProjectionNode())
	if stdErr != nil {
		return emptyResult, qerr.InternalErr("unable to create model slice: %v", stdErr)
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
