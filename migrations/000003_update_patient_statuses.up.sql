-- Миграция для обновления статусов пациентов на новые значения по спецификации

-- Обновляем существующие статусы
UPDATE patients SET status = 'DRAFT' WHERE status = 'NEW';
UPDATE patients SET status = 'IN_PROGRESS' WHERE status = 'PREPARATION';
UPDATE patients SET status = 'PENDING_REVIEW' WHERE status = 'REVIEW_NEEDED';
UPDATE patients SET status = 'SCHEDULED' WHERE status = 'SURGERY_SCHEDULED';
UPDATE patients SET status = 'NEEDS_CORRECTION' WHERE status = 'REJECTED';

-- Обновляем историю статусов
UPDATE patient_status_histories SET from_status = 'DRAFT' WHERE from_status = 'NEW';
UPDATE patient_status_histories SET from_status = 'IN_PROGRESS' WHERE from_status = 'PREPARATION';
UPDATE patient_status_histories SET from_status = 'PENDING_REVIEW' WHERE from_status = 'REVIEW_NEEDED';
UPDATE patient_status_histories SET from_status = 'SCHEDULED' WHERE from_status = 'SURGERY_SCHEDULED';
UPDATE patient_status_histories SET from_status = 'NEEDS_CORRECTION' WHERE from_status = 'REJECTED';

UPDATE patient_status_histories SET to_status = 'DRAFT' WHERE to_status = 'NEW';
UPDATE patient_status_histories SET to_status = 'IN_PROGRESS' WHERE to_status = 'PREPARATION';
UPDATE patient_status_histories SET to_status = 'PENDING_REVIEW' WHERE to_status = 'REVIEW_NEEDED';
UPDATE patient_status_histories SET to_status = 'SCHEDULED' WHERE to_status = 'SURGERY_SCHEDULED';
UPDATE patient_status_histories SET to_status = 'NEEDS_CORRECTION' WHERE to_status = 'REJECTED';

-- Обновляем default значение для новых записей
ALTER TABLE patients ALTER COLUMN status SET DEFAULT 'DRAFT';
