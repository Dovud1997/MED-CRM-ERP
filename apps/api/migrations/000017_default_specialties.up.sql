INSERT INTO specialty_catalog(organization_id,name)
SELECT o.id,s.name
FROM organizations o
CROSS JOIN (VALUES
  ('Врач-педиатр'),
  ('Врач-хирург'),
  ('Врач-стоматолог'),
  ('Врач УЗИ'),
  ('Врач-гинеколог'),
  ('Логопед')
) AS s(name)
ON CONFLICT DO NOTHING;
