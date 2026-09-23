// Package domain はアプリケーションのドメインモデルと永続化契約を定義します。
package domain

import "context"

// UserRepository はユーザーの永続化操作を定義します。
type UserRepository interface {
	Save(ctx context.Context, user *User) error
	FindByID(ctx context.Context, id int64) (*User, error)
}

// OrderRepository は注文の永続化操作を定義します。
type OrderRepository interface {
	Save(ctx context.Context, order *Order) error
}
