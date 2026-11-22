-- Migration: 000012_create_focus_sessions_table.up.sql
CREATE TABLE IF NOT EXISTS focus_sessions (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    task_id UUID,
    session_type VARCHAR(20) NOT NULL,
    started_at TIMESTAMPTZ NOT NULL,
    ended_at TIMESTAMPTZ,
    duration_minutes INTEGER NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    interruption_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_focus_sessions_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT fk_focus_sessions_task FOREIGN KEY (task_id) REFERENCES tasks (id) ON DELETE SET NULL,
    CONSTRAINT chk_focus_session_type CHECK (session_type IN ('focus', 'break')),
    CONSTRAINT chk_focus_session_status CHECK (status IN ('active', 'completed', 'abandoned')),
    CONSTRAINT chk_focus_session_duration CHECK (duration_minutes >= 0),
    CONSTRAINT chk_focus_session_interruptions CHECK (interruption_count >= 0)
);

CREATE INDEX IF NOT EXISTS idx_focus_sessions_user_id ON focus_sessions (user_id);
CREATE INDEX IF NOT EXISTS idx_focus_sessions_user_type_status ON focus_sessions (user_id, session_type, status);
CREATE INDEX IF NOT EXISTS idx_focus_sessions_started_at ON focus_sessions (started_at);
