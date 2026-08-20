INSERT INTO permissions(code,description) VALUES
('messages:read','Read internal messages'),
('messages:write','Send internal messages')
ON CONFLICT(code) DO NOTHING;

INSERT INTO role_permissions(role_id,permission_id)
SELECT r.id,p.id FROM roles r CROSS JOIN permissions p
WHERE p.code IN ('messages:read','messages:write')
  AND (r.code='OWNER' OR r.code='RECEPTIONIST' OR r.code LIKE 'DOCTOR_%')
ON CONFLICT DO NOTHING;

CREATE TABLE internal_messages(
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  organization_id uuid NOT NULL REFERENCES organizations(id),
  sender_id uuid NOT NULL REFERENCES users(id),
  recipient_id uuid NOT NULL REFERENCES users(id),
  body text NOT NULL CHECK(char_length(body) BETWEEN 1 AND 4000),
  read_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  CHECK(sender_id<>recipient_id)
);
CREATE INDEX idx_internal_messages_dialog ON internal_messages(organization_id,sender_id,recipient_id,created_at DESC);
CREATE INDEX idx_internal_messages_unread ON internal_messages(organization_id,recipient_id,created_at DESC) WHERE read_at IS NULL;
