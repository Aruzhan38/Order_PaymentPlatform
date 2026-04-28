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
	GetStats(ctx context.Context) (int64, int64, int64, int64, error)
}

type EventPublisher interface {
	PublishPaymentCompleted(ctx context.Context, orderID string, amount int64, customerEmail string, status string) error
}
type PaymentUsecase struct {
	repo      PaymentRepository
	publisher EventPublisher
}

func NewPaymentUsecase(repo PaymentRepository, publisher EventPublisher) *PaymentUsecase {
	return &PaymentUsecase{
		repo:      repo,
		publisher: publisher,
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

	if status == "Authorized" && u.publisher != nil {
		customerEmail := "user@example.com"

		if err := u.publisher.PublishPaymentCompleted(
			ctx,
			payment.OrderID,
			payment.Amount,
			customerEmail,
			payment.Status,
		); err != nil {
			return nil, err
		}
	}

	return payment, nil
}

func (u *PaymentUsecase) GetPaymentByOrderID(ctx context.Context, orderID string) (*domain.Payment, error) {
	return u.repo.GetByOrderID(ctx, orderID)
}

func (u *PaymentUsecase) GetStats(ctx context.Context) (int64, int64, int64, int64, error) {
	return u.repo.GetStats(ctx)
}
