package usecase

import (
	"Order_PaymentPlatform/internal/domain"
	"context"
	"errors"
	"github.com/google/uuid"
	"time"
)

type OrderRepository interface {
	Create(ctx context.Context, order *domain.Order) error
	GetByID(ctx context.Context, id string) (*domain.Order, error)
	UpdateStatus(ctx context.Context, id string, status string) error
}

type PaymentClient interface {
	AuthorizePayment(ctx context.Context, orderID string, amount int64) (string, string, error)
}

type OrderUsecase struct {
	repo          OrderRepository
	PaymentClient PaymentClient
}

func NewOrderUsecase(repo OrderRepository, paymentClient PaymentClient) *OrderUsecase {
	return &OrderUsecase{
		repo:          repo,
		PaymentClient: paymentClient,
	}
}

func (u *OrderUsecase) CreateOrder(ctx context.Context, customerID, itemName string, amount int64) (*domain.Order, error) {
	if amount <= 0 {
		return nil, errors.New("amount must be greater than 0")
	}

	order := &domain.Order{
		ID:         uuid.NewString(),
		CustomerID: customerID,
		ItemName:   itemName,
		Amount:     amount,
		Status:     "Pending",
		CreatedAt:  time.Now(),
	}

	if err := u.repo.Create(ctx, order); err != nil {
		return nil, err
	}
	status, _, err := u.PaymentClient.AuthorizePayment(ctx, order.ID, order.Amount)
	if err != nil {
		_ = u.repo.UpdateStatus(ctx, order.ID, "Failed")
		return nil, err
	}

	if status == "Authorized" {
		order.Status = "Paid"
		_ = u.repo.UpdateStatus(ctx, order.ID, "Paid")
	} else {
		order.Status = "Failed"
		_ = u.repo.UpdateStatus(ctx, order.ID, "Failed")
	}

	return order, nil

}

func (u *OrderUsecase) GetOrderByID(ctx context.Context, id string) (*domain.Order, error) {
	return u.repo.GetByID(ctx, id)
}

func (u *OrderUsecase) CancelOrder(ctx context.Context, id string) (*domain.Order, error) {
	order, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if order.Status != "Pending" {
		return nil, errors.New("only pending orders can be cancelled")
	}

	if err := u.repo.UpdateStatus(ctx, id, "Cancelled"); err != nil {
		return nil, err
	}

	order.Status = "Cancelled"
	return order, nil
}
