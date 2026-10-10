-- PulseWatch Database Migration: 008_enhance_monitors_and_alerts.down.sql
-- Description: Revert threshold additions and alert deliveries table

-- Revert view to 006 definition
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

DROP TRIGGER IF EXISTS trg_alert_deliveries_updated_at ON alert_deliveries;
DROP TABLE IF EXISTS alert_deliveries;

DROP INDEX IF EXISTS idx_heartbeats_job_id;
ALTER TABLE heartbeats DROP COLUMN IF EXISTS job_id;

ALTER TABLE monitors
    DROP COLUMN IF EXISTS failure_threshold,
    DROP COLUMN IF EXISTS recovery_threshold,
    DROP COLUMN IF EXISTS consecutive_failures,
    DROP COLUMN IF EXISTS consecutive_successes;
