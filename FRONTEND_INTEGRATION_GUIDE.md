# 🚀 Frontend Integration Guide - Oculus-Feldsher API

**Версия API:** 2.1.0
**Base URL:** `http://localhost:8080`
**Дата:** 2026-02-26

---

## 📋 Содержание

1. [Быстрый старт](#быстрый-старт)
2. [Аутентификация](#аутентификация)
3. [RBAC и права доступа](#rbac-и-права-доступа)
4. [State Machine пациентов](#state-machine-пациентов)
5. [Основные endpoints](#основные-endpoints)
6. [Медицинские стандарты](#медицинские-стандарты)
7. [Интеграции с внешними системами](#интеграции-с-внешними-системами)
8. [Обработка ошибок](#обработка-ошибок)
9. [Примеры кода](#примеры-кода)

---

## 🚀 Быстрый старт

### 1. Регистрация и вход

```javascript
// Регистрация нового пользователя
const registerResponse = await fetch('http://localhost:8080/api/v1/auth/register', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    email: 'doctor@example.com',
    password: 'securepassword',
    name: 'Иван Иванов',
    first_name: 'Иван',
    last_name: 'Иванов',
    phone: '+79991234567',
    role: 'DISTRICT_DOCTOR',
    district_id: 1  // ОБЯЗАТЕЛЬНО: ID района
  })
});

const { access_token, refresh_token, user } = await registerResponse.json();

// Вход
const loginResponse = await fetch('http://localhost:8080/api/v1/auth/login', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    email: 'doctor@example.com',
    password: 'securepassword'
  })
});

const { access_token, refresh_token, user } = await loginResponse.json();
```

### 2. Использование токена

```javascript
// Все защищённые запросы требуют токен в заголовке
const response = await fetch('http://localhost:8080/api/v1/patients', {
  headers: {
    'Authorization': `Bearer ${access_token}`
  }
});
```

---

## 🔐 Аутентификация

### Endpoints

| Endpoint | Метод | Описание |
|----------|-------|----------|
| `/api/v1/auth/register` | POST | Регистрация нового пользователя |
| `/api/v1/auth/login` | POST | Вход в систему |
| `/api/v1/auth/me` | GET | Получить текущего пользователя |
| `/api/v1/auth/refresh` | POST | Обновить access token |
| `/api/v1/auth/logout` | POST | Выход из системы |

### Refresh Token Flow

```javascript
// Когда access_token истекает (обычно через 1 час)
const refreshResponse = await fetch('http://localhost:8080/api/v1/auth/refresh', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    refresh_token: refresh_token
  })
});

const { access_token: newAccessToken } = await refreshResponse.json();
```

### Автоматическое обновление токена

```javascript
async function fetchWithAuth(url, options = {}) {
  let response = await fetch(url, {
    ...options,
    headers: {
      ...options.headers,
      'Authorization': `Bearer ${getAccessToken()}`
    }
  });

  // Если 401, попробовать обновить токен
  if (response.status === 401) {
    const newToken = await refreshAccessToken();
    response = await fetch(url, {
      ...options,
      headers: {
        ...options.headers,
        'Authorization': `Bearer ${newToken}`
      }
    });
  }

  return response;
}
```

---

## 👥 RBAC и права доступа

### Роли в системе

| Роль | Код | Описание |
|------|-----|----------|
| Районный врач | `DISTRICT_DOCTOR` | Создаёт и ведёт пациентов |
| Хирург | `SURGEON` | Одобряет пациентов, планирует операции |
| Колл-центр | `CALL_CENTER` | Только просмотр пациентов |
| Пациент | `PATIENT` | Просмотр своих данных, создание комментариев |
| Администратор | `ADMIN` | Полный доступ |

### Вход для пациентов

Пациенты входят через специальный endpoint с кодом доступа:

```javascript
const patientLoginResponse = await fetch('http://localhost:8080/api/v1/auth/patient-login', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    access_code: 'a1b2c3d4' // Код из карты пациента
  })
});

const { access_token, refresh_token, user } = await patientLoginResponse.json();
// user.role === "PATIENT"
```

### Матрица прав доступа

| Действие | DISTRICT_DOCTOR | SURGEON | CALL_CENTER | PATIENT | ADMIN |
|----------|-----------------|---------|-------------|---------|-------|
| Создать пациента | ✅ | ❌ | ❌ | ❌ | ✅ |
| Просмотр своих данных | ✅ | ✅ | ✅ | ✅ | ✅ |
| Просмотр всех пациентов | ✅ | ✅ | ✅ | ❌ | ✅ |
| Обновить пациента | ✅ | ✅ | ❌ | ❌ | ✅ |
| Удалить пациента | ❌ | ❌ | ❌ | ❌ | ✅ |
| Сменить статус | ✅ | ✅ | ❌ | ❌ | ✅ |
| Одобрить пациента | ❌ | ✅ | ❌ | ❌ | ✅ |
| Создать комментарий | ✅ | ✅ | ❌ | ✅ | ✅ |
| Просмотр чек-листа | ✅ | ✅ | ✅ | ✅ | ✅ |
| Создать район | ❌ | ❌ | ❌ | ❌ | ✅ |

### Проверка прав на фронтенде

```javascript
function canCreatePatient(userRole) {
  return ['DISTRICT_DOCTOR', 'ADMIN'].includes(userRole);
}

function canDeletePatient(userRole) {
  return userRole === 'ADMIN';
}

function canApprovePatient(userRole) {
  return ['SURGEON', 'ADMIN'].includes(userRole);
}

function canCreateComment(userRole) {
  return ['DISTRICT_DOCTOR', 'SURGEON', 'PATIENT', 'ADMIN'].includes(userRole);
}

function canViewAllPatients(userRole) {
  return ['DISTRICT_DOCTOR', 'SURGEON', 'CALL_CENTER', 'ADMIN'].includes(userRole);
}

function isPatient(userRole) {
  return userRole === 'PATIENT';
}
```

---

## 🏥 Мобильное приложение для пациентов

### Вход пациента

Пациенты входят через код доступа (access_code), который они получают от врача:

```javascript
async function patientLogin(accessCode) {
  const response = await fetch('http://localhost:8080/api/v1/auth/patient-login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      access_code: accessCode
    })
  });

  if (!response.ok) {
    throw new Error('Неверный код доступа');
  }

  const { access_token, refresh_token, user } = await response.json();

  // Сохранить токены
  localStorage.setItem('access_token', access_token);
  localStorage.setItem('refresh_token', refresh_token);
  localStorage.setItem('user', JSON.stringify(user));

  return user;
}
```

### Что может делать пациент

**Просмотр своих данных:**
```javascript
// Получить информацию о себе
const response = await fetchWithAuth('http://localhost:8080/api/v1/auth/me');
const { data: patient } = await response.json();

// Просмотр своего статуса
const statusResponse = await fetchWithAuth(`http://localhost:8080/api/v1/patients/${patient.id}`);
const { data: patientData } = await statusResponse.json();
```

**Просмотр чек-листа:**
```javascript
// Получить свой чек-лист
const checklistResponse = await fetchWithAuth(
  `http://localhost:8080/api/v1/checklists/patient/${patient.id}`
);
const { data: checklist } = await checklistResponse.json();

// Прогресс подготовки
const progressResponse = await fetchWithAuth(
  `http://localhost:8080/api/v1/checklists/patient/${patient.id}/progress`
);
const { data: progress } = await progressResponse.json();
// progress = { completed_count: 10, total_count: 15, percentage: 66.67 }
```

**Добавление пунктов в чек-лист (районный врач или хирург):**
```javascript
// Врач или хирург может добавить дополнительные обследования к стандартному чек-листу
async function addChecklistItem(patientId, itemData) {
  const response = await fetchWithAuth('http://localhost:8080/api/v1/checklists', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      patient_id: patientId,
      name: itemData.name,                    // Обязательно: "Консультация кардиолога"
      description: itemData.description,      // Опционально: "При наличии гипертонии"
      category: itemData.category,            // Опционально: "Заключения"
      is_required: itemData.isRequired,       // Опционально: true/false
      expires_in_days: itemData.expiresInDays // Опционально: 30
    })
  });

  return await response.json();
}

// Пример использования
await addChecklistItem(patientId, {
  name: "Консультация кардиолога",
  description: "При наличии гипертонии или ИБС",
  category: "Заключения",
  isRequired: true,
  expiresInDays: 30
});
```

**Отметка выполнения пункта чек-листа:**
```javascript
// Врач отмечает пункт как выполненный когда пациент приносит результаты
async function markChecklistItemCompleted(itemId, result, notes) {
  const response = await fetchWithAuth(`http://localhost:8080/api/v1/checklists/${itemId}`, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      status: 'COMPLETED',
      result: result,  // Опционально: "Гемоглобин 140 г/л"
      notes: notes     // Опционально: "Все показатели в норме"
    })
  });

  return await response.json();
}
```

**⚠️ ВАЖНО: Автоматический переход статуса**

При обновлении пунктов чек-листа система автоматически проверяет выполнение всех **обязательных** пунктов:

- Когда все обязательные пункты (`is_required: true`) отмечены как `COMPLETED`, статус пациента **автоматически** меняется с `IN_PROGRESS` на `PENDING_REVIEW`
- Опциональные пункты (`is_required: false`) **не влияют** на автопереход
- Автопереход происходит сразу после обновления любого пункта чек-листа
- Создается запись в истории статусов с комментарием "Все обязательные пункты чек-листа выполнены"
- Отправляются уведомления хирургам о необходимости проверки

**Пример для UI:**
```javascript
// После обновления пункта чек-листа, проверьте статус пациента
async function updateChecklistAndRefresh(itemId, data) {
  // Обновить пункт
  await markChecklistItemCompleted(itemId, data.result, data.notes);

  // Перезагрузить данные пациента, т.к. статус мог измениться
  const patientResponse = await fetchWithAuth(`http://localhost:8080/api/v1/patients/${patientId}`);
  const { data: updatedPatient } = await patientResponse.json();

  // Проверить, изменился ли статус
  if (updatedPatient.status === 'PENDING_REVIEW') {
    // Показать уведомление пользователю
    showNotification('Все обязательные пункты выполнены! Пациент отправлен на проверку хирургу.');
  }

  return updatedPatient;
}
```

**📱 ВАЖНО: Уведомления пациентам в Telegram**

Все операции с чек-листом автоматически отправляют уведомления пациенту через Telegram бота:

**При создании пункта чек-листа** (`POST /api/v1/checklists`):
- Пациент получает уведомление о новом пункте, который необходимо выполнить
- Уведомление содержит название пункта, описание и срок выполнения (если указан)

**При обновлении статуса пункта** (`PATCH /api/v1/checklists/:id`):
- `IN_PROGRESS` — уведомление о начале выполнения пункта
- `COMPLETED` — уведомление о завершении пункта
- `REJECTED` — уведомление об отклонении с указанием причины

**При проверке хирургом** (`PUT /api/v1/checklists/:id/review`):
- Одобрение пункта — уведомление с комментарием хирурга (если указан)
- Отклонение пункта — уведомление с обязательным комментарием о причине отклонения и необходимых исправлениях

**При завершении всех обязательных пунктов**:
- Когда все обязательные пункты выполнены и статус меняется на `PENDING_REVIEW`, пациент получает уведомление о том, что подготовка завершена

**Примечание**: Уведомления отправляются автоматически на стороне бэкенда через Celery задачи. Фронтенд не требует дополнительных действий для отправки уведомлений — достаточно выполнить стандартные API запросы для работы с чек-листом. Для получения уведомлений пациент должен быть зарегистрирован в Telegram-боте системы.

**Создание комментариев:**
```javascript
// Пациент может задавать вопросы врачу
async function askDoctor(patientId, question) {
  const response = await fetchWithAuth('http://localhost:8080/api/v1/comments', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      patient_id: patientId,
      body: question
    })
  });

  return await response.json();
}

