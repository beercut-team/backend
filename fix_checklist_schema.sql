-- Remove default value from is_required column
ALTER TABLE checklist_items ALTER COLUMN is_required DROP DEFAULT;
