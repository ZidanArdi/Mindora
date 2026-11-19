-- Migration: 000013_create_lifestyle_records_table.up.sql
CREATE TABLE IF NOT EXISTS lifestyle_records (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    record_date DATE NOT NULL,
    sleep_hours NUMERIC(4,2),
    phone_usage_hours NUMERIC(4,2),
    social_media_hours NUMERIC(4,2),
    youtube_hours NUMERIC(4,2),
    gaming_hours NUMERIC(4,2),
    coffee_intake_mg INTEGER,
    exercise_minutes INTEGER,
    stress_level INTEGER,
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT fk_lifestyle_records_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT uq_lifestyle_records_user_date UNIQUE (user_id, record_date),
    CONSTRAINT chk_lifestyle_stress_level CHECK (stress_level IS NULL OR (stress_level >= 1 AND stress_level <= 10)),
    CONSTRAINT chk_lifestyle_sleep CHECK (sleep_hours IS NULL OR (sleep_hours >= 0 AND sleep_hours <= 24)),
    CONSTRAINT chk_lifestyle_phone CHECK (phone_usage_hours IS NULL OR (phone_usage_hours >= 0 AND phone_usage_hours <= 24)),
    CONSTRAINT chk_lifestyle_social CHECK (social_media_hours IS NULL OR (social_media_hours >= 0 AND social_media_hours <= 24)),
    CONSTRAINT chk_lifestyle_youtube CHECK (youtube_hours IS NULL OR (youtube_hours >= 0 AND youtube_hours <= 24)),
    CONSTRAINT chk_lifestyle_gaming CHECK (gaming_hours IS NULL OR (gaming_hours >= 0 AND gaming_hours <= 24)),
    CONSTRAINT chk_lifestyle_coffee CHECK (coffee_intake_mg IS NULL OR coffee_intake_mg >= 0),
    CONSTRAINT chk_lifestyle_exercise CHECK (exercise_minutes IS NULL OR exercise_minutes >= 0)
);

CREATE INDEX IF NOT EXISTS idx_lifestyle_records_user_date ON lifestyle_records (user_id, record_date);
CREATE INDEX IF NOT EXISTS idx_lifestyle_records_deleted_at ON lifestyle_records (deleted_at);
