# Oculus-Feldsher Backend

Медицинская платформа для удалённой подготовки офтальмологических пациентов к операции.

## Описание проекта

Система предназначена для управления процессом подготовки пациентов к офтальмологическим операциям. Включает управление пациентами, чек-листами обследований, расчёт ИОЛ, планирование операций, уведомления через Telegram, поддержку медицинских стандартов (ICD-10, SNOMED-CT, LOINC) и интеграции с внешними системами (ЕМИАС, РИАМС).

## Технологический стек

- **Язык**: Go 1.23
- **Фреймворк**: Gin (HTTP), GORM (ORM)
- **База данных**: PostgreSQL 16
- **Кэш**: Redis 7
- **Хранилище файлов**: MinIO (S3-совместимое) или локальная ФС
- **Аутентификация**: JWT (access + refresh токены)
- **PDF генерация**: go-pdf/fpdf
- **Планировщик**: robfig/cron/v3
- **Telegram бот**: go-telegram-bot-api/v5
- **Документация API**: Scalar API Reference на `/docs`

## Архитектура

Проект следует принципам Clean Architecture:

```
domain → repository → service → handler
```

- **Domain** (`internal/domain/`) — модели сущностей и DTO
- **Repository** (`internal/repository/`) — доступ к данным с интерфейсами
- **Service** (`internal/service/`) — бизнес-логика
- **Handler** (`internal/handler/`) — HTTP обработчики
- **Middleware** (`internal/middleware/`) — аутентификация (JWT) и RBAC
- **Server** (`internal/server/`) — конфигурация роутера и DI

## Роли пользователей

- **ADMIN** — полный доступ ко всем функциям
- **SURGEON** — просмотр чек-листов, планирование операций
- **DISTRICT_DOCTOR** — создание и управление пациентами в своём районе
- **PATIENT** — ограниченный доступ к своим данным
- **CALL_CENTER** — управление пациентами и координация

**Важно**: При регистрации все пользователи должны указать `district_id` (ID района).

## Установка и запуск

### Требования

- Go 1.23+
- Docker и Docker Compose
- PostgreSQL 16
- Redis 7
- MinIO (опционально)

### Быстрый старт

1. Клонируйте репозиторий:
```bash
git clone <repository-url>
cd backend
```

2. Скопируйте файл окружения:
```bash
cp .env.example .env
```

3. Настройте переменные окружения в `.env`:
```env
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=oculus_db
DB_SSLMODE=disable

# JWT
JWT_SECRET=your-secret-key-change-in-production
JWT_ACCESS_EXPIRY=15m
JWT_REFRESH_EXPIRY=168h

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# MinIO
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin
MINIO_BUCKET=oculus-media
MINIO_USE_SSL=false

# Telegram (опционально)
TELEGRAM_BOT_TOKEN=your-bot-token
```

4. Запустите инфраструктуру через Docker:
```bash
docker-compose up -d
```

5. Запустите приложение:
```bash
go run ./cmd/api
```

6. (Опционально) Заполните тестовыми данными:
```bash
go run ./cmd/seed
```

## Основные команды

```bash
# Сборка всех модулей
go build ./...

# Запуск API сервера
go run ./cmd/api

# Заполнение тестовыми данными
go run ./cmd/seed

# Запуск всех сервисов через Docker
docker-compose up

# Остановка сервисов
docker-compose down

# Просмотр логов
docker-compose logs -f

# Запуск тестов
go test ./...
```

## Структура проекта

```
backend/
├── cmd/
│   ├── api/          # Точка входа API сервера
│   └── seed/         # Скрипт заполнения тестовыми данными
├── internal/
│   ├── config/       # Загрузка конфигурации (Viper)
│   ├── domain/       # Модели данных и DTO
│   ├── handler/      # HTTP обработчики
│   ├── middleware/   # Middleware (Auth, RBAC)
│   ├── repository/   # Репозитории для работы с БД
│   ├── service/      # Бизнес-логика
│   │   └── formulas/ # Формулы расчёта ИОЛ (SRK/T, Haigis, Hoffer Q)
│   └── server/       # Настройка роутера и DI
├── pkg/
│   ├── database/     # Подключение к PostgreSQL и Redis
│   ├── storage/      # Абстракция хранилища файлов (MinIO/Local)
│   ├── telegram/     # Telegram бот
│   └── logger/       # Настройка логирования (Zerolog)
├── .env.example      # Пример файла окружения
├── docker-compose.yml
└── README.md
```

