// Package txmanager はデータベーストランザクションの開始と完了を管理します。
package txmanager

import (
	"context"
	"database/sql"
	"fmt"
	"sample/application"
)

// txKey はトランザクションをコンテキストへ格納するための非公開キーです。
type txKey struct{}

type sqlTxManager struct {
	db *sql.DB
}

// NewSqlTxManager はSQLデータベース用のトランザクション管理器を生成します。
func NewSqlTxManager(db *sql.DB) application.TxManager {
	return &sqlTxManager{db: db}
}

// Do はトランザクションを開始し、コンテキストへ埋め込んで処理を実行します。
func (m *sqlTxManager) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}

	// リポジトリが同じトランザクションを利用できるようにします。
	txCtx := context.WithValue(ctx, txKey{}, tx)

	if err := fn(txCtx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("err: %v, rollback err: %w", err, rbErr)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit tx: %w", err)
	}
	return nil
}

// ExtractTx はコンテキストからトランザクションを取得します。
func ExtractTx(ctx context.Context) *sql.Tx {
	if tx, ok := ctx.Value(txKey{}).(*sql.Tx); ok {
		return tx
	}
	return nil
}
