# Phase 1 Inventory Report — Oculus-Feldsher Backend

**Date:** 2026-02-26
**Status:** NO CODE CHANGES (Inventory Only)

---

## CRITICAL FINDING: Technology Stack Mismatch

**SPEC REQUIRES:** Django + DRF + PostgreSQL + Celery + Redis + drf-spectacular
**ACTUAL IMPLEMENTATION:** Go + Gin + GORM + PostgreSQL + Redis + Telegram Bot

This is a **fundamental architectural mismatch**. The entire backend is implemented in Go, not Python/Django.

---

## 1. Current Architecture Overview

### Tech Stack
- **Language:** Go 1.23.0
- **Web Framework:** Gin (github.com/gin-gonic/gin v1.10.1)
- **ORM:** GORM (gorm.io/gorm v1.25.12)
- **Database:** PostgreSQL (gorm.io/driver/postgres v1.5.9)
- **Auth:** JWT (github.com/golang-jwt/jwt/v5 v5.2.1)
- **Storage:** MinIO or Local filesystem
- **Cache/Queue:** Redis (github.com/redis/go-redis/v9 v9.18.0)
- **Scheduler:** Cron (github.com/robfig/cron/v3 v3.0.1)
- **Notifications:** Telegram Bot API (github.com/go-telegram-bot-api/telegram-bot-api/v5 v5.5.1)
- **PDF Generation:** fpdf (github.com/go-pdf/fpdf v0.9.0)

### Project Structure
```
backend/
├── cmd/
│   ├── api/              # Main application entry point
│   ├── fix-access-codes/ # Utility commands
│   ├── reset-db/
│   └── seed/
├── internal/
│   ├── config/           # Configuration management
│   ├── domain/           # Domain models (entities)
│   ├── handler/          # HTTP handlers (controllers)
│   ├── middleware/       # Auth, RBAC middleware
│   ├── repository/       # Data access layer
│   ├── server/           # Router setup
│   └── service/          # Business logic
│       └── formulas/     # IOL calculation formulas
├── pkg/
│   ├── database/         # Database connection
│   ├── logger/           # Logging setup
│   ├── storage/          # File storage abstraction
│   └── telegram/         # Telegram bot
├── migrations/           # SQL migrations (2 files)
├── openapi.json          # OpenAPI 3.0.3 spec
└── uploads/              # Local file storage
```

---

## 2. Modules and URL Routing

### API Base Path
`/api/v1/*`

### Endpoints by Module

#### Auth (`/api/v1/auth`)
- `POST /register` - User registration
- `POST /login` - User login
- `POST /patient-login` - Patient login via access code
- `POST /telegram-token-login` - Telegram bot token login
- `POST /refresh` - Refresh JWT token
- `GET /me` - Get current user (protected)
- `POST /logout` - Logout (protected)

#### Districts (`/api/v1/districts`)
- `GET /` - List districts
- `GET /:id` - Get district by ID
- `POST /` - Create district (ADMIN only)
- `PATCH /:id` - Update district (ADMIN only)
- `DELETE /:id` - Delete district (ADMIN only)

#### Patients (`/api/v1/patients`)
- `GET /` - List patients (filtered by role)
- `GET /dashboard` - Dashboard stats
- `GET /:id` - Get patient by ID
- `GET /public/:accessCode` - **PUBLIC** status endpoint (no auth)
- `POST /` - Create patient (DISTRICT_DOCTOR, ADMIN)
- `PATCH /:id` - Update patient
- `DELETE /:id` - Delete patient (ADMIN only)
- `POST /:id/status` - Change patient status
- `POST /:id/regenerate-code` - Regenerate access code (ADMIN only)

#### Checklists (`/api/v1/checklists`)
- `GET /patient/:patientId` - Get checklist for patient
- `GET /patient/:patientId/progress` - Get progress summary
- `PATCH /:id` - Update checklist item
- `POST /:id/review` - Review item (SURGEON, ADMIN)

#### Media (`/api/v1/media`)
- `POST /upload` - Upload file
- `GET /patient/:patientId` - Get files for patient
- `GET /:id/download` - Download file (auth required)
- `GET /:id/download-url` - Get download URL
- `GET /:id/thumbnail` - Get thumbnail
- `DELETE /:id` - Delete file

#### IOL Calculator (`/api/v1/iol`)
- `POST /calculate` - Calculate IOL power
- `GET /patient/:patientId/history` - Get calculation history

