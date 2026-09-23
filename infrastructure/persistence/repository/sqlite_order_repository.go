package repository

import (
	"context"
	"database/sql"
	"sample/domain"
)

type sqliteOrderRepository struct {
	db *sql.DB
}

// NewSQLiteOrderRepository はSQLite用の注文リポジトリを生成します。
func NewSQLiteOrderRepository(db *sql.DB) *sqliteOrderRepository {
	return &sqliteOrderRepository{db: db}
}

// Save は注文を保存し、採番されたIDを注文へ設定します。
func (r *sqliteOrderRepository) Save(ctx context.Context, order *domain.Order) error {
	query := `INSERT INTO orders (user_id, amount) VALUES ($1, $2)`
	result, err := getDB(ctx, r.db).ExecContext(ctx, query, order.UserID, order.Amount)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	order.ID = id

	return nil
}
