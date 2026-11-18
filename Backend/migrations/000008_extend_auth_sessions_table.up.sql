-- Migration: 000008_extend_auth_sessions_table.up.sql
ALTER TABLE auth_sessions ADD COLUMN IF NOT EXISTS family_id UUID;
ALTER TABLE auth_sessions ADD COLUMN IF NOT EXISTS ip_address VARCHAR(45);
ALTER TABLE auth_sessions ADD COLUMN IF NOT EXISTS user_agent TEXT;

CREATE INDEX IF NOT EXISTS idx_auth_sessions_family_id ON auth_sessions (family_id);
