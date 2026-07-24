CREATE TABLE todoapp.payments (
    id UUID PRIMARY KEY,
    account_user_id INTEGER NOT NULL REFERENCES todoapp.accounts(user_id) ON DELETE CASCADE,
    idempotency_key VARCHAR(64) NOT NULL,
    provider VARCHAR(20) NOT NULL DEFAULT 'yookassa',
    provider_payment_id VARCHAR(100) UNIQUE,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    amount_minor BIGINT NOT NULL,
    currency VARCHAR(3) NOT NULL,
    confirmation_url TEXT,
    provider_attached_at TIMESTAMPTZ,
    credited_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    UNIQUE (account_user_id, idempotency_key),
    CHECK (provider = 'yookassa'),
    CHECK (status IN ('pending', 'succeeded', 'canceled')),
    CHECK (amount_minor > 0),
    CHECK (currency = 'RUB'),
    CHECK (
        (status = 'succeeded' AND credited_at IS NOT NULL)
        OR
        (status <> 'succeeded' AND credited_at IS NULL)
    )
);

CREATE INDEX payments_account_created_idx
    ON todoapp.payments (account_user_id, created_at DESC);
