# API Документация

## Базовый URL

```
http://localhost:8080/api/v1
```

## Аутентификация

Большинство endpoints требуют JWT токен в заголовке:

```
Authorization: Bearer <access_token>
```

## Коды ответов

- `200 OK` — Успешный запрос
- `201 Created` — Ресурс создан
- `400 Bad Request` — Неверные данные запроса
- `401 Unauthorized` — Требуется аутентификация
- `403 Forbidden` — Недостаточно прав
- `404 Not Found` — Ресурс не найден
- `409 Conflict` — Конфликт данных
- `500 Internal Server Error` — Внутренняя ошибка сервера

## Формат ответа

### Успешный ответ

```json
{
  "success": true,
  "data": { ... }
}
```

### Ответ с пагинацией

```json
{
  "success": true,
  "data": [ ... ],
  "meta": {
    "page": 1,
    "limit": 20,
    "total": 100,
    "total_pages": 5
  }
}
```

### Ответ с ошибкой

```json
{
  "success": false,
  "error": "Описание ошибки"
}
```

---

## Аутентификация

### Регистрация

```http
POST /auth/register
Content-Type: application/json

{
  "email": "doctor@example.com",
  "password": "SecurePass123",
  "name": "Иванов Иван Иванович",
  "first_name": "Иван",
  "last_name": "Иванов",
  "middle_name": "Иванович",
  "phone": "+79001234567",
  "role": "DISTRICT_DOCTOR"
}
```

**Роли**: `ADMIN`, `SURGEON`, `DISTRICT_DOCTOR`, `PATIENT`

**Ответ**:
```json
{
  "success": true,
  "data": {
    "access_token": "eyJhbGc...",
    "refresh_token": "eyJhbGc...",
    "user": {
      "id": 1,
      "email": "doctor@example.com",
      "name": "Иванов Иван Иванович",
      "role": "DISTRICT_DOCTOR",
      "is_active": true
    }
  }
}
```

### Вход

```http
POST /auth/login
Content-Type: application/json

{
  "email": "doctor@example.com",
  "password": "SecurePass123"
}
```

### Обновление токена

```http
POST /auth/refresh
Content-Type: application/json

{
  "refresh_token": "eyJhbGc..."
}
```

### Выход

```http
POST /auth/logout
Authorization: Bearer <access_token>
```

### Текущий пользователь

```http
GET /auth/me
Authorization: Bearer <access_token>
```

---

## Пациенты

### Создать пациента

```http
POST /patients
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "first_name": "Петр",
  "last_name": "Петров",
  "middle_name": "Петрович",
  "date_of_birth": "1960-05-15",
  "phone": "+79001234567",
  "email": "patient@example.com",
  "address": "г. Москва, ул. Ленина, д. 1",
  "snils": "123-456-789 00",
  "passport_series": "1234",
  "passport_number": "567890",
  "policy_number": "1234567890123456",
  "diagnosis": "Катаракта правого глаза",
  "operation_type": "PHACO",
  "eye": "RIGHT",
  "district_id": 1,
  "notes": "Дополнительные заметки"
}
```

**Типы операций**: `PHACO`, `ECCE`, `ICCE`, `LASER`
**Глаз**: `RIGHT`, `LEFT`, `BOTH`

### Список пациентов

```http
GET /patients?page=1&limit=20&search=Петров&status=PREPARATION
Authorization: Bearer <access_token>
```

**Параметры**:
- `page` — номер страницы (по умолчанию 1)
- `limit` — количество на странице (по умолчанию 20, макс 100)
- `search` — поиск по ФИО
- `status` — фильтр по статусу

**Статусы пациента**:
- `NEW` — Новый
- `PREPARATION` — Подготовка
- `REVIEW_NEEDED` — Требуется проверка
- `APPROVED` — Одобрен
- `SURGERY_SCHEDULED` — Операция запланирована
- `COMPLETED` — Завершено
- `REJECTED` — Отклонён

### Получить пациента

```http
GET /patients/:id
Authorization: Bearer <access_token>
```

### Обновить пациента

```http
PUT /patients/:id
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "phone": "+79009999999",
  "email": "newemail@example.com",
  "notes": "Обновлённые заметки"
}
```

### Изменить статус

```http
PUT /patients/:id/status
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "status": "APPROVED",
  "comment": "Все анализы в норме"
}
```

### Публичный доступ

```http
GET /patients/public/:accessCode
```

Не требует аутентификации. Возвращает ограниченную информацию.

### Статистика

```http
GET /patients/dashboard
Authorization: Bearer <access_token>
```

Возвращает количество пациентов по статусам.

### Фильтрация по ролям

Dashboard endpoint применяет фильтрацию в зависимости от роли пользователя:

- **DISTRICT_DOCTOR**: Видит статистику только по своим пациентам (где doctor_id = user.ID)
- **SURGEON**: Видит статистику только по пациентам со статусом >= PENDING_REVIEW (PENDING_REVIEW, APPROVED, NEEDS_CORRECTION, SCHEDULED, COMPLETED)
- **ADMIN**: Видит статистику по всем пациентам

