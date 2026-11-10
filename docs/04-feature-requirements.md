# 04 - Feature & Functional Requirements

**Product:** Mindora — Smart Student Productivity Application  
**Document Type:** Feature Inventory & Functional Specifications  
**Status:** Requirements Analysis (Phase 4)

---

## 1. Feature Inventory

| ID | Feature Name | Target Role | Priority | Description |
|---|---|---|---|---|
| **FEAT-01** | User Authentication & Role Selection | All Roles | Critical | Secure login, session management, and role-based redirect (User vs Admin). |
| **FEAT-02** | Task Management (CRUD & Prioritization) | User | Critical | Add, view, categorize, schedule, and mark tasks as completed. |
| **FEAT-03** | Focus Session & Pomodoro Timer | User | High | Set study timer durations, link sessions to specific tasks, track focus minutes. |
| **FEAT-04** | AI Planning Assistant | User | High | Natural language productivity recommendations and day-planning prompts. |
| **FEAT-05** | Calendar & Upcoming Schedule | User | Medium | Schedule visualization, upcoming academic deadlines, and meeting reminders. |
| **FEAT-06** | Workspace & Team Collaboration | User / Admin | High | Multi-member workspaces (e.g., "UI Design", "A-team Class") with shared tasks. |
| **FEAT-07** | Admin Dashboard & KPI Monitoring | Admin | High | High-level metrics: total users, active tasks, completion rates, workspace activity. |
| **FEAT-08** | Member & Role Access Control | Admin | High | Managing organizational members, assigning roles (Admin, Manager, Member), toggling status. |
| **FEAT-09** | Student Lifestyle & Habit Logging (ML Prep) | User | Future | Recording study hours, sleep duration, exercise, and digital distraction metrics. |

---

## 2. Detailed Functional Specifications

### 2.1 FEAT-01: Authentication & Role Selection
- **Inputs:** `email` (string, valid RFC 5322 format), `password` (string, min 8 characters), `rememberMe` (boolean), `role` (enum: `'user'` | `'admin'`).
- **Validation Rules:** Email must be non-empty and well-formed. Password cannot be blank.
- **Success State:** Redirects to `/` if `role == 'user'` or `/admin` if `role == 'admin'`.
- **Error State:** Displays inline error alert: *"Invalid email or password."*

### 2.2 FEAT-02: Task Management
- **Inputs:** `title` (string, 1-255 characters), `category` (string, e.g. Inbox, Work, Study, Health), `scheduleTime` (optional timestamp), `priority` (low, medium, high).
- **Interactions:**
  - Checkbox toggle marks task completed/uncompleted immediately with optimistic UI update.
  - Category selector assigns tag color and metadata.
- **States:**
  - *Loading:* Skeleton task rows.
  - *Empty:* *"All clear! No tasks scheduled for today."*

### 2.3 FEAT-03: Focus Session Timer
- **Inputs:** `durationMinutes` (integer between 5 and 120, default 30), `taskId` (optional link to active task).
- **Controls:** `Minus` (decrement 5m), `Plus` (increment 5m), `Start` / `Pause` / `Reset`.
- **Behavior:** Runs active countdown timer; visually reflects running status; stores completed focus interval for productivity analytics.

### 2.4 FEAT-04: AI Planning Assistant
- **Inputs:** Natural language prompt or predefined quick chips (*"Plan my day"*, *"Prioritize tasks"*, *"Create focus plan"*).
- **Expected Output:** Structured suggestions prioritizing academic deadlines, suggesting realistic study intervals, and minimizing burnout.
- **Guardrails:** AI responses must be supportive and non-judgmental; must not generate ungrounded statistics.

### 2.5 FEAT-07 & FEAT-08: Admin Management & Privacy Principles
- **Admin Visibility:**
  - Aggregated team statistics (Total members, Active tasks count, Organization-wide completion percentage).
  - Member access list (Name, email, assigned role, account status).
- **Strict Privacy Rule:**
  - Admin users CANNOT view individual student task titles, private study notes, mood logs, or personal schedule details.
  - Auditing is strictly confined to aggregated counts to respect academic and personal privacy.

---

## 3. Open Questions & Technical Clarifications

1. **AI / LLM Integration Provider:** Will the AI Planning Assistant utilize a cloud API (e.g. Gemini / OpenAI) or a self-hosted local model?
2. **Machine Learning Pipeline Interface:** The student productivity framework outlines a Random Forest Regressor running in Python (FastAPI). The interface contract between Golang Fiber and Python ML API will be specified in Phase 5.
