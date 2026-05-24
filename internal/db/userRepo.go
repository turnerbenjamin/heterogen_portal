package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"

	"github.com/google/uuid"
)

type Crypt interface {
	GenerateFromPassword(password []byte, cost int) ([]byte, error)
	CompareHashAndPassword(hashed, password []byte) error
}

type UserRepo interface {
	Close()
	UpsertUser(context context.Context, oid, givenName, familyName, userName, emailAddress string) (*User, error)
	RetrieveUserById(id string) (*User, error)
	RetrieveUserByOid(id string) (*User, error)
}

type userRepo struct {
	db         *sql.DB
	statements map[statementKey]*sql.Stmt
	crypt      Crypt
}

type statementKey = int

const (
	STMT_KEY_UPSERT_USER statementKey = iota
	STMT_KEY_RETRIEVE_USER_BY_ID
	STMT_KEY_RETRIEVE_USER_BY_OID
)

const (
	SQL_SERVER_ERR_UNIQUE_CONSTRAINT = 2627
	SQL_SERVER_ERR_UNIQUE_INDEX      = 2601
)

var (
	ErrUserNotFound     = errors.New("user not found")
	ErrGivenNameEmpty   = errors.New("given name cannot be empty")
	ErrFamilyNameEmpty  = errors.New("family name cannot be empty")
	ErrUserNameEmpty    = errors.New("user name cannot be empty")
	ErrEmailEmpty       = errors.New("email cannot be empty")
	ErrGivenNameTooLong = fmt.Errorf(
		"given cannot exceed %d chars",
		DB_CONSTRAINT_GIVEN_NAME_MAX,
	)
	ErrFamilyNameTooLong = fmt.Errorf(
		"family name cannot exceed %d chars",
		DB_CONSTRAINT_GIVEN_NAME_MAX,
	)
	ErrUserNameTooLong = fmt.Errorf(
		"user name cannot exceed %d chars",
		DB_CONSTRAINT_USER_NAME_MAX,
	)
	ErrEmailTooLong = fmt.Errorf(
		"email cannot exceed %d chars",
		DB_CONSTRAINT_GIVEN_NAME_MAX,
	)
)

func BuildUserRepo(db *sql.DB, crypt Crypt) UserRepo {
	ur := userRepo{
		db:         db,
		crypt:      crypt,
		statements: make(map[statementKey]*sql.Stmt),
	}
	return &ur
}

func (r *userRepo) Close() {
	for _, stmt := range r.statements {
		err := stmt.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error closing prepared statement: %v", err)
		}
	}
}

func (r *userRepo) UpsertUser(ctx context.Context, oid, givenName, familyName, userName, emailAddress string) (*User, error) {

	var err error
	id := uuid.New().String()

	query, ok := r.statements[STMT_KEY_UPSERT_USER]
	if !ok {
		query, err = r.db.Prepare(
			`
			BEGIN TRAN;
			BEGIN TRY
			DECLARE @out TABLE (
				id NVARCHAR(36),
				oid NVARCHAR(36),
				given_name NVARCHAR(64),
				family_name NVARCHAR(64),
				user_name NVARCHAR(128),
				email_address NVARCHAR(320),
				created_at DATETIME2,
				updated_at DATETIME2
			);

			-- try update
			UPDATE hg.users
			SET 
				given_name = @givenName, 
				family_name = @familyName, 
				user_name = @userName, 
				email_address = @emailAddress, 
				updated_at = SYSUTCDATETIME()
			OUTPUT 
				inserted.id, 
				inserted.oid, 
				inserted.given_name, 
				inserted.family_name, 
				inserted.user_name, 
				inserted.email_address, 
				inserted.created_at, 
				inserted.updated_at INTO @out
			WHERE oid = @oid;

			IF @@ROWCOUNT = 0
			BEGIN
			-- insert if not found
			INSERT INTO hg.users (
				id, 
				oid, 
				given_name, 
				family_name, 
				user_name, 
				email_address, 
				created_at, 
				updated_at
			)
			OUTPUT 
				inserted.id, 
				inserted.oid, 
				inserted.given_name, 
				inserted.family_name, 
				inserted.user_name, 
				inserted.email_address, 
				inserted.created_at, 
				inserted.updated_at INTO @out
			VALUES (
				@id, 
				@oid, 
				@givenName, 
				@familyName, 
				@userName, 
				@emailAddress, 
				SYSUTCDATETIME(), 
				SYSUTCDATETIME());
			END

			SELECT id, oid, given_name, family_name, user_name, email_address, created_at, updated_at FROM @out;

			COMMIT TRAN;
			END TRY
			BEGIN CATCH
				IF @@TRANCOUNT > 0
					ROLLBACK TRAN;
				THROW;
			END CATCH;
			`,
		)
		if err != nil {
			return nil, err
		}
		r.statements[STMT_KEY_UPSERT_USER] = query
	}

	err = r.preCheckConstraintViolations(givenName, familyName, userName, emailAddress)
	if err != nil {
		return nil, err
	}

	row := query.QueryRowContext(ctx,
		sql.Named("id", id),
		sql.Named("oid", oid),
		sql.Named("givenName", givenName),
		sql.Named("familyName", familyName),
		sql.Named("userName", userName),
		sql.Named("emailAddress", emailAddress),
	)
	return parseUserFromQueryResponse(row)
}

