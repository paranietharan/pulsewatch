package repository

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"os"
	"testing"
)

func TestOutboxRetainsIdentityAfterFailedPublish(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL required")
	}
	ctx := context.Background()
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatal(err)
	}
	cfg.MaxConns = 1
	db, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	_, err = db.Exec(ctx, `CREATE TEMP TABLE auth_outbox_events (LIKE public.auth_outbox_events INCLUDING ALL)`)
	if err != nil {
		t.Fatal(err)
	}
	id := uuid.NewString()
	if _, err = db.Exec(ctx, `INSERT INTO auth_outbox_events(id,event_type,routing_key,payload) VALUES($1,'OPEN','incident','{}')`, id); err != nil {
		t.Fatal(err)
	}
	down := errors.New("broker unavailable")
	attempts := 0
	publish := func(_ context.Context, _ string, _ []byte, messageID string) error {
		attempts++
		if messageID != id {
			t.Fatal("message identity changed")
		}
		if attempts == 1 {
			return down
		}
		return nil
	}
	repo := NewOutboxRepository(db)
	if _, err := repo.DispatchOne(ctx, publish); !errors.Is(err, down) {
		t.Fatalf("unexpected failure: %v", err)
	}
	if _, err = db.Exec(ctx, "UPDATE auth_outbox_events SET ready_at=now()"); err != nil {
		t.Fatal(err)
	}
	if ok, err := NewOutboxRepository(db).DispatchOne(ctx, publish); err != nil || !ok {
		t.Fatal(ok, err)
	}
	if ok, err := repo.DispatchOne(ctx, publish); err != nil || ok {
		t.Fatal(ok, err)
	}
}
