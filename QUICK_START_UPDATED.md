# Quick Start Guide — Updated Backend

## Что изменилось

### Новые статусы пациентов
Старые статусы заменены на новые по спецификации:
- `NEW` → `DRAFT`
- `PREPARATION` → `IN_PROGRESS`
- `REVIEW_NEEDED` → `PENDING_REVIEW`
- `SURGERY_SCHEDULED` → `SCHEDULED`
- `REJECTED` → `NEEDS_CORRECTION`
- Добавлен: `CANCELLED`

### Новый endpoint для оффлайн-режима
```bash
POST /api/v1/patients/:id/batch-update
```

### Изменённый публичный endpoint
```bash
GET /api/public/status/:code  # было: /api/v1/patients/public/:accessCode
```

## Быстрый старт

### 1. Применить миграции

```bash
# Если используете migrate tool
migrate -path migrations -database "postgres://user:pass@localhost:5432/dbname?sslmode=disable" up

# Или напрямую через psql
psql -U postgres -d oculus_feldsher < migrations/000003_update_patient_statuses.up.sql
psql -U postgres -d oculus_feldsher < migrations/000004_create_operation_types.up.sql
```

### 2. Пересобрать и запустить

```bash
# Собрать
go build -o api ./cmd/api
go build -o seed ./cmd/seed

# Заполнить тестовыми данными (опционально)
./seed

# Запустить
./api
```

### 3. Проверить работу

```bash
# Проверить health
curl http://localhost:8080/api/v1/ping

# Проверить документацию
open http://localhost:8080/docs

# Проверить публичный endpoint
curl http://localhost:8080/api/public/status/a1b2c3d4
```

## Примеры использования новых фич

### Batch Update (оффлайн-режим)

```bash
curl -X POST http://localhost:8080/api/v1/patients/1/batch-update \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "patient": {
      "diagnosis": "Обновлённый диагноз"
    },
    "status": {
      "status": "APPROVED",
      "comment": "Готов к операции"
    },
    "checklist": [
      {
        "id": 1,
        "status": "COMPLETED",
        "result": "Норма"
      }
    ],
    "timestamp": "2026-02-26T12:00:00Z"
  }'
```

### Смена статуса с валидацией

```bash
# Допустимый переход: IN_PROGRESS → PENDING_REVIEW
curl -X POST http://localhost:8080/api/v1/patients/1/status \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "status": "PENDING_REVIEW",
    "comment": "Все анализы готовы"
  }'

# Недопустимый переход вернёт ошибку
curl -X POST http://localhost:8080/api/v1/patients/1/status \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "status": "COMPLETED",
    "comment": "Попытка пропустить этапы"
  }'
# Ответ: {"error": "недопустимый переход статуса: IN_PROGRESS → COMPLETED"}
```

## State Machine Flow

```
DRAFT
  ↓
IN_PROGRESS
  ↓
PENDING_REVIEW
  ↓         ↓
APPROVED  NEEDS_CORRECTION
  ↓         ↓
SCHEDULED   ← (возврат в IN_PROGRESS)
  ↓
COMPLETED

Из любого статуса → CANCELLED
```

## Тестирование

```bash
# Запустить все тесты
go test ./...

# Только state machine тесты
go test ./internal/domain -v -run TestValidateStatusTransition

# Только IOL формулы
go test ./internal/service/formulas -v
```

## Audit Logging

Все мутации (POST/PUT/PATCH/DELETE) автоматически логируются в таблицу `audit_logs`:
- User ID
- Action (CREATE/UPDATE/DELETE)
- Entity (patients, districts, etc.)
- Old/New values (JSON)
- IP address
- Timestamp

Просмотр логов:
```sql
SELECT * FROM audit_logs
WHERE entity = 'patients' AND entity_id = 1
ORDER BY created_at DESC;
```

## Troubleshooting

### Миграции не применяются
```bash
# Проверить текущую версию
migrate -path migrations -database "postgres://..." version

# Откатить последнюю миграцию
migrate -path migrations -database "postgres://..." down 1

# Применить заново
migrate -path migrations -database "postgres://..." up
```

### Старые статусы в БД
Миграция `000003_update_patient_statuses.up.sql` автоматически обновляет существующие записи.

### Тесты не проходят
```bash
# Очистить кэш
go clean -testcache

# Запустить с verbose
go test ./... -v
```

## Полезные ссылки

- API документация: http://localhost:8080/docs
- OpenAPI схема: http://localhost:8080/openapi.json
- Admin панель: http://localhost:8080/admin
- Публичный статус: http://localhost:8080/patient

## Что дальше?

1. Обновить фронтенд для работы с новыми статусами
2. Протестировать batch-update в оффлайн-режиме
3. Настроить мониторинг audit logs
4. Добавить больше тестов (опционально)
