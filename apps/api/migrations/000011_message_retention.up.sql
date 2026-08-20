ALTER TABLE retention_policies ADD COLUMN IF NOT EXISTS message_retention_months integer NOT NULL DEFAULT 3 CHECK(message_retention_months >= 3);
