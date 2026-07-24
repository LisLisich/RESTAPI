CREATE TABLE todoapp.refresh_tokens (
    id BIGSERIAL PRIMARY KEY,
    account_user_id INTEGER NOT NULL REFERENCES todoapp.accounts(user_id) ON DELETE CASCADE,
    family_id VARCHAR(100) NOT NULL,
    token_hash BYTEA NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL,
    CHECK (expires_at > created_at),
    CHECK (used_at IS NULL OR used_at >= created_at),
    CHECK (revoked_at IS NULL OR revoked_at >= created_at)
);

CREATE INDEX refresh_tokens_active_family_idx
    ON todoapp.refresh_tokens (family_id, expires_at)
    WHERE used_at IS NULL AND revoked_at IS NULL;
