-- Remove indexes
DROP INDEX IF EXISTS idx_patients_oms_policy;
DROP INDEX IF EXISTS idx_patients_medical_metadata;

-- Remove columns
ALTER TABLE patients DROP COLUMN IF EXISTS gender;
ALTER TABLE patients DROP COLUMN IF EXISTS oms_policy;
ALTER TABLE patients DROP COLUMN IF EXISTS medical_metadata;