Эта фильтрация согласована с поведением List endpoint (/api/v1/patients).

---

## Чек-листы

### Получить чек-лист пациента

```http
GET /checklists/patient/:patientId
Authorization: Bearer <access_token>
```

### Создать пункт чек-листа

```http
POST /checklists
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "patient_id": 1,
  "name": "Консультация кардиолога",
  "description": "При наличии гипертонии или ИБС",
  "category": "Заключения",
  "is_required": true,
  "expires_in_days": 30
}
```

**Доступ**: Районный врач, хирург, администратор

### Обновить пункт чек-листа

```http
PUT /checklists/:id
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "status": "COMPLETED",
  "result": "В норме",
  "notes": "Дополнительные заметки"
}
```

**Статусы пункта**: `PENDING`, `IN_PROGRESS`, `COMPLETED`, `REJECTED`

### Проверить пункт (хирург)

```http
PUT /checklists/:id/review
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "status": "COMPLETED",
  "review_note": "Результаты приняты"
}
```

### Прогресс выполнения

```http
GET /checklists/patient/:patientId/progress
Authorization: Bearer <access_token>
```

**Ответ**:
```json
{
  "success": true,
  "data": {
    "total": 15,
    "completed": 10,
    "required": 12,
    "required_completed": 8,
    "percentage": 66.67
  }
}
```

### Автоматический переход статуса

**Важно**: При обновлении пунктов чек-листа система автоматически проверяет выполнение всех **обязательных** пунктов.

Когда все обязательные пункты отмечены как `COMPLETED`:
- Статус пациента автоматически меняется с `IN_PROGRESS` на `PENDING_REVIEW`
- Создается запись в истории статусов
- Отправляются уведомления хирургам о необходимости проверки

**Примечание**:
- Чек-лист содержит обязательные (`is_required: true`) и опциональные (`is_required: false`) пункты
- Автопереход срабатывает только при выполнении всех обязательных пунктов
- Опциональные пункты не влияют на автоматическую смену статуса
- Для операции PHACOEMULSIFICATION: 13 обязательных + 2 опциональных пункта

### Уведомления в Telegram

Пациенты получают автоматические уведомления в Telegram при работе с чек-листами:

**При создании пункта чек-листа**:
- Пациент получает уведомление о новом пункте, который необходимо выполнить
- Уведомление содержит название пункта, описание и срок выполнения (если указан)

**При обновлении статуса пункта**:
- Изменение статуса на `IN_PROGRESS` — уведомление о начале выполнения
- Изменение статуса на `COMPLETED` — уведомление о завершении пункта
- Изменение статуса на `REJECTED` — уведомление об отклонении с указанием причины

**При проверке хирургом**:
- Одобрение пункта — уведомление с комментарием хирурга (если указан)
- Отклонение пункта — уведомление с обязательным комментарием о причине отклонения и необходимых исправлениях

**Примечание**: Для получения уведомлений пациент должен быть зарегистрирован в Telegram-боте системы и связать свой аккаунт с профилем пациента.

---

## Медиафайлы

### Загрузить файл

```http
POST /media/upload
Authorization: Bearer <access_token>
Content-Type: multipart/form-data

file: <binary>
patient_id: 1
category: "analysis"
```

**Категории**: `analysis`, `document`, `photo`, `general`
**Допустимые типы**: JPG, PNG, PDF
**Максимальный размер**: 20 МБ

### Файлы пациента

```http
GET /media/patient/:patientId
Authorization: Bearer <access_token>
```

### Скачать файл

```http
GET /media/:id/download
Authorization: Bearer <access_token>
```

Возвращает presigned URL для скачивания.

### Миниатюра

```http
GET /media/:id/thumbnail
Authorization: Bearer <access_token>
```

Доступно только для изображений.

### Удалить файл

```http
DELETE /media/:id
Authorization: Bearer <access_token>
```

---

## Расчёт ИОЛ

### Рассчитать силу ИОЛ

```http
POST /iol/calculate
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "patient_id": 1,
  "eye": "RIGHT",
  "axial_length": 23.5,
  "keratometry1": 43.5,
  "keratometry2": 44.0,
  "acd": 3.2,
  "target_refraction": -0.5,
  "formula": "SRKT",
  "a_constant": 118.4
}
```

**Формулы**: `SRKT`, `HAIGIS`, `HOFFERQ`

**Ответ**:
```json
{
  "success": true,
  "data": {
    "id": 1,
    "patient_id": 1,
    "eye": "RIGHT",
    "iol_power": 21.5,
    "predicted_refraction": -0.48,
    "formula": "SRKT",
    "warnings": "",
    "created_at": "2026-02-26T10:00:00Z"
  }
}
```

### История расчётов

```http
GET /iol/patient/:patientId/history
Authorization: Bearer <access_token>
```

