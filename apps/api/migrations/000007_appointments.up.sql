CREATE EXTENSION IF NOT EXISTS btree_gist;

CREATE TABLE appointments(
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  organization_id uuid NOT NULL REFERENCES organizations(id),
  branch_id uuid NOT NULL REFERENCES branches(id),
  patient_id uuid NOT NULL REFERENCES patients(id) ON DELETE RESTRICT,
  employee_id uuid NOT NULL REFERENCES employees(id) ON DELETE RESTRICT,
  starts_at timestamptz NOT NULL,
  ends_at timestamptz NOT NULL,
  status text NOT NULL DEFAULT 'scheduled' CHECK(status IN ('scheduled','confirmed','arrived','in_progress','completed','cancelled','no_show')),
  reason text,
  notes text,
  created_by uuid NOT NULL REFERENCES users(id),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  version integer NOT NULL DEFAULT 1,
  CHECK(ends_at > starts_at),
  CHECK(ends_at <= starts_at + interval '4 hours')
);

ALTER TABLE appointments ADD CONSTRAINT appointments_doctor_no_overlap
  EXCLUDE USING gist (organization_id WITH =, employee_id WITH =, tstzrange(starts_at,ends_at,'[)') WITH &&)
  WHERE (status NOT IN ('cancelled','no_show'));
CREATE INDEX idx_appointments_org_start ON appointments(organization_id,starts_at);
CREATE INDEX idx_appointments_patient ON appointments(patient_id,starts_at DESC);