#### Surgeries (`/api/v1/surgeries`)
- `GET /` - List surgeries
- `GET /:id` - Get surgery by ID
- `POST /` - Schedule surgery (SURGEON, ADMIN)
- `PATCH /:id` - Update surgery (SURGEON, ADMIN)
- `DELETE /:id` - Delete surgery (SURGEON, ADMIN)

#### Comments (`/api/v1/comments`)
- `POST /` - Create comment
- `GET /patient/:patientId` - Get comments for patient
- `POST /patient/:patientId/read` - Mark comments as read

#### Notifications (`/api/v1/notifications`)
- `GET /` - List notifications
- `GET /unread-count` - Get unread count
- `POST /` - Create notification
- `POST /:id/read` - Mark as read
- `POST /read-all` - Mark all as read

#### Print/PDF (`/api/v1/print`)
- `GET /patient/:patientId/routing-sheet` - Generate routing sheet PDF
- `GET /patient/:patientId/checklist-report` - Generate checklist report PDF

#### Sync (`/api/v1/sync`)
- `POST /push` - Push offline changes
- `GET /pull` - Pull server changes

#### Admin (`/api/v1/admin`)
- `GET /users` - List users (ADMIN only)
- `GET /stats` - System statistics (ADMIN only)

#### Documentation
- `GET /openapi.json` - OpenAPI schema
- `GET /docs` - Scalar API documentation UI
- `GET /admin` - Admin panel HTML
- `GET /patient` - Patient public status page
- `GET /patient/login` - Patient login page
- `GET /patient/portal` - Patient portal page

---

## 3. Domain Models and Fields

### User
```go
ID             uint
Email          string (unique, not null)
PasswordHash   string (not null)
Name           string (not null)
FirstName      string
LastName       string
MiddleName     string
Phone          string (indexed)
Role           Role (DISTRICT_DOCTOR, SURGEON, PATIENT, ADMIN, CALL_CENTER)
DistrictID     *uint (FK to District)
Specialization string
LicenseNumber  string
TelegramChatID *int64 (indexed)
IsActive       bool (default: true)
RefreshToken   string (indexed)
CreatedAt      time.Time
UpdatedAt      time.Time
```

### Patient
```go
ID             uint
AccessCode     string (unique, not null)
FirstName      string (not null)
LastName       string (not null)
MiddleName     string
DateOfBirth    time.Time
Phone          string
Email          string
Address        string
SNILs          string
PassportSeries string
PassportNumber string
PolicyNumber   string
Diagnosis      string (text)
OperationType  OperationType (PHACOEMULSIFICATION, ANTIGLAUCOMA, VITRECTOMY)
Eye            string (OD, OS, OU)
Status         PatientStatus (indexed, default: NEW)
DoctorID       uint (FK to User, indexed, not null)
SurgeonID      *uint (FK to User, indexed)
DistrictID     uint (FK to District, indexed)
Notes          string (text)
SurgeryDate    *time.Time
CreatedAt      time.Time
UpdatedAt      time.Time
```

**PatientStatus values:**
- NEW
- PREPARATION
- REVIEW_NEEDED
- APPROVED
- SURGERY_SCHEDULED
- COMPLETED
- REJECTED

### PatientStatusHistory
```go
ID         uint
PatientID  uint (indexed, not null)
FromStatus PatientStatus
ToStatus   PatientStatus (not null)
ChangedBy  uint
Comment    string (text)
CreatedAt  time.Time
```

### District (Ulus)
```go
ID        uint
Name      string (not null, unique)
Region    string (not null)
Code      string (unique)
Timezone  string (default: Europe/Moscow)
CreatedAt time.Time
UpdatedAt time.Time
```

### ChecklistTemplate
```go
ID            uint
OperationType OperationType (indexed, not null)
Name          string (not null)
Description   string (text)
Category      string
IsRequired    bool (default: true)
ExpiresInDays int
SortOrder     int
CreatedAt     time.Time
```

### ChecklistItem
```go
ID          uint
PatientID   uint (indexed, not null)
TemplateID  uint (indexed)
Name        string (not null)
Description string (text)
Category    string
IsRequired  bool (default: true)
Status      ChecklistItemStatus (indexed, default: PENDING)
Result      string (text)
Notes       string (text)
CompletedAt *time.Time
CompletedBy *uint
ReviewedBy  *uint
ReviewNote  string (text)
ExpiresAt   *time.Time
MediaID     *uint
CreatedAt   time.Time
UpdatedAt   time.Time
```

**ChecklistItemStatus values:**
- PENDING
- IN_PROGRESS
- COMPLETED
- REJECTED
- EXPIRED

