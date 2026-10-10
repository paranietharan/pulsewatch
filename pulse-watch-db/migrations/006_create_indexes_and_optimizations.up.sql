-- PulseWatch Database Migration: 006_create_indexes_and_optimizations.up.sql
-- Description: Automatic updated_at triggers and analytical views for monitors

-- Generic trigger function for updating updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply trigger to users
DROP TRIGGER IF EXISTS trg_users_updated_at ON users;
CREATE TRIGGER trg_users_updated_at
BEFORE UPDATE ON users
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- Apply trigger to monitors
DROP TRIGGER IF EXISTS trg_monitors_updated_at ON monitors;
CREATE TRIGGER trg_monitors_updated_at
BEFORE UPDATE ON monitors
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- Apply trigger to incidents
DROP TRIGGER IF EXISTS trg_incidents_updated_at ON incidents;
CREATE TRIGGER trg_incidents_updated_at
BEFORE UPDATE ON incidents
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- Apply trigger to alert_channels
DROP TRIGGER IF EXISTS trg_alert_channels_updated_at ON alert_channels;
CREATE TRIGGER trg_alert_channels_updated_at
BEFORE UPDATE ON alert_channels
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- View for Dashboard: Summary of active monitors with latest incident count
CREATE OR REPLACE VIEW v_monitor_dashboard_summary AS
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
