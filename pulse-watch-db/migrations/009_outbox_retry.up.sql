ALTER TABLE auth_outbox_events ADD COLUMN ready_at TIMESTAMPTZ NOT NULL DEFAULT now();
CREATE TABLE notification_receipts(message_id text PRIMARY KEY, delivered_at timestamptz NOT NULL DEFAULT now());
