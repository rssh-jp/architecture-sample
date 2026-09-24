package txmanager_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"sample/infrastructure/txmanager"

	_ "modernc.org/sqlite"
)

func newTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = db.Close() })
	if _, err := db.Exec(`CREATE TABLE records (value TEXT NOT NULL)`); err != nil {
		t.Fatalf("create table error = %v", err)
	}
	return db
}

func TestDoCommitsAndProvidesTransaction(t *testing.T) {
	db := newTestDB(t)
	manager := txmanager.NewSqlTxManager(db)

	err := manager.Do(context.Background(), func(ctx context.Context) error {
		if txmanager.ExtractTx(ctx) == nil {
			t.Fatal("ExtractTx() = nil inside transaction")
		}
		_, err := txmanager.ExtractTx(ctx).ExecContext(ctx, `INSERT INTO records (value) VALUES ('committed')`)
		return err
	})
	if err != nil {
		t.Fatalf("Do() error = %v", err)
	}

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM records`).Scan(&count); err != nil {
		t.Fatalf("count query error = %v", err)
	}
	if count != 1 {
		t.Fatalf("record count = %d, want 1", count)
	}
}

func TestDoRollsBackOnCallbackError(t *testing.T) {
	db := newTestDB(t)
	manager := txmanager.NewSqlTxManager(db)
	callbackErr := errors.New("callback failed")

	err := manager.Do(context.Background(), func(ctx context.Context) error {
		if _, err := txmanager.ExtractTx(ctx).ExecContext(ctx, `INSERT INTO records (value) VALUES ('rolled back')`); err != nil {
			return err
		}
		return callbackErr
	})
	if !errors.Is(err, callbackErr) {
		t.Fatalf("error = %v, want %v", err, callbackErr)
	}

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM records`).Scan(&count); err != nil {
		t.Fatalf("count query error = %v", err)
	}
	if count != 0 {
		t.Fatalf("record count = %d, want 0", count)
	}
}

func TestExtractTxReturnsNilOutsideTransaction(t *testing.T) {
	if got := txmanager.ExtractTx(context.Background()); got != nil {
		t.Fatalf("ExtractTx() = %v outside transaction, want nil", got)
	}
}
