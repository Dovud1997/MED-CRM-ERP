CREATE TABLE patient_clinical_profiles(
  patient_id uuid PRIMARY KEY REFERENCES patients(id) ON DELETE RESTRICT,
  organization_id uuid NOT NULL REFERENCES organizations(id),
  blood_group text CHECK(blood_group IN ('O+','O-','A+','A-','B+','B-','AB+','AB-','unknown')) DEFAULT 'unknown',
  updated_by uuid REFERENCES users(id),
  updated_at timestamptz NOT NULL DEFAULT now(),
  version integer NOT NULL DEFAULT 1
);
CREATE TABLE patient_allergies(
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),patient_id uuid NOT NULL REFERENCES patients(id) ON DELETE RESTRICT,
  organization_id uuid NOT NULL REFERENCES organizations(id),allergen text NOT NULL,reaction text,severity text NOT NULL CHECK(severity IN ('mild','moderate','severe','unknown')),
  is_active boolean NOT NULL DEFAULT true,recorded_by uuid REFERENCES users(id),created_at timestamptz NOT NULL DEFAULT now(),updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE clinical_history_entries(
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),patient_id uuid NOT NULL REFERENCES patients(id) ON DELETE RESTRICT,
  organization_id uuid NOT NULL REFERENCES organizations(id),branch_id uuid REFERENCES branches(id),author_id uuid NOT NULL REFERENCES users(id),
  entry_type text NOT NULL CHECK(entry_type IN ('visit','diagnosis','prescription','procedure','note')),
  occurred_at timestamptz NOT NULL,complaints text,diagnosis text,treatment text,notes text,
  created_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE retention_policies(
  organization_id uuid PRIMARY KEY REFERENCES organizations(id),medical_record_years integer NOT NULL DEFAULT 3 CHECK(medical_record_years>=3),updated_at timestamptz NOT NULL DEFAULT now()
);
INSERT INTO retention_policies(organization_id) SELECT id FROM organizations ON CONFLICT DO NOTHING;
CREATE INDEX idx_allergies_patient ON patient_allergies(patient_id) WHERE is_active;
CREATE INDEX idx_history_patient_time ON clinical_history_entries(patient_id,occurred_at DESC);
INSERT INTO permissions(code,description) VALUES('clinical:read','Read clinical records'),('clinical:write','Manage clinical records'),('backup:export','Export encrypted backup') ON CONFLICT DO NOTHING;
