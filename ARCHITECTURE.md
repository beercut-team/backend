# Архитектура проекта

## Обзор

Проект построен на принципах Clean Architecture с чёткой изоляцией слоёв и зависимостей.

## Структура слоёв

```
┌─────────────────────────────────────────┐
│           HTTP Handlers                 │
│  (Gin Controllers, Middleware)          │
└──────────────┬──────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────┐
│          Business Logic                 │
│         (Services Layer)                │
└──────────────┬──────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────┐
│         Data Access Layer               │
│        (Repositories)                   │
└──────────────┬──────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────┐
│          Domain Models                  │
│    (Entities, DTOs, Interfaces)         │
└─────────────────────────────────────────┘
```

## Слои приложения

### 1. Domain Layer (`internal/domain/`)

**Назначение**: Определяет бизнес-сущности и правила

**Содержит**:
- Модели данных (User, Patient, Surgery и т.д.)
- DTO (Data Transfer Objects) для API
- Бизнес-константы и перечисления
- Валидационные правила

**Принципы**:
- Не зависит от других слоёв
- Содержит только бизнес-логику
- Не содержит технических деталей (БД, HTTP и т.д.)

**Пример**:
```go
type Patient struct {
    ID            uint
    AccessCode    string
    FirstName     string
    LastName      string
    Status        PatientStatus
    OperationType OperationType
    // ...
}
```

### 2. Repository Layer (`internal/repository/`)

**Назначение**: Абстракция доступа к данным

**Содержит**:
- Интерфейсы репозиториев
- Реализации для работы с БД (GORM)
- SQL-запросы и операции с данными

**Принципы**:
- Один репозиторий на одну сущность
- Интерфейсы определяют контракт
- Скрывает детали реализации БД

**Пример**:
```go
type PatientRepository interface {
    Create(ctx context.Context, patient *domain.Patient) error
    FindByID(ctx context.Context, id uint) (*domain.Patient, error)
    FindAll(ctx context.Context, filters PatientFilters, offset, limit int) ([]domain.Patient, int64, error)
    Update(ctx context.Context, patient *domain.Patient) error
}
```

### 3. Service Layer (`internal/service/`)

**Назначение**: Бизнес-логика приложения

**Содержит**:
- Интерфейсы сервисов
- Реализации бизнес-операций
- Оркестрация между репозиториями
- Валидация и трансформация данных

**Принципы**:
- Один сервис на один домен
- Использует репозитории через интерфейсы
- Содержит всю бизнес-логику
- Не знает о HTTP/транспорте

**Пример**:
```go
type PatientService interface {
    Create(ctx context.Context, req domain.CreatePatientRequest, doctorID uint) (*domain.Patient, error)
    GetByID(ctx context.Context, id uint) (*domain.Patient, error)
    ChangeStatus(ctx context.Context, id uint, req domain.PatientStatusRequest, changedBy uint) error
}
```

### 4. Handler Layer (`internal/handler/`)

**Назначение**: HTTP обработчики запросов

**Содержит**:
- HTTP handlers (Gin)
- Парсинг запросов
- Формирование ответов
- Обработка ошибок HTTP

**Принципы**:
- Тонкий слой между HTTP и бизнес-логикой
- Делегирует работу сервисам
- Не содержит бизнес-логику
- Отвечает за HTTP-специфичные вещи

**Пример**:
```go
func (h *PatientHandler) Create(c *gin.Context) {
    var req domain.CreatePatientRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        BadRequest(c, err.Error())
        return
    }

    patient, err := h.svc.Create(c.Request.Context(), req, userID)
    if err != nil {
        Error(c, http.StatusBadRequest, err.Error())
        return
    }

    Success(c, http.StatusCreated, patient)
}
```

### 5. Middleware Layer (`internal/middleware/`)

**Назначение**: Обработка сквозной функциональности

**Содержит**:
- Аутентификация (JWT)
- Авторизация (RBAC)
- Логирование
- CORS

**Принципы**:
- Выполняется до handlers
- Модифицирует context
- Может прервать цепочку обработки

### 6. Server Layer (`internal/server/`)

**Назначение**: Конфигурация и запуск сервера

**Содержит**:
- Настройка роутера
- Регистрация маршрутов
- Dependency Injection
- Инициализация компонентов

## Поток данных

### Типичный запрос

```
1. HTTP Request
   ↓
2. Middleware (Auth, RBAC)
   ↓
3. Handler (парсинг, валидация)
   ↓
4. Service (бизнес-логика)
   ↓
5. Repository (доступ к БД)
   ↓
6. Database
   ↓
7. Repository (маппинг данных)
   ↓
8. Service (трансформация)
   ↓
9. Handler (формирование ответа)
   ↓
10. HTTP Response
```

### Пример: Создание пациента

```
POST /api/v1/patients
↓
Auth Middleware → проверка JWT токена
↓
RBAC Middleware → проверка роли DISTRICT_DOCTOR
↓
PatientHandler.Create()
  ├─ Парсинг JSON
  ├─ Валидация данных
  └─ Вызов PatientService.Create()
      ├─ Генерация AccessCode
      ├─ Вызов PatientRepository.Create()
      │   └─ INSERT в БД
      ├─ Генерация чек-листа
      ├─ Изменение статуса
      └─ Создание истории статусов
↓
HTTP 201 Created + JSON
```

