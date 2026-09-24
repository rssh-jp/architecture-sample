package application_test

import (
	"context"
	"errors"
	"testing"

	"sample/application"
	"sample/domain"
)

type fakeTxManager struct {
	calls        int
	callbackCtx  context.Context
	callbackErr  error
	callbackRuns int
}

func (m *fakeTxManager) Do(ctx context.Context, fn func(context.Context) error) error {
	m.calls++
	m.callbackCtx = ctx
	callbackErr := fn(ctx)
	m.callbackRuns++
	if callbackErr != nil {
		return callbackErr
	}
	return m.callbackErr
}

type fakeUserRepository struct {
	saved      []*domain.User
	findResult *domain.User
	findErr    error
	saveErr    error
	findCalls  int
	saveCalls  int
	lastFindID int64
}

func (r *fakeUserRepository) Save(_ context.Context, user *domain.User) error {
	r.saveCalls++
	if r.saveErr != nil {
		return r.saveErr
	}
	r.saved = append(r.saved, user)
	if user.ID == 0 {
		user.ID = 10
	}
	return nil
}

func (r *fakeUserRepository) FindByID(_ context.Context, id int64) (*domain.User, error) {
	r.findCalls++
	r.lastFindID = id
	if r.findErr != nil {
		return nil, r.findErr
	}
	return r.findResult, nil
}

type fakeOrderRepository struct {
	saved     []*domain.Order
	saveErr   error
	saveCalls int
}

func (r *fakeOrderRepository) Save(_ context.Context, order *domain.Order) error {
	r.saveCalls++
	if r.saveErr != nil {
		return r.saveErr
	}
	r.saved = append(r.saved, order)
	return nil
}

type fakeQueryService struct {
	result *application.UserOrderSummaryDTO
	err    error
	calls  int
	lastID int64
}

func (s *fakeQueryService) GetUserOrderSummary(_ context.Context, userID int64) (*application.UserOrderSummaryDTO, error) {
	s.calls++
	s.lastID = userID
	return s.result, s.err
}

func newTestUserUseCase(txm *fakeTxManager, users *fakeUserRepository, orders *fakeOrderRepository, query *fakeQueryService) *application.UserUseCase {
	return application.NewUserUseCase(txm, users, orders, query)
}

func TestRegisterUserWithInitialOrder(t *testing.T) {
	txm := &fakeTxManager{}
	users := &fakeUserRepository{}
	orders := &fakeOrderRepository{}
	useCase := newTestUserUseCase(txm, users, orders, &fakeQueryService{})

	err := useCase.RegisterUserWithInitialOrder(context.Background(), "Alice", "alice@example.com", 5000)
	if err != nil {
		t.Fatalf("RegisterUserWithInitialOrder() error = %v", err)
	}
	if txm.calls != 1 || txm.callbackRuns != 1 {
		t.Fatalf("transaction calls = %d, callback runs = %d, want 1 and 1", txm.calls, txm.callbackRuns)
	}
	if len(users.saved) != 1 || users.saved[0].Name != "Alice" {
		t.Fatalf("saved users = %+v, want Alice", users.saved)
	}
	if len(orders.saved) != 1 {
		t.Fatalf("saved orders = %d, want 1", len(orders.saved))
	}
	if got := orders.saved[0]; got.UserID != 10 || got.Amount != 5000 {
		t.Fatalf("saved order = %+v, want user ID 10 and amount 5000", got)
	}
}

func TestRegisterUserWithInitialOrderStopsWhenUserSaveFails(t *testing.T) {
	userErr := errors.New("save user failed")
	txm := &fakeTxManager{}
	users := &fakeUserRepository{saveErr: userErr}
	orders := &fakeOrderRepository{}
	useCase := newTestUserUseCase(txm, users, orders, &fakeQueryService{})

	err := useCase.RegisterUserWithInitialOrder(context.Background(), "Alice", "alice@example.com", 5000)
	if !errors.Is(err, userErr) {
		t.Fatalf("error = %v, want %v", err, userErr)
	}
	if orders.saveCalls != 0 {
		t.Fatalf("order save calls = %d, want 0", orders.saveCalls)
	}
}

func TestRegisterUserWithInitialOrderPropagatesOrderError(t *testing.T) {
	orderErr := errors.New("save order failed")
	txm := &fakeTxManager{}
	users := &fakeUserRepository{}
	orders := &fakeOrderRepository{saveErr: orderErr}
	useCase := newTestUserUseCase(txm, users, orders, &fakeQueryService{})

	err := useCase.RegisterUserWithInitialOrder(context.Background(), "Alice", "alice@example.com", 5000)
	if !errors.Is(err, orderErr) {
		t.Fatalf("error = %v, want %v", err, orderErr)
	}
}

func TestRegisterOrder(t *testing.T) {
	txm := &fakeTxManager{}
	users := &fakeUserRepository{findResult: &domain.User{ID: 7, Name: "Alice"}}
	orders := &fakeOrderRepository{}
	useCase := newTestUserUseCase(txm, users, orders, &fakeQueryService{})

	err := useCase.RegisterOrder(context.Background(), 7, 3000)
	if err != nil {
		t.Fatalf("RegisterOrder() error = %v", err)
	}
	if users.findCalls != 1 || users.lastFindID != 7 {
		t.Fatalf("find calls = %d, ID = %d, want 1 and 7", users.findCalls, users.lastFindID)
	}
	if len(orders.saved) != 1 || orders.saved[0].UserID != 7 || orders.saved[0].Amount != 3000 {
		t.Fatalf("saved orders = %+v, want user ID 7 and amount 3000", orders.saved)
	}
}

func TestRegisterOrderPropagatesUserError(t *testing.T) {
	userErr := errors.New("user not found")
	txm := &fakeTxManager{}
	users := &fakeUserRepository{findErr: userErr}
	orders := &fakeOrderRepository{}
	useCase := newTestUserUseCase(txm, users, orders, &fakeQueryService{})

	err := useCase.RegisterOrder(context.Background(), 99, 3000)
	if !errors.Is(err, userErr) {
		t.Fatalf("error = %v, want %v", err, userErr)
	}
	if orders.saveCalls != 0 {
		t.Fatalf("order save calls = %d, want 0", orders.saveCalls)
	}
}

func TestGetSummary(t *testing.T) {
	expected := &application.UserOrderSummaryDTO{UserID: 7, UserName: "Alice", TotalOrders: 2, TotalAmount: 8000}
	query := &fakeQueryService{result: expected}
	useCase := newTestUserUseCase(&fakeTxManager{}, &fakeUserRepository{}, &fakeOrderRepository{}, query)

	got, err := useCase.GetSummary(context.Background(), 7)
	if err != nil {
		t.Fatalf("GetSummary() error = %v", err)
	}
	if got != expected || query.calls != 1 || query.lastID != 7 {
		t.Fatalf("result = %+v, calls = %d, ID = %d, want delegated result", got, query.calls, query.lastID)
	}
}
