DROP INDEX IF EXISTS inventory_assets_room_idx;
ALTER TABLE inventory_assets DROP COLUMN IF EXISTS room_id;
