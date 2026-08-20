CREATE TABLE patients(
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  organization_id uuid NOT NULL REFERENCES organizations(id),
  home_branch_id uuid REFERENCES branches(id),
  first_name text NOT NULL,
  last_name text NOT NULL,
  middle_name text,
  birth_date date NOT NULL,
  gender text NOT NULL CHECK(gender IN ('female','male')),
  phone_hash char(64) NOT NULL,
  phone_encrypted bytea NOT NULL,
  passport_encrypted bytea,
  permanent_address_encrypted bytea,
  guardian_name text,
  guardian_phone_encrypted bytea,
  notes text,
  created_by uuid REFERENCES users(id),
  deleted_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  version integer NOT NULL DEFAULT 1
);
CREATE UNIQUE INDEX idx_patients_org_phone ON patients(organization_id,phone_hash) WHERE deleted_at IS NULL;
CREATE INDEX idx_patients_org_name ON patients(organization_id,last_name,first_name) WHERE deleted_at IS NULL;
