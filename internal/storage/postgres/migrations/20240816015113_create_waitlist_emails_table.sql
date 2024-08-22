-- +goose Up
-- +goose StatementBegin
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

-- +goose StatementEnd