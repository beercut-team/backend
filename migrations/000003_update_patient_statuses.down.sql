-- Откат миграции статусов пациентов

-- Возвращаем старые статусы
UPDATE patients SET status = 'NEW' WHERE status = 'DRAFT';
UPDATE patients SET status = 'PREPARATION' WHERE status = 'IN_PROGRESS';
UPDATE patients SET status = 'REVIEW_NEEDED' WHERE status = 'PENDING_REVIEW';
UPDATE patients SET status = 'SURGERY_SCHEDULED' WHERE status = 'SCHEDULED';
UPDATE patients SET status = 'REJECTED' WHERE status = 'NEEDS_CORRECTION';

-- Возвращаем историю статусов
UPDATE patient_status_histories SET from_status = 'NEW' WHERE from_status = 'DRAFT';
UPDATE patient_status_histories SET from_status = 'PREPARATION' WHERE from_status = 'IN_PROGRESS';
UPDATE patient_status_histories SET from_status = 'REVIEW_NEEDED' WHERE from_status = 'PENDING_REVIEW';
UPDATE patient_status_histories SET from_status = 'SURGERY_SCHEDULED' WHERE from_status = 'SCHEDULED';
UPDATE patient_status_histories SET from_status = 'REJECTED' WHERE from_status = 'NEEDS_CORRECTION';

UPDATE patient_status_histories SET to_status = 'NEW' WHERE to_status = 'DRAFT';
UPDATE patient_status_histories SET to_status = 'PREPARATION' WHERE to_status = 'IN_PROGRESS';
UPDATE patient_status_histories SET to_status = 'REVIEW_NEEDED' WHERE to_status = 'PENDING_REVIEW';
UPDATE patient_status_histories SET to_status = 'SURGERY_SCHEDULED' WHERE to_status = 'SCHEDULED';
UPDATE patient_status_histories SET to_status = 'REJECTED' WHERE to_status = 'NEEDS_CORRECTION';

-- Возвращаем старый default
ALTER TABLE patients ALTER COLUMN status SET DEFAULT 'NEW';
