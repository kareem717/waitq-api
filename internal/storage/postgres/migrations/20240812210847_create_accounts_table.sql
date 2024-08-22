-- +goose Up
-- +goose StatementBegin
CREATE TYPE parsed_email AS (
    "domain" VARCHAR(255),
    "local" VARCHAR(64),
    "tld" TEXT,
    "host" TEXT
);

CREATE FUNCTION parse_email (email VARCHAR(320)) RETURNS parsed_email AS $$
DECLARE
    domain VARCHAR(255);
    local VARCHAR(64);
    tld TEXT;
    host TEXT;
BEGIN
    -- Example parsing logic
    SELECT split_part(email, '@', 2) INTO domain;
    SELECT split_part(email, '@', 1) INTO local;
    SELECT split_part(split_part(email, '@', 2), '.', 2) INTO tld;
    SELECT split_part(split_part(email, '@', 2), '.', 1) INTO host;

    RETURN (domain, local, tld, host);
END;
$$ LANGUAGE plpgsql IMMUTABLE;

CREATE TABLE
    accounts (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
        user_id UUID NOT NULL REFERENCES auth.users (id) ON DELETE RESTRICT ON UPDATE CASCADE,
        NAME TEXT NOT NULL,
        stripe_customer_id TEXT NOT NULL,
        email VARCHAR(320) NOT NULL,
        parsed_email parsed_email GENERATED ALWAYS AS (parse_email (email)) STORED NOT NULL,
        created_at timestamptz NOT NULL DEFAULT CLOCK_TIMESTAMP(),
        updated_at timestamptz,
        deleted_at timestamptz
    );

CREATE UNIQUE INDEX unique_user_id_not_deleted ON accounts (user_id)
WHERE
    deleted_at IS NULL;

CREATE TRIGGER sync_account_updated_at BEFORE
UPDATE ON accounts FOR EACH ROW
EXECUTE FUNCTION sync_updated_at_column ();

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TRIGGER sync_account_updated_at ON accounts;

DROP INDEX IF EXISTS unique_user_id_not_deleted;

DROP TABLE accounts;

DROP FUNCTION parse_email;

DROP TYPE parsed_email;

-- +goose StatementEnd