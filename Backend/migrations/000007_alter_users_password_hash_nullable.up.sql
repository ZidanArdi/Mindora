-- Migration: 000007_alter_users_password_hash_nullable.up.sql
ALTER TABLE users ALTER COLUMN password_hash DROP NOT NULL;
