CREATE TABLE specialty_catalog (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  name text NOT NULL,
  is_active boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE(organization_id,name),
  CHECK(length(trim(name)) BETWEEN 2 AND 120)
);
INSERT INTO specialty_catalog(organization_id,name)
SELECT DISTINCT organization_id,trim(specialty) FROM employees
WHERE specialty IS NOT NULL AND length(trim(specialty))>=2
ON CONFLICT DO NOTHING;
INSERT INTO permissions(code,description) VALUES
('specialists:read','Read specialist directory'),('specialists:write','Manage specialist directory')
ON CONFLICT DO NOTHING;
