package db

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
)

type QueryRepo struct {
	ctx context.Context
	db  *sql.DB
}

func BuildQueryRepo(ctx context.Context, db *sql.DB) *QueryRepo {
	return &QueryRepo{
		ctx: ctx,
		db:  db,
	}
}

func (r *QueryRepo) ExecuteJsonRequest(
	ctx context.Context,
	queryStatementStr string,
	args []any,
) ([]byte, error) {
	stmt, err := r.db.Prepare(queryStatementStr)
	if err != nil {
		return nil, err
	}

	rows, err := stmt.QueryContext(ctx, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return parseJson(rows)
}

func (r *QueryRepo) ExecuteJsonRequestWithCount(
	ctx context.Context,
	queryStatementStr string,
	countStatementStr string,
	sharedArgs []any,
) ([]byte, *int64, error) {
	combinedStatements := fmt.Sprintf("%s %s", queryStatementStr, countStatementStr)
	stmt, err := r.db.Prepare(combinedStatements)
	if err != nil {
		return nil, nil, err
	}

	rows, err := stmt.QueryContext(ctx, sharedArgs...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	json, err := parseJson(rows)
	if err != nil {
		return nil, nil, err
	}

	if !rows.NextResultSet() {
		return nil, nil, fmt.Errorf("unable to a access count statement results")
	}

	count, err := parseCount(rows)
	if err != nil {
		return nil, nil, err
	}

	return json, count, err
}

func parseJson(rows *sql.Rows) ([]byte, error) {
	var buf bytes.Buffer
	var rowString string

	for rows.Next() {
		if err := rows.Scan(&rowString); err != nil {
			return nil, err
		}
		// 2. Write the string directly into the byte buffer
		_, err := buf.WriteString(rowString)
		if err != nil {
			return nil, err
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func parseCount(rows *sql.Rows) (*int64, error) {
	if !rows.Next() {
		return nil, fmt.Errorf("unable to access count result from rows")
	}

	var totalCount int64
	err := rows.Scan(&totalCount)
	if err != nil {
		return nil, err
	}
	return &totalCount, nil
}
