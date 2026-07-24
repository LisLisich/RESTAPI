CREATE TABLE todoapp.wallets (
    account_user_id INTEGER PRIMARY KEY REFERENCES todoapp.accounts(user_id) ON DELETE CASCADE,
    balance_minor BIGINT NOT NULL DEFAULT 0,
    currency VARCHAR(3) NOT NULL DEFAULT 'RUB',
    version BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    CHECK (balance_minor >= 0),
    CHECK (currency = 'RUB')
);

CREATE TABLE todoapp.ledger_transactions (
    id UUID PRIMARY KEY,
    reference_type VARCHAR(50) NOT NULL,
    reference_id VARCHAR(100) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    UNIQUE (reference_type, reference_id)
);

CREATE TABLE todoapp.ledger_postings (
    id BIGSERIAL PRIMARY KEY,
    transaction_id UUID NOT NULL REFERENCES todoapp.ledger_transactions(id) ON DELETE CASCADE,
    account_code VARCHAR(150) NOT NULL,
    amount_minor BIGINT NOT NULL,
    currency VARCHAR(3) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    CHECK (amount_minor <> 0),
    CHECK (currency = 'RUB')
);

CREATE INDEX ledger_postings_account_idx
    ON todoapp.ledger_postings (account_code, id);

CREATE FUNCTION todoapp.assert_ledger_transaction_balanced()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
DECLARE
    checked_transaction_id UUID;
    posting_count BIGINT;
    posting_sum NUMERIC;
BEGIN
    checked_transaction_id := COALESCE(NEW.transaction_id, OLD.transaction_id);

    IF NOT EXISTS (
        SELECT 1
        FROM todoapp.ledger_transactions
        WHERE id = checked_transaction_id
    ) THEN
        RETURN NULL;
    END IF;

    SELECT COUNT(*), COALESCE(SUM(amount_minor::NUMERIC), 0)
    INTO posting_count, posting_sum
    FROM todoapp.ledger_postings
    WHERE transaction_id = checked_transaction_id;

    IF posting_count < 2 OR posting_sum <> 0 THEN
        RAISE EXCEPTION
            'ledger transaction % must contain at least two postings with zero sum',
            checked_transaction_id
            USING ERRCODE = '23514';
    END IF;

    RETURN NULL;
END;
$$;

CREATE CONSTRAINT TRIGGER ledger_postings_balanced
AFTER INSERT OR UPDATE OR DELETE ON todoapp.ledger_postings
DEFERRABLE INITIALLY DEFERRED
FOR EACH ROW
EXECUTE FUNCTION todoapp.assert_ledger_transaction_balanced();

CREATE FUNCTION todoapp.assert_new_ledger_transaction_balanced()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
DECLARE
    posting_count BIGINT;
    posting_sum NUMERIC;
BEGIN
    SELECT COUNT(*), COALESCE(SUM(amount_minor::NUMERIC), 0)
    INTO posting_count, posting_sum
    FROM todoapp.ledger_postings
    WHERE transaction_id = NEW.id;

    IF posting_count < 2 OR posting_sum <> 0 THEN
        RAISE EXCEPTION
            'ledger transaction % must contain at least two postings with zero sum',
            NEW.id
            USING ERRCODE = '23514';
    END IF;

    RETURN NULL;
END;
$$;

CREATE CONSTRAINT TRIGGER ledger_transaction_balanced
AFTER INSERT ON todoapp.ledger_transactions
DEFERRABLE INITIALLY DEFERRED
FOR EACH ROW
EXECUTE FUNCTION todoapp.assert_new_ledger_transaction_balanced();
