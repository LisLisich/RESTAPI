CREATE TABLE todoapp.notifications (
    id BIGSERIAL PRIMARY KEY,
    account_user_id INTEGER NOT NULL REFERENCES todoapp.accounts(user_id) ON DELETE CASCADE,
    outbox_event_id BIGINT NOT NULL REFERENCES todoapp.outbox_events(id) ON DELETE CASCADE,
    title VARCHAR(150) NOT NULL,
    body TEXT NOT NULL,
    read_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (account_user_id, outbox_event_id)
);

CREATE INDEX notifications_account_created_idx
    ON todoapp.notifications (account_user_id, created_at DESC);
