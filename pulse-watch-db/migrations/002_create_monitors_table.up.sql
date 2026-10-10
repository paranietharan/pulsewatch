-- PulseWatch Database Migration: 002_create_monitors_table.up.sql
-- Description: Create monitors table for tracking endpoints and probe configurations

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'http_method') THEN
        CREATE TYPE http_method AS ENUM ('GET', 'POST', 'PUT', 'PATCH', 'DELETE', 'HEAD', 'OPTIONS');
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'monitor_status') THEN
        CREATE TYPE monitor_status AS ENUM ('PENDING', 'UP', 'DOWN', 'DEGRADED');
    END IF;
END$$;

CREATE TABLE IF NOT EXISTS monitors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(150) NOT NULL,
    url TEXT NOT NULL,
    method http_method NOT NULL DEFAULT 'GET',
    interval_seconds INTEGER NOT NULL DEFAULT 60 CHECK (interval_seconds >= 10),
    timeout_seconds INTEGER NOT NULL DEFAULT 10 CHECK (timeout_seconds > 0 AND timeout_seconds <= interval_seconds),
    expected_status_code INTEGER NOT NULL DEFAULT 200,
    headers JSONB NOT NULL DEFAULT '{}'::jsonb,
    body TEXT,
    status monitor_status NOT NULL DEFAULT 'PENDING',
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
CREATE INDEX IF NOT EXISTS idx_monitors_scheduler ON monitors (next_check_at ASC)
INCLUDE (id, url, method, interval_seconds, timeout_seconds, expected_status_code)
WHERE is_active = TRUE;
