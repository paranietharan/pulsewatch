-- PulseWatch Database Migration: 005_create_alert_channels_table.down.sql
-- Description: Drop monitor_alert_channels and alert_channels tables

DROP TABLE IF EXISTS monitor_alert_channels CASCADE;
DROP TABLE IF EXISTS alert_channels CASCADE;