## API Endpoints

### Аутентификация
- `POST /api/v1/auth/register` — Регистрация пользователя
- `POST /api/v1/auth/login` — Вход в систему
- `POST /api/v1/auth/refresh` — Обновление токена
- `POST /api/v1/auth/logout` — Выход из системы
- `GET /api/v1/auth/me` — Получить текущего пользователя

### Пациенты
- `POST /api/v1/patients` — Создать пациента
- `GET /api/v1/patients` — Список пациентов (с фильтрами)
- `GET /api/v1/patients/:id` — Получить пациента по ID
- `PUT /api/v1/patients/:id` — Обновить данные пациента
- `PUT /api/v1/patients/:id/status` — Изменить статус пациента
- `GET /api/v1/patients/public/:accessCode` — Публичный доступ по коду
- `GET /api/v1/patients/dashboard` — Статистика по пациентам

### Чек-листы
- `GET /api/v1/checklists/patient/:patientId` — Чек-лист пациента
- `POST /api/v1/checklists` — Создать пункт чек-листа (районный врач, хирург, администратор)
- `PUT /api/v1/checklists/:id` — Обновить пункт чек-листа
- `PUT /api/v1/checklists/:id/review` — Проверить пункт (хирург)
- `GET /api/v1/checklists/patient/:patientId/progress` — Прогресс выполнения

**Автоматический переход статуса**: При выполнении всех обязательных пунктов чек-листа статус пациента автоматически меняется с `IN_PROGRESS` на `PENDING_REVIEW`.

### Медиафайлы
- `POST /api/v1/media/upload` — Загрузить файл
- `GET /api/v1/media/patient/:patientId` — Файлы пациента
- `GET /api/v1/media/:id/download` — Скачать файл
- `GET /api/v1/media/:id/thumbnail` — Миниатюра изображения
- `DELETE /api/v1/media/:id` — Удалить файл

### Расчёт ИОЛ
- `POST /api/v1/iol/calculate` — Рассчитать силу ИОЛ
- `GET /api/v1/iol/patient/:patientId/history` — История расчётов

### Операции
- `POST /api/v1/surgeries` — Запланировать операцию
- `GET /api/v1/surgeries` — Список операций хирурга
- `GET /api/v1/surgeries/:id` — Получить операцию
- `PUT /api/v1/surgeries/:id` — Обновить операцию

### Комментарии
- `POST /api/v1/comments` — Создать комментарий
- `GET /api/v1/comments/patient/:patientId` — Комментарии пациента
- `PUT /api/v1/comments/patient/:patientId/read` — Отметить как прочитанные

### Уведомления
- `GET /api/v1/notifications` — Список уведомлений
- `PUT /api/v1/notifications/:id/read` — Отметить как прочитанное
- `PUT /api/v1/notifications/read-all` — Отметить все как прочитанные
- `GET /api/v1/notifications/unread-count` — Количество непрочитанных

### Печать
- `GET /api/v1/print/routing-sheet/:patientId` — Маршрутный лист (PDF)
- `GET /api/v1/print/checklist-report/:patientId` — Отчёт по чек-листу (PDF)

### Районы
- `POST /api/v1/districts` — Создать район
- `GET /api/v1/districts` — Список районов
- `GET /api/v1/districts/:id` — Получить район
- `PUT /api/v1/districts/:id` — Обновить район
- `DELETE /api/v1/districts/:id` — Удалить район

### Синхронизация (для мобильных приложений)
- `POST /api/v1/sync/push` — Отправить изменения
- `GET /api/v1/sync/pull?since=<timestamp>` — Получить изменения

### Медицинские стандарты
- `GET /api/v1/medical-codes/icd10/search?q=<query>` — Поиск кодов диагнозов ICD-10
- `GET /api/v1/medical-codes/snomed/search?q=<query>` — Поиск кодов процедур SNOMED-CT
- `GET /api/v1/medical-codes/loinc/search?q=<query>` — Поиск кодов наблюдений LOINC
- `POST /api/v1/patients/:id/medical-metadata` — Обновить медицинские метаданные пациента

