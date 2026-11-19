-- Migration: 000011_create_study_sessions_table.up.sql
CREATE TABLE IF NOT EXISTS study_sessions (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    task_id UUID,
    subject VARCHAR(100) NOT NULL,
    duration_seconds INTEGER NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    started_at TIMESTAMPTZ NOT NULL,
    ended_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_study_sessions_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT fk_study_sessions_task FOREIGN KEY (task_id) REFERENCES tasks (id) ON DELETE SET NULL,
    CONSTRAINT chk_study_sessions_status CHECK (status IN ('active', 'paused', 'completed', 'cancelled')),
    CONSTRAINT chk_study_sessions_duration CHECK (duration_seconds >= 0)
);

CREATE INDEX IF NOT EXISTS idx_study_sessions_user_id ON study_sessions (user_id);
CREATE INDEX IF NOT EXISTS idx_study_sessions_user_status ON study_sessions (user_id, status);
CREATE INDEX IF NOT EXISTS idx_study_sessions_started_at ON study_sessions (started_at);
