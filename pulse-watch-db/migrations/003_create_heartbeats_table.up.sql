-- PulseWatch Database Migration: 003_create_heartbeats_table.up.sql
-- Description: Create heartbeats table for probe result time-series data

CREATE TABLE IF NOT EXISTS heartbeats (
    id BIGSERIAL PRIMARY KEY,
    monitor_id UUID NOT NULL REFERENCES monitors(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL, -- UP, DOWN, DEGRADED
    status_code INTEGER,
    response_time_ms INTEGER NOT NULL,
    dns_lookup_ms INTEGER DEFAULT 0,
    tcp_connect_ms INTEGER DEFAULT 0,
    tls_handshake_ms INTEGER DEFAULT 0,
    error_message TEXT,
    region VARCHAR(50) NOT NULL DEFAULT 'default',
    checked_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_heartbeats_monitor_checked_at ON heartbeats(monitor_id, checked_at DESC);
CREATE INDEX IF NOT EXISTS idx_heartbeats_checked_at ON heartbeats(checked_at DESC);
