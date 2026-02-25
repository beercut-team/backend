# Как проверить и отладить seed на сервере

## 1. Проверить логи GitHub Actions

```bash
# Открыть в браузере
https://github.com/beercut-team/backend/actions
```

Посмотреть последний workflow run, проверить шаги:
- ✅ Build Docker Image
- ✅ Run seed (idempotent, includes AutoMigrate)
- ✅ Fix access codes for existing patients
- ✅ Run container

## 2. Проверить логи контейнера на сервере

```bash
# SSH на сервер
ssh your-server

# Проверить запущен ли контейнер
docker ps | grep peak-it-backend

# Посмотреть логи контейнера
docker logs peak-it-backend

# Посмотреть последние 50 строк
docker logs peak-it-backend --tail 50

# Следить за логами в реальном времени
docker logs peak-it-backend -f
```

## 3. Проверить данные в БД

```bash
# Подключиться к PostgreSQL контейнеру
docker exec -it postgres psql -U postgres -d peakit

# Проверить пациентов и их коды
SELECT id, first_name, last_name, access_code FROM patients;

# Проверить сколько пациентов без кодов
SELECT COUNT(*) FROM patients WHERE access_code IS NULL OR access_code = '';

# Выйти
\q
```

## 4. Вручную запустить seed (если CI не сработал)

```bash
# На сервере
docker run --rm \
  --network app-network \
  -e "DB_HOST=postgres" \
  -e "DB_PORT=5432" \
  -e "DB_USER=postgres" \
  -e "DB_PASSWORD=your_password" \
  -e "DB_NAME=peakit" \
  -e "DB_SSLMODE=disable" \
  peak-it-backend:latest ./seed
```

## 5. Вручную запустить fix-access-codes

```bash
# На сервере
docker run --rm \
  --network app-network \
  -e "DB_HOST=postgres" \
  -e "DB_PORT=5432" \
  -e "DB_USER=postgres" \
  -e "DB_PASSWORD=your_password" \
  -e "DB_NAME=peakit" \
  -e "DB_SSLMODE=disable" \
  peak-it-backend:latest ./fix-access-codes
```

## 6. Полный сброс БД (если нужно начать с нуля)

```bash
# На сервере
docker run --rm \
  --network app-network \
  -e "DB_HOST=postgres" \
  -e "DB_PORT=5432" \
  -e "DB_USER=postgres" \
  -e "DB_PASSWORD=your_password" \
  -e "DB_NAME=peakit" \
  -e "DB_SSLMODE=disable" \
  peak-it-backend:latest ./reset-db

# Затем seed
docker run --rm \
  --network app-network \
  -e "DB_HOST=postgres" \
  -e "DB_PORT=5432" \
  -e "DB_USER=postgres" \
  -e "DB_PASSWORD=your_password" \
  -e "DB_NAME=peakit" \
  -e "DB_SSLMODE=disable" \
  peak-it-backend:latest ./seed
```

## 7. Проверить API напрямую

```bash
# Получить токен
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"admin123"}' \
  | jq -r '.data.access_token')

# Проверить пациента
curl -s http://localhost:8080/api/v1/patients/1 \
  -H "Authorization: Bearer $TOKEN" \
  | jq '.data.access_code'

# Должен вернуть код, а не null
```

## Типичные проблемы

### Проблема: "connection refused"
- Контейнер postgres не запущен
- Неправильная сеть (app-network)

```bash
docker network ls
docker network inspect app-network
```

### Проблема: "authentication failed"
- Неправильный пароль в secrets
- Проверить GitHub Secrets: Settings → Secrets → Actions

### Проблема: Seed запустился, но данных нет
- Проверить логи seed: `docker logs peak-it-backend | grep seed`
- Возможно ошибка при создании данных

### Проблема: access_code всё ещё NULL
- fix-access-codes не запустился или упал
- Запустить вручную (см. пункт 5)
