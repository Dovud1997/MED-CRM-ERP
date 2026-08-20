CREATE TABLE doctor_schedules(
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  organization_id uuid NOT NULL REFERENCES organizations(id),
  branch_id uuid NOT NULL REFERENCES branches(id),
  employee_id uuid NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
  weekday smallint NOT NULL CHECK(weekday BETWEEN 1 AND 7),
  is_working boolean NOT NULL DEFAULT true,
  start_time time,
  end_time time,
  break_start time,
  break_end time,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  version integer NOT NULL DEFAULT 1,
  UNIQUE(organization_id,employee_id,weekday),
  CHECK((NOT is_working AND start_time IS NULL AND end_time IS NULL AND break_start IS NULL AND break_end IS NULL) OR
        (is_working AND start_time IS NOT NULL AND end_time IS NOT NULL AND end_time > start_time)),
  CHECK((break_start IS NULL AND break_end IS NULL) OR
        (is_working AND break_start IS NOT NULL AND break_end IS NOT NULL AND break_end > break_start AND break_start >= start_time AND break_end <= end_time))
);
CREATE INDEX idx_doctor_schedules_employee ON doctor_schedules(organization_id,employee_id,weekday);