// Просмотр ответов врача
const commentsResponse = await fetchWithAuth(
  `http://localhost:8080/api/v1/comments/patient/${patient.id}`
);
const { data: comments } = await commentsResponse.json();
```

**Уведомления:**

**ВАЖНО:** Уведомления получают врачи, хирурги и пациенты. Пациенты получают уведомления через Telegram бот при работе с чек-листами. Уведомления для врачей и хирургов содержат имя пациента для контекста.

```javascript
// Получить уведомления текущего пользователя
const notificationsResponse = await fetchWithAuth(
  'http://localhost:8080/api/v1/notifications'
);
const { data: notifications } = await notificationsResponse.json();

// Количество непрочитанных
const unreadResponse = await fetchWithAuth(
  'http://localhost:8080/api/v1/notifications/unread-count'
);
const { data: { count } } = await unreadResponse.json();

// Пример уведомления:
// {
//   "id": 1,
//   "user_id": 2,  // ID врача или хирурга
//   "type": "STATUS_CHANGE",
//   "title": "Статус пациента изменен",
//   "body": "Пациент Алексеева Туяра: статус изменен на Одобрено, готов к операции",
//   "entity_type": "patient",
//   "entity_id": 1,  // ID пациента
//   "is_read": false,
//   "created_at": "2026-02-26T15:30:00Z"
// }
```

**Кто получает уведомления:**

| Событие | Лечащий врач | Хирург | Пациент |
|---------|--------------|--------|---------|
| Изменение статуса | ✅ Всегда | ✅ Если назначен | ❌ Нет |
| Изменение диагноза | ✅ Всегда | ❌ Нет | ❌ Нет |
| Новый комментарий | ✅ Если упомянут | ✅ Если упомянут | ❌ Нет |

**Типы уведомлений:**
- `STATUS_CHANGE` - изменение статуса пациента
- `COMMENT_ADDED` - новый комментарий
- `CHECKLIST_UPDATED` - обновление чек-листа
- `SURGERY_SCHEDULED` - назначена дата операции
```