### Media
```go
ID            uint
PatientID     uint (indexed, not null)
UploadedBy    uint
FileName      string (not null)
OriginalName  string (not null)
ContentType   string (not null)
Size          int64
StoragePath   string (not null)
ThumbnailPath string
Category      string (indexed)
CreatedAt     time.Time
```

### IOLCalculation
```go
ID                  uint
PatientID           uint (indexed, not null)
Eye                 string (not null)
AxialLength         float64 (not null)
Keratometry1        float64 (not null)
Keratometry2        float64 (not null)
ACD                 float64
TargetRefraction    float64
Formula             string (not null) // "SRKT", "Haigis", "HofferQ"
IOLPower            float64
PredictedRefraction float64
AConstant           float64
CalculatedBy        uint
Warnings            string (text)
CreatedAt           time.Time
```

### Surgery
```go
ID            uint
PatientID     uint (indexed, not null)
SurgeonID     uint (indexed, not null)
ScheduledDate time.Time (not null)
OperationType OperationType (not null)
Eye           string
Status        SurgeryStatus (default: SCHEDULED)
Notes         string (text)
CreatedAt     time.Time
UpdatedAt     time.Time
```

**SurgeryStatus values:**
- SCHEDULED
- COMPLETED
- CANCELLED

### Comment
```go
ID        uint
PatientID uint (indexed, not null)
AuthorID  uint (not null)
ParentID  *uint (indexed) // for threading
Body      string (text, not null)
IsUrgent  bool (default: false)
IsRead    bool (default: false)
CreatedAt time.Time
UpdatedAt time.Time
```

### Notification
```go
ID         uint
UserID     uint (indexed, not null)
Type       NotificationType (not null)
Title      string (not null)
Body       string (text)
EntityType string
EntityID   uint
IsRead     bool (indexed, default: false)
CreatedAt  time.Time
```

**NotificationType values:**
- STATUS_CHANGE
- NEW_COMMENT
- SURGERY_SCHEDULED
- CHECKLIST_EXPIRY
- SURGERY_REMINDER

### AuditLog
```go
ID        uint
UserID    uint (indexed)
Action    string (indexed, not null)
Entity    string (indexed, not null)
EntityID  uint
OldValue  string (text)
NewValue  string (text)
IP        string
CreatedAt time.Time
```

### SyncQueue
```go
ID         uint
UserID     uint (indexed, not null)
Entity     string (indexed, not null)
EntityID   uint (not null)
Action     string (not null) // CREATE, UPDATE, DELETE
Payload    string (text)
ClientTime time.Time
ServerTime time.Time
Synced     bool (indexed, default: false)
```

### TelegramToken
```go
ID        uint
Token     string (unique, not null)
UserID    *uint
PatientID *uint
ExpiresAt time.Time
CreatedAt time.Time
```

---

## 4. Permissions Matrix (RBAC)

### Middleware Implementation
- **Auth Middleware:** `middleware.Auth()` - validates JWT from Authorization header or query param
- **Role Middleware:** `middleware.RequireRole(roles...)` - checks user role

### Role-Based Access

| Endpoint | DISTRICT_DOCTOR | SURGEON | PATIENT | CALL_CENTER | ADMIN |
|----------|----------------|---------|---------|-------------|-------|
| **Auth** |
| Register, Login, Refresh | ✓ | ✓ | ✓ | ✓ | ✓ |
| Patient Login (access code) | ✓ | ✓ | ✓ | ✓ | ✓ |
| **Districts** |
| List, Get | ✓ | ✓ | ✓ | ✓ | ✓ |
| Create, Update, Delete | - | - | - | - | ✓ |
| **Patients** |
| List | Own only | Review+ | - | - | All |
| Create | ✓ | - | - | - | ✓ |
| Update | ✓ | ✓ | - | - | ✓ |
| Delete | - | - | - | - | ✓ |
| Change Status | ✓ | ✓ | - | - | ✓ |
| Regenerate Code | - | - | - | - | ✓ |
| **Checklists** |
| View, Update | ✓ | ✓ | - | - | ✓ |
| Review | - | ✓ | - | - | ✓ |
| **Surgeries** |
| List, View | ✓ | ✓ | - | - | ✓ |
| Create, Update, Delete | - | ✓ | - | - | ✓ |
| **Admin** |
| Users, Stats | - | - | - | - | ✓ |

**Notes:**
- District doctors see only their own patients
- Surgeons see patients with status >= REVIEW_NEEDED
- Public endpoint `/patients/public/:accessCode` has NO auth requirement
- Media download requires auth but no specific role check (relies on patient access)

---

## 5. OpenAPI Schema Summary

**File:** `openapi.json`
**Version:** OpenAPI 3.0.3
**API Version:** 2.0.0
**Title:** Oculus-Feldsher API

