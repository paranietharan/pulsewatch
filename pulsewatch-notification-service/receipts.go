package main

import (
	"context"
	"database/sql"
	"errors"
	amqp "github.com/rabbitmq/amqp091-go"
	"log/slog"
)

type deliveryResult int

const (
	retry deliveryResult = iota
	discard
	delivered
)

// A receipt and the broker ack are ordered so a replay after commit skips the
// external send. An external provider can still accept a request before a crash.
func withReceipt(ctx context.Context, db *sql.DB, id string, deliver func() deliveryResult) (deliveryResult, error) {
	if id == "" {
		return deliver(), nil
	} // Compatibility with older producers.
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return retry, err
	}
	defer tx.Rollback()
	var claimed string
	err = tx.QueryRowContext(ctx, "INSERT INTO notification_receipts(message_id) VALUES($1) ON CONFLICT DO NOTHING RETURNING message_id", id).Scan(&claimed)
	if errors.Is(err, sql.ErrNoRows) {
		return delivered, nil
	}
	if err != nil {
		return retry, err
	}
	result := deliver()
	if result != delivered {
		return result, nil
	}
	if err := tx.Commit(); err != nil {
		return retry, err
	}
	return delivered, nil
}

func settle(ctx context.Context, db *sql.DB, d amqp.Delivery, logger *slog.Logger, deliver func() deliveryResult) {
	result, err := withReceipt(ctx, db, d.MessageId, deliver)
	if err != nil {
		logger.Error("Notification receipt failed", "error", err)
	}
	if result == delivered {
		err = d.Ack(false)
	} else {
		err = d.Nack(false, result == retry)
	}
	if err != nil {
		logger.Error("Notification acknowledgement failed", "error", err)
	}
}
