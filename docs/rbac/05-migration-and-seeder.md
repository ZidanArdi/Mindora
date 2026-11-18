# 05 — Mindora Migrations and Seeder Guide

**Engine:** Go / PostgreSQL 18.6  
**Migrations Path:** `Backend/migrations/`  
**Tooling:** Version-Controlled Go SQL Migrator (`Backend/internal/database/migrator.go`, `Backend/cmd/migrate/main.go`)  

---

## 1. Migration Strategy & Tool Selection Rationale

### Selection Analysis
1. **Existing State:** The `Backend/migrations/` directory was previously empty, and no external migration binaries (such as `golang-migrate` or `goose`) were installed in the developer environment.
2. **Chosen Approach:** An in-tree, version-controlled SQL migration engine was engineered (`migrator.go`) backed by a dedicated CLI tool (`cmd/migrate/main.go`).
3. **Rationale:**
   - **Zero External Dependencies:** Eliminates requirement for developers to download or configure external CLI binaries on Windows/Linux.
   - **Explicit SQL Control:** Uses pure PostgreSQL `.up.sql` and `.down.sql` files for precise constraints (such as `LOWER(email)` functional unique indexes and check constraints) rather than relying on GORM AutoMigrate.
   - **Transaction Isolation:** Each migration runs inside an explicit SQL transaction (`BEGIN ... COMMIT`).
   - **State Tracking:** Tracks applied versions via `schema_migrations (version, name, applied_at)`.
   - **Standard Tool Compatibility:** Migration file naming matches `golang-migrate` standard (`000001_name.up.sql`), maintaining 100% future interoperability.

---

## 2. Migration Inventory

| Version | Migration Name | Up File | Down File | Description |
|---|---|---|---|---|
| `000001` | `create_roles_table` | `000001_create_roles_table.up.sql` | `000001_create_roles_table.down.sql` | Creates `roles` table and unique name index. |
| `000002` | `create_permissions_table` | `000002_create_permissions_table.up.sql` | `000002_create_permissions_table.down.sql` | Creates `permissions` table and `(resource, action)` constraint. |
| `000003` | `create_role_permissions_table`| `000003_create_role_permissions_table.up.sql` | `000003_create_role_permissions_table.down.sql` | Creates composite junction table with RESTRICT foreign keys. |
| `000004` | `create_users_table` | `000004_create_users_table.up.sql` | `000004_create_users_table.down.sql` | Creates `users` table with status CHECK, soft delete, and lower email index. |
| `000005` | `create_auth_sessions_table` | `000005_create_auth_sessions_table.up.sql` | `000005_create_auth_sessions_table.down.sql` | Creates `auth_sessions` table for refresh token hashes and revocation. |
| `000006` | `create_auth_identities_table` | `000006_create_auth_identities_table.up.sql` | `000006_create_auth_identities_table.down.sql` | Creates federated OAuth/OIDC identity mapping table. |
| `000007` | `alter_users_password_hash_nullable` | `000007_alter_users_password_hash_nullable.up.sql` | `000007_alter_users_password_hash_nullable.down.sql` | Makes `password_hash` column nullable for OAuth-only accounts. |
| `000008` | `extend_auth_sessions_table` | `000008_extend_auth_sessions_table.up.sql` | `000008_extend_auth_sessions_table.down.sql` | Adds `family_id`, `ip_address`, and `user_agent` to `auth_sessions`. |
| `000009` | `add_distributed_refresh_idempotency` | `000009_add_distributed_refresh_idempotency.up.sql` | `000009_add_distributed_refresh_idempotency.down.sql` | Adds `replaced_by_id`, `rotation_op_id`, `encrypted_response`, and `response_expires_at` for distributed coordination. |

---

## 3. Migration CLI Usage

From the `Backend/` directory:

### Check Migration Status
```bash
go run ./cmd/migrate -action status
```

### Apply All Pending Migrations
```bash
go run ./cmd/migrate -action up
```

### Roll Back Last Migration
```bash
go run ./cmd/migrate -action down -steps 1
```

---

## 4. RBAC Seeder Architecture & Execution

The seeder (`internal/database/seeder.go`) executes within an atomic database transaction.

### Features
1. **Deterministic UUIDs:** Uses UUIDv5 derived from a Mindora base namespace (`e07f7b58-86d3-4e89-b883-9366e60d3d5f`). This guarantees identical UUIDs across environments:
   - `user` Role UUID: `027e3141-0fec-5655-8ed2-b9ec6f37281c`
   - `admin` Role UUID: `9955ee39-659f-521e-b08d-99bf0339373c`
2. **Conflict Handling:** Uses PostgreSQL `ON CONFLICT (name) DO UPDATE` for roles, `ON CONFLICT (resource, action) DO UPDATE` for permissions, and `ON CONFLICT DO NOTHING` for role assignments.
3. **Idempotency:** Can be executed repeatedly without generating duplicate rows or errors.
4. **Security:** Does **NOT** seed user credentials or admin passwords.

### Execute Seeder
```bash
go run ./cmd/migrate -action seed
```

---

## 5. Database Verification Commands

Verify tables and relations via PostgreSQL CLI:

```bash
# List all tables
\dt

# Count permissions and assignments
SELECT r.name, count(rp.permission_id) 
FROM roles r 
JOIN role_permissions rp ON r.id = rp.role_id 
GROUP BY r.name;
```
