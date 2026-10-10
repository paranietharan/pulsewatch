-- PulseWatch Database Migration: 008_enhance_monitors_and_alerts.up.sql
-- Description: Add thresholds, job tracking, and alert delivery audit tables

-- 1. Add threshold and consecutive count columns to monitors table
ALTER TABLE monitors
    ADD COLUMN IF NOT EXISTS failure_threshold INTEGER NOT NULL DEFAULT 3 CHECK (failure_threshold >= 1),
    ADD COLUMN IF NOT EXISTS recovery_threshold INTEGER NOT NULL DEFAULT 2 CHECK (recovery_threshold >= 1),
    ADD COLUMN IF NOT EXISTS consecutive_failures INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS consecutive_successes INTEGER NOT NULL DEFAULT 0;

-- 2. Add job_id to heartbeats for end-to-end tracing and deduplication
ALTER TABLE heartbeats
    ADD COLUMN IF NOT EXISTS job_id UUID;

CREATE INDEX IF NOT EXISTS idx_heartbeats_job_id ON heartbeats(job_id) WHERE job_id IS NOT NULL;

-- 3. Create alert deliveries table to track notification outcomes and retries
CREATE TABLE IF NOT EXISTS alert_deliveries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    incident_id UUID NOT NULL REFERENCES incidents(id) ON DELETE CASCADE,
    alert_channel_id UUID REFERENCES alert_channels(id) ON DELETE SET NULL,
    channel_type alert_channel_type NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'SENDING', 'DELIVERED', 'FAILED')),
    attempt_count INTEGER NOT NULL DEFAULT 0,
    last_error TEXT,
    payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_alert_deliveries_incident ON alert_deliveries(incident_id);
CREATE INDEX IF NOT EXISTS idx_alert_deliveries_status ON alert_deliveries(status) WHERE status IN ('PENDING', 'SENDING');

-- Apply updated_at trigger to alert_deliveries
DROP TRIGGER IF EXISTS trg_alert_deliveries_updated_at ON alert_deliveries;
CREATE TRIGGER trg_alert_deliveries_updated_at
BEFORE UPDATE ON alert_deliveries
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- 4. Drop and recreate dashboard view to include enhanced fields
DROP VIEW IF EXISTS v_monitor_dashboard_summary;
CREATE VIEW v_monitor_dashboard_summary AS
SELECT 
    m.id,
    m.user_id,
    m.name,
    m.url,
    m.method,
    m.interval_seconds,
    m.timeout_seconds,
    m.failure_threshold,
    m.recovery_threshold,
    m.consecutive_failures,
    m.consecutive_successes,
    m.status,
    m.is_active,
    m.ssl_check_enabled,
    m.last_checked_at,
    m.next_check_at,
    m.uptime_24h,
    m.uptime_7d,
    m.uptime_30d,
    m.avg_response_time_ms,
    (
        SELECT COUNT(*)
        FROM incidents i
        WHERE i.monitor_id = m.id AND i.status = 'OPEN'
    ) AS open_incidents_count
FROM monitors m;
