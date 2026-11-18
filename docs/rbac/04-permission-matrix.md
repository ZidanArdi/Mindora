# 04 — Mindora RBAC Permission Matrix

**Classification:** Role-Based Access Control Mapping  
**Target Roles:** `Guest` (Unauthenticated), `User`, `Admin`  

---

## 1. Granular Permission Catalog

The Mindora RBAC system defines exactly 13 granular permissions:

| # | Permission Code | Resource | Action | Description | Assigned Roles |
|---|---|---|---|---|---|
| 1 | `profile.read` | profile | read | View personal profile information | User |
| 2 | `profile.update` | profile | update | Update personal profile details | User |
| 3 | `tasks.create` | tasks | create | Create personal productivity tasks | User |
| 4 | `tasks.read` | tasks | read | View own personal productivity tasks | User |
| 5 | `tasks.update` | tasks | update | Update own personal productivity tasks | User |
| 6 | `tasks.delete` | tasks | delete | Delete own personal productivity tasks | User |
| 7 | `users.create` | users | create | Provision new student user accounts | Admin |
| 8 | `users.read` | users | read | View student user account listings | Admin |
| 9 | `users.update` | users | update | Update student account status/metadata | Admin |
| 10 | `users.deactivate`| users | deactivate| Deactivate student user accounts | Admin |
| 11 | `users.activate` | users | activate | Reactivate student user accounts | Admin |
| 12 | `system.monitor` | system | monitor | View system status, health, and metrics | Admin |
| 13 | `audit_logs.read`| audit_logs | read | Inspect security audit and access logs | Admin |

---

## 2. Role-to-Permission Matrix

| Permission | Guest | User | Admin |
|---|:---:|:---:|:---:|
| `profile.read` | ❌ | ✅ | ❌ |
| `profile.update` | ❌ | ✅ | ❌ |
| `tasks.create` | ❌ | ✅ | ❌ |
| `tasks.read` | ❌ | ✅ | ❌ (Cannot view personal tasks) |
| `tasks.update` | ❌ | ✅ | ❌ |
| `tasks.delete` | ❌ | ✅ | ❌ |
| `users.create` | ❌ | ❌ | ✅ |
| `users.read` | ❌ | ❌ | ✅ |
| `users.update` | ❌ | ❌ | ✅ |
| `users.deactivate`| ❌ | ❌ | ✅ |
| `users.activate` | ❌ | ❌ | ✅ |
| `system.monitor` | ❌ (Health only) | ❌ | ✅ |
| `audit_logs.read` | ❌ | ❌ | ✅ |

---

## 3. Special Boundary & Governance Rules

1. **Strict Admin Privacy Boundary:**
   - Administrators do **not** inherit student productivity permissions (`tasks.*`).
   - Administrators cannot view personal student tasks, notes, or focus session details.
   - Any administrator report must be generated using aggregated statistical views (`COUNT`, `AVG`).
2. **User Isolation Policy:**
   - Standard users cannot access administrative endpoints (`/admin/*`).
   - Every personal resource access must be scoped to the authenticated caller's identity (`WHERE user_id = :authenticated_user_id`).
3. **Immutability of Role Assignment via Ordinary CRUD:**
   - Users cannot change their own role.
   - User creation API defaults strictly to `user` role.
   - Ordinary Admin CRUD cannot assign the `admin` role. Provisioning of administrative accounts requires dedicated CLI tools or elevated administrative procedures.
