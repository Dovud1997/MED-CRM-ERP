CREATE TABLE inventory_items(
  id uuid PRIMARY KEY,
  organization_id uuid NOT NULL REFERENCES organizations(id),
  sku text NOT NULL,
  name text NOT NULL,
  category text NOT NULL DEFAULT 'MEDICINE',
  unit text NOT NULL DEFAULT 'шт.',
  manufacturer text,
  barcode text,
  min_stock numeric(14,3) NOT NULL DEFAULT 0 CHECK(min_stock>=0),
  is_active boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT now(),updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE(organization_id,sku)
);
CREATE TABLE inventory_batches(
  id uuid PRIMARY KEY,
  organization_id uuid NOT NULL REFERENCES organizations(id),
  branch_id uuid NOT NULL REFERENCES branches(id),
  item_id uuid NOT NULL REFERENCES inventory_items(id),
  lot_number text NOT NULL,
  expires_on date,
  quantity numeric(14,3) NOT NULL DEFAULT 0 CHECK(quantity>=0),
  purchase_price_uzs bigint NOT NULL DEFAULT 0 CHECK(purchase_price_uzs>=0),
  created_at timestamptz NOT NULL DEFAULT now(),updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE(organization_id,branch_id,item_id,lot_number)
);
CREATE TABLE inventory_movements(
  id uuid PRIMARY KEY,
  organization_id uuid NOT NULL REFERENCES organizations(id),
  branch_id uuid NOT NULL REFERENCES branches(id),
  item_id uuid NOT NULL REFERENCES inventory_items(id),
  batch_id uuid REFERENCES inventory_batches(id),
  movement_type text NOT NULL CHECK(movement_type IN ('RECEIPT','ISSUE','WRITE_OFF','TRANSFER_IN','TRANSFER_OUT','ADJUSTMENT')),
  quantity numeric(14,3) NOT NULL CHECK(quantity>0),
  reference_id uuid,
  reason text,
  performed_by uuid NOT NULL REFERENCES users(id),
  created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX inventory_batches_lookup_idx ON inventory_batches(organization_id,branch_id,item_id,expires_on);
CREATE INDEX inventory_movements_created_idx ON inventory_movements(organization_id,created_at DESC);
INSERT INTO permissions(code,description) VALUES('inventory:read','Read inventory and stock movements'),('inventory:write','Manage inventory and stock movements') ON CONFLICT DO NOTHING;
INSERT INTO role_permissions(role_id,permission_id) SELECT r.id,p.id FROM roles r CROSS JOIN permissions p WHERE r.code IN ('OWNER','ADMIN','ACCOUNTANT','MANAGER') AND p.code IN ('inventory:read','inventory:write') ON CONFLICT DO NOTHING;
