package main

import (
	"context"
	"database/sql"
	"github.com/google/uuid"
	"os"
	"testing"
)

func TestCommittedReceiptSuppressesReplayAndFailuresRetry(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL with migration 009 required")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	id := uuid.NewString()
	ctx := context.Background()
	calls := 0
	send := func() deliveryResult {
		calls++
		if calls == 1 {
			return retry
		}
		return delivered
	}
	for _, want := range []deliveryResult{retry, delivered, delivered} {
		result, err := withReceipt(ctx, db, id, send)
		if err != nil || result != want {
			t.Fatalf("result %v want %v: %v", result, want, err)
		}
	}
	if calls != 2 {
		t.Fatalf("sent %d times", calls)
	}
}
