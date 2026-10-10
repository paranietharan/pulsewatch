-- PulseWatch Database Migration: 006_create_indexes_and_optimizations.down.sql
-- Description: Drop analytical views, triggers, and trigger function

DROP VIEW IF EXISTS v_monitor_dashboard_summary;
DROP TRIGGER IF EXISTS trg_alert_channels_updated_at ON alert_channels;
DROP TRIGGER IF EXISTS trg_incidents_updated_at ON incidents;
DROP TRIGGER IF EXISTS trg_monitors_updated_at ON monitors;
DROP TRIGGER IF EXISTS trg_users_updated_at ON users;
DROP FUNCTION IF EXISTS update_updated_at_column();
