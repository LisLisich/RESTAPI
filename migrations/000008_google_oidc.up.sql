CREATE TABLE todoapp.google_login_attempts (
    id BIGSERIAL PRIMARY KEY,
    state_hash BYTEA NOT NULL UNIQUE,
    nonce TEXT NOT NULL,
    code_verifier TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    consumed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL,
    CHECK (expires_at > created_at)
);

CREATE TABLE todoapp.oauth_identities (
    id BIGSERIAL PRIMARY KEY,
    account_user_id INTEGER NOT NULL REFERENCES todoapp.accounts(user_id) ON DELETE CASCADE,
    provider VARCHAR(20) NOT NULL,
    subject VARCHAR(255) NOT NULL,
    email_snapshot VARCHAR(254) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    UNIQUE (provider, subject),
    UNIQUE (account_user_id, provider),
    CHECK (provider = 'google')
);
