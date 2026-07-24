package query

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/turnerbenjamin/heterogen_portal/internal/model"
	"golang.org/x/sync/errgroup"
)

type Query struct {
	ctx           context.Context
	queryExecutor func(ctx context.Context, statementStr string, args []any) (jsonResult []byte, err error)
	tableData     model.TableMetadata
	queryBuilder  *sqlQueryBuilder
}

type nestedQueryResult struct {
	link    *TraversalStep
	results []model.TableModel
}

func NewQuery(
	ctx context.Context,
	exectuteQuery func(ctx context.Context, statementStr string, args []any) (jsonResult []byte, err error),
	resource string,
	operations []QueryOperation,
) (*Query, error) {
	resourceModel := model.GetTableModel(resource)
	if resourceModel == nil {
		return nil, fmt.Errorf("no resource found for %s", resource)
	}
	resourceMetadata := resourceModel.GetMetadata()

	queryBuilder, err := NewSqlQueryBuilder(
		&resourceMetadata,
		operations,
	)
	if err != nil {
		return nil, err
	}

	q := &Query{
		ctx:           ctx,
		queryExecutor: exectuteQuery,
		tableData:     resourceModel.GetMetadata(),
		queryBuilder:  queryBuilder,
	}
	return q, nil
}

func (q *Query) Execute() ([]model.TableModel, error) {
	return q.executeQuery(q.queryBuilder)
}

func (q *Query) executeQuery(qb *sqlQueryBuilder) ([]model.TableModel, error) {
	resourceModel := model.GetTableModel(qb.rootResource.GetResourceShortName())
	if resourceModel == nil {
		return nil, fmt.Errorf(
			"unable to access model for %s",
			qb.rootResource.GetResourceShortName(),
		)
	}

	query, err := qb.Build()
	if err != nil {
		return nil, err
	}

	json, err := q.queryExecutor(q.ctx, query.Statement, query.Args)
	if err != nil {
		return nil, err
	}

	queryResults, err := resourceModel.NewSlice(json)
	if err != nil {
		return nil, err
	}

	nestedQueryResults := make(map[string]*nestedQueryResult)
	var mu sync.Mutex

	// Execute any nested queries
	if len(queryResults) > 0 && len(qb.nestedQueries) > 0 {
		g, _ := errgroup.WithContext(q.ctx)

		for _, nestedQuery := range qb.nestedQueries {

			g.Go(func() error {
				link := nestedQuery.link
				nestedQueryBuilder := nestedQuery.queryBuilder

				joinOnValues, err := q.getJoinOnValues(link, queryResults)
				if err != nil {
					return err
				}

				err = nestedQueryBuilder.AddAssociatedWithParentFilter(
					link,
					joinOnValues,
				)
				if err != nil {
					return err
				}

				results, err := q.executeQuery(nestedQueryBuilder)
				if err != nil {
					return err
				}

				mu.Lock()
				nestedQueryResults[string(link.Relationship.Id)] = &nestedQueryResult{
					link:    link,
					results: results,
				}
				mu.Unlock()
				return nil
			})

		}

		err := g.Wait()
		if err != nil {
			return nil, err
		}

		// Attach nested queries to main results
		for _, result := range nestedQueryResults {
			link := result.link
			nestedResults := result.results

			switch link.Relationship.Type {
			case model.RelationshipManyToOne:
				err := attachNestedResultsForManyToOneQuery(link.Relationship, queryResults, nestedResults)
				if err != nil {
					return nil, err
				}
			case model.RelationshipOneToMany:
				err := attachNestedResultsForOneToManyQuery(link.Relationship, queryResults, nestedResults)
				if err != nil {
					return nil, err
				}
			default:
				return nil, fmt.Errorf(
					"unsupported relationship type %v",
					link.Relationship.Type,
				)
			}
		}
	}

	return queryResults, nil
}

func (q Query) getJoinOnValues(
	link *TraversalStep,
	fromResults []model.TableModel,
) ([]string, error) {
	seen := map[string]struct{}{}
	joinValues := make([]string, 0, len(fromResults))

	for _, fromResult := range fromResults {
		value, err := fromResult.GetJoinOnValue(link.Relationship.Id)
		if err != nil {
			return nil, err
		}

		if _, exists := seen[value]; !exists {
			joinValues = append(joinValues, value)
		}
		seen[value] = struct{}{}
	}
	return joinValues, nil
}

func attachNestedResultsForManyToOneQuery(relationship *model.Relationship, parentResults, nestedResults []model.TableModel) error {
	nestedResultsMap := map[string]model.TableModel{}
	for _, nestedResult := range nestedResults {
		joinOnValue, err := nestedResult.GetJoinOnValue(relationship.Id)
		if err != nil {
			return err
		}
		nestedResultsMap[joinOnValue] = nestedResult
	}

	for _, result := range parentResults {
		joinOnValue, err := result.GetJoinOnValue(relationship.Id)
		if err != nil {
			return err
		}
		related, ok := nestedResultsMap[joinOnValue]
		if !ok || related == nil {
			continue
		}
		result.SetRelationshipField(relationship.Id, related)
	}
	return nil
}

func attachNestedResultsForOneToManyQuery(relationship *model.Relationship, parentResults, nestedResults []model.TableModel) error {
	parentResultsMap := map[string]model.TableModel{}
	for _, parentResult := range parentResults {
		joinOnValue, err := parentResult.GetJoinOnValue(relationship.Id)
		if err != nil {
			return err
		}
		parentResultsMap[joinOnValue] = parentResult
	}

	for _, nestedResult := range nestedResults {
		joinOnValue, err := nestedResult.GetJoinOnValue(relationship.Id)
		if err != nil {
			return err
		}
		parent, ok := parentResultsMap[joinOnValue]
		if !ok || parent == nil {
			return errors.New("unable to join results")
		}
		parent.SetRelationshipField(relationship.Id, nestedResult)
	}
	return nil
}
