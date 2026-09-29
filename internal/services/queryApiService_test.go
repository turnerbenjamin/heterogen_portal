package services

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/turnerbenjamin/heterogen_portal/internal/accesspolicies"
	"github.com/turnerbenjamin/querystack"
	"github.com/turnerbenjamin/querystack/querymodel"
)

func TestNewQueryApiService_ShouldConfigureTheExecutorFactoryCorrectly(t *testing.T) {
	repo := NewMockRepository(t)
	paginationTokenSigner := NewMockPayloadSigner(t)
	paginationTokenSecret := []byte("super-secret")
	f := &mockQueryExecutorFactory{
		t:             t,
		queryExecutor: NewMockQueryExecutor(t),
	}

	_, err := NewQueryApiService(
		repo,
		f.builder,
		paginationTokenSigner,
		paginationTokenSecret,
	)

	require.NoError(t, err)

	assert.Equal(t, repo, f.repository)
	assert.Equal(t, schema, f.schema)
	assert.Equal(t, paginationTokenSigner, f.paginationtokenSigner)
	assert.Equal(t, paginationTokenSecret, f.paginationTokenSecret)
	assert.Equal(t, querymodel.SqlFlavorAzureSql, f.sqlFlavor)
	assert.Equal(t, uint32(100), f.queryConfig.DefaultPageSize)
	assert.Equal(t, uint32(100_000), f.queryConfig.MaxRecordsPerPage)
	assert.Equal(t, uint8(10), f.queryConfig.MaxDepth)
}

func TestNewQueryApiService_ShouldReturnErrorsReturnedFromTheQueryExecutorFactory(t *testing.T) {
	repo := NewMockRepository(t)
	paginationTokenSigner := NewMockPayloadSigner(t)
	paginationTokenSecret := []byte("super-secret")

	expectedErr := errors.New("query executor factory err")

	f := &mockQueryExecutorFactory{
		t:             t,
		queryExecutor: NewMockQueryExecutor(t),
		builderErr:    expectedErr,
	}

	_, err := NewQueryApiService(
		repo,
		f.builder,
		paginationTokenSigner,
		paginationTokenSecret,
	)

	require.EqualError(t, err, expectedErr.Error())
}

func TestExecuteQuery_ShouldCallQueryExecutorWithTheAnonymousAccessPolicyAndReturnTheResults(t *testing.T) {
	repo := NewMockRepository(t)
	paginationTokenSigner := NewMockPayloadSigner(t)
	paginationTokenSecret := []byte("super-secret")

	qe := NewMockQueryExecutor(t)
	f := &mockQueryExecutorFactory{
		t:             t,
		queryExecutor: qe,
	}

	ctx := context.TODO()
	resource := "test-resource"
	queryString := "test-query-string"

	expectedResults := &querymodel.ExecuteResult{}
	qe.
		EXPECT().
		Execute(
			ctx,
			resource,
			accesspolicies.AnonymousAccessPolicy,
			queryString,
		).
		Return(expectedResults, nil)

	s, err := NewQueryApiService(
		repo,
		f.builder,
		paginationTokenSigner,
		paginationTokenSecret,
	)
	require.NoError(t, err)

	r, err := s.ExecuteQuery(ctx, resource, queryString)

	require.NoError(t, err)
	assert.Equal(t, expectedResults, r)
}

func TestExecuteQuery_ShouldReturnErrorsFromTheQueryExecutor(t *testing.T) {
	repo := NewMockRepository(t)
	paginationTokenSigner := NewMockPayloadSigner(t)
	paginationTokenSecret := []byte("super-secret")

	qe := NewMockQueryExecutor(t)
	f := &mockQueryExecutorFactory{
		t:             t,
		queryExecutor: qe,
	}

	ctx := context.TODO()
	resource := "test-resource"
	queryString := "test-query-string"

	expectedErr := errors.New("query executor error")
	qe.
		EXPECT().
		Execute(
			ctx,
			resource,
			accesspolicies.AnonymousAccessPolicy,
			queryString,
		).
		Return(nil, expectedErr)

	s, err := NewQueryApiService(
		repo,
		f.builder,
		paginationTokenSigner,
		paginationTokenSecret,
	)
	require.NoError(t, err)

	r, err := s.ExecuteQuery(ctx, resource, queryString)

	require.Nil(t, r)
	assert.EqualError(t, err, expectedErr.Error())
}

type mockQueryExecutorFactory struct {
	t             testing.TB
	queryExecutor querystack.QueryExecutor

	builderErr error

	repository            querymodel.Repository
	schema                querymodel.Schema
	paginationtokenSigner querymodel.PayloadSigner
	paginationTokenSecret []byte
	sqlFlavor             querymodel.SqlFlavour
	queryConfig           querymodel.QueryConfig
}

func (f *mockQueryExecutorFactory) builder(
	repository querymodel.Repository,
	schema querymodel.Schema,
	paginationTokenSigner querymodel.PayloadSigner,
	paginationTokenSecret []byte,
	sqlFlavor querymodel.SqlFlavour,
	queryConfig querymodel.QueryConfig,
) (querystack.QueryExecutor, error) {
	f.t.Helper()

	if repository != nil {
		f.repository = repository
	}

	if schema != nil {
		f.schema = schema
	}

	if paginationTokenSigner != nil {
		f.paginationtokenSigner = paginationTokenSigner
	}

	f.paginationTokenSecret = paginationTokenSecret
	f.sqlFlavor = sqlFlavor
	f.queryConfig = queryConfig

	if f.builderErr != nil {
		return nil, f.builderErr
	}

	return f.queryExecutor, nil
}
