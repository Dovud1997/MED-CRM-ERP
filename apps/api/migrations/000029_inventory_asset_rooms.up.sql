ALTER TABLE inventory_assets
  ADD COLUMN room_id uuid REFERENCES inpatient_rooms(id);

CREATE INDEX inventory_assets_room_idx
  ON inventory_assets(organization_id, room_id)
  WHERE room_id IS NOT NULL;
