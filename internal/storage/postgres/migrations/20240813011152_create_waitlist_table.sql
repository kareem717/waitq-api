-- +goose Up
-- +goose StatementBegin
CREATE TABLE
    waitlists (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
        account_id UUID NOT NULL,
        NAME TEXT NOT NULL,
        anon_key TEXT NOT NULL UNIQUE,
        service_key TEXT NOT NULL UNIQUE,
        jwt_secret VARCHAR(512) NOT NULL CHECK (LENGTH(jwt_secret)>31),
        created_at timestamptz NOT NULL DEFAULT CLOCK_TIMESTAMP(),
        updated_at timestamptz,
        deleted_at timestamptz
    );

CREATE TRIGGER sync_waitlist_updated_at BEFORE
UPDATE ON waitlists FOR EACH ROW
EXECUTE FUNCTION sync_updated_at_column ();

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE waitlists;

-- +goose StatementEnd