### Публичный статус (без авторизации)

Для QR-кодов и публичных ссылок:

```javascript
// Любой может посмотреть статус по коду доступа
const publicResponse = await fetch(
  `http://localhost:8080/api/public/status/${accessCode}`
);
const { data: publicStatus } = await publicResponse.json();

// publicStatus содержит:
// - patient_name: "Иван И." (скрыто отчество)
// - status: "SCHEDULED"
// - surgery_date: "2026-03-15T10:00:00Z"
// - checklist_progress: { completed: 12, total: 15 }
```

### Пример мобильного приложения

```javascript
// PatientApp.jsx
function PatientApp() {
  const [patient, setPatient] = useState(null);
  const [checklist, setChecklist] = useState([]);
  const [progress, setProgress] = useState(null);

  useEffect(() => {
    async function loadPatientData() {
      // Получить данные пациента
      const meResponse = await fetchWithAuth('http://localhost:8080/api/v1/auth/me');
      const { data: patientData } = await meResponse.json();
      setPatient(patientData);

      // Загрузить чек-лист
      const checklistResponse = await fetchWithAuth(
        `http://localhost:8080/api/v1/checklists/patient/${patientData.id}`
      );
      const { data: checklistData } = await checklistResponse.json();
      setChecklist(checklistData);

      // Загрузить прогресс
      const progressResponse = await fetchWithAuth(
        `http://localhost:8080/api/v1/checklists/patient/${patientData.id}/progress`
      );
      const { data: progressData } = await progressResponse.json();
      setProgress(progressData);
    }

    loadPatientData();
  }, []);

  if (!patient) return <div>Загрузка...</div>;

  return (
    <div>
      <h1>Привет, {patient.first_name}!</h1>
      <StatusCard status={patient.status} />
      <ProgressBar
        completed={progress?.completed_count}
        total={progress?.total_count}
      />
      <ChecklistItems items={checklist} />
      <CommentsSection patientId={patient.id} />
    </div>
  );
}
```

---

## 🔄 State Machine пациентов

### Диаграмма переходов

```
DRAFT → IN_PROGRESS → PENDING_REVIEW → APPROVED → SCHEDULED → COMPLETED
                            ↓
                      NEEDS_CORRECTION
                            ↓
                       IN_PROGRESS

