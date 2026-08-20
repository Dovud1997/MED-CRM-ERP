DROP INDEX IF EXISTS cash_transactions_purpose_idx;
ALTER TABLE cash_transactions DROP COLUMN IF EXISTS payment_purpose;
