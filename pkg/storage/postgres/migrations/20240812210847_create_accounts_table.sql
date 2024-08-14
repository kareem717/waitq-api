-- +goose Up
-- +goose StatementBegin
CREATE TABLE
    accounts (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
        user_id UUID NOT NULL REFERENCES auth.users (id) ON DELETE RESTRICT ON UPDATE CASCADE,
        username TEXT NOT NULL,
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

-- +goose StatementEnd