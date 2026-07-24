CREATE TABLE todoapp.sessions (
    id BIGSERIAL PRIMARY KEY,
    account_user_id INTEGER NOT NULL REFERENCES todoapp.accounts(user_id) ON DELETE CASCADE,
    token_hash BYTEA NOT NULL UNIQUE,
    csrf_hash BYTEA NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL,
    last_seen_at TIMESTAMPTZ NOT NULL,
    CHECK (expires_at > created_at),
    CHECK (revoked_at IS NULL OR revoked_at >= created_at),
    CHECK (last_seen_at >= created_at)
);

CREATE INDEX sessions_active_account_idx
    ON todoapp.sessions (account_user_id, expires_at)
    WHERE revoked_at IS NULL;
