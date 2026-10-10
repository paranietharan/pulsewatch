-- PulseWatch Database Migration: 003_create_heartbeats_table.up.sql
-- Description: Create heartbeats table for probe result time-series data

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'heartbeat_status') THEN
        CREATE TYPE heartbeat_status AS ENUM ('UP', 'DOWN', 'DEGRADED');
    END IF;
END$$;

CREATE TABLE IF NOT EXISTS heartbeats (
    id BIGSERIAL PRIMARY KEY,
    monitor_id UUID NOT NULL REFERENCES monitors(id) ON DELETE CASCADE,
    status heartbeat_status NOT NULL,
    status_code INTEGER,
    response_time_ms INTEGER NOT NULL,
    dns_lookup_ms INTEGER DEFAULT 0,
    tcp_connect_ms INTEGER DEFAULT 0,
    tls_handshake_ms INTEGER DEFAULT 0,
    error_message TEXT,
    region VARCHAR(50) NOT NULL DEFAULT 'default',
    checked_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_heartbeats_monitor_graph ON heartbeats(monitor_id, checked_at DESC)
INCLUDE (status, response_time_ms, status_code);
CREATE INDEX IF NOT EXISTS idx_heartbeats_checked_at ON heartbeats(checked_at DESC);
