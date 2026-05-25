package db

import (
	"database/sql"

	_ "github.com/microsoft/go-mssqldb/azuread"
)

func SetUpDB(dsn string) (*sql.DB, error) {

	db, err := connectToDb(dsn)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func connectToDb(dsn string) (*sql.DB, error) {
	db, err := sql.Open("azuresql", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}
