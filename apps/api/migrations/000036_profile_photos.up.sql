CREATE TABLE profile_photos(
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  organization_id uuid NOT NULL REFERENCES organizations(id),
  entity_type text NOT NULL CHECK(entity_type IN ('patient','employee')),
  entity_id uuid NOT NULL,
  content_type text NOT NULL CHECK(content_type IN ('image/jpeg','image/png','image/webp')),
  byte_size integer NOT NULL CHECK(byte_size > 0 AND byte_size <= 5242880),
  content bytea NOT NULL,
  uploaded_by uuid NOT NULL REFERENCES users(id),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE(organization_id,entity_type,entity_id)
);
CREATE INDEX idx_profile_photos_entity ON profile_photos(organization_id,entity_type,entity_id);
