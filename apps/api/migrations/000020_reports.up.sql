INSERT INTO permissions(code,description) VALUES ('reports:read','Read clinic operational reports') ON CONFLICT DO NOTHING;
INSERT INTO role_permissions(role_id,permission_id)
SELECT r.id,p.id FROM roles r CROSS JOIN permissions p
WHERE r.code IN ('OWNER','ACCOUNTANT','MANAGER') AND p.code='reports:read'
ON CONFLICT DO NOTHING;
