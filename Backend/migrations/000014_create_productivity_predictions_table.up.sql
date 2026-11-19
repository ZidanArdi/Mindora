-- Migration: 000014_create_productivity_predictions_table.up.sql
CREATE TABLE IF NOT EXISTS productivity_predictions (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    predicted_score NUMERIC(5,2) NOT NULL,
    model_version VARCHAR(50) NOT NULL,
    study_hours NUMERIC(4,2) NOT NULL DEFAULT 0.00,
    sleep_hours NUMERIC(4,2),
    phone_usage_hours NUMERIC(4,2),
    stress_level INTEGER,
    features_snapshot JSONB NOT NULL DEFAULT '{}'::jsonb,
    prediction_date DATE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_productivity_predictions_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT chk_productivity_predictions_score CHECK (predicted_score >= 0.00 AND predicted_score <= 100.00),
    CONSTRAINT chk_productivity_predictions_study CHECK (study_hours >= 0.00 AND study_hours <= 24.00),
    CONSTRAINT chk_productivity_predictions_sleep CHECK (sleep_hours IS NULL OR (sleep_hours >= 0.00 AND sleep_hours <= 24.00)),
    CONSTRAINT chk_productivity_predictions_phone CHECK (phone_usage_hours IS NULL OR (phone_usage_hours >= 0.00 AND phone_usage_hours <= 24.00)),
    CONSTRAINT chk_productivity_predictions_stress CHECK (stress_level IS NULL OR (stress_level >= 1 AND stress_level <= 10))
);

CREATE INDEX IF NOT EXISTS idx_productivity_predictions_user_date ON productivity_predictions (user_id, prediction_date);
CREATE INDEX IF NOT EXISTS idx_productivity_predictions_user_created ON productivity_predictions (user_id, created_at DESC);
