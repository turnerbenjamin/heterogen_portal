-- +goose Up
-- *********

-- +goose StatementBegin
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
        user_name NVARCHAR(128) NOT NULL CHECK (LEN(user_name) > 0 AND LEN(user_name) <= 128),
        email_address NVARCHAR(320) NOT NULL CHECK (LEN(email_address) > 0 AND LEN(email_address) <= 320),
        created_at DATETIME2 NOT NULL CONSTRAINT DF_hg_users_created_at DEFAULT SYSUTCDATETIME(),
        updated_at DATETIME2 NOT NULL CONSTRAINT DF_hg_users_updated_at DEFAULT SYSUTCDATETIME()
    );
END;
-- +goose StatementEnd



-- +goose Down
-- ***********

-- +goose StatementBegin
IF OBJECT_ID('hg.users', 'U') IS NOT NULL
BEGIN
    DROP TABLE hg.users;
END;

IF EXISTS (SELECT 1 FROM sys.schemas WHERE name = 'hg')
BEGIN
    EXEC('DROP SCHEMA hg');
END;
-- +goose StatementEnd