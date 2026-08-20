DROP TABLE IF EXISTS specialty_catalog;
DELETE FROM permissions WHERE code IN ('specialists:read','specialists:write');
