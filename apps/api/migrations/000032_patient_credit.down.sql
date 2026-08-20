DROP INDEX IF EXISTS accounting_obligations_patient_idx;
ALTER TABLE accounting_obligations DROP COLUMN IF EXISTS source_transaction_id;
ALTER TABLE accounting_obligations DROP COLUMN IF EXISTS patient_id;
ALTER TABLE cash_transactions DROP CONSTRAINT cash_transactions_transaction_type_check;
ALTER TABLE cash_transactions DROP CONSTRAINT cash_transactions_payment_method_check;
ALTER TABLE cash_transactions ADD CONSTRAINT cash_transactions_transaction_type_check CHECK(transaction_type IN ('PAYMENT','REFUND','EXPENSE'));
ALTER TABLE cash_transactions ADD CONSTRAINT cash_transactions_payment_method_check CHECK(payment_method IN ('CASH','CARD','TRANSFER'));
