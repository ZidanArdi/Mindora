-- Migration: 000004_create_users_table.up.sql
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    role_id UUID NOT NULL,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(255) NOT NULL,
    password_hash TEXT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    email_verified_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT fk_users_role FOREIGN KEY (role_id) REFERENCES roles (id) ON DELETE RESTRICT,
    CONSTRAINT chk_users_status CHECK (status IN ('active', 'inactive', 'suspended'))
);

-- Case-insensitive email uniqueness reserving email even when account is soft-deleted
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email_lower ON users (LOWER(email));
CREATE INDEX IF NOT EXISTS idx_users_role_id ON users (role_id);
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users (deleted_at);
