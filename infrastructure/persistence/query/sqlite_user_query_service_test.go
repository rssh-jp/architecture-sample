package query_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"sample/infrastructure/persistence/query"

	_ "modernc.org/sqlite"
)

func newQueryTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = db.Close() })
	_, err = db.Exec(`
		CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT NOT NULL);
		CREATE TABLE orders (id INTEGER PRIMARY KEY, user_id INTEGER NOT NULL, amount INTEGER NOT NULL);
		INSERT INTO users (id, name) VALUES (1, 'Alice'), (2, 'Bob');
		INSERT INTO orders (id, user_id, amount) VALUES (1, 1, 5000), (2, 1, 3000);`)
	if err != nil {
		t.Fatalf("setup error = %v", err)
	}
	return db
}

func TestSQLiteUserQueryServiceReturnsSummary(t *testing.T) {
	service := query.NewSQLiteUserQueryService(newQueryTestDB(t))

	got, err := service.GetUserOrderSummary(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetUserOrderSummary() error = %v", err)
	}
	want := struct {
		id     int64
		name   string
		orders int
		amount int
	}{1, "Alice", 2, 8000}
	if got.UserID != want.id || got.UserName != want.name || got.TotalOrders != want.orders || got.TotalAmount != want.amount {
		t.Fatalf("summary = %+v, want ID=%d name=%q orders=%d amount=%d", got, want.id, want.name, want.orders, want.amount)
	}
}

func TestSQLiteUserQueryServiceReturnsZeroForUserWithoutOrders(t *testing.T) {
	service := query.NewSQLiteUserQueryService(newQueryTestDB(t))

	got, err := service.GetUserOrderSummary(context.Background(), 2)
	if err != nil {
		t.Fatalf("GetUserOrderSummary() error = %v", err)
	}
	if got.UserID != 2 || got.UserName != "Bob" || got.TotalOrders != 0 || got.TotalAmount != 0 {
		t.Fatalf("summary = %+v, want Bob with zero orders and amount", got)
	}
}

func TestSQLiteUserQueryServiceReturnsNotFound(t *testing.T) {
	service := query.NewSQLiteUserQueryService(newQueryTestDB(t))

	_, err := service.GetUserOrderSummary(context.Background(), 999)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("error = %v, want sql.ErrNoRows", err)
	}
}
