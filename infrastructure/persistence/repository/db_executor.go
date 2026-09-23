// Package repository はドメインモデルの永続化処理を担当します。
package repository

import (
	"context"
	"database/sql"
	"sample/infrastructure/txmanager"
)

// DBExecutor は通常のDB接続とトランザクションに共通する実行操作です。
type DBExecutor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// getDB はトランザクションがコンテキストにあればそれを、なければDB接続を返します。
func getDB(ctx context.Context, db *sql.DB) DBExecutor {
	if tx := txmanager.ExtractTx(ctx); tx != nil {
		return tx
	}
	return db
}
