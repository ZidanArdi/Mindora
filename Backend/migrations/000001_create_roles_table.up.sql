-- Migration: 000001_create_roles_table.up.sql
CREATE TABLE IF NOT EXISTS roles (
    id UUID PRIMARY KEY,
    name VARCHAR(30) NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_roles_name UNIQUE (name)
);

CREATE INDEX IF NOT EXISTS idx_roles_name ON roles (name);