## Dependency Injection

### Принцип

Зависимости передаются через конструкторы, а не создаются внутри компонентов.

### Пример

```go
// Создание зависимостей
patientRepo := repository.NewPatientRepository(db)
checklistRepo := repository.NewChecklistRepository(db)

// Инъекция в сервис
patientService := service.NewPatientService(patientRepo, checklistRepo)

// Инъекция в handler
patientHandler := handler.NewPatientHandler(patientService)
```

### Преимущества

- Легко тестировать (mock зависимости)
- Слабая связанность
- Гибкость замены реализаций

## Паттерны проектирования

### 1. Repository Pattern

Абстракция доступа к данным через интерфейсы.

```go
type UserRepository interface {
    FindByEmail(ctx context.Context, email string) (*domain.User, error)
}
```

### 2. Service Pattern

Инкапсуляция бизнес-логики в сервисы.

```go
type AuthService interface {
    Login(ctx context.Context, req domain.LoginRequest) (*domain.AuthResponse, error)
}
```

### 3. Factory Pattern

Создание объектов через фабричные функции.

```go
func NewPatientService(repo repository.PatientRepository) PatientService {
    return &patientService{repo: repo}
}
```

### 4. Strategy Pattern

Разные стратегии расчёта ИОЛ (SRK/T, Haigis, Hoffer Q).

```go
switch formula {
case "SRKT":
    power, ref = formulas.SRKT(...)
case "HAIGIS":
    power, ref = formulas.Haigis(...)
}
```

## Обработка ошибок

### Принципы

1. Ошибки возвращаются, не паникуют
2. Контекстная информация добавляется на каждом уровне
3. Пользовательские сообщения на русском
4. Технические детали в логах

### Пример

```go
// Repository
func (r *patientRepo) FindByID(ctx context.Context, id uint) (*domain.Patient, error) {
    var patient domain.Patient
    if err := r.db.First(&patient, id).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, errors.New("пациент не найден")
        }
        return nil, err
    }
    return &patient, nil
}

// Service
func (s *patientService) GetByID(ctx context.Context, id uint) (*domain.Patient, error) {
    p, err := s.repo.FindByID(ctx, id)
    if err != nil {
        return nil, err // пробрасываем ошибку выше
    }
    return p, nil
}

// Handler
func (h *PatientHandler) GetByID(c *gin.Context) {
    patient, err := h.svc.GetByID(c.Request.Context(), id)
    if err != nil {
        NotFound(c, err.Error()) // возвращаем 404 с сообщением
        return
    }
    Success(c, http.StatusOK, patient)
}
```

## Тестирование

### Unit тесты

Тестируют отдельные компоненты с mock зависимостями.

```go
func TestPatientService_Create(t *testing.T) {
    mockRepo := &MockPatientRepository{}
    service := NewPatientService(mockRepo, nil)

    patient, err := service.Create(ctx, req, doctorID)

    assert.NoError(t, err)
    assert.NotNil(t, patient)
}
```

### Integration тесты

Тестируют взаимодействие с реальной БД.

```go
func TestPatientRepository_Create(t *testing.T) {
    db := setupTestDB()
    repo := NewPatientRepository(db)

    err := repo.Create(ctx, patient)

    assert.NoError(t, err)
}
```

## Безопасность

### Аутентификация

- JWT токены (access + refresh)
- Короткое время жизни access токена (15 мин)
- Длительное время жизни refresh токена (7 дней)

### Авторизация

- RBAC (Role-Based Access Control)
- Проверка на уровне middleware
- Фильтрация данных по ролям

### Защита данных

- Хеширование паролей (bcrypt)
- Валидация входных данных
- Защита от SQL-инъекций (GORM)
- Ограничение размера файлов

## Масштабируемость

### Горизонтальное масштабирование

- Stateless приложение
- Shared Redis для сессий
- Shared PostgreSQL
- Load balancer (Nginx)

### Вертикальное масштабирование

- Оптимизация запросов к БД
- Индексы в PostgreSQL
- Кэширование в Redis
- Connection pooling

## Мониторинг

### Логирование

- Структурированные логи (Zerolog)
- Уровни: Debug, Info, Warn, Error
- Контекстная информация

### Метрики

- Время ответа API
- Количество запросов
- Ошибки и их типы
- Использование ресурсов

## Лучшие практики

1. **Один файл = одна ответственность**
2. **Интерфейсы для абстракции**
3. **Контекст для отмены операций**
4. **Defer для освобождения ресурсов**
5. **Константы вместо magic numbers**
6. **Валидация на границах системы**
7. **Логирование важных событий**
8. **Обработка всех ошибок**

## Дальнейшее развитие

### Возможные улучшения

- GraphQL API
- WebSocket для real-time уведомлений
- Микросервисная архитектура
- Event-driven architecture
- CQRS для сложных запросов
- Elasticsearch для полнотекстового поиска
