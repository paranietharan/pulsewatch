-- PulseWatch Database Migration: 001_create_users_table.down.sql
-- Description: Drop users table

DROP TABLE IF EXISTS users CASCADE;
DROP TYPE IF EXISTS user_role;