### Security Schemes
```json
"securitySchemes": {
  "bearerAuth": {
    "type": "http",
    "scheme": "bearer",
    "bearerFormat": "JWT"
  }
}
```

### Tags (11 total)
- Auth
- Districts
- Patients
- Checklists
- Media
- IOL
- Surgeries
- Comments
- Notifications
- Print
- Sync
- System

### Endpoint Count
Approximately 40+ endpoints documented in OpenAPI schema.

### Documentation UI
- **Scalar UI** available at `/docs`
- Loads schema from `/openapi.json`

---

## 6. State Machine Implementation

### Patient Status Flow

**Current Implementation:**
```
NEW → PREPARATION → REVIEW_NEEDED → (APPROVED | REJECTED)
                                      ↓
                              SURGERY_SCHEDULED → COMPLETED
```

**Transitions handled in:**
- `internal/service/patient_service.go` - ChangeStatus method
- Status history tracked in `PatientStatusHistory` table

**Validation:**
- Status transitions are validated in service layer
- History is logged for audit trail
- Notifications sent on status changes

**Missing from spec:**
- No explicit "draft" status (starts at NEW)
- No "in_progress" status
- No "pending_review" status (uses REVIEW_NEEDED)
- No "needs_correction" status (uses REJECTED)
- No "cancelled" status for patients (only for surgeries)

---

## 7. Key Features Implementation Status

### ✅ Implemented
1. **JWT Authentication** - golang-jwt/jwt/v5
2. **Role-based Access Control** - 5 roles (DISTRICT_DOCTOR, SURGEON, PATIENT, ADMIN, CALL_CENTER)
3. **District (Ulus) Management** - Full CRUD with FK relationships
4. **Patient Management** - CRUD with status workflow
5. **Checklist System** - Auto-generation from templates based on OperationType
6. **Media Upload/Download** - With auth gate, MinIO or local storage
7. **IOL Calculation** - SRK/T, Haigis, HofferQ formulas implemented
8. **Surgery Scheduling** - Full CRUD
9. **Comments System** - Threaded comments with urgent flag
10. **Notifications** - In-app notifications with types
11. **Audit Logging** - AuditLog model exists
12. **Public Status Endpoint** - `/api/v1/patients/public/:accessCode` (no auth)
13. **Offline Sync** - Push/pull endpoints with SyncQueue
14. **PDF Generation** - Routing sheet and checklist reports
15. **Telegram Bot Integration** - Notifications via Telegram
16. **Scheduler** - Cron-based background tasks

### ❌ Missing or Different from Spec

1. **Technology Stack** - Go instead of Django+DRF
2. **Celery** - Using Go cron scheduler instead
3. **drf-spectacular** - Using manually maintained openapi.json
4. **Batch Update Endpoint** - No `/api/cases/{id}/batch-update/` endpoint
5. **State Machine** - Different status names (NEW vs draft, REVIEW_NEEDED vs pending_review)
6. **OperationType Model** - Hardcoded constants, not a database table with checklist_template field
7. **PreparationCase** - Called "Patient" in this implementation
8. **MediaFile** - Called "Media" in this implementation
9. **Tests** - Only 1 test file found (`formulas_test.go`)

---

## 8. Database Migrations

**Location:** `migrations/`

**Files:**
1. `000001_create_users.up.sql` (empty)
2. `000001_create_users.down.sql` (empty)
3. `000002_add_telegram_chat_id.up.sql` - Adds telegram_chat_id to users
4. `000002_add_telegram_chat_id.down.sql` - Removes telegram_chat_id

**Note:** GORM AutoMigrate is likely used instead of SQL migrations.

---

## 9. Testing Status

**Test Files Found:** 1
- `internal/service/formulas/formulas_test.go`

**Coverage:** Minimal
- Only IOL formula calculations are tested
- No integration tests
- No API endpoint tests
- No auth/permission tests
- No state machine tests

---

## 10. GAP REPORT: Spec vs Current Implementation

### CRITICAL GAPS

#### 1. **Wrong Technology Stack**
- **Spec:** Django + DRF + Python
- **Current:** Go + Gin
- **Impact:** Complete rewrite required to match spec
- **Files:** Entire codebase

#### 2. **Missing Celery**
- **Spec:** Celery + Redis for async tasks (image compression, notifications)
- **Current:** Go cron scheduler + Telegram bot
- **Impact:** Different async architecture
- **Files:** `internal/service/scheduler_service.go`

