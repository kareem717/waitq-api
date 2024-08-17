-- +goose Up
-- +goose StatementBegin
CREATE TYPE parsed_email AS (
    DOMAIN TEXT,
    local_part TEXT,
    tld TEXT,
    HOST TEXT,
    plain_address TEXT
);

CREATE FUNCTION parse_email (email TEXT) RETURNS parsed_email AS $$
DECLARE
    domain TEXT;
    local_part TEXT;
    tld TEXT;
    host TEXT;
    plain_address TEXT;
BEGIN
    -- Example parsing logic
    SELECT split_part(email, '@', 2) INTO domain;
    SELECT split_part(email, '@', 1) INTO local_part;
    SELECT split_part(split_part(email, '@', 2), '.', 2) INTO tld;
    SELECT split_part(split_part(email, '@', 2), '.', 1) INTO host;
    plain_address := email;

    RETURN (domain, local_part, tld, host, plain_address);
END;
$$ LANGUAGE plpgsql IMMUTABLE;

CREATE TABLE
    waitlist_emails (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
        waitlist_id UUID NOT NULL,
        email VARCHAR(320) NOT NULL,
        parsed_email parsed_email GENERATED ALWAYS AS (parse_email (email)) STORED NOT NULL,
        unsubscribed_at timestamptz,
        created_at timestamptz NOT NULL DEFAULT CLOCK_TIMESTAMP(),
        updated_at timestamptz,
        deleted_at timestamptz
    );

CREATE UNIQUE INDEX unique_waitlist_id_email_not_deleted_or_unsubscribed ON waitlist_emails (waitlist_id, email)
WHERE
    deleted_at IS NULL
    AND unsubscribed_at IS NULL;

CREATE TRIGGER sync_waitlist_emails_updated_at BEFORE
UPDATE ON waitlist_emails FOR EACH ROW
EXECUTE FUNCTION sync_updated_at_column ();

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE waitlist_emails;

DROP FUNCTION parse_email;

DROP TYPE parsed_email;

-- +goose StatementEnd