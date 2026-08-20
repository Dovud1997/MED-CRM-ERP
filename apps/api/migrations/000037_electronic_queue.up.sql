INSERT INTO permissions(code,description) VALUES
('queue:read','View electronic queue'),
('queue:manage','Manage electronic queue'),
('queue:call','Call patients from queue'),
('queue:settings','Manage queue settings'),
('queue:media','Manage queue media'),
('queue:audio','Manage queue audio'),
('queue:display','Manage queue displays')
ON CONFLICT(code) DO NOTHING;

INSERT INTO role_permissions(role_id,permission_id)
SELECT r.id,p.id FROM roles r CROSS JOIN permissions p
WHERE r.code IN ('OWNER','ADMIN') AND p.code LIKE 'queue:%'
ON CONFLICT DO NOTHING;
INSERT INTO role_permissions(role_id,permission_id)
SELECT r.id,p.id FROM roles r CROSS JOIN permissions p
WHERE (r.code='RECEPTION' OR r.code='RECEPTIONIST') AND p.code IN ('queue:read','queue:manage','queue:call')
ON CONFLICT DO NOTHING;
INSERT INTO role_permissions(role_id,permission_id)
SELECT r.id,p.id FROM roles r CROSS JOIN permissions p
WHERE r.code LIKE 'DOCTOR_%' AND p.code IN ('queue:read','queue:call')
ON CONFLICT DO NOTHING;

CREATE TABLE queue_rooms(
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  organization_id uuid NOT NULL REFERENCES organizations(id),
  branch_id uuid NOT NULL REFERENCES branches(id),
  room_number text NOT NULL,
  name text,
  floor integer,
  department text,
  employee_id uuid REFERENCES employees(id),
  is_active boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE(organization_id,branch_id,room_number)
);

CREATE TABLE queue_displays(
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  organization_id uuid NOT NULL REFERENCES organizations(id),
  branch_id uuid NOT NULL REFERENCES branches(id),
  name text NOT NULL,
  slug text NOT NULL,
  token_hash char(64) NOT NULL UNIQUE,
  token_prefix text NOT NULL,
  is_active boolean NOT NULL DEFAULT true,
  language text NOT NULL DEFAULT 'ru' CHECK(language IN ('ru','uz','en')),
  voice_language text NOT NULL DEFAULT 'ru' CHECK(voice_language IN ('ru','uz','en')),
  privacy_mode text NOT NULL DEFAULT 'NUMBER_ONLY' CHECK(privacy_mode IN ('NUMBER_ONLY','FIRST_NAME_INITIAL','FULL_NAME')),
  settings jsonb NOT NULL DEFAULT '{"primaryColor":"#e80d4f","secondaryColor":"#17202a","backgroundColor":"#f7f9fb","textColor":"#17202a","showDoctor":true,"showSpecialty":true,"showClock":true,"showDate":true,"showSeconds":false,"showMedia":true,"showTicker":true,"showRecentCalls":true,"recentCallCount":5,"callDisplaySeconds":12,"videoCallMode":"OVERLAY","videoAnnouncementVolume":0}'::jsonb,
  created_by uuid NOT NULL REFERENCES users(id),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE(organization_id,slug)
);

CREATE TABLE queue_display_rooms(
  display_id uuid NOT NULL REFERENCES queue_displays(id) ON DELETE CASCADE,
  room_id uuid NOT NULL REFERENCES queue_rooms(id) ON DELETE CASCADE,
  PRIMARY KEY(display_id,room_id)
);

