DROP TABLE notification_receipts;
ALTER TABLE auth_outbox_events DROP COLUMN ready_at;