Из любого статуса → CANCELLED
```

### Валидные переходы

| Из статуса | В статус | Кто может |
|------------|----------|-----------|
| DRAFT | IN_PROGRESS | DISTRICT_DOCTOR, ADMIN |
| IN_PROGRESS | PENDING_REVIEW | DISTRICT_DOCTOR, ADMIN |
| PENDING_REVIEW | APPROVED | SURGEON, ADMIN |
| PENDING_REVIEW | NEEDS_CORRECTION | SURGEON, ADMIN |
| NEEDS_CORRECTION | IN_PROGRESS | DISTRICT_DOCTOR, ADMIN |
| APPROVED | SCHEDULED | SURGEON, ADMIN |
| SCHEDULED | COMPLETED | SURGEON, ADMIN |
| Любой | CANCELLED | SURGEON, ADMIN |

### Смена статуса

```javascript
async function changePatientStatus(patientId, newStatus, comment) {
  const response = await fetchWithAuth(
    `http://localhost:8080/api/v1/patients/${patientId}/status`,
    {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        status: newStatus,
        comment: comment
      })
    }
  );

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error); // "недопустимый переход статуса: X → Y"
  }

  return await response.json();
}
```

### Валидация на фронтенде

```javascript
const VALID_TRANSITIONS = {
  'DRAFT': ['IN_PROGRESS', 'CANCELLED'],
  'IN_PROGRESS': ['PENDING_REVIEW', 'CANCELLED'],
  'PENDING_REVIEW': ['APPROVED', 'NEEDS_CORRECTION', 'CANCELLED'],
  'APPROVED': ['SCHEDULED', 'CANCELLED'],
  'NEEDS_CORRECTION': ['IN_PROGRESS', 'CANCELLED'],
  'SCHEDULED': ['COMPLETED', 'CANCELLED'],
  'COMPLETED': [],
  'CANCELLED': []
};