### Интеграции с внешними системами
- `POST /api/v1/integrations/emias/patients/:id/export` — Экспорт пациента в ЕМИАС
- `POST /api/v1/integrations/emias/patients/:id/case` — Создать случай в ЕМИАС
- `GET /api/v1/integrations/emias/patients/:id/status` — Статус синхронизации с ЕМИАС
- `POST /api/v1/integrations/riams/patients/:id/export` — Экспорт пациента в РИАМС
- `GET /api/v1/integrations/riams/patients/:id/status` — Статус синхронизации с РИАМС
- `GET /api/v1/integrations/riams/regions` — Список поддерживаемых регионов РИАМС

### Администрирование
- `GET /api/v1/admin/users` — Список пользователей
- `GET /api/v1/admin/stats` — Общая статистика системы

## Формулы расчёта ИОЛ

Система поддерживает следующие формулы:

- **SRK/T** — универсальная формула для большинства случаев
- **Haigis** — требует измерение глубины передней камеры (ACD)
- **Hoffer Q** — оптимизирована для коротких глаз

## Telegram бот

Бот позволяет пациентам отслеживать статус подготовки:

- `/start <код_доступа>` — Привязать к карте пациента
- `/status` — Проверить текущий статус
- `/help` — Справка по командам

**Автоматические уведомления**: Пациенты получают уведомления в Telegram при:
- Добавлении нового пункта в чек-лист
- Изменении статуса пункта чек-листа
- Проверке пункта хирургом

## Переменные окружения

| Переменная | Описание | По умолчанию |
|-----------|----------|--------------|
| `PORT` | Порт API сервера | `8080` |
| `DB_HOST` | Хост PostgreSQL | `localhost` |
| `DB_PORT` | Порт PostgreSQL | `5432` |
| `DB_USER` | Пользователь БД | `postgres` |
| `DB_PASSWORD` | Пароль БД | - |
| `DB_NAME` | Имя БД | `oculus_db` |
| `DB_SSLMODE` | SSL режим | `disable` |
| `JWT_SECRET` | Секретный ключ JWT | - |
| `JWT_ACCESS_EXPIRY` | Время жизни access токена | `15m` |
| `JWT_REFRESH_EXPIRY` | Время жизни refresh токена | `168h` |
| `REDIS_HOST` | Хост Redis | `localhost` |
| `REDIS_PORT` | Порт Redis | `6379` |
| `MINIO_ENDPOINT` | Endpoint MinIO | `localhost:9000` |
| `MINIO_BUCKET` | Имя bucket | `oculus-media` |
| `TELEGRAM_BOT_TOKEN` | Токен Telegram бота | - |

## Разработка

### Документация

- **[FRONTEND_INTEGRATION_GUIDE.md](FRONTEND_INTEGRATION_GUIDE.md)** — Полное руководство по интеграции с API для фронтенд-разработчиков
- **[STATUS_TRANSITIONS.md](STATUS_TRANSITIONS.md)** — Карта переходов статусов пациентов с визуальными диаграммами
- **[API Documentation](http://localhost:8080/docs)** — Интерактивная документация API (Scalar)
- **[OpenAPI Spec](http://localhost:8080/openapi.json)** — OpenAPI 3.0 спецификация

### Добавление нового endpoint

1. Создайте DTO в `internal/domain/`
2. Добавьте методы в интерфейс репозитория
3. Реализуйте бизнес-логику в сервисе
4. Создайте handler в `internal/handler/`
5. Зарегистрируйте роут в `internal/server/server.go`

### Миграции

GORM автоматически создаёт и обновляет схему БД при запуске. Для production рекомендуется использовать инструменты миграций (например, golang-migrate).

## Безопасность

- Все пароли хешируются с использованием bcrypt
- JWT токены с коротким временем жизни
- RBAC для контроля доступа
- Валидация входных данных
- Защита от SQL-инъекций через GORM

## Лицензия

Proprietary

## Контакты

Для вопросов и предложений обращайтесь к команде разработки.
