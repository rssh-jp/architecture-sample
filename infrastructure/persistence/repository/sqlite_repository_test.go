package repository_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"sample/domain"
	"sample/infrastructure/persistence/repository"
	"sample/infrastructure/txmanager"

	_ "modernc.org/sqlite"
)

func newRepositoryTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = db.Close() })
	_, err = db.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			email TEXT NOT NULL
		);
		CREATE TABLE orders (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			amount INTEGER NOT NULL
		);`)
	if err != nil {
		t.Fatalf("create tables error = %v", err)
	}
	return db
}

func TestSQLiteUserRepositorySaveAndFind(t *testing.T) {
	db := newRepositoryTestDB(t)
	repo := repository.NewSQLiteUserRepository(db)
	user := &domain.User{Name: "Alice", Email: "alice@example.com"}

	if err := repo.Save(context.Background(), user); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	if user.ID == 0 {
		t.Fatal("Save() did not set user ID")
	}

	got, err := repo.FindByID(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if got.ID != user.ID || got.Name != user.Name || got.Email != user.Email {
		t.Fatalf("FindByID() = %+v, want %+v", got, user)
	}
}

func TestSQLiteOrderRepositorySave(t *testing.T) {
	db := newRepositoryTestDB(t)
	users := repository.NewSQLiteUserRepository(db)
	orders := repository.NewSQLiteOrderRepository(db)
	user := &domain.User{Name: "Alice", Email: "alice@example.com"}
	if err := users.Save(context.Background(), user); err != nil {
		t.Fatalf("user Save() error = %v", err)
	}
	order := &domain.Order{UserID: user.ID, Amount: 5000}

	if err := orders.Save(context.Background(), order); err != nil {
		t.Fatalf("order Save() error = %v", err)
	}
	if order.ID == 0 {
		t.Fatal("Save() did not set order ID")
	}

	var amount int
	if err := db.QueryRow(`SELECT amount FROM orders WHERE id = ?`, order.ID).Scan(&amount); err != nil {
		t.Fatalf("order query error = %v", err)
	}
	if amount != 5000 {
		t.Fatalf("amount = %d, want 5000", amount)
	}
}

func TestRepositoriesUseSameTransactionAndRollback(t *testing.T) {
	db := newRepositoryTestDB(t)
	manager := txmanager.NewSqlTxManager(db)
	users := repository.NewSQLiteUserRepository(db)
	orders := repository.NewSQLiteOrderRepository(db)
	callbackErr := errors.New("rollback")

	err := manager.Do(context.Background(), func(ctx context.Context) error {
		user := &domain.User{Name: "Alice", Email: "alice@example.com"}
		if err := users.Save(ctx, user); err != nil {
			return err
		}
		if err := orders.Save(ctx, &domain.Order{UserID: user.ID, Amount: 5000}); err != nil {
			return err
		}
		return callbackErr
	})
	if !errors.Is(err, callbackErr) {
		t.Fatalf("error = %v, want %v", err, callbackErr)
	}

	var usersCount, ordersCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&usersCount); err != nil {
		t.Fatalf("users count error = %v", err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM orders`).Scan(&ordersCount); err != nil {
		t.Fatalf("orders count error = %v", err)
	}
	if usersCount != 0 || ordersCount != 0 {
		t.Fatalf("counts = users:%d orders:%d, want 0 and 0", usersCount, ordersCount)
	}
}

func TestSQLiteUserRepositoryFindByIDReturnsNotFound(t *testing.T) {
	db := newRepositoryTestDB(t)
	repo := repository.NewSQLiteUserRepository(db)

	_, err := repo.FindByID(context.Background(), 999)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("error = %v, want sql.ErrNoRows", err)
	}
}
