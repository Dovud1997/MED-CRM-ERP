CREATE TABLE cash_shifts(
 id uuid PRIMARY KEY, organization_id uuid NOT NULL REFERENCES organizations(id), branch_id uuid NOT NULL REFERENCES branches(id),
 opened_by uuid NOT NULL REFERENCES users(id), closed_by uuid REFERENCES users(id), status text NOT NULL DEFAULT 'OPEN' CHECK(status IN ('OPEN','CLOSED')),
 opening_cash_uzs bigint NOT NULL DEFAULT 0 CHECK(opening_cash_uzs>=0), closing_cash_uzs bigint CHECK(closing_cash_uzs>=0),
 expected_cash_uzs bigint, variance_uzs bigint, notes text, opened_at timestamptz NOT NULL DEFAULT now(), closed_at timestamptz
);
CREATE UNIQUE INDEX cash_shift_one_open_idx ON cash_shifts(organization_id,branch_id) WHERE status='OPEN';
CREATE INDEX cash_shifts_history_idx ON cash_shifts(organization_id,opened_at DESC);
CREATE TABLE cash_transactions(
 id uuid PRIMARY KEY, organization_id uuid NOT NULL REFERENCES organizations(id), branch_id uuid NOT NULL REFERENCES branches(id), shift_id uuid NOT NULL REFERENCES cash_shifts(id),
 patient_id uuid REFERENCES patients(id), service_id uuid REFERENCES services(id), related_transaction_id uuid REFERENCES cash_transactions(id),
 transaction_type text NOT NULL CHECK(transaction_type IN ('PAYMENT','REFUND','EXPENSE')), payment_method text NOT NULL CHECK(payment_method IN ('CASH','CARD','TRANSFER')),
 amount_uzs bigint NOT NULL CHECK(amount_uzs>0), receipt_number text NOT NULL, description text NOT NULL, created_by uuid NOT NULL REFERENCES users(id), created_at timestamptz NOT NULL DEFAULT now(),
 UNIQUE(organization_id,receipt_number)
);
CREATE INDEX cash_transactions_history_idx ON cash_transactions(organization_id,branch_id,created_at DESC);
CREATE INDEX cash_transactions_shift_idx ON cash_transactions(shift_id,created_at DESC);
INSERT INTO role_permissions(role_id,permission_id) SELECT r.id,p.id FROM roles r CROSS JOIN permissions p WHERE r.code IN ('OWNER','ADMIN','ACCOUNTANT','MANAGER','CASHIER') AND p.code IN ('finance:read','finance:write') ON CONFLICT DO NOTHING;
