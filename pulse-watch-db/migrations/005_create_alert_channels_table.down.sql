-- PulseWatch Database Migration: 005_create_alert_channels_table.down.sql
-- Description: Drop monitor_alert_channels, alert_channels tables and alert_channel_type ENUM

DROP TABLE IF EXISTS monitor_alert_channels CASCADE;
DROP TABLE IF EXISTS alert_channels CASCADE;
DROP TYPE IF EXISTS alert_channel_type;
