-- Add medical metadata JSONB column
ALTER TABLE patients ADD COLUMN IF NOT EXISTS medical_metadata JSONB;

-- Create GIN index for fast JSONB queries
CREATE INDEX IF NOT EXISTS idx_patients_medical_metadata ON patients USING GIN (medical_metadata);

-- Add OMS policy number for integrations
ALTER TABLE patients ADD COLUMN IF NOT EXISTS oms_policy VARCHAR(16);

-- Add gender field for FHIR/integrations
ALTER TABLE patients ADD COLUMN IF NOT EXISTS gender VARCHAR(10);

-- Add index on OMS policy for lookups
CREATE INDEX IF NOT EXISTS idx_patients_oms_policy ON patients (oms_policy);
