# State Machine & Batch Update — Changelog

## ✅ Выполнено

### 1. State Machine по спецификации
- **Новые статусы:**
  - `DRAFT` (черновик)
  - `IN_PROGRESS` (в процессе подготовки)
  - `PENDING_REVIEW` (ожидает проверки)
  - `APPROVED` (одобрено)
  - `NEEDS_CORRECTION` (требуется доработка)
  - `SCHEDULED` (операция запланирована)
  - `COMPLETED` (завершено)
  - `CANCELLED` (отменено)

- **Валидация переходов:** `ValidateStatusTransition()` в `internal/domain/patient.go`
- **Допустимые переходы:**
  - DRAFT → IN_PROGRESS, CANCELLED
  - IN_PROGRESS → PENDING_REVIEW, CANCELLED
  - PENDING_REVIEW → APPROVED, NEEDS_CORRECTION, CANCELLED
  - NEEDS_CORRECTION → IN_PROGRESS, CANCELLED
  - APPROVED → SCHEDULED, CANCELLED
  - SCHEDULED → COMPLETED, CANCELLED
  - COMPLETED, CANCELLED — финальные статусы

- **Миграция:** `migrations/000003_update_patient_statuses.up.sql`
  - Обновляет существующие статусы в БД
  - Обновляет историю статусов
  - Меняет default на DRAFT

- **Обновлены файлы:**
  - `internal/domain/patient.go` — новые статусы + валидация
  - `internal/service/patient_service.go` — валидация при смене статуса
  - `internal/service/checklist_service.go` — автопереход IN_PROGRESS → PENDING_REVIEW
  - `internal/service/surgery_service.go` — APPROVED → SCHEDULED
  - `internal/handler/patient_handler.go` — RBAC для хирургов
  - `cmd/seed/main.go` — тестовые данные с новыми статусами

### 2. Batch Update Endpoint
- **Endpoint:** `POST /api/v1/patients/:id/batch-update`
- **Функционал:**
  - Обновление данных пациента
  - Смена статуса
  - Массовое обновление чек-листа
  - Обнаружение конфликтов по timestamp
  - Автопереход статуса при завершении чек-листа

- **Request:**
```json
{
  "patient": { /* UpdatePatientRequest */ },
  "status": { "status": "APPROVED", "comment": "..." },
  "checklist": [
    { "id": 1, "status": "COMPLETED", "result": "...", "notes": "..." }
  ],
  "timestamp": "2026-02-26T12:00:00Z"
}
```

- **Response:**
```json
{
  "success": true,
  "patient": { /* Patient object */ },
  "conflicts": [],
  "updated_items": 5,
  "message": "Пакетное обновление выполнено успешно"
}
```

- **Файлы:**
  - `internal/domain/patient.go` — BatchUpdateRequest, BatchUpdateResponse
  - `internal/service/patient_service.go` — BatchUpdate()
  - `internal/handler/patient_handler.go` — BatchUpdate()
  - `internal/server/server.go` — роут добавлен

### 3. Публичный Endpoint по спеке
- **Старый путь:** `/api/v1/patients/public/:accessCode`
- **Новый путь:** `/api/public/status/:code` ✅
- **Без авторизации** — доступен всем

### 4. OperationType как таблица
- **Модель:** `OperationTypeModel` в `internal/domain/operation_type.go`
- **Таблица:** `operation_types` с полями:
  - code (PHACOEMULSIFICATION, ANTIGLAUCOMA, VITRECTOMY)
  - name, description
  - checklist_template (JSON)
  - is_active
- **Миграция:** `migrations/000004_create_operation_types.up.sql`
- **Seed данные:** 3 типа операций предзаполнены

**Примечание:** Старые константы `OperationType` (string) сохранены для обратной совместимости. Новая модель `OperationTypeModel` для динамического управления.

---

## 🔄 Следующие шаги

1. ✅ State Machine — DONE
2. ✅ Batch Update — DONE
3. ✅ Публичный endpoint — DONE
4. ✅ OperationType таблица — DONE
5. ⏳ Audit Logging — активировать
6. ⏳ Тесты — написать для критичных флоу

---

## 🚀 Как применить

```bash
# Применить миграции
psql -U user -d dbname -f migrations/000003_update_patient_statuses.up.sql
psql -U user -d dbname -f migrations/000004_create_operation_types.up.sql

# Или через migrate tool
migrate -path migrations -database "postgres://..." up

# Пересобрать
go build ./cmd/api

# Запустить
./api
```

## 📝 API Changes

### Новый endpoint
```
POST /api/v1/patients/:id/batch-update
Authorization: Bearer <token>
Content-Type: application/json

{
  "patient": { "diagnosis": "Updated diagnosis" },
  "status": { "status": "APPROVED", "comment": "Ready for surgery" },
  "checklist": [
    { "id": 1, "status": "COMPLETED", "result": "Normal" }
  ],
  "timestamp": "2026-02-26T12:00:00Z"
}
```

### Изменённый endpoint
```
GET /api/public/status/:code  (было: /api/v1/patients/public/:accessCode)
```

### Новые статусы
- DRAFT, IN_PROGRESS, PENDING_REVIEW, APPROVED, NEEDS_CORRECTION, SCHEDULED, COMPLETED, CANCELLED
