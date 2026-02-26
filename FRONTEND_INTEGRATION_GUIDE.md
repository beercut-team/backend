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
6. [Обработка ошибок](#обработка-ошибок)
7. [Примеры кода](#примеры-кода)

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
    role: 'DISTRICT_DOCTOR'
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
| Администратор | `ADMIN` | Полный доступ |

### Матрица прав доступа

| Действие | DISTRICT_DOCTOR | SURGEON | CALL_CENTER | ADMIN |
|----------|-----------------|---------|-------------|-------|
| Создать пациента | ✅ | ❌ | ❌ | ✅ |
| Просмотр пациентов | ✅ | ✅ | ✅ | ✅ |
| Обновить пациента | ✅ | ✅ | ❌ | ✅ |
| Удалить пациента | ❌ | ❌ | ❌ | ✅ |
| Сменить статус | ✅ | ✅ | ❌ | ✅ |
| Одобрить пациента | ❌ | ✅ | ❌ | ✅ |
| Создать район | ❌ | ❌ | ❌ | ✅ |

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
