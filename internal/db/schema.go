package db

import "database/sql"

const (
	DB_CONSTRAINT_GIVEN_NAME_MAX  = 64
	DB_CONSTRAINT_FAMILY_NAME_MAX = 64
	DB_CONSTRAINT_USER_NAME_MAX   = 128
	DB_CONSTRAINT_EMAIL_MAX       = 320
)

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
