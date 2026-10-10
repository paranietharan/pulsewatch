-- PulseWatch Database Migration: 002_create_monitors_table.down.sql
-- Description: Drop monitors table and related ENUM types

DROP TABLE IF EXISTS monitors CASCADE;
DROP TYPE IF EXISTS monitor_status;
DROP TYPE IF EXISTS http_method;