function canTransitionTo(currentStatus, newStatus) {
  return VALID_TRANSITIONS[currentStatus]?.includes(newStatus) || false;
}

function getAvailableStatuses(currentStatus, userRole) {
  const validStatuses = VALID_TRANSITIONS[currentStatus] || [];

  // Фильтруем по правам пользователя
  if (userRole === 'DISTRICT_DOCTOR') {
    return validStatuses.filter(s =>
      ['IN_PROGRESS', 'PENDING_REVIEW'].includes(s)
    );
  }

  if (userRole === 'SURGEON') {
    return validStatuses; // Хирург может все переходы
  }

  return [];
}
```

---

## 📡 Основные Endpoints

### Пациенты

#### Создать пациента

```javascript
POST /api/v1/patients
Authorization: Bearer {token}
Content-Type: application/json

{
  "first_name": "Иван",
  "last_name": "Иванов",
  "middle_name": "Петрович",
  "birth_date": "1980-01-15",
  "phone": "+79991234567",
  "email": "patient@example.com",
  "district_id": 1,
  "diagnosis": "Катаракта правого глаза",
  "operation_type": "PHACOEMULSIFICATION",
  "eye": "OD"
}

// Response
{
  "success": true,
  "data": {
    "id": 1,
    "access_code": "a1b2c3d4",
    "status": "IN_PROGRESS",
    ...
  }
}
```

#### Получить список пациентов

```javascript
GET /api/v1/patients?page=1&limit=20&status=IN_PROGRESS&search=Иванов
Authorization: Bearer {token}

