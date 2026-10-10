-- PulseWatch Database Migration: 003_create_heartbeats_table.down.sql
-- Description: Drop heartbeats table and heartbeat_status ENUM

DROP TABLE IF EXISTS heartbeats CASCADE;
DROP TYPE IF EXISTS heartbeat_status;
