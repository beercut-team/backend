# Oculus-Feldsher Backend

## Project Overview
Medical platform for remote ophthalmological patient preparation (Go + Gin + GORM).

## Architecture
Clean Architecture: `domain → repository → service → handler`
- **Domain**: Entity models and DTOs (`internal/domain/`)
- **Repository**: Data access with interface-based design (`internal/repository/`)
- **Service**: Business logic implementing interfaces (`internal/service/`)
- **Handler**: HTTP handlers orchestrating services (`internal/handler/`)
- **Middleware**: Auth (JWT) + RBAC (`internal/middleware/`)
- **Server**: Router config and DI wiring (`internal/server/server.go`)

## Tech Stack
- **Language**: Go 1.23
- **Framework**: Gin (HTTP), GORM (ORM)
- **Database**: PostgreSQL 16
- **Cache**: Redis 7
- **Object Storage**: MinIO (or local FS for dev)
- **Auth**: JWT (access + refresh tokens, role claims)
- **PDF**: go-pdf/fpdf
- **Scheduler**: robfig/cron/v3
- **Telegram**: go-telegram-bot-api/v5
- **Docs**: Scalar API Reference at `/docs`

## Key Commands
```bash
go build ./...           # Build everything
go run ./cmd/api         # Start the API server
go run ./cmd/seed        # Seed test data
docker-compose up        # Start all services
```

## Roles
- `ADMIN` — full access
- `SURGEON` — reviews checklists, schedules surgeries
- `DISTRICT_DOCTOR` — creates/manages patients in their district
- `PATIENT` — limited access

## API Base URL
`/api/v1/`

## Environment
Copy `.env.example` to `.env` and configure. Key vars: DB_*, JWT_*, REDIS_*, MINIO_*, TELEGRAM_BOT_TOKEN

## Module Structure
- `cmd/api/` — Main server entrypoint
- `cmd/seed/` — Seed test data
- `internal/config/` — Config loading (Viper)
- `internal/domain/` — All domain models
- `internal/handler/` — HTTP handlers + response helpers
- `internal/middleware/` — Auth + RBAC middleware
- `internal/repository/` — Database repositories
- `internal/service/` — Business logic services
- `internal/service/formulas/` — IOL calculation formulas (SRK/T, Haigis, Hoffer Q)
- `internal/server/` — Router + DI wiring
- `pkg/database/` — PostgreSQL + Redis connections
- `pkg/storage/` — File storage abstraction (MinIO + local)
- `pkg/telegram/` — Telegram bot
- `pkg/logger/` — Zerolog setup
