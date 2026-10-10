-- PulseWatch Database Migration: 004_create_incidents_table.down.sql
-- Description: Drop incident updates, incidents tables and their ENUM types

DROP TABLE IF EXISTS incident_updates CASCADE;
DROP TABLE IF EXISTS incidents CASCADE;
DROP TYPE IF EXISTS incident_update_status;
DROP TYPE IF EXISTS incident_status;
