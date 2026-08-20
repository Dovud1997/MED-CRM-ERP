ALTER TABLE cash_transactions DROP CONSTRAINT cash_transactions_transaction_type_check;
ALTER TABLE cash_transactions DROP CONSTRAINT cash_transactions_payment_method_check;
ALTER TABLE cash_transactions ADD CONSTRAINT cash_transactions_transaction_type_check CHECK(transaction_type IN ('PAYMENT','REFUND','EXPENSE','DEBT'));
ALTER TABLE cash_transactions ADD CONSTRAINT cash_transactions_payment_method_check CHECK(payment_method IN ('CASH','CARD','TRANSFER','DEBT'));

ALTER TABLE accounting_obligations ADD COLUMN patient_id uuid REFERENCES patients(id);
ALTER TABLE accounting_obligations ADD COLUMN source_transaction_id uuid UNIQUE REFERENCES cash_transactions(id);
CREATE INDEX accounting_obligations_patient_idx ON accounting_obligations(organization_id,patient_id,status);
