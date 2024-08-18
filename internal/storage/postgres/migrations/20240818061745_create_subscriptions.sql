-- +goose Up
-- +goose StatementBegin
CREATE TABLE
    subscriptions (
        id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
        NAME TEXT NOT NULL,
        stripe_product_id TEXT NOT NULL,
        stripe_price_id TEXT NOT NULL,
        created_at timestamptz NOT NULL DEFAULT CLOCK_TIMESTAMP(),
        updated_at timestamptz,
        deleted_at timestamptz
    );

CREATE TRIGGER sync_subscription_updated_at BEFORE
UPDATE ON subscriptions FOR EACH ROW
EXECUTE FUNCTION sync_updated_at_column ();

ALTER TABLE accounts
ADD COLUMN subscription_id UUID REFERENCES subscriptions (id) ON DELETE SET NULL ON UPDATE CASCADE;

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
ALTER TABLE accounts
DROP COLUMN subscription_id;

DROP TABLE subscriptions;

-- +goose StatementEnd