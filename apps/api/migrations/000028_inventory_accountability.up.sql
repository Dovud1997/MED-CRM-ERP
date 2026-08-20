ALTER TABLE inventory_batches ADD COLUMN responsible_employee_id uuid REFERENCES employees(id);

CREATE TABLE inventory_assets(
 id uuid PRIMARY KEY,organization_id uuid NOT NULL REFERENCES organizations(id),branch_id uuid NOT NULL REFERENCES branches(id),item_id uuid NOT NULL REFERENCES inventory_items(id),
 inventory_number text NOT NULL,serial_number text,status text NOT NULL DEFAULT 'AVAILABLE' CHECK(status IN('AVAILABLE','IN_USE','MAINTENANCE','RETIRED')),
 assigned_employee_id uuid REFERENCES employees(id),responsible_employee_id uuid REFERENCES employees(id),acquired_on date,warranty_until date,notes text,
 created_at timestamptz NOT NULL DEFAULT now(),updated_at timestamptz NOT NULL DEFAULT now(),UNIQUE(organization_id,inventory_number)
);
CREATE TABLE inventory_counts(
 id uuid PRIMARY KEY,organization_id uuid NOT NULL REFERENCES organizations(id),branch_id uuid NOT NULL REFERENCES branches(id),status text NOT NULL DEFAULT 'DRAFT' CHECK(status IN('DRAFT','COMPLETED','CANCELLED')),
 started_by uuid NOT NULL REFERENCES users(id),completed_by uuid REFERENCES users(id),started_at timestamptz NOT NULL DEFAULT now(),completed_at timestamptz,notes text
);
CREATE TABLE inventory_count_lines(
 id uuid PRIMARY KEY,count_id uuid NOT NULL REFERENCES inventory_counts(id) ON DELETE CASCADE,batch_id uuid NOT NULL REFERENCES inventory_batches(id),expected_quantity numeric(14,3) NOT NULL,actual_quantity numeric(14,3),difference numeric(14,3),UNIQUE(count_id,batch_id)
);
CREATE INDEX inventory_assets_branch_idx ON inventory_assets(organization_id,branch_id,status);
CREATE INDEX inventory_counts_branch_idx ON inventory_counts(organization_id,branch_id,started_at DESC);
