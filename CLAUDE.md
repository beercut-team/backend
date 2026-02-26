# Oculus-Feldsher Backend — Claude Operating Rules

You are working in an existing repository. DO NOT rewrite the project from scratch.
Goal: verify current backend matches the specification below, fix mismatches with minimal diffs, and update API docs using Scalar.

## Hard Rules

- Always start with an inventory (what exists now) before changing code.
- Prefer minimal changes and incremental commits.
- No mass refactors unless required to match the spec.
- Keep existing folder structure unless it clearly blocks the spec.
- All API changes must update OpenAPI and Scalar docs in the same PR.

## Current Target Spec (Source of Truth)

This repo must implement the "Oculus-Feldsher" backend spec:

- Django + DRF, PostgreSQL, JWT (simplejwt)
- Celery + Redis (image compression, notifications)
- drf-spectacular OpenAPI
- Roles: district_doctor, surgeon, patient, call_center
- Core models: User (custom), Patient, OperationType, PreparationCase, ChecklistItem, MediaFile, IOLCalculation, Comment, Notification, AuditLog, Ulus
- State machine: draft → in_progress → pending_review → (approved | needs_correction) → scheduled → completed; can go cancelled anytime
- Public endpoint: /api/public/status/{short_code}/ (no auth)
- Batch endpoint: POST /api/cases/{id}/batch-update/ for offline mode
- Files must NOT be served by direct links without auth checks

The detailed spec text is in the conversation (treat it as canonical).

## Required Work Plan (Must Follow)

### Phase 1 — Inventory (NO code changes)

1. List apps/modules and URLs:
   - show installed apps, urls.py routing, viewsets
2. List current models and fields (especially User, Patient, PreparationCase, ChecklistItem)
3. List permissions matrix currently implemented
4. Generate current OpenAPI schema and summarize endpoints
5. Produce a GAP REPORT: "Spec vs Current", with exact file references

### Phase 2 — Fixes (Minimal diffs)

Fix gaps in this order:

1. Auth/roles correctness (User.role, JWT, /api/auth/\* endpoints)
2. Ulus as FK table and filters by ulus everywhere required
3. PreparationCase state machine + transitions endpoints
4. Checklist autogeneration from OperationType.checklist_template
5. Media upload/download with auth gate
6. Batch-update endpoint for offline mode
7. IOL calculation SRK/T + Haigis, persistence
8. Notifications + audit logging
9. Tests for critical flows (auth, role access, state transitions, public status)

### Phase 3 — Docs (Scalar)

- Generate OpenAPI via drf-spectacular.
- Update Scalar docs from the OpenAPI file (single source).
- Ensure endpoints/params/response schemas match actual implementation.

## Commands You Can Run

- python manage.py check
- python manage.py makemigrations
- python manage.py migrate
- python manage.py test
- python manage.py spectacular --file openapi.json
- pytest (if used)
- celery worker/beat (only if needed for tests)
- grep, head, find

## Output Format Expectations

- For any change: explain "why" in one sentence and show file diffs.
- Always keep a running checklist of remaining gaps.
- Do not produce long prose. Prefer bullet points and exact paths.

## Acceptance Criteria

- All required endpoints exist and enforce permissions exactly as spec.
- OpenAPI schema matches reality and includes security schemes.
- Scalar docs are updated and render without errors.
- Minimal but meaningful tests are green.
