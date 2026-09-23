package domain

import (
	"time"
)

// User はユーザーを表すドメインエンティティです。
type User struct {
	ID        int64
	Name      string
	Email     string
	CreatedAt time.Time
}

// Order はユーザーに紐づく注文を表すドメインエンティティです。
type Order struct {
	ID        int64
	UserID    int64
	Amount    int
	CreatedAt time.Time
}
