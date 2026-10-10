-- PulseWatch Database Migration: 007_create_auth_tables.down.sql
-- Description: Revert authentication tables and columns

DROP TABLE IF EXISTS auth_outbox_events CASCADE;
DROP TABLE IF EXISTS password_reset_tokens CASCADE;
DROP TABLE IF EXISTS email_verification_tokens CASCADE;
DROP TABLE IF EXISTS refresh_tokens CASCADE;

ALTER TABLE users 
DROP COLUMN IF EXISTS last_login_at,
DROP COLUMN IF EXISTS is_verified;
