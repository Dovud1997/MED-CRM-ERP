INSERT INTO service_providers(organization_id,service_id,branch_id,employee_id)
SELECT organization_id,service_id,branch_id,employee_id
FROM service_prices
WHERE employee_id IS NOT NULL
ON CONFLICT DO NOTHING;
