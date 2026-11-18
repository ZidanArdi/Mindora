-- Migration: 000009_add_distributed_refresh_idempotency.up.sql
-- Adds idempotency and distributed concurrency tracking columns to auth_sessions table.

ALTER TABLE auth_sessions ADD COLUMN IF NOT EXISTS replaced_by_id UUID;
ALTER TABLE auth_sessions ADD COLUMN IF NOT EXISTS rotation_op_id VARCHAR(64);
ALTER TABLE auth_sessions ADD COLUMN IF NOT EXISTS encrypted_response TEXT;
ALTER TABLE auth_sessions ADD COLUMN IF NOT EXISTS response_expires_at TIMESTAMPTZ;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'fk_auth_sessions_replaced_by'
    ) THEN
        ALTER TABLE auth_sessions 
        ADD CONSTRAINT fk_auth_sessions_replaced_by 
        FOREIGN KEY (replaced_by_id) REFERENCES auth_sessions (id) ON DELETE SET NULL;
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_auth_sessions_rotation_op_id ON auth_sessions (rotation_op_id);
CREATE INDEX IF NOT EXISTS idx_auth_sessions_replaced_by_id ON auth_sessions (replaced_by_id);
