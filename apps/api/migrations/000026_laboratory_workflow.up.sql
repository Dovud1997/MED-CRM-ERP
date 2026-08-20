CREATE TABLE laboratory_test_catalog(
  id uuid PRIMARY KEY,
  organization_id uuid NOT NULL REFERENCES organizations(id),
  name text NOT NULL,
  code text,
  specimen text NOT NULL DEFAULT 'Кровь',
  description text,
  is_active boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE(organization_id,name)
);

CREATE TABLE laboratory_orders(
  id uuid PRIMARY KEY,
  organization_id uuid NOT NULL REFERENCES organizations(id),
  branch_id uuid NOT NULL REFERENCES branches(id),
  patient_id uuid NOT NULL REFERENCES patients(id),
  ordering_employee_id uuid REFERENCES employees(id),
  status text NOT NULL DEFAULT 'NEW' CHECK(status IN ('NEW','SAMPLE_COLLECTED','IN_PROGRESS','COMPLETED','CANCELLED')),
  priority text NOT NULL DEFAULT 'ROUTINE' CHECK(priority IN ('ROUTINE','URGENT')),
  clinical_note text,
  source text NOT NULL DEFAULT 'INTERNAL' CHECK(source IN ('INTERNAL','EXTERNAL')),
  attachment_id uuid REFERENCES patient_lab_attachments(id),
  result_note text,
  requested_by uuid NOT NULL REFERENCES users(id),
  completed_by uuid REFERENCES users(id),
  completed_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE laboratory_order_items(
  id uuid PRIMARY KEY,
  order_id uuid NOT NULL REFERENCES laboratory_orders(id) ON DELETE CASCADE,
  test_id uuid NOT NULL REFERENCES laboratory_test_catalog(id),
  result_value text,
  unit text,
  reference_range text,
  patient_lab_result_id uuid REFERENCES patient_lab_results(id),
  UNIQUE(order_id,test_id)
);

CREATE INDEX idx_lab_orders_org_status_created ON laboratory_orders(organization_id,status,created_at DESC);
CREATE INDEX idx_lab_orders_patient ON laboratory_orders(patient_id,created_at DESC);

INSERT INTO laboratory_test_catalog(id,organization_id,name,code,specimen,description)
SELECT gen_random_uuid(),o.id,x.name,x.code,x.specimen,x.description
FROM organizations o CROSS JOIN (VALUES
 ('Общий анализ крови','CBC','Кровь','Основные показатели крови и лейкоцитарная формула'),
 ('Общий анализ мочи','UA','Моча','Физико-химические и микроскопические показатели'),
 ('Биохимический анализ крови','BIO','Кровь','Базовая биохимическая панель'),
 ('Глюкоза крови','GLU','Кровь','Определение уровня глюкозы'),
 ('С-реактивный белок','CRP','Кровь','Маркер воспаления'),
 ('ТТГ','TSH','Кровь','Тиреотропный гормон')
) AS x(name,code,specimen,description)
ON CONFLICT DO NOTHING;
