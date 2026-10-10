-- PulseWatch Database Migration: 002_create_monitors_table.up.sql
-- Description: Create monitors table for tracking endpoints and probe configurations

CREATE TABLE IF NOT EXISTS monitors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(150) NOT NULL,
    url TEXT NOT NULL,
    method VARCHAR(10) NOT NULL DEFAULT 'GET',
    interval_seconds INTEGER NOT NULL DEFAULT 60,
    timeout_seconds INTEGER NOT NULL DEFAULT 10,
    expected_status_code INTEGER NOT NULL DEFAULT 200,
    headers JSONB NOT NULL DEFAULT '{}'::jsonb,
    body TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING', -- PENDING, UP, DOWN, DEGRADED
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    ssl_check_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    last_checked_at TIMESTAMPTZ,
    next_check_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    uptime_24h NUMERIC(5, 2) NOT NULL DEFAULT 100.00,
    uptime_7d NUMERIC(5, 2) NOT NULL DEFAULT 100.00,
    uptime_30d NUMERIC(5, 2) NOT NULL DEFAULT 100.00,
    avg_response_time_ms INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_monitors_user_id ON monitors(user_id);
CREATE INDEX IF NOT EXISTS idx_monitors_status ON monitors(status);
CREATE INDEX IF NOT EXISTS idx_monitors_scheduler ON monitors(next_check_at, is_active) WHERE is_active = TRUE;
