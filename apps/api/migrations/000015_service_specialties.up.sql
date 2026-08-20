CREATE TABLE service_specialties (
  organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  service_id uuid NOT NULL REFERENCES services(id) ON DELETE CASCADE,
  branch_id uuid NOT NULL REFERENCES branches(id) ON DELETE CASCADE,
  specialty text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY(service_id, branch_id, specialty),
  CHECK(length(trim(specialty)) BETWEEN 2 AND 120)
);
CREATE INDEX service_specialties_lookup_idx ON service_specialties(organization_id,branch_id,specialty);
