CREATE EXTENSION IF NOT EXISTS btree_gist;

CREATE TABLE inpatient_rooms (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  branch_id uuid NOT NULL REFERENCES branches(id),
  number text NOT NULL,
  name text NOT NULL DEFAULT '',
  room_type text NOT NULL CHECK (room_type IN ('GENERAL','STANDARD','COMFORT','LUX','ICU','PEDIATRIC','MATERNITY')),
  floor integer NOT NULL DEFAULT 1 CHECK (floor BETWEEN -5 AND 100),
  patient_gender text NOT NULL DEFAULT 'ANY' CHECK (patient_gender IN ('ANY','FEMALE','MALE','CHILDREN')),
  amenities text NOT NULL DEFAULT '',
  price_per_day_uzs bigint NOT NULL CHECK (price_per_day_uzs >= 0),
  is_active boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (organization_id, branch_id, number)
);

CREATE TABLE inpatient_beds (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  room_id uuid NOT NULL REFERENCES inpatient_rooms(id),
  number text NOT NULL,
  operational_status text NOT NULL DEFAULT 'AVAILABLE' CHECK (operational_status IN ('AVAILABLE','CLEANING','MAINTENANCE')),
  is_active boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (room_id, number)
);

CREATE TABLE inpatient_bookings (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  branch_id uuid NOT NULL REFERENCES branches(id),
  room_id uuid NOT NULL REFERENCES inpatient_rooms(id),
  bed_id uuid NOT NULL REFERENCES inpatient_beds(id),
  patient_id uuid NOT NULL REFERENCES patients(id),
  attending_employee_id uuid REFERENCES employees(id),
  check_in timestamptz NOT NULL,
  check_out timestamptz NOT NULL,
  status text NOT NULL DEFAULT 'RESERVED' CHECK (status IN ('RESERVED','CHECKED_IN','DISCHARGED','CANCELLED')),
  price_per_day_uzs bigint NOT NULL CHECK (price_per_day_uzs >= 0),
  deposit_uzs bigint NOT NULL DEFAULT 0 CHECK (deposit_uzs >= 0),
  discount_uzs bigint NOT NULL DEFAULT 0 CHECK (discount_uzs >= 0),
  notes text NOT NULL DEFAULT '',
  created_by uuid REFERENCES users(id),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CHECK (check_out > check_in)
);

ALTER TABLE inpatient_bookings ADD CONSTRAINT inpatient_bed_booking_no_overlap
  EXCLUDE USING gist (bed_id WITH =, tstzrange(check_in, check_out, '[)') WITH &&)
  WHERE (status IN ('RESERVED','CHECKED_IN'));

CREATE INDEX inpatient_rooms_branch_idx ON inpatient_rooms(organization_id, branch_id, is_active);
CREATE INDEX inpatient_bookings_period_idx ON inpatient_bookings(organization_id, check_in, check_out, status);

INSERT INTO permissions(code,description) VALUES
('inpatient:read','Read rooms, beds and inpatient bookings'),
('inpatient:write','Manage rooms, beds and inpatient bookings')
ON CONFLICT DO NOTHING;

INSERT INTO role_permissions(role_id,permission_id)
SELECT r.id,p.id FROM roles r CROSS JOIN permissions p
WHERE r.code IN ('OWNER','ADMIN','RECEPTION') AND p.code IN ('inpatient:read','inpatient:write')
ON CONFLICT DO NOTHING;
