-- +goose Up

-- +goose StatementBegin
EXEC('CREATE SCHEMA hg');

CREATE TABLE hg.users (
    id UNIQUEIDENTIFIER NOT NULL PRIMARY KEY DEFAULT (NEWID()),
    oid UNIQUEIDENTIFIER NOT NULL UNIQUE DEFAULT (NEWID()),
    given_name NVARCHAR(64) NOT NULL CHECK (LEN(given_name) > 0 AND LEN(given_name) <= 64),
    family_name NVARCHAR(64) NOT NULL CHECK (LEN(family_name) > 0 AND LEN(family_name) <= 64),
    display_name NVARCHAR(128) NOT NULL UNIQUE CHECK (LEN(display_name) > 0 AND LEN(display_name) <= 128),
    email_address NVARCHAR(320) NOT NULL CHECK (LEN(email_address) > 0 AND LEN(email_address) <= 320),
    created_at DATETIMEOFFSET(7) NOT NULL CONSTRAINT DF_hg_users_created_at DEFAULT SYSUTCDATETIME(),
    modified_at DATETIMEOFFSET(7) NOT NULL CONSTRAINT DF_hg_users_modified_at DEFAULT SYSUTCDATETIME()
);
CREATE INDEX IX_user_oid
    ON hg.users(oid);

CREATE INDEX IX_user_display_name
    ON hg.users(display_name);
-- +goose StatementEnd



-- +goose Down

-- +goose StatementBegin
DROP TABLE hg.users;
EXEC('DROP SCHEMA hg');
-- +goose StatementEnd