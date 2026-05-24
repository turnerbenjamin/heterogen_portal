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

	if err = initAdminTable(db); err != nil {
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

type User struct {
	ID           string       `json:"id"`
	OID          string       `json:"oid"`
	GivenName    string       `json:"given_name"`
	FamilyName   string       `json:"family_name"`
	UserName     string       `json:"user_name"`
	EmailAddress string       `json:"email_address"`
	CreatedAt    sql.NullTime `json:"created_at"`
	UpdatedAt    sql.NullTime `json:"updated_at"`
}

const (
	DB_CONSTRAINT_GIVEN_NAME_MAX  = 64
	DB_CONSTRAINT_FAMILY_NAME_MAX = 64
	DB_CONSTRAINT_USER_NAME_MAX   = 128
	DB_CONSTRAINT_EMAIL_MAX       = 320
)

func initAdminTable(db *sql.DB) error {
	query := `
IF NOT EXISTS (SELECT 1 FROM sys.schemas WHERE name = 'hg')
BEGIN
	EXEC('CREATE SCHEMA hg');
END;

IF NOT EXISTS (
	SELECT 1 FROM sys.tables t
	JOIN sys.schemas s ON t.schema_id = s.schema_id
	WHERE t.name = 'users' AND s.name = 'hg'
)
BEGIN
	CREATE TABLE hg.users (
		id NVARCHAR(36) NOT NULL PRIMARY KEY,
		oid NVARCHAR(36) NOT NULL UNIQUE,
		given_name NVARCHAR(64) NOT NULL CHECK (LEN(given_name) > 0 AND LEN(given_name) <= 64),
		family_name NVARCHAR(64) NOT NULL CHECK (LEN(family_name) > 0 AND LEN(family_name) <= 64),
		user_name NVARCHAR(128) NOT NULL CHECK (LEN(user_name) > 0 AND LEN(user_name) <= 64),
		email_address NVARCHAR(320) NOT NULL CHECK (LEN(email_address) > 0 AND LEN(email_address) <= 320),
		created_at DATETIME2 NOT NULL CONSTRAINT DF_hg_users_created_at DEFAULT SYSUTCDATETIME(),
		updated_at DATETIME2 NOT NULL CONSTRAINT DF_hg_users_updated_at DEFAULT SYSUTCDATETIME()
	);
END;`

	_, err := db.Exec(query)
	return err
}
