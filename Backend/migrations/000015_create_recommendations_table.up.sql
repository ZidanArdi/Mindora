-- Up Migration: 000015_create_recommendations_table.up.sql

CREATE TABLE IF NOT EXISTS recommendations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    rule_id VARCHAR(50) NOT NULL,
    category VARCHAR(50) NOT NULL,
    priority VARCHAR(20) NOT NULL DEFAULT 'medium',
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    explanation TEXT NOT NULL,
    evidence JSONB NOT NULL DEFAULT '{}'::jsonb,
    action_route VARCHAR(100),
    action_label VARCHAR(100),
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    evaluation_period_start TIMESTAMPTZ NOT NULL,
    evaluation_period_end TIMESTAMPTZ NOT NULL,
    evaluation_key VARCHAR(150) NOT NULL,
    dismissed_until TIMESTAMPTZ,
    valid_until TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_recommendations_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT chk_recommendations_priority CHECK (priority IN ('critical', 'high', 'medium', 'low')),
    CONSTRAINT chk_recommendations_status CHECK (status IN ('pending', 'dismissed', 'completed', 'expired', 'invalidated'))
);

-- Index 1: Enforce at most ONE active pending recommendation per user per rule (Idempotency)
CREATE UNIQUE INDEX IF NOT EXISTS uq_recommendations_user_rule_pending 
    ON recommendations (user_id, rule_id) 
    WHERE status = 'pending';

-- Index 2: Query index for active valid user recommendations
CREATE INDEX IF NOT EXISTS idx_recommendations_user_status_valid 
    ON recommendations (user_id, status, valid_until);

-- Index 3: Index for audit trails and history listing
CREATE INDEX IF NOT EXISTS idx_recommendations_user_created 
    ON recommendations (user_id, created_at DESC);

-- Index 4: Evaluation key lookup for cooldown and period deduplication
CREATE INDEX IF NOT EXISTS idx_recommendations_user_eval_key 
    ON recommendations (user_id, evaluation_key);
