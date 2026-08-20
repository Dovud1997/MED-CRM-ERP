CREATE TABLE patient_lab_attachments(
  id uuid PRIMARY KEY,
  patient_id uuid NOT NULL REFERENCES patients(id),
  organization_id uuid NOT NULL REFERENCES organizations(id),
  original_name text NOT NULL,
  content_type text NOT NULL CHECK(content_type IN ('image/jpeg','image/png','image/webp','application/pdf')),
  byte_size integer NOT NULL CHECK(byte_size > 0 AND byte_size <= 10485760),
  content bytea NOT NULL,
  image_quality text NOT NULL DEFAULT 'not_checked' CHECK(image_quality IN ('accepted','not_checked')),
  uploaded_by uuid NOT NULL REFERENCES users(id),
  created_at timestamptz NOT NULL DEFAULT now()
);

ALTER TABLE patient_lab_results
  ADD COLUMN attachment_id uuid REFERENCES patient_lab_attachments(id);

CREATE INDEX idx_lab_attachments_patient ON patient_lab_attachments(patient_id,created_at DESC);
CREATE INDEX idx_lab_results_attachment ON patient_lab_results(attachment_id);