---

## Операции

### Запланировать операцию

```http
POST /surgeries
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "patient_id": 1,
  "scheduled_date": "2026-03-15",
  "notes": "Плановая операция"
}
```

### Список операций

```http
GET /surgeries?page=1&limit=20
Authorization: Bearer <access_token>
```

### Получить операцию

```http
GET /surgeries/:id
Authorization: Bearer <access_token>
```

### Обновить операцию

```http
PUT /surgeries/:id
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "scheduled_date": "2026-03-20",
  "status": "COMPLETED",
  "notes": "Операция прошла успешно"
}
```

**Статусы операции**: `SCHEDULED`, `IN_PROGRESS`, `COMPLETED`, `CANCELLED`

---

## Комментарии

### Создать комментарий

```http
POST /comments
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "patient_id": 1,
  "body": "Требуется повторный анализ крови",
  "is_urgent": true,
  "parent_id": null
}
```

### Комментарии пациента

```http
GET /comments/patient/:patientId
Authorization: Bearer <access_token>
```

### Отметить как прочитанные

```http
PUT /comments/patient/:patientId/read
Authorization: Bearer <access_token>
```

---

## Уведомления

### Список уведомлений

```http
GET /notifications?page=1&limit=20
Authorization: Bearer <access_token>
```

### Отметить как прочитанное

```http
PUT /notifications/:id/read
Authorization: Bearer <access_token>
```

### Отметить все как прочитанные

```http
PUT /notifications/read-all
Authorization: Bearer <access_token>
```

### Количество непрочитанных

```http
GET /notifications/unread-count
Authorization: Bearer <access_token>
```

---

## Печать

### Маршрутный лист

```http
GET /print/routing-sheet/:patientId
Authorization: Bearer <access_token>
```

Возвращает PDF файл.

### Отчёт по чек-листу

```http
GET /print/checklist-report/:patientId
Authorization: Bearer <access_token>
```

Возвращает PDF файл.

---

## Районы

### Создать район

```http
POST /districts
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "name": "Центральный район",
  "region": "Московская область",
  "code": "MSK-01",
  "timezone": "Europe/Moscow"
}
```

### Список районов

```http
GET /districts?page=1&limit=20&search=Центральный
Authorization: Bearer <access_token>
```

### Получить район

```http
GET /districts/:id
Authorization: Bearer <access_token>
```

### Обновить район

```http
PUT /districts/:id
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "name": "Новое название"
}
```

### Удалить район

```http
DELETE /districts/:id
Authorization: Bearer <access_token>
```

---

## Синхронизация

### Отправить изменения

```http
POST /sync/push
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "changes": [
    {
      "entity": "patient",
      "action": "update",
      "data": { ... }
    }
  ]
}
```

### Получить изменения

```http
GET /sync/pull?since=2026-02-26T10:00:00Z
Authorization: Bearer <access_token>
```

---

## Администрирование

### Список пользователей

```http
GET /admin/users
Authorization: Bearer <access_token>
```

Требуется роль `ADMIN`.

### Статистика системы

```http
GET /admin/stats
Authorization: Bearer <access_token>
```

**Ответ**:
```json
{
  "users": 50,
  "patients": 200,
  "districts": 10,
  "surgeries": 150
}
```

---

## Примеры использования

### Полный цикл работы с пациентом

1. **Регистрация врача**
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "doctor@example.com",
    "password": "SecurePass123",
    "name": "Иванов И.И.",
    "role": "DISTRICT_DOCTOR"
  }'
```

2. **Вход**
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "doctor@example.com",
    "password": "SecurePass123"
  }'
```

3. **Создание пациента**
```bash
curl -X POST http://localhost:8080/api/v1/patients \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "Петр",
    "last_name": "Петров",
    "operation_type": "PHACO",
    "eye": "RIGHT"
  }'
```

4. **Обновление чек-листа**
```bash
curl -X PUT http://localhost:8080/api/v1/checklists/1 \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "status": "COMPLETED",
    "result": "В норме"
  }'
```

5. **Планирование операции**
```bash
curl -X POST http://localhost:8080/api/v1/surgeries \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "patient_id": 1,
    "scheduled_date": "2026-03-15"
  }'
```

---

## Коды ошибок

| Сообщение | Причина | Решение |
|-----------|---------|---------|
| `отсутствует заголовок авторизации` | Не передан токен | Добавьте заголовок Authorization |
| `недействительный или просроченный токен` | Токен истёк | Обновите токен через /auth/refresh |
| `этот email уже зарегистрирован` | Email занят | Используйте другой email |
| `неверный email или пароль` | Неверные данные | Проверьте учётные данные |
| `пациент не найден` | Неверный ID | Проверьте ID пациента |
| `не все обязательные пункты чек-листа выполнены` | Чек-лист не завершён | Завершите обязательные пункты |

---

## Документация Scalar

Интерактивная документация API доступна по адресу:

```
http://localhost:8080/docs
```
