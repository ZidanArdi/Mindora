-- Migration: 000009_add_distributed_refresh_idempotency.down.sql
-- Rolls back idempotency and distributed concurrency tracking columns on auth_sessions table.

DROP INDEX IF EXISTS idx_auth_sessions_replaced_by_id;
DROP INDEX IF EXISTS idx_auth_sessions_rotation_op_id;
ALTER TABLE auth_sessions DROP CONSTRAINT IF EXISTS fk_auth_sessions_replaced_by;
ALTER TABLE auth_sessions DROP COLUMN IF EXISTS response_expires_at;
ALTER TABLE auth_sessions DROP COLUMN IF EXISTS encrypted_response;
ALTER TABLE auth_sessions DROP COLUMN IF EXISTS rotation_op_id;
ALTER TABLE auth_sessions DROP COLUMN IF EXISTS replaced_by_id;
