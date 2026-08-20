DROP TABLE IF EXISTS patient_lab_results,patient_vaccinations;
ALTER TABLE patient_clinical_profiles DROP COLUMN IF EXISTS weight_kg, DROP COLUMN IF EXISTS height_cm;
