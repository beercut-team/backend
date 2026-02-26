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
go build ./...                      # Build everything
go run ./cmd/api                    # Start the API server
go run ./cmd/seed                   # Seed test data
go run ./cmd/fix-access-codes       # Generate access codes for old patients
docker-compose up                   # Start all services
```

## Roles & Permissions
- **ADMIN** — full system access, user management, all CRUD operations
- **SURGEON** — reviews checklists, schedules surgeries, approves patients
- **DISTRICT_DOCTOR** — creates/manages patients in their district, fills checklists
- **PATIENT** — limited public access via access code (no authentication required)

## Patient Access Code System

### Overview
Each patient receives a unique 8-character hex access code upon creation. This code enables:
- **Public status tracking** without authentication
- **Telegram bot integration** for real-time notifications
- **Secure patient identification** without exposing personal data

### Access Code Generation
- Auto-generated on patient creation: `domain.GenerateAccessCode()`
- Format: 8-character hex string (e.g., `a1b2c3d4`)
- Stored in `patients.access_code` (unique index)

### Public Access Points

#### 1. Web Interface
**URL**: `/patient?code=<access_code>`
- Beautiful responsive UI with Tailwind CSS
- Shows: patient name, status, surgery date, status history
- No authentication required
- Mobile-friendly design

#### 2. Public API
**Endpoint**: `GET /api/v1/patients/public/:accessCode`
- Returns: `PatientPublicResponse` (limited fields)
- No auth token required
- Used by web interface and external integrations

#### 3. Telegram Bot
**Commands**:
- `/start <access_code>` — bind patient to Telegram chat
- `/status` — check current preparation status
- `/login` — get one-time login link for patient portal
- Automatic notifications on status changes

**Patient Authentication Flow**:
1. **Direct login**: `POST /api/v1/auth/patient-login` with `{"access_code": "a1b2c3d4"}`
2. **Telegram login**:
   - Patient uses `/login` command in Telegram bot
   - Bot generates one-time token (valid 15 min)
   - Patient clicks link: `/patient/portal?token=<token>`
   - Frontend calls `POST /api/v1/auth/telegram-token-login` with token
   - Returns JWT tokens for authenticated access

### Admin Features
- Access code displayed prominently in patient card
- Copy-to-clipboard functionality
- Direct links for Telegram and web access
- Example: `/patient?code=a1b2c3d4`

## API Base URL
`/api/v1/`

## Public Endpoints (No Auth)
- `POST /api/v1/auth/register` — user registration
- `POST /api/v1/auth/login` — user login (email + password)
- `POST /api/v1/auth/patient-login` — patient login by access code (returns JWT tokens)
- `POST /api/v1/auth/telegram-token-login` — patient login via Telegram one-time token
- `POST /api/v1/auth/refresh` — refresh access token
- `GET /api/v1/patients/public/:accessCode` — patient status by code (no auth)
- `GET /patient` — patient web interface
- `GET /patient/login` — patient login page
- `GET /patient/portal` — patient portal (requires auth)
- `GET /admin` — admin panel UI
- `GET /docs` — API documentation (Scalar)

## Protected Endpoints (Auth Required)
All other endpoints require JWT token in `Authorization: Bearer <token>` header.

## Environment Variables
Copy `.env.example` to `.env` and configure:
- `DB_*` — PostgreSQL connection
- `JWT_SECRET` — JWT signing key
- `REDIS_*` — Redis connection
- `MINIO_*` — Object storage (optional, falls back to local FS)
- `TELEGRAM_BOT_TOKEN` — Telegram bot API token

## Module Structure
- `cmd/api/` — Main server entrypoint
- `cmd/seed/` — Seed test data (districts, users, patients)
- `cmd/fix-access-codes/` — Utility to generate codes for existing patients
- `internal/config/` — Config loading (Viper)
- `internal/domain/` — All domain models and DTOs
- `internal/handler/` — HTTP handlers + response helpers
- `internal/middleware/` — Auth (JWT) + RBAC middleware
- `internal/repository/` — Database repositories (interface-based)
- `internal/service/` — Business logic services
- `internal/service/formulas/` — IOL calculation formulas (SRK/T, Haigis, Hoffer Q)
- `internal/server/` — Router + DI wiring + embedded HTML pages
- `pkg/database/` — PostgreSQL + Redis connections
- `pkg/storage/` — File storage abstraction (MinIO + local)
- `pkg/telegram/` — Telegram bot with patient notifications
- `pkg/logger/` — Zerolog setup

## Telegram Bot Integration

### Patient Commands
- `/start <code>` — bind access code to chat
- `/status` — view current preparation status
- Receives automatic notifications on status changes

### Doctor Commands
- `/register <email>` — bind doctor account to Telegram
- `/mypatients` — list assigned patients
- Receives notifications about new patients and reviews

### Notification Types
- New patient assigned to doctor
- Patient status changed
- Checklist ready for surgeon review
- Surgery scheduled
