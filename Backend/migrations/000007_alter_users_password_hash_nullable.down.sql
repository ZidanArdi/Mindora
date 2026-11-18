-- Migration: 000007_alter_users_password_hash_nullable.down.sql
ALTER TABLE users ALTER COLUMN password_hash SET NOT NULL;
