// Package usecase はアプリケーションのユースケースと参照用データを定義します。
package usecase

import "context"

// UserOrderSummaryDTO はユーザーと注文の集計結果を受け取る参照用DTOです。
type UserOrderSummaryDTO struct {
	UserID      int64  `json:"user_id"`
	UserName    string `json:"user_name"`
	TotalOrders int    `json:"total_orders"`
	TotalAmount int    `json:"total_amount"`
}

// UserQueryService は参照処理用の問い合わせサービスを定義します。
type UserQueryService interface {
	GetUserOrderSummary(ctx context.Context, userID int64) (*UserOrderSummaryDTO, error)
}
