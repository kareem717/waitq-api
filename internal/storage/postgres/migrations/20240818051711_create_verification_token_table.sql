-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION pg_cron;

CREATE TABLE
    verification_tokens (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
        token BIGINT UNIQUE NOT NULL,
        created_at timestamptz NOT NULL DEFAULT CLOCK_TIMESTAMP(),
        expires_at timestamptz NOT NULL
    );

SELECT
    cron.schedule (
        'verification_token_cleanup',
        '0 0 * * *',
        $$DELETE FROM verification_tokens WHERE expires_at < NOW()$$
    );

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
SELECT
    cron.unschedule ('verification_token_cleanup');

DROP TABLE verification_tokens;

DROP EXTENSION pg_cron;

-- +goose StatementEnd