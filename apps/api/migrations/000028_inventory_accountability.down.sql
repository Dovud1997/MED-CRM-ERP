DROP TABLE IF EXISTS inventory_count_lines,inventory_counts,inventory_assets;
ALTER TABLE inventory_batches DROP COLUMN IF EXISTS responsible_employee_id;
