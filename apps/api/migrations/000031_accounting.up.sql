CREATE TABLE accounting_categories(
 id uuid PRIMARY KEY, organization_id uuid NOT NULL REFERENCES organizations(id), name text NOT NULL,
 entry_type text NOT NULL CHECK(entry_type IN ('INCOME','EXPENSE')), is_active boolean NOT NULL DEFAULT true,
 created_at timestamptz NOT NULL DEFAULT now(), UNIQUE(organization_id,name,entry_type)
);
CREATE TABLE accounting_entries(
 id uuid PRIMARY KEY, organization_id uuid NOT NULL REFERENCES organizations(id), branch_id uuid NOT NULL REFERENCES branches(id),
 category_id uuid REFERENCES accounting_categories(id), entry_type text NOT NULL CHECK(entry_type IN ('INCOME','EXPENSE')),
 amount_uzs bigint NOT NULL CHECK(amount_uzs>0), payment_method text NOT NULL CHECK(payment_method IN ('CASH','CARD','TRANSFER')),
 occurred_on date NOT NULL DEFAULT CURRENT_DATE, description text NOT NULL, document_number text, counterparty text,
 created_by uuid NOT NULL REFERENCES users(id), created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX accounting_entries_period_idx ON accounting_entries(organization_id,occurred_on DESC,branch_id);

CREATE TABLE accounting_obligations(
 id uuid PRIMARY KEY, organization_id uuid NOT NULL REFERENCES organizations(id), branch_id uuid NOT NULL REFERENCES branches(id),
 obligation_type text NOT NULL CHECK(obligation_type IN ('RECEIVABLE','PAYABLE','EMPLOYEE')),
 employee_id uuid REFERENCES employees(id), counterparty text, description text NOT NULL, amount_uzs bigint NOT NULL CHECK(amount_uzs>0),
 paid_uzs bigint NOT NULL DEFAULT 0 CHECK(paid_uzs>=0), due_on date, status text NOT NULL DEFAULT 'OPEN' CHECK(status IN ('OPEN','PARTIAL','PAID','CANCELLED')),
 created_by uuid NOT NULL REFERENCES users(id), created_at timestamptz NOT NULL DEFAULT now(), updated_at timestamptz NOT NULL DEFAULT now(),
 CHECK(paid_uzs<=amount_uzs)
);
CREATE INDEX accounting_obligations_status_idx ON accounting_obligations(organization_id,status,due_on);

CREATE TABLE payroll_rules(
 id uuid PRIMARY KEY, organization_id uuid NOT NULL REFERENCES organizations(id), employee_id uuid NOT NULL REFERENCES employees(id),
 rule_type text NOT NULL CHECK(rule_type IN ('FIXED','PERCENT','MIXED')), fixed_amount_uzs bigint NOT NULL DEFAULT 0 CHECK(fixed_amount_uzs>=0),
 percent_rate numeric(5,2) NOT NULL DEFAULT 0 CHECK(percent_rate>=0 AND percent_rate<=100), is_active boolean NOT NULL DEFAULT true,
 created_by uuid NOT NULL REFERENCES users(id), updated_at timestamptz NOT NULL DEFAULT now(), UNIQUE(organization_id,employee_id)
);
CREATE TABLE payroll_accruals(
 id uuid PRIMARY KEY, organization_id uuid NOT NULL REFERENCES organizations(id), branch_id uuid REFERENCES branches(id), employee_id uuid NOT NULL REFERENCES employees(id),
 period_month date NOT NULL, revenue_base_uzs bigint NOT NULL DEFAULT 0 CHECK(revenue_base_uzs>=0), fixed_amount_uzs bigint NOT NULL DEFAULT 0,
 percent_amount_uzs bigint NOT NULL DEFAULT 0, bonus_uzs bigint NOT NULL DEFAULT 0, deduction_uzs bigint NOT NULL DEFAULT 0,
 total_uzs bigint NOT NULL CHECK(total_uzs>=0), paid_uzs bigint NOT NULL DEFAULT 0 CHECK(paid_uzs>=0), status text NOT NULL DEFAULT 'ACCRUED' CHECK(status IN ('ACCRUED','PARTIAL','PAID')),
 notes text, created_by uuid NOT NULL REFERENCES users(id), created_at timestamptz NOT NULL DEFAULT now(), UNIQUE(organization_id,employee_id,period_month), CHECK(paid_uzs<=total_uzs)
);
CREATE INDEX payroll_accruals_period_idx ON payroll_accruals(organization_id,period_month DESC,status);

INSERT INTO accounting_categories(id,organization_id,name,entry_type)
SELECT gen_random_uuid(),o.id,x.name,x.kind FROM organizations o CROSS JOIN (VALUES
 ('Оплата медицинских услуг','INCOME'),('Стационар','INCOME'),('Лаборатория','INCOME'),('Прочий доход','INCOME'),
 ('Заработная плата','EXPENSE'),('Закупка лекарств и материалов','EXPENSE'),('Аренда','EXPENSE'),('Коммунальные услуги','EXPENSE'),('Налоги','EXPENSE'),('Ремонт и обслуживание','EXPENSE'),('Прочий расход','EXPENSE')
) x(name,kind) ON CONFLICT DO NOTHING;
INSERT INTO role_permissions(role_id,permission_id) SELECT r.id,p.id FROM roles r CROSS JOIN permissions p WHERE r.code IN ('OWNER','ADMIN','ACCOUNTANT') AND p.code IN ('finance:read','finance:write') ON CONFLICT DO NOTHING;
