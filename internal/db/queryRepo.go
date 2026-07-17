package db

import (
	"bytes"
	"context"
	"database/sql"
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

func (r *QueryRepo) Execute(ctx context.Context, queryString string) ([]byte, error) {
	rows, err := r.db.QueryContext(ctx, queryString)
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
