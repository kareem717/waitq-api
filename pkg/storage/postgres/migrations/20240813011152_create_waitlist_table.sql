-- +goose Up
-- +goose StatementBegin
CREATE TABLE
    waitlists (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
        account_id UUID NOT NULL,
        NAME TEXT NOT NULL,
        created_at timestamptz NOT NULL DEFAULT CLOCK_TIMESTAMP(),
        updated_at timestamptz,
        deleted_at timestamptz
    );

CREATE TRIGGER sync_waitlist_updated_at BEFORE
UPDATE ON waitlists FOR EACH ROW
EXECUTE FUNCTION sync_updated_at_column ();

CREATE TABLE
    waitlist_emails (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
        waitlist_id UUID NOT NULL,
        email VARCHAR(360) NOT NULL,
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
DROP TABLE IF EXISTS waitlist_emails;

DROP TABLE IF EXISTS waitlists;

-- +goose StatementEnd