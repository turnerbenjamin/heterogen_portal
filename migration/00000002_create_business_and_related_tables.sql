-- +goose Up

EXEC('CREATE SCHEMA hg_enum');

CREATE TABLE hg_enum.businesses_business_type (
    enum_value TINYINT NOT NULL,
    display_name NVARCHAR(255) NOT NULL,

    CONSTRAINT PK_hg_enum_businesses_business_type
        PRIMARY KEY (enum_value),

    CONSTRAINT UQ_hg_enum_businesses_business_type_display_name
        UNIQUE (display_name)
);

INSERT INTO hg_enum.businesses_business_type (
    enum_value,
    display_name
)
VALUES
    (1, N'Merchant'),
    (2, N'Grower');

CREATE TABLE hg.businesses (
    id NVARCHAR(36) NOT NULL PRIMARY KEY DEFAULT (NEWID()),

    reference NVARCHAR(8) NOT NULL UNIQUE,
    trading_name NVARCHAR(255) NOT NULL UNIQUE,

    logo_url NVARCHAR(2048) NULL,

    description NVARCHAR(4000) NULL,

    business_type TINYINT NOT NULL,

    cph_number NVARCHAR(11) NULL,
    email_address NVARCHAR(320) NULL,
    contact_number NVARCHAR(50) NULL,
    website_url NVARCHAR(2048) NULL,

    address_line_1 NVARCHAR(255) NOT NULL,
    address_line_2 NVARCHAR(255) NULL,
    town NVARCHAR(100) NULL,
    county NVARCHAR(100) NULL,
    country NVARCHAR(100) NULL,
    postcode NVARCHAR(20) NOT NULL,

    location GEOGRAPHY NOT NULL
        CONSTRAINT CK_Businesses_Location_IsPoint
        CHECK (Location.STGeometryType() = 'Point'),

    created_at DATETIMEOFFSET(7) NOT NULL,
    created_by_id NVARCHAR(36) NOT NULL,
    modified_at DATETIMEOFFSET(7) NOT NULL,
    modified_by_id NVARCHAR(36) NOT NULL,

    CONSTRAINT FK_business_business_type
        FOREIGN KEY (business_type)
        REFERENCES hg_enum.businesses_business_type(enum_value),

    CONSTRAINT FK_business_created_by_id
        FOREIGN KEY (created_by_id)
        REFERENCES hg.users(id),

    CONSTRAINT FK_business_modified_by_id
        FOREIGN KEY (modified_by_id)
        REFERENCES hg.users(id)
);

EXEC sys.sp_addextendedproperty
    @name = N'Enum',
    @value = N'businesses_business_type',
    @level0type = N'SCHEMA',
    @level0name = N'hg',
    @level1type = N'TABLE',
    @level1name = N'businesses',
    @level2type = N'COLUMN',
    @level2name = N'business_type';

CREATE TABLE hg.farm_fields (
    id NVARCHAR(36) NOT NULL PRIMARY KEY DEFAULT (NEWID()),

    reference NVARCHAR(255) NOT NULL,

    business_id NVARCHAR(36) NOT NULL,

    location GEOGRAPHY NOT NULL
        CONSTRAINT CK_FarmFields_Location_IsPoint
        CHECK (Location.STGeometryType() = 'Point'),

    created_at DATETIMEOFFSET(7) NOT NULL,
    created_by_id NVARCHAR(36) NOT NULL,
    modified_at DATETIMEOFFSET(7) NOT NULL,
    modified_by_id NVARCHAR(36) NOT NULL,

    CONSTRAINT FK_field_business
        FOREIGN KEY (business_id)
        REFERENCES hg.businesses(id)
        ON DELETE CASCADE,

    CONSTRAINT FK_field_created_by_id
        FOREIGN KEY (created_by_id)
        REFERENCES hg.users(id),

    CONSTRAINT FK_field_modified_by_id
        FOREIGN KEY (modified_by_id)
        REFERENCES hg.users(id)
);

CREATE INDEX IX_businesses_trading_name
    ON hg.businesses(trading_name);

CREATE INDEX IX_farm_fields_business
    ON hg.farm_fields(business_id);

CREATE SPATIAL INDEX SIX_businesses_location
ON hg.businesses(location);

CREATE SPATIAL INDEX SIX_farm_fields_location
ON hg.farm_fields(location);

-- +goose Down

DROP TABLE IF EXISTS hg.farm_fields;
DROP TABLE IF EXISTS hg.businesses;
DROP TABLE IF EXISTS hg_enum.businesses_business_type;
EXEC('DROP SCHEMA hg_enum');