// Response
{
  "success": true,
  "data": {
    "items": [...],
    "total": 100,
    "page": 1,
    "limit": 20
  }
}
```

#### Batch Update (для оффлайн-режима)

```javascript
POST /api/v1/patients/{id}/batch-update
Authorization: Bearer {token}
Content-Type: application/json

{
  "patient": {
    "diagnosis": "Обновлённый диагноз",
    "notes": "Дополнительные заметки"
  },
  "status": {
    "status": "PENDING_REVIEW",
    "comment": "Готов к проверке"
  },
  "checklist_updates": [
    {
      "id": 1,
      "status": "COMPLETED",
      "notes": "Анализы сданы"
    }
  ],
  "timestamp": "2026-02-26T12:00:00Z"
}

// Response
{
  "success": true,
  "data": {
    "updated_items": 3,
    "conflicts": []
  }
}
```

### Публичный статус (без авторизации)

```javascript
GET /api/public/status/{access_code}

// Response
{
  "success": true,
  "data": {
    "patient_name": "Иван И.",
    "status": "SCHEDULED",
    "surgery_date": "2026-03-15T10:00:00Z",
    "checklist_progress": {
      "completed": 12,
      "total": 15
    }
  }
}
```

---

## 🏥 Медицинские стандарты

### Поиск кодов диагнозов ICD-10

```javascript
GET /api/v1/medical-codes/icd10/search?q=катаракта
Authorization: Bearer {token}

// Response
{
  "success": true,
  "data": [
    {
      "code": "H25.1",
      "display": "Старческая ядерная катаракта",
      "system": "http://hl7.org/fhir/sid/icd-10"
    },
    {
      "code": "H25.0",
      "display": "Старческая начальная катаракта",
      "system": "http://hl7.org/fhir/sid/icd-10"
    }
  ],
  "count": 2
}
```

### Поиск кодов процедур SNOMED-CT

```javascript
GET /api/v1/medical-codes/snomed/search?q=факоэмульсификация
Authorization: Bearer {token}

// Response
{
  "success": true,
  "data": [
    {
      "code": "397544007",
      "display": "Факоэмульсификация катаракты",
      "system": "http://snomed.info/sct"
    }
  ],
  "count": 1
}
```

### Поиск кодов наблюдений LOINC

```javascript
GET /api/v1/medical-codes/loinc/search?q=длина
Authorization: Bearer {token}

// Response
{
  "success": true,
  "data": [
    {
      "code": "79893-4",
      "display": "Длина оси глаза",
      "system": "http://loinc.org"
    }
  ],
  "count": 1
}
```

### Обновление медицинских метаданных пациента

```javascript
POST /api/v1/patients/{id}/medical-metadata
Authorization: Bearer {token}
Content-Type: application/json

{
  "diagnosis_codes": [
    {
      "code": "H25.1",
      "display": "Старческая ядерная катаракта",
      "system": "http://hl7.org/fhir/sid/icd-10"
    }
  ],
  "procedure_codes": [
    {
      "code": "397544007",
      "display": "Факоэмульсификация катаракты",
      "system": "http://snomed.info/sct"
    }
  ],
  "observations": [
    {
      "code": "79893-4",
      "display": "Длина оси глаза",
      "system": "http://loinc.org",
      "value": "23.5",
      "unit": "mm",
      "observed_at": "2026-02-26T10:00:00Z"
    }
  ]
}

