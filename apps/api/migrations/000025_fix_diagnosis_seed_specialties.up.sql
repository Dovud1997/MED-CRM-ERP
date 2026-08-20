-- Remove general-therapy diagnoses that were accidentally matched to specialties
-- whose names also end with "терапевт".
DELETE FROM diagnosis_catalog d
USING specialty_catalog s
WHERE d.specialty_id = s.id
  AND s.organization_id = d.organization_id
  AND lower(s.name) IN ('врач-физиотерапевт', 'мануальный терапевт')
  AND d.icd10_code IN ('J06.9', 'I10', 'D50.9');
