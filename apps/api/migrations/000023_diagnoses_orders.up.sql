CREATE TABLE patient_diagnoses(
  id uuid PRIMARY KEY,
  patient_id uuid NOT NULL REFERENCES patients(id),
  organization_id uuid NOT NULL REFERENCES organizations(id),
  branch_id uuid REFERENCES branches(id),
  appointment_id uuid REFERENCES appointments(id),
  diagnosis_name text NOT NULL,
  icd10_code text,
  diagnosis_type text NOT NULL CHECK(diagnosis_type IN ('PRIMARY','SECONDARY')),
  certainty text NOT NULL CHECK(certainty IN ('PRELIMINARY','CONFIRMED')),
  status text NOT NULL DEFAULT 'ACTIVE' CHECK(status IN ('ACTIVE','RESOLVED','RULED_OUT')),
  diagnosed_on date NOT NULL,
  notes text,
  created_by uuid NOT NULL REFERENCES users(id),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE patient_clinical_orders(
  id uuid PRIMARY KEY,
  patient_id uuid NOT NULL REFERENCES patients(id),
  organization_id uuid NOT NULL REFERENCES organizations(id),
  diagnosis_id uuid NOT NULL REFERENCES patient_diagnoses(id),
  branch_id uuid REFERENCES branches(id),
  appointment_id uuid REFERENCES appointments(id),
  order_type text NOT NULL CHECK(order_type IN ('MEDICATION','LAB','IMAGING','REFERRAL','PROCEDURE','FOLLOW_UP')),
  title text NOT NULL,
  dosage text,
  frequency text,
  duration_days integer CHECK(duration_days IS NULL OR duration_days BETWEEN 1 AND 3650),
  instructions text NOT NULL,
  start_on date NOT NULL,
  end_on date,
  status text NOT NULL DEFAULT 'ACTIVE' CHECK(status IN ('ACTIVE','COMPLETED','CANCELLED')),
  created_by uuid NOT NULL REFERENCES users(id),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CHECK(end_on IS NULL OR end_on >= start_on)
);

CREATE INDEX idx_diagnoses_patient ON patient_diagnoses(patient_id,diagnosed_on DESC,created_at DESC);
CREATE INDEX idx_orders_patient ON patient_clinical_orders(patient_id,status,start_on DESC);
CREATE INDEX idx_orders_diagnosis ON patient_clinical_orders(diagnosis_id,created_at DESC);
