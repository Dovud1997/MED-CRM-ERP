DROP TABLE IF EXISTS service_prices, services, service_categories CASCADE;
DELETE FROM permissions WHERE code IN ('services:read','services:write');
