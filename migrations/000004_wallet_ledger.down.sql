DROP TRIGGER ledger_transaction_balanced ON todoapp.ledger_transactions;
DROP TRIGGER ledger_postings_balanced ON todoapp.ledger_postings;
DROP FUNCTION todoapp.assert_new_ledger_transaction_balanced();
DROP FUNCTION todoapp.assert_ledger_transaction_balanced();
DROP TABLE todoapp.ledger_postings;
DROP TABLE todoapp.ledger_transactions;
DROP TABLE todoapp.wallets;
