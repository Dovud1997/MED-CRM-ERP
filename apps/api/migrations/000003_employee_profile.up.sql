ALTER TABLE employees
  ADD COLUMN middle_name text,
  ADD COLUMN passport_encrypted bytea,
  ADD COLUMN permanent_address_encrypted bytea,
  ADD COLUMN phone_encrypted bytea,
  ADD COLUMN public_email text,
  ADD COLUMN telegram_url text;

CREATE UNIQUE INDEX idx_employees_org_public_email
  ON employees(organization_id, lower(public_email))
  WHERE public_email IS NOT NULL;