CREATE TABLE queue_entries(
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  organization_id uuid NOT NULL REFERENCES organizations(id),
  branch_id uuid NOT NULL REFERENCES branches(id),
  appointment_id uuid REFERENCES appointments(id),
  patient_id uuid NOT NULL REFERENCES patients(id),
  employee_id uuid REFERENCES employees(id),
  room_id uuid REFERENCES queue_rooms(id),
  queue_date date NOT NULL DEFAULT (now() AT TIME ZONE 'Asia/Tashkent')::date,
  sequence_number integer NOT NULL,
  display_number text NOT NULL,
  queue_scope text NOT NULL DEFAULT 'CLINIC' CHECK(queue_scope IN ('CLINIC','DOCTOR','ROOM','DEPARTMENT','SERVICE')),
  scope_key text NOT NULL DEFAULT 'clinic',
  status text NOT NULL DEFAULT 'WAITING' CHECK(status IN ('WAITING','CALLED','IN_SERVICE','COMPLETED','SKIPPED','CANCELLED','RECALLED')),
  priority integer NOT NULL DEFAULT 0,
  called_at timestamptz,
  service_started_at timestamptz,
  completed_at timestamptz,
  created_by uuid NOT NULL REFERENCES users(id),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE(organization_id,branch_id,queue_date,scope_key,sequence_number)
);
CREATE INDEX idx_queue_entries_live ON queue_entries(organization_id,branch_id,queue_date,status,priority DESC,sequence_number);

CREATE TABLE queue_call_history(
  id bigserial PRIMARY KEY,
  organization_id uuid NOT NULL REFERENCES organizations(id),
  queue_entry_id uuid NOT NULL REFERENCES queue_entries(id),
  display_id uuid REFERENCES queue_displays(id),
  called_by uuid NOT NULL REFERENCES users(id),
  call_type text NOT NULL CHECK(call_type IN ('CALL','RECALL','SKIP','START','COMPLETE','CANCEL')),
  status text NOT NULL,
  room_number text,
  created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_queue_call_history_entry ON queue_call_history(queue_entry_id,created_at DESC);

CREATE TABLE queue_media_assets(
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  organization_id uuid NOT NULL REFERENCES organizations(id),
  kind text NOT NULL CHECK(kind IN ('AUDIO','VIDEO','BANNER','IMAGE','LOGO','QR','NOTIFICATION')),
  title text NOT NULL,
  asset_key text,
  language text CHECK(language IS NULL OR language IN ('ru','uz','en')),
  content_type text NOT NULL,
  file_name text NOT NULL,
  storage_path text NOT NULL,
  byte_size bigint NOT NULL CHECK(byte_size>0),
  sort_order integer NOT NULL DEFAULT 0,
  is_active boolean NOT NULL DEFAULT true,
  starts_at timestamptz,
  ends_at timestamptz,
  duration_seconds integer,
  created_by uuid NOT NULL REFERENCES users(id),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE(organization_id,asset_key,language)
);

CREATE TABLE queue_display_media(
  display_id uuid NOT NULL REFERENCES queue_displays(id) ON DELETE CASCADE,
  media_id uuid NOT NULL REFERENCES queue_media_assets(id) ON DELETE CASCADE,
  sort_order integer NOT NULL DEFAULT 0,
  PRIMARY KEY(display_id,media_id)
);

CREATE TABLE queue_audio_templates(
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  organization_id uuid NOT NULL REFERENCES organizations(id),
  name text NOT NULL,
  language text NOT NULL CHECK(language IN ('ru','uz','en')),
  template text NOT NULL,
  is_active boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE(organization_id,language,name)
);

CREATE TABLE queue_ticker_messages(
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  organization_id uuid NOT NULL REFERENCES organizations(id),
  text text NOT NULL CHECK(char_length(text) BETWEEN 1 AND 500),
  language text NOT NULL DEFAULT 'ru' CHECK(language IN ('ru','uz','en')),
  speed integer NOT NULL DEFAULT 40 CHECK(speed BETWEEN 10 AND 200),
  sort_order integer NOT NULL DEFAULT 0,
  is_active boolean NOT NULL DEFAULT true,
  starts_at timestamptz,
  ends_at timestamptz,
  created_by uuid NOT NULL REFERENCES users(id),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

