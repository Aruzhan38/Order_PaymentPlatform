package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"order-service/internal/domain"
	"order-service/internal/stream"
)

type OrderRepository interface {
	Create(ctx context.Context, order *domain.Order) error
	GetByID(ctx context.Context, id string) (*domain.Order, error)
	UpdateStatus(ctx context.Context, id string, status string) error
}

type PaymentClient interface {
	ProcessPayment(orderID string, amount int64) (string, error)
}

type OrderUsecase struct {
	repo          OrderRepository
	paymentClient PaymentClient
	streamManager *stream.OrderStreamManager
}

func NewOrderUsecase(
	repo OrderRepository,
	paymentClient PaymentClient,
	streamManager *stream.OrderStreamManager,
) *OrderUsecase {
	return &OrderUsecase{
		repo:          repo,
		paymentClient: paymentClient,
		streamManager: streamManager,
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

	u.streamManager.Publish(stream.StatusUpdate{
		OrderID: order.ID,
		Status:  "Pending",
		Message: "Order created",
	})

	status, err := u.paymentClient.ProcessPayment(order.ID, order.Amount)
	if err != nil {
		if err := u.repo.UpdateStatus(ctx, order.ID, "Failed"); err != nil {
			return nil, err
		}

		order.Status = "Failed"

		u.streamManager.Publish(stream.StatusUpdate{
			OrderID: order.ID,
			Status:  "Failed",
			Message: "Payment service unavailable or request failed",
		})

		return nil, err
	}

	if status == "Authorized" {
		if err := u.repo.UpdateStatus(ctx, order.ID, "Paid"); err != nil {
			return nil, err
		}

		order.Status = "Paid"

		u.streamManager.Publish(stream.StatusUpdate{
			OrderID: order.ID,
			Status:  "Paid",
			Message: "Payment authorized",
		})
	} else {
		if err := u.repo.UpdateStatus(ctx, order.ID, "Failed"); err != nil {
			return nil, err
		}

		order.Status = "Failed"

		u.streamManager.Publish(stream.StatusUpdate{
			OrderID: order.ID,
			Status:  "Failed",
			Message: "Payment declined",
		})
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

	u.streamManager.Publish(stream.StatusUpdate{
		OrderID: order.ID,
		Status:  "Cancelled",
		Message: "Order cancelled",
	})

	return order, nil
}
