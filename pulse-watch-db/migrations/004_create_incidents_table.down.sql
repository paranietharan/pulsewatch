-- PulseWatch Database Migration: 004_create_incidents_table.down.sql
-- Description: Drop incident updates and incidents tables

DROP TABLE IF EXISTS incident_updates CASCADE;
DROP TABLE IF EXISTS incidents CASCADE;
