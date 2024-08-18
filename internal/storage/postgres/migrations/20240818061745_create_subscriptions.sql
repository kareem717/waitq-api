-- +goose Up
-- +goose StatementBegin
CREATE TABLE
    subscriptions (
        id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
        NAME TEXT NOT NULL,
        stripe_product_id TEXT NOT NULL,
        created_at timestamptz NOT NULL DEFAULT CLOCK_TIMESTAMP(),
        updated_at timestamptz,
        deleted_at timestamptz
    );

CREATE TRIGGER sync_subscription_updated_at BEFORE
UPDATE ON subscriptions FOR EACH ROW
EXECUTE FUNCTION sync_updated_at_column ();

CREATE TABLE
    account_subscriptions (
        account_id UUID NOT NULL PRIMARY KEY REFERENCES accounts (id) ON DELETE CASCADE ON UPDATE CASCADE,
        subscription_id UUID NOT NULL REFERENCES subscriptions (id) ON DELETE CASCADE ON UPDATE CASCADE,
        stripe_customer_id TEXT NOT NULL,
        stripe_subscription_id TEXT NOT NULL,
        created_at timestamptz NOT NULL DEFAULT CLOCK_TIMESTAMP(),
        updated_at timestamptz,
        deleted_at timestamptz
    );

CREATE TRIGGER sync_account_subscription_updated_at BEFORE
UPDATE ON account_subscriptions FOR EACH ROW
EXECUTE FUNCTION sync_updated_at_column ();

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE account_subscriptions;

DROP TABLE subscriptions;

-- +goose StatementEnd