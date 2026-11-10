# 06 - Database Planning & Schema Design

**Database Engine:** PostgreSQL 18  
**ORM Tool:** GORM (Go)  
**Status:** Conceptual Planning (No migrations or tables created during this stage)

---

## 1. Candidate Entities & Relationships

### 1.1 Entity: `users`
- **Purpose:** Central account entity for authentication and identification.
- **Attributes:**
  - `id`: `UUID` (Primary Key, default `gen_random_uuid()`)
  - `name`: `VARCHAR(100)` NOT NULL
  - `email`: `VARCHAR(255)` UNIQUE NOT NULL
  - `password_hash`: `VARCHAR(255)` NOT NULL
  - `role`: `VARCHAR(20)` NOT NULL (Values: `'student'`, `'admin'`, `'manager'`)
  - `avatar_url`: `VARCHAR(500)` NULL
  - `created_at`: `TIMESTAMPTZ` NOT NULL DEFAULT `NOW()`
  - `updated_at`: `TIMESTAMPTZ` NOT NULL DEFAULT `NOW()`
  - `deleted_at`: `TIMESTAMPTZ` NULL (Soft delete support)

### 1.2 Entity: `workspaces`
- **Purpose:** Collaborative spaces (e.g., "UI Design", "A-team Class").
- **Attributes:**
  - `id`: `UUID` (Primary Key)
  - `name`: `VARCHAR(100)` NOT NULL
  - `icon_type`: `VARCHAR(50)` DEFAULT `'folder'`
  - `owner_id`: `UUID` REFERENCES `users(id)` ON DELETE RESTRICT
  - `created_at`: `TIMESTAMPTZ` NOT NULL DEFAULT `NOW()`

### 1.3 Entity: `workspace_members`
- **Purpose:** Many-to-many junction between users and workspaces with role permissions.
- **Attributes:**
  - `id`: `UUID` (Primary Key)
  - `workspace_id`: `UUID` REFERENCES `workspaces(id)` ON DELETE CASCADE
  - `user_id`: `UUID` REFERENCES `users(id)` ON DELETE CASCADE
  - `role`: `VARCHAR(20)` NOT NULL (Values: `'admin'`, `'manager'`, `'member'`)
  - `status`: `VARCHAR(20)` DEFAULT `'active'` (Values: `'active'`, `'away'`, `'inactive'`)
  - `joined_at`: `TIMESTAMPTZ` NOT NULL DEFAULT `NOW()`
  - Unique constraint on `(workspace_id, user_id)`.

### 1.4 Entity: `tasks`
- **Purpose:** Personal and workspace tasks.
- **Attributes:**
  - `id`: `UUID` (Primary Key)
  - `user_id`: `UUID` REFERENCES `users(id)` ON DELETE CASCADE (Data owner)
  - `workspace_id`: `UUID` NULL REFERENCES `workspaces(id)` (NULL for personal tasks)
  - `title`: `VARCHAR(255)` NOT NULL
  - `category`: `VARCHAR(50)` NOT NULL DEFAULT `'Inbox'`
  - `category_color`: `VARCHAR(20)` DEFAULT `'#3B82F6'`
  - `scheduled_time`: `TIMESTAMPTZ` NULL
  - `is_completed`: `BOOLEAN` NOT NULL DEFAULT `FALSE`
  - `priority`: `VARCHAR(20)` DEFAULT `'medium'`
  - `created_at`: `TIMESTAMPTZ` NOT NULL DEFAULT `NOW()`
  - `completed_at`: `TIMESTAMPTZ` NULL
  - `deleted_at`: `TIMESTAMPTZ` NULL (Soft delete)

### 1.5 Entity: `focus_sessions`
- **Purpose:** Recorded Pomodoro and study sessions for focus tracking.
- **Attributes:**
  - `id`: `UUID` (Primary Key)
  - `user_id`: `UUID` REFERENCES `users(id)` ON DELETE CASCADE
  - `task_id`: `UUID` NULL REFERENCES `tasks(id)` ON DELETE SET NULL
  - `duration_minutes`: `INTEGER` NOT NULL DEFAULT 30
  - `started_at`: `TIMESTAMPTZ` NOT NULL
  - `ended_at`: `TIMESTAMPTZ` NOT NULL
  - `is_completed`: `BOOLEAN` NOT NULL DEFAULT `TRUE`

### 1.6 Entity: `lifestyle_records` (For Machine Learning Integration)
- **Purpose:** Capturing lifestyle and study habit variables planned in the research framework.
- **Attributes:**
  - `id`: `UUID` (Primary Key)
  - `user_id`: `UUID` REFERENCES `users(id)` ON DELETE CASCADE
  - `recorded_date`: `DATE` NOT NULL
  - `study_hours`: `NUMERIC(4,2)` NOT NULL DEFAULT 0.00
  - `sleep_hours`: `NUMERIC(4,2)` NOT NULL DEFAULT 0.00
  - `physical_activity_hours`: `NUMERIC(4,2)` DEFAULT 0.00
  - `digital_distraction_hours`: `NUMERIC(4,2)` DEFAULT 0.00
  - `predicted_productivity_score`: `NUMERIC(5,2)` NULL
  - `features_jsonb`: `JSONB` NULL (Raw features vector)
  - `created_at`: `TIMESTAMPTZ` NOT NULL DEFAULT `NOW()`

---

## 2. Privacy & Data Protection Safeguards

1. **Isolation of Personal vs Workspace Data:** Personal tasks (`workspace_id IS NULL`) are strictly filtered out of workspace reporting.
2. **Aggregated Reporting Views:** Administrative reporting views (e.g. `vw_team_productivity_summary`) compute counts and percentage aggregates only, ensuring administrators cannot query task text or personal notes.
3. **Soft-Delete Auditability:** All task and user deletions use `deleted_at`, preserving referential integrity for historic academic analytics.
