ALTER TABLE patient_lab_results DROP COLUMN IF EXISTS attachment_id;
DROP TABLE IF EXISTS patient_lab_attachments;