// Response
{
  "success": true,
  "message": "Медицинские метаданные обновлены"
}
```

---

## 🔗 Интеграции с внешними системами

### ЕМИАС (Москва)

#### Экспорт пациента в ЕМИАС

```javascript
POST /api/v1/integrations/emias/patients/{id}/export
Authorization: Bearer {token}

// Response
{
  "success": true,
  "external_id": "EMIAS-a1b2c3d4",
  "message": "Пациент успешно экспортирован в ЕМИАС"
}
```

#### Создание случая в ЕМИАС

```javascript
POST /api/v1/integrations/emias/patients/{id}/case
Authorization: Bearer {token}
Content-Type: application/json

{
  "surgery_date": "2026-03-15",
  "procedure_code": "397544007",
  "diagnosis_code": "H25.1"
}

// Response
{
  "success": true,
  "external_id": "CASE-e5f6g7h8",
  "message": "Случай успешно создан в ЕМИАС"
}
```

#### Получение статуса синхронизации с ЕМИАС

```javascript
GET /api/v1/integrations/emias/patients/{id}/status
Authorization: Bearer {token}

// Response
{
  "success": true,
  "patient_id": "EMIAS-a1b2c3d4",
  "case_id": "CASE-e5f6g7h8",
  "status": "synced",
  "last_sync_at": "2026-02-26T12:00:00Z"
}
```

### РИАМС (Региональные системы)

#### Получение списка регионов

```javascript
GET /api/v1/integrations/riams/regions
Authorization: Bearer {token}

// Response
{
  "success": true,
  "data": [
    { "code": "77", "name": "Москва" },
    { "code": "78", "name": "Санкт-Петербург" },
    { "code": "50", "name": "Московская область" }
  ],
  "count": 10
}
```

#### Экспорт пациента в РИАМС

```javascript
POST /api/v1/integrations/riams/patients/{id}/export
Authorization: Bearer {token}
Content-Type: application/json

{
  "region_code": "77"
}

// Response
{
  "success": true,
  "external_id": "RIAMS-77-a1b2c3d4",
  "message": "Пациент успешно экспортирован в РИАМС"
}
```

#### Получение статуса синхронизации с РИАМС

```javascript
GET /api/v1/integrations/riams/patients/{id}/status
Authorization: Bearer {token}

// Response
{
  "success": true,
  "patient_id": "RIAMS-77-a1b2c3d4",
  "region_code": "77",
  "status": "synced",
  "last_sync_at": "2026-02-26T12:00:00Z"
}
```

### Валидация перед экспортом

Перед экспортом в ЕМИАС или РИАМС система автоматически проверяет:
- ФИО пациента (обязательно)
- Дата рождения (обязательно)
- СНИЛС (предупреждение, если отсутствует)
- Полис ОМС (предупреждение, если отсутствует)

Если валидация не пройдена, API вернёт ошибку с деталями:

```javascript
{
  "success": false,
  "error": "валидация не пройдена",
  "errors": ["Дата рождения обязательна"],
  "warnings": ["СНИЛС не указан", "Полис ОМС не указан"]
}
```

---

## ⚠️ Обработка ошибок

### Формат ошибок

```javascript
{
  "success": false,
  "error": "описание ошибки"
}
```

### HTTP коды

| Код | Значение | Действие |
|-----|----------|----------|
| 200 | OK | Успешно |
| 400 | Bad Request | Проверить данные запроса |
| 401 | Unauthorized | Обновить токен или войти заново |
| 403 | Forbidden | Недостаточно прав |
| 404 | Not Found | Ресурс не найден |
| 409 | Conflict | Конфликт данных (batch update) |
| 500 | Server Error | Ошибка сервера |

### Обработка ошибок

```javascript
async function handleApiCall(apiFunction) {
  try {
    const response = await apiFunction();

    if (!response.ok) {
      const error = await response.json();

      switch (response.status) {
        case 401:
          // Перенаправить на логин
          redirectToLogin();
          break;
        case 403:
          showError('Недостаточно прав для этого действия');
          break;
        case 404:
          showError('Ресурс не найден');
          break;
        default:
          showError(error.error || 'Произошла ошибка');
      }

      return null;
    }

    return await response.json();
  } catch (error) {
    showError('Ошибка сети. Проверьте подключение.');
    return null;
  }
}
```

---

## 💻 Примеры кода

### React Hook для работы с API

```javascript
import { useState, useEffect } from 'react';

