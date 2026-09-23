// Package query は参照処理用のデータ取得を担当します。
package query

import (
	"context"
	"database/sql"
	"sample/usecase"
)

type sqliteUserQueryService struct {
	db *sql.DB
}

// NewSQLiteUserQueryService はSQLite用の参照サービスを生成します。
func NewSQLiteUserQueryService(db *sql.DB) usecase.UserQueryService {
	return &sqliteUserQueryService{db: db}
}

// GetUserOrderSummary はユーザーと注文を結合し、注文数と合計金額を取得します。
func (q *sqliteUserQueryService) GetUserOrderSummary(ctx context.Context, userID int64) (*usecase.UserOrderSummaryDTO, error) {
	// 注文がないユーザーも取得できるように左外部結合を使用します。
	query := `
		SELECT
			u.id,
			u.name,
			COUNT(o.id),
			COALESCE(SUM(o.amount), 0)
		FROM users u
		LEFT JOIN orders o ON u.id = o.user_id
		WHERE u.id = $1
		GROUP BY u.id, u.name
	`

	dto := &usecase.UserOrderSummaryDTO{}
	err := q.db.QueryRowContext(ctx, query, userID).Scan(
		&dto.UserID,
		&dto.UserName,
		&dto.TotalOrders,
		&dto.TotalAmount,
	)
	if err != nil {
		return nil, err
	}

	return dto, nil
}
