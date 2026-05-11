package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
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

type OrderCache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
}

type OrderUsecase struct {
	repo          OrderRepository
	paymentClient PaymentClient
	streamManager *stream.OrderStreamManager
	cache         OrderCache
	cacheTTL      time.Duration
}

func NewOrderUsecase(
	repo OrderRepository,
	paymentClient PaymentClient,
	streamManager *stream.OrderStreamManager,
	cache OrderCache,
	cacheTTL time.Duration,
) *OrderUsecase {
	return &OrderUsecase{
		repo:          repo,
		paymentClient: paymentClient,
		streamManager: streamManager,
		cache:         cache,
		cacheTTL:      cacheTTL,
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

		if u.cache != nil {
			_ = u.cache.Delete(ctx, fmt.Sprintf("order:%s", order.ID))
		}

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

		if u.cache != nil {
			_ = u.cache.Delete(ctx, fmt.Sprintf("order:%s", order.ID))
		}

		u.streamManager.Publish(stream.StatusUpdate{
			OrderID: order.ID,
			Status:  "Failed",
			Message: "Payment declined",
		})
	}

	return order, nil
}

func (u *OrderUsecase) GetOrderByID(ctx context.Context, id string) (*domain.Order, error) {
	cacheKey := fmt.Sprintf("order:%s", id)

	if u.cache != nil {
		cachedOrder, err := u.cache.Get(ctx, cacheKey)
		if err == nil && cachedOrder != "" {
			log.Println("CACHE HIT")
			var order domain.Order
			if err := json.Unmarshal([]byte(cachedOrder), &order); err == nil {
				return &order, nil
			}
		}

		log.Println("CACHE MISS")
	}

	order, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if u.cache != nil {
		orderJSON, err := json.Marshal(order)
		if err == nil {
			_ = u.cache.Set(ctx, cacheKey, string(orderJSON), u.cacheTTL)
		}
	}

	return order, nil
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

	if u.cache != nil {
		_ = u.cache.Delete(ctx, fmt.Sprintf("order:%s", id))
	}
	return order, nil
}
