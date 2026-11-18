# 02 — Mindora RBAC Architecture Specification

**System:** Mindora Student Productivity Application  
**Component:** Security & Role-Based Access Control Architecture  
**Status:** Approved & Implemented in Database (Phase 2)  

---

## 1. Approved Architecture Overview

The backend maintains the **Standard Go Project Layout**:

```text
Backend/
├── cmd/
│   ├── api/
│   │   └── main.go            # Application entry point, DB init, graceful shutdown
│   └── migrate/
│       └── main.go            # Migration & seeder CLI runner
├── internal/
│   ├── config/                # Environment & connection pool configuration
│   ├── database/              # GORM setup, connection manager, migrator, seeder
│   ├── handlers/              # HTTP transport handlers
│   ├── middleware/            # CORS, future JWT auth, RBAC guard, rate limiting
│   ├── models/                # GORM entity definitions (Role, Permission, User, etc.)
│   ├── repositories/          # Data access layer & tenant/user scoping
│   ├── routes/                # Route registration and route grouping
│   └── services/              # Business logic & domain orchestration
├── pkg/
│   ├── hash/                  # Password hashing (Phase 3)
│   ├── jwt/                   # Token management (Phase 3)
│   ├── response/              # Standard JSON response envelope
│   └── validator/             # Request validation (Phase 3)
├── migrations/                # Version-controlled SQL migration scripts
└── docs/rbac/                 # Architecture & database specifications
```

---

## 2. Core Access Contexts & Principles

Mindora enforces access control across three primary operational contexts:

### 2.1 Guest (Public Context)
- **Nature:** Unauthenticated visitor.
- **Persistence:** **Guest is NOT a database role**. It does not exist in the `roles` table.
- **Capabilities:**
  - View public landing pages and public health endpoints (`/api/health`).
  - Access registration (`/api/v1/auth/register`) and login (`/api/v1/auth/login`).
- **Restrictions:**
  - Strictly blocked from any protected user or administrative resources.

### 2.2 User (Student Role)
- **Nature:** Authenticated student user account.
- **Persistence:** Persisted system role in `roles` with `name = 'user'`.
- **Capabilities:**
  - Manage personal profile (`profile.read`, `profile.update`).
  - Full CRUD on personal tasks (`tasks.create`, `tasks.read`, `tasks.update`, `tasks.delete`).
  - Access personal productivity tools (Pomodoro focus timer, AI assistant).
- **Restrictions:**
  - Cannot access administrative routes (`/api/v1/admin/*`).
  - Cannot read, update, or delete other users' private records (enforced via `user_id = ctx_user_id`).
  - Cannot change own role or assign roles.

### 2.3 Admin (System Management Role)
- **Nature:** Authenticated system administrator.
- **Persistence:** Persisted system role in `roles` with `name = 'admin'`.
- **Capabilities:**
  - View user accounts (`users.read`).
  - Create user accounts (`users.create`).
  - Update user accounts and activate/deactivate accounts (`users.update`, `users.activate`, `users.deactivate`).
  - Monitor system health, performance, and audit logs (`system.monitor`, `audit_logs.read`).
- **Strict Architectural Boundaries:**
  - **Cannot view or query private user productivity records** (personal tasks, private notes). Admin queries strictly consume aggregated SQL metrics (`COUNT`, `AVG`).
  - **Cannot modify roles or permissions**.
  - **Cannot create Admin accounts via user CRUD**. Admin provisioning is strictly controlled.
  - **Cannot assign arbitrary roles**.

---

## 3. Dual-Layer Authorization Strategy

Mindora separates role permissions from data ownership:

```text
Incoming Request -> [Layer 1: RBAC Middleware] -> [Layer 2: Data Ownership Scoping] -> Database
```

1. **Layer 1: Route-Level RBAC Middleware (Coarse-Grained):**
   - Validates that the caller holds the required permission or role (e.g. `tasks.read`).
2. **Layer 2: Repository-Level Ownership Scoping (Fine-Grained):**
   - Injects the authenticated `user_id` from token context into every personal data query (`WHERE user_id = :authenticated_user_id`).
   - Prevents Insecure Direct Object References (IDOR).

---

## 4. Future Bot Protection: Cloudflare Turnstile Integration

As mandated in product architecture decisions:
- Cloudflare Turnstile will be integrated during future frontend-backend auth wiring.
- Target endpoints: `/api/v1/auth/login` and `/api/v1/auth/register`.
- **Workflow:**
  1. Frontend renders Turnstile widget and obtains a client verification token.
  2. Frontend sends `cf_turnstile_token` in login/registration request payloads.
  3. Backend verifies token with Cloudflare API using server-side secret before processing credentials.
  4. Combined with IP-based rate limiting on the backend.
- Turnstile requires zero database tables.