function usePatients(filters = {}) {
  const [patients, setPatients] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    async function fetchPatients() {
      try {
        setLoading(true);
        const params = new URLSearchParams(filters);
        const response = await fetchWithAuth(
          `http://localhost:8080/api/v1/patients?${params}`
        );

        if (!response.ok) throw new Error('Failed to fetch');

        const data = await response.json();
        setPatients(data.data.items);
      } catch (err) {
        setError(err.message);
      } finally {
        setLoading(false);
      }
    }

    fetchPatients();
  }, [JSON.stringify(filters)]);

  return { patients, loading, error };
}

// Использование
function PatientList() {
  const { patients, loading, error } = usePatients({
    status: 'IN_PROGRESS'
  });

  if (loading) return <div>Загрузка...</div>;
  if (error) return <div>Ошибка: {error}</div>;

  return (
    <ul>
      {patients.map(p => (
        <li key={p.id}>{p.first_name} {p.last_name}</li>
      ))}
    </ul>
  );
}
```

### Оффлайн-режим с синхронизацией

```javascript
class OfflineQueue {
  constructor() {
    this.queue = JSON.parse(localStorage.getItem('offline_queue') || '[]');
  }

  add(action) {
    this.queue.push({
      ...action,
      timestamp: new Date().toISOString()
    });
    this.save();
  }

  async sync() {
    const results = [];

    for (const action of this.queue) {
      try {
        if (action.type === 'batch_update') {
          const response = await fetchWithAuth(
            `http://localhost:8080/api/v1/patients/${action.patientId}/batch-update`,
            {
              method: 'POST',
              headers: { 'Content-Type': 'application/json' },
              body: JSON.stringify(action.data)
            }
          );

          if (response.ok) {
            results.push({ success: true, action });
          } else {
            const error = await response.json();
            results.push({ success: false, action, error });
          }
        }
      } catch (error) {
        results.push({ success: false, action, error: error.message });
      }
    }

    // Удалить успешные из очереди
    this.queue = this.queue.filter((_, i) => !results[i].success);
    this.save();

    return results;
  }

  save() {
    localStorage.setItem('offline_queue', JSON.stringify(this.queue));
  }
}

// Использование
const offlineQueue = new OfflineQueue();

// При оффлайн-изменениях
offlineQueue.add({
  type: 'batch_update',
  patientId: 1,
  data: { patient: { diagnosis: 'Updated' } }
});

// При восстановлении связи
window.addEventListener('online', async () => {
  const results = await offlineQueue.sync();
  console.log('Синхронизация завершена:', results);
});
```

---

## 📚 Дополнительные ресурсы

- **OpenAPI Schema:** `/openapi.json`
- **Scalar Docs:** `http://localhost:8080/docs`
- **Admin Panel:** `http://localhost:8080/admin`
- **Patient Portal:** `http://localhost:8080/patient`

---

## ✅ Чеклист интеграции

- [ ] Реализована аутентификация (login, register, refresh)
- [ ] Настроен автоматический refresh токенов
- [ ] Реализована проверка RBAC на фронтенде
- [ ] Добавлена валидация state machine переходов
- [ ] Реализован оффлайн-режим с batch-update
- [ ] Настроена обработка ошибок
- [ ] Добавлены loading states
- [ ] Реализована пагинация списков
- [ ] Добавлены фильтры и поиск
- [ ] Протестированы все основные флоу

---

**Готово к использованию! 🚀**

Все endpoints протестированы и работают. Backend готов к интеграции с фронтендом.
