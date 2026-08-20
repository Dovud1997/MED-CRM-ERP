CREATE TABLE patient_imaging_studies(
  id uuid PRIMARY KEY,
  patient_id uuid NOT NULL REFERENCES patients(id),
  organization_id uuid NOT NULL REFERENCES organizations(id),
  modality text NOT NULL CHECK(modality IN ('MRI','CT','XRAY')),
  performed_on date NOT NULL,
  body_area text NOT NULL,
  diagnosis text NOT NULL,
  conclusion text,
  notes text,
  original_name text NOT NULL,
  content_type text NOT NULL CHECK(content_type IN ('image/jpeg','image/png','image/webp','application/pdf','application/dicom')),
  byte_size integer NOT NULL CHECK(byte_size > 0 AND byte_size <= 26214400),
  content bytea NOT NULL,
  created_by uuid NOT NULL REFERENCES users(id),
  created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_imaging_patient_date ON patient_imaging_studies(patient_id,performed_on DESC,created_at DESC);
