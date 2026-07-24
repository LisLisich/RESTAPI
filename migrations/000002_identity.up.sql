CREATE TABLE todoapp.accounts (
    user_id INTEGER PRIMARY KEY REFERENCES todoapp.users(id) ON DELETE CASCADE,
    version BIGINT NOT NULL DEFAULT 1,
    email VARCHAR(254) NOT NULL UNIQUE,
    status VARCHAR(32) NOT NULL,
    email_verified_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    CHECK (email = lower(trim(email))),
    CHECK (status IN ('pending_verification', 'active', 'disabled')),
    CHECK (
        (status = 'pending_verification' AND email_verified_at IS NULL)
        OR
        (status IN ('active', 'disabled') AND email_verified_at IS NOT NULL)
    )
);

CREATE TABLE todoapp.password_credentials (
    account_user_id INTEGER PRIMARY KEY REFERENCES todoapp.accounts(user_id) ON DELETE CASCADE,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE todoapp.account_tokens (
    id BIGSERIAL PRIMARY KEY,
    account_user_id INTEGER NOT NULL REFERENCES todoapp.accounts(user_id) ON DELETE CASCADE,
    purpose VARCHAR(32) NOT NULL,
    token_hash BYTEA NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    consumed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL,
    CHECK (purpose IN ('email_verification', 'password_reset')),
    CHECK (expires_at > created_at),
    CHECK (consumed_at IS NULL OR consumed_at >= created_at)
);

CREATE INDEX account_tokens_active_idx
    ON todoapp.account_tokens (account_user_id, purpose, expires_at)
    WHERE consumed_at IS NULL;

CREATE TABLE todoapp.outbox_events (
    id BIGSERIAL PRIMARY KEY,
    topic VARCHAR(100) NOT NULL,
    aggregate_type VARCHAR(50) NOT NULL,
    aggregate_id VARCHAR(100) NOT NULL,
    payload JSONB NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    attempts INTEGER NOT NULL DEFAULT 0,
    available_at TIMESTAMPTZ NOT NULL,
    published_at TIMESTAMPTZ,
    last_error TEXT,
    created_at TIMESTAMPTZ NOT NULL,
    CHECK (status IN ('pending', 'processing', 'published', 'dead_letter')),
    CHECK (attempts >= 0),
    CHECK (
        (status = 'published' AND published_at IS NOT NULL)
        OR
        (status <> 'published')
    )
);

CREATE INDEX outbox_events_pending_idx
    ON todoapp.outbox_events (available_at, id)
    WHERE status IN ('pending', 'processing');
