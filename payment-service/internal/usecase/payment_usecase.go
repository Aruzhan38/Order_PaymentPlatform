package usecase

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"payment-service/internal/domain"
)

type PaymentRepository interface {
	Create(ctx context.Context, payment *domain.Payment) error
	GetByOrderID(ctx context.Context, orderID string) (*domain.Payment, error)
}

type PaymentUsecase struct {
	repo PaymentRepository
}

func NewPaymentUsecase(repo PaymentRepository) *PaymentUsecase {
	return &PaymentUsecase{
		repo: repo,
	}
}

func (u *PaymentUsecase) CreatePayment(ctx context.Context, orderID string, amount int64) (*domain.Payment, error) {
	if orderID == "" {
		return nil, errors.New("order_id is required")
	}

	if amount <= 0 {
		return nil, errors.New("amount must be greater than 0")
	}

	status := "Authorized"
	if amount > 100000 {
		status = "Declined"
	}

	payment := &domain.Payment{
		ID:            uuid.NewString(),
		OrderID:       orderID,
		TransactionID: uuid.NewString(),
		Amount:        amount,
		Status:        status,
	}

	if err := u.repo.Create(ctx, payment); err != nil {
		return nil, err
	}

	return payment, nil
}

func (u *PaymentUsecase) GetPaymentByOrderID(ctx context.Context, orderID string) (*domain.Payment, error) {
	return u.repo.GetByOrderID(ctx, orderID)
}
