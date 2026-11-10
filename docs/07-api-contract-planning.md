# 07 - API Contract Planning & Interface Specifications

**Protocol:** HTTPS / REST API  
**Payload Format:** JSON  
**Base URL:** `/api/v1`  
**Status:** Conceptual Planning (No backend implementation in current stage)

---

## 1. Standard API Response Structure

All endpoints return a uniform JSON envelope adhering to `shared/response`:

### 1.1 Success Response (`HTTP 200 / 201`)
```json
{
  "success": true,
  "message": "Operation completed successfully",
  "data": { ... }
}
```

### 1.2 Error Response (`HTTP 400 / 401 / 403 / 404 / 500`)
```json
{
  "success": false,
  "message": "Descriptive error message",
  "error_code": "RESOURCE_NOT_FOUND"
}
```

---

## 2. Planned API Endpoints

### 2.1 System & Health Checks
- `GET /api/v1/health` → System status check (Active).
- `GET /api/v1/health/database` → PostgreSQL ping check (Active).

### 2.2 Authentication Module (`/auth`)
- `POST /api/v1/auth/login`
  - **Body:** `{ "email": "john@mindora.app", "password": "...", "role": "user" }`
  - **Response:** `{ "token": "jwt...", "user": { "id": "...", "name": "John", "role": "user" } }`
- `POST /api/v1/auth/register`
  - **Body:** `{ "name": "John", "email": "...", "password": "..." }`
- `POST /api/v1/auth/refresh`
  - **Header:** `Authorization: Bearer <refresh_token>`

### 2.3 Task Management Module (`/tasks`)
- `GET /api/v1/tasks?filter=today`
  - **Role:** User (Scoped to authenticated `user_id`)
  - **Response:** Array of `TaskItem` objects.
- `POST /api/v1/tasks`
  - **Body:** `{ "title": "Client meeting prep", "category": "Work", "scheduled_time": "2026-12-02T10:20:00Z" }`
- `PATCH /api/v1/tasks/:id/toggle`
  - **Body:** `{ "is_completed": true }`
- `DELETE /api/v1/tasks/:id`

### 2.4 Focus Timer Module (`/focus`)
- `POST /api/v1/focus/sessions`
  - **Body:** `{ "task_id": "tsk_01", "duration_minutes": 30, "started_at": "...", "ended_at": "..." }`
  - **Response:** Records completed study session.

### 2.5 AI Assistant & Planning Module (`/ai`)
- `POST /api/v1/ai/plan-day`
  - **Body:** `{ "prompt": "Plan my day based on current pending tasks" }`
  - **Response:** `{ "recommendations": ["..."], "focus_intervals": [...] }`

### 2.6 Admin Module (`/admin`)
- `GET /api/v1/admin/overview`
  - **Role:** Admin only
  - **Response:** Aggregated KPIs:
    ```json
    {
      "total_members": 248,
      "member_growth_percentage": 12.5,
      "active_tasks": 1429,
      "task_growth_percentage": 8.2,
      "completion_rate": 84.6,
      "active_workspaces": 16
    }
    ```
- `GET /api/v1/admin/members`
  - **Role:** Admin only
  - **Response:** Member list with roles and aggregated task completion counts (strictly no private task data).
