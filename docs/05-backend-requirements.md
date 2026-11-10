# 05 - Backend Requirements & Architectural Planning

**Product:** Mindora — Smart Student Productivity Application  
**Engine:** Go Fiber v2 (REST API)  
**Database ORM:** GORM with PostgreSQL 18 driver  
**Status:** Architecture Planning (No backend code changes made during this stage)

---

## 1. Architectural Role & Responsibilities

The Mindora backend serves as the core orchestration engine, responsible for:
1. **Authentication & Session Management:** JWT token issuance, verification, and role-based access control (`User` vs `Admin`).
2. **Business Logic & Validation:** Enforcing task ownership, workspace role constraints, and privacy boundaries.
3. **Data Persistence:** Managing ACID transactions and entity relations via GORM and PostgreSQL.
4. **AI & ML Service Orchestration:** Safely brokering requests to external LLM providers and the internal Python ML prediction service without exposing sensitive credentials or user PII.

---

## 2. Candidate Backend Modules

In alignment with the modular structure established in previous stages, the backend will be organized into feature-driven modules:

```text
Backend/
├── main.go
├── config/             # Environment & configuration loader
├── database/           # PostgreSQL GORM connection & migration runner
├── middleware/         # CORS, JWT Authentication, Role Guard, Rate Limiter
├── health/             # System and database health checks
├── auth/               # Registration, Login, Token Refresh, Password Reset
├── users/              # User profiles, preferences, student academic info
├── tasks/              # Task CRUD, status toggling, categories, prioritization
├── focus/              # Focus timer sessions, completed focus durations
├── workspaces/         # Multi-member collaborative workspaces, invitations
├── analytics/          # Aggregated productivity metrics & chart time-series
├── admin/              # Administrative overview KPIs, member access management
├── ai/                 # Proxy to LLM recommendations & Python ML Regressor
└── shared/
    ├── response/       # Standardized JSON response payload { success, message, data }
    └── errors/         # Application error codes & HTTP status mappers
```

---

## 3. Data Flow & Security Rules

1. **Authentication Guard:** All endpoints under `/api/v1/*` (except `/api/v1/health` and `/api/v1/auth/*`) require a valid Bearer JWT.
2. **Data Ownership Scoping:** Queries for tasks, focus logs, and schedules must always include `WHERE user_id = :current_user_id`.
3. **Admin Privacy Isolation:** The `admin` module endpoints must only query aggregated SQL views (e.g. `COUNT(tasks.id)`, `AVG(completion_rate)`). Under no circumstance may an endpoint return raw personal task descriptions or student private notes to an admin requester.
4. **Graceful Degradation:** If external LLM or Python ML services become unavailable, the backend returns cached or deterministic heuristic suggestions rather than failing the user request.
