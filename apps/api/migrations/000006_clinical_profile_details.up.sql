ALTER TABLE patient_clinical_profiles
  ADD COLUMN height_cm numeric(5,2) CHECK(height_cm BETWEEN 30 AND 250),
  ADD COLUMN weight_kg numeric(6,2) CHECK(weight_kg BETWEEN 0.5 AND 500);

CREATE TABLE patient_vaccinations(
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  patient_id uuid NOT NULL REFERENCES patients(id) ON DELETE RESTRICT,
  organization_id uuid NOT NULL REFERENCES organizations(id),
  vaccine_name text NOT NULL,
  administered_on date NOT NULL,
  dose text,
  batch_number text,
  notes text,
  recorded_by uuid NOT NULL REFERENCES users(id),
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE patient_lab_results(
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  patient_id uuid NOT NULL REFERENCES patients(id) ON DELETE RESTRICT,
  organization_id uuid NOT NULL REFERENCES organizations(id),
  analysis_type text NOT NULL CHECK(analysis_type IN ('blood','urine')),
  collected_on date NOT NULL,
  test_name text NOT NULL,
  result_value text NOT NULL,
  unit text,
  reference_range text,
  notes text,
  recorded_by uuid NOT NULL REFERENCES users(id),
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_vaccinations_patient_date ON patient_vaccinations(patient_id,administered_on DESC);
CREATE INDEX idx_lab_results_patient_type_date ON patient_lab_results(patient_id,analysis_type,collected_on DESC);
