DROP TABLE IF EXISTS inventory_movements,inventory_batches,inventory_items;
DELETE FROM permissions WHERE code IN ('inventory:read','inventory:write');
