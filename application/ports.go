package application

import "context"

// TxManager はアプリケーション処理のトランザクション境界を定義します。
type TxManager interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}
