-- +goose Up
-- +goose StatementBegin
INSERT INTO
    subscriptions (
        NAME,
        stripe_product_id,
        permission_tier,
        max_people_per_waitlist,
        max_waitlists
    )
VALUES
    ('Basic', 'prod_QgjRVWYFhoY4yL', 1, 800, 2),
    ('Pro', 'prod_QgPI36BtTgwz5s', 2, 2500, 4),
    ('Super', 'prod_QgPJXG1HfYXzWQ', 3, 500000, 85);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DELETE FROM subscriptions
WHERE
    NAME IN ('Basic', 'Pro', 'Super');

-- +goose StatementEnd