func (r *userRepo) RetrieveUserById(id string) (*User, error) {
	return r.retrieveUser(STMT_KEY_RETRIEVE_USER_BY_ID, id)
}

func (r *userRepo) RetrieveUserByOid(oid string) (*User, error) {
	return r.retrieveUser(STMT_KEY_RETRIEVE_USER_BY_OID, oid)
}

func (r *userRepo) retrieveUser(
	statementKey statementKey,
	identifier string,
) (*User, error) {
	if statementKey != STMT_KEY_RETRIEVE_USER_BY_ID &&
		statementKey != STMT_KEY_RETRIEVE_USER_BY_OID {
		return nil, errors.New("unsupported statement key")
	}

	_, ok := r.statements[STMT_KEY_RETRIEVE_USER_BY_ID]
	if !ok {
		query, err := (*r).db.Prepare(
			`SELECT id, oid, given_name, family_name, user_name, email_address, created_at, updated_at
			 FROM hg.users WHERE id = @identifier`,
		)
		if err != nil {
			return nil, err
		}
		r.statements[STMT_KEY_RETRIEVE_USER_BY_ID] = query
	}

	_, ok = r.statements[STMT_KEY_RETRIEVE_USER_BY_OID]
	if !ok {
		query, err := (*r).db.Prepare(
			`SELECT id, oid, given_name, family_name, user_name, email_address, created_at, updated_at
			 FROM hg.users WHERE oid = @identifier`,
		)
		if err != nil {
			return nil, err
		}
		r.statements[STMT_KEY_RETRIEVE_USER_BY_OID] = query
	}

	query := r.statements[statementKey]

	row := query.QueryRow(sql.Named("identifier", identifier))
	return parseUserFromQueryResponse(row)
}

func parseUserFromQueryResponse(res *sql.Row) (*User, error) {
	u := &User{}
	err := res.Scan(
		&u.ID,
		&u.OID,
		&u.GivenName,
		&u.FamilyName,
		&u.UserName,
		&u.EmailAddress,
		&u.CreatedAt,
		&u.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}

	return u, err
}

func (r *userRepo) preCheckConstraintViolations(givenName, familyName, username, email string) error {
	givenNameLen := len(givenName)
	if givenNameLen == 0 {
		return ErrGivenNameEmpty
	}
	if givenNameLen > DB_CONSTRAINT_GIVEN_NAME_MAX {
		return ErrGivenNameTooLong
	}

	familyNameLen := len(familyName)
	if familyNameLen == 0 {
		return ErrFamilyNameEmpty
	}
	if familyNameLen > DB_CONSTRAINT_FAMILY_NAME_MAX {
		return ErrFamilyNameTooLong
	}

	emailLen := len(email)
	if emailLen == 0 {
		return ErrEmailEmpty
	}
	if familyNameLen > DB_CONSTRAINT_EMAIL_MAX {
		return ErrEmailTooLong
	}

	return nil
}