#### 3. **Missing drf-spectacular**
- **Spec:** Auto-generated OpenAPI from DRF
- **Current:** Manually maintained `openapi.json`
- **Impact:** Schema can drift from implementation
- **Files:** `openapi.json`

### MAJOR GAPS

#### 4. **State Machine Differences**
- **Spec:** draft → in_progress → pending_review → (approved | needs_correction) → scheduled → completed | cancelled
- **Current:** NEW → PREPARATION → REVIEW_NEEDED → (APPROVED | REJECTED) → SURGERY_SCHEDULED → COMPLETED
- **Impact:** Different status names, missing states
- **Files:** `internal/domain/patient.go`

#### 5. **Missing Batch Update Endpoint**
- **Spec:** `POST /api/cases/{id}/batch-update/` for offline mode
- **Current:** Generic sync push/pull endpoints
- **Impact:** Different offline sync approach
- **Files:** Need to add to `internal/handler/patient_handler.go`

#### 6. **OperationType Not a Table**
- **Spec:** OperationType model with checklist_template field
- **Current:** Hardcoded constants, templates in code
- **Impact:** Cannot dynamically manage operation types
- **Files:** `internal/domain/patient.go`, `internal/domain/checklist.go`

#### 7. **Model Naming**
- **Spec:** PreparationCase, MediaFile
- **Current:** Patient, Media
- **Impact:** Semantic difference, spec uses "case" terminology
- **Files:** `internal/domain/patient.go`, `internal/domain/media.go`

### MINOR GAPS

#### 8. **Insufficient Tests**
- **Spec:** Tests for critical flows
- **Current:** Only formula tests
- **Impact:** Low confidence in correctness
- **Files:** Need tests in all packages

#### 9. **Audit Logging Not Active**
- **Spec:** AuditLog for all mutations
- **Current:** Model exists but not used
- **Files:** `internal/domain/audit.go`, `internal/repository/audit_repo.go`

#### 10. **Public Endpoint Path**
- **Spec:** `/api/public/status/{short_code}/`
- **Current:** `/api/v1/patients/public/:accessCode`
- **Impact:** Different URL structure
- **Files:** `internal/server/server.go`

---

## 11. Recommendations

### Option A: Rewrite in Django (Match Spec)
- **Effort:** 4-6 weeks
- **Risk:** High (complete rewrite)
- **Benefit:** Matches spec exactly

### Option B: Update Spec to Match Go Implementation
- **Effort:** 1-2 days (documentation)
- **Risk:** Low
- **Benefit:** Spec reflects reality

### Option C: Hybrid Approach
- **Effort:** 2-3 weeks
- **Risk:** Medium
- **Benefit:** Keep Go, fix critical gaps (state machine, batch endpoint, tests)

---

## 12. Files Reference

### Core Domain Models
- `internal/domain/user.go`
- `internal/domain/patient.go`
- `internal/domain/district.go`
- `internal/domain/checklist.go`
- `internal/domain/media.go`
- `internal/domain/iol.go`
- `internal/domain/surgery.go`
- `internal/domain/comment.go`
- `internal/domain/notification.go`
- `internal/domain/audit.go`
- `internal/domain/sync.go`

### Handlers (Controllers)
- `internal/handler/auth_handler.go`
- `internal/handler/patient_handler.go`
- `internal/handler/district_handler.go`
- `internal/handler/checklist_handler.go`
- `internal/handler/media_handler.go`
- `internal/handler/iol_handler.go`
- `internal/handler/surgery_handler.go`
- `internal/handler/comment_handler.go`
- `internal/handler/notification_handler.go`
- `internal/handler/print_handler.go`
- `internal/handler/sync_handler.go`
- `internal/handler/admin_handler.go`

### Services (Business Logic)
- `internal/service/auth_service.go`
- `internal/service/patient_service.go`
- `internal/service/district_service.go`
- `internal/service/checklist_service.go`
- `internal/service/media_service.go`
- `internal/service/iol_service.go`
- `internal/service/surgery_service.go`
- `internal/service/comment_service.go`
- `internal/service/notification_service.go`
- `internal/service/pdf_service.go`
- `internal/service/sync_service.go`
- `internal/service/token_service.go`
- `internal/service/scheduler_service.go`
- `internal/service/formulas/srkt.go`
- `internal/service/formulas/haigis.go`
- `internal/service/formulas/hofferq.go`

### Middleware
- `internal/middleware/auth.go`
- `internal/middleware/rbac.go`

### Router
- `internal/server/server.go`

### Configuration
- `internal/config/config.go`
- `.env.example`

---

## END OF PHASE 1 INVENTORY

**Next Steps:** Await decision on Option A, B, or C before proceeding to Phase 2.
