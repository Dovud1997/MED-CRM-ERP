DROP TABLE IF EXISTS internal_messages;
DELETE FROM permissions WHERE code IN ('messages:read','messages:write');
