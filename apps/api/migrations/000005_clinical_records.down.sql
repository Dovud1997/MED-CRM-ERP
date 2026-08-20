DELETE FROM permissions WHERE code IN ('clinical:read','clinical:write','backup:export');
DROP TABLE IF EXISTS retention_policies,clinical_history_entries,patient_allergies,patient_clinical_profiles;
