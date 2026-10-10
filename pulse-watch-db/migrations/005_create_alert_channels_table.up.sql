-- PulseWatch Database Migration: 005_create_alert_channels_table.up.sql
-- Description: Create alert channels and monitor associations for automated notifications

CREATE TABLE IF NOT EXISTS alert_channels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(30) NOT NULL, -- email, webhook, slack, pagerduty
    name VARCHAR(100) NOT NULL,
    config JSONB NOT NULL DEFAULT '{}'::jsonb, -- e.g. {"email": "ops@example.com"} or {"url": "https://api.webhook.com"}
    is_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_alert_channels_user_id ON alert_channels(user_id);

CREATE TABLE IF NOT EXISTS monitor_alert_channels (
    monitor_id UUID NOT NULL REFERENCES monitors(id) ON DELETE CASCADE,
    alert_channel_id UUID NOT NULL REFERENCES alert_channels(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (monitor_id, alert_channel_id)
);

CREATE INDEX IF NOT EXISTS idx_monitor_alert_channels_channel ON monitor_alert_channels(alert_channel_id);
