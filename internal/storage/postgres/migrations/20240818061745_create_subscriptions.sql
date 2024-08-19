-- +goose Up
-- +goose StatementBegin
CREATE TABLE
    subscriptions (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
        NAME TEXT NOT NULL UNIQUE,
        stripe_product_id TEXT NOT NULL UNIQUE,
        permission_tier INT NOT NULL CHECK (permission_tier>0) UNIQUE,
        max_people_per_waitlist INT NOT NULL,
        max_waitlists INT NOT NULL,
        created_at timestamptz NOT NULL DEFAULT CLOCK_TIMESTAMP(),
        updated_at timestamptz,
        deleted_at timestamptz
    );

CREATE TRIGGER sync_subscription_updated_at BEFORE
UPDATE ON subscriptions FOR EACH ROW
EXECUTE FUNCTION sync_updated_at_column ();

CREATE TABLE
    account_subscriptions (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
        account_id UUID NOT NULL REFERENCES accounts (id) ON DELETE CASCADE ON UPDATE CASCADE,
        subscription_id UUID NOT NULL REFERENCES subscriptions (id) ON DELETE CASCADE ON UPDATE CASCADE,
        stripe_customer_id TEXT NOT NULL,
        stripe_subscription_id TEXT NOT NULL,
        stripe_price_id TEXT NOT NULL,
        created_at timestamptz NOT NULL DEFAULT CLOCK_TIMESTAMP(),
        updated_at timestamptz,
        deleted_at timestamptz
    );

CREATE UNIQUE INDEX account_subscriptions_account_id_idx ON account_subscriptions (account_id)
WHERE
    deleted_at IS NULL;

CREATE TRIGGER sync_account_subscription_updated_at BEFORE
UPDATE ON account_subscriptions FOR EACH ROW
EXECUTE FUNCTION sync_updated_at_column ();

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE account_subscriptions;

DROP TABLE subscriptions;

-- +goose StatementEnd