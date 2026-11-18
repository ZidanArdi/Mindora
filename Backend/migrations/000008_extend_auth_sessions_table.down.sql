-- Migration: 000008_extend_auth_sessions_table.down.sql
DROP INDEX IF EXISTS idx_auth_sessions_family_id;
ALTER TABLE auth_sessions DROP COLUMN IF EXISTS user_agent;
ALTER TABLE auth_sessions DROP COLUMN IF EXISTS ip_address;
ALTER TABLE auth_sessions DROP COLUMN IF EXISTS family_id;
