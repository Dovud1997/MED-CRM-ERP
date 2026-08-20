CREATE TABLE patient_charges(
 id uuid PRIMARY KEY, organization_id uuid NOT NULL REFERENCES organizations(id), branch_id uuid NOT NULL REFERENCES branches(id),
 patient_id uuid NOT NULL REFERENCES patients(id), source_type text NOT NULL CHECK(source_type IN ('APPOINTMENT','LABORATORY','INPATIENT','PHARMACY','DIAGNOSTICS','MANUAL')),
 source_id uuid, description text NOT NULL, amount_uzs bigint NOT NULL CHECK(amount_uzs>0), paid_uzs bigint NOT NULL DEFAULT 0 CHECK(paid_uzs>=0 AND paid_uzs<=amount_uzs),
 status text NOT NULL DEFAULT 'UNPAID' CHECK(status IN ('UNPAID','PARTIAL','PAID','CANCELLED')), charged_at timestamptz NOT NULL DEFAULT now(), created_by uuid NOT NULL REFERENCES users(id),
 UNIQUE(organization_id,source_type,source_id)
);
CREATE INDEX patient_charges_account_idx ON patient_charges(organization_id,patient_id,status,charged_at DESC);
CREATE TABLE cash_payment_allocations(
 transaction_id uuid NOT NULL REFERENCES cash_transactions(id), charge_id uuid NOT NULL REFERENCES patient_charges(id),
 amount_uzs bigint NOT NULL CHECK(amount_uzs>0), created_at timestamptz NOT NULL DEFAULT now(), PRIMARY KEY(transaction_id,charge_id)
);
ALTER TABLE laboratory_test_catalog ADD COLUMN price_uzs bigint NOT NULL DEFAULT 0 CHECK(price_uzs>=0);
