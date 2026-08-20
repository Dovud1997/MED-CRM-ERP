DROP INDEX IF EXISTS idx_employees_org_public_email;
ALTER TABLE employees
  DROP COLUMN IF EXISTS telegram_url,
  DROP COLUMN IF EXISTS public_email,
  DROP COLUMN IF EXISTS phone_encrypted,
  DROP COLUMN IF EXISTS permanent_address_encrypted,
  DROP COLUMN IF EXISTS passport_encrypted,
  DROP COLUMN IF EXISTS middle_name;
