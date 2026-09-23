package usecase

import (
	"context"
	"sample/domain"
	"sample/infrastructure/txmanager"
)

// UserUseCase はユーザーと注文に関するアプリケーション処理をまとめます。
type UserUseCase struct {
	txManager    txmanager.TxManager
	userRepo     domain.UserRepository
	orderRepo    domain.OrderRepository
	queryService UserQueryService
}

func NewUserUseCase(
	txm txmanager.TxManager,
	ur domain.UserRepository,
	or domain.OrderRepository,
	qs UserQueryService,
) *UserUseCase {
	// 各依存関係をユースケースへ注入します。
	return &UserUseCase{
		txManager:    txm,
		userRepo:     ur,
		orderRepo:    or,
		queryService: qs,
	}
}

// RegisterUserWithInitialOrder はユーザーと初回注文を同一トランザクションで登録します。
func (u *UserUseCase) RegisterUserWithInitialOrder(ctx context.Context, name, email string, initialAmount int) error {
	// ユーザー登録と注文登録を同じトランザクション境界で実行します。
	return u.txManager.Do(ctx, func(txCtx context.Context) error {
		user := &domain.User{Name: name, Email: email}
		if err := u.userRepo.Save(txCtx, user); err != nil {
			return err
		}

		order := &domain.Order{UserID: user.ID, Amount: initialAmount}
		if err := u.orderRepo.Save(txCtx, order); err != nil {
			return err
		}

		return nil
	})
}

// RegisterOrder は既存ユーザーへ注文を追加します。
func (u *UserUseCase) RegisterOrder(ctx context.Context, userid int64, initialAmount int) error {
	// ユーザーの存在確認と注文登録を同じトランザクションで実行します。
	return u.txManager.Do(ctx, func(txCtx context.Context) error {
		user, err := u.userRepo.FindByID(txCtx, userid)
		if err != nil {
			return err
		}

		order := &domain.Order{UserID: user.ID, Amount: initialAmount}
		if err := u.orderRepo.Save(txCtx, order); err != nil {
			return err
		}

		return nil
	})
}

// GetSummary は参照サービスからユーザーの注文集計を取得します。
func (u *UserUseCase) GetSummary(ctx context.Context, userID int64) (*UserOrderSummaryDTO, error) {
	return u.queryService.GetUserOrderSummary(ctx, userID)
}
