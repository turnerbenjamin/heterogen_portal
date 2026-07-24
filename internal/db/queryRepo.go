package db

import (
	"bytes"
	"context"
	"database/sql"
)

/* TEST QUERIES
http://localhost:8080/api/v0.1/businesses?select=trading_name,location

http://localhost:8080/api/v0.1/businesses?select=trading_name&filter=farm_fields_businesses_business_id/any(reference%20startswith%20%27t%27) and created_by_id/full_name contains 'bot'&expand=farm_fields_businesses_business_id(select=reference)

*/

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

func (r *QueryRepo) Execute(ctx context.Context, statementStr string, args []any) ([]byte, error) {
	stmt, err := r.db.Prepare(statementStr)
	if err != nil {
		return nil, err
	}

	rows, err := stmt.QueryContext(ctx, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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

	// 3. Fallback to an empty JSON array if no data was returned
	if buf.Len() == 0 {
		return []byte("[]"), nil
	}

	// 4. Return the underlying raw byte slice safely
	return buf.Bytes(), nil
}
