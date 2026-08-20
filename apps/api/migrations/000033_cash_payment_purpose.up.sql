ALTER TABLE cash_transactions ADD COLUMN payment_purpose text NOT NULL DEFAULT 'SERVICE'
  CHECK(payment_purpose IN ('SERVICE','LABORATORY','MEDICINE','SUPPLEMENT','INPATIENT','DIAGNOSTICS','OTHER'));
CREATE INDEX cash_transactions_purpose_idx ON cash_transactions(organization_id,payment_purpose,created_at DESC);
