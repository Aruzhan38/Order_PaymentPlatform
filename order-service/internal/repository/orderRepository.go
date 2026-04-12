package repository

import (
	"context"
	"database/sql"
	"order-service/internal/domain"
)

type OrderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) Create(ctx context.Context, order *domain.Order) error {
	query := `
		INSERT INTO orders (id, customer_id, item_name, amount, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.ExecContext(ctx, query,
		order.ID, order.CustomerID, order.ItemName,
		order.Amount, order.Status, order.CreatedAt,
	)
	return err
}

func (r *OrderRepository) GetByID(ctx context.Context, id string) (*domain.Order, error) {
	query := `
		SELECT id, customer_id, item_name, amount, status, created_at
		FROM orders WHERE id = $1
	`
	row := r.db.QueryRowContext(ctx, query, id)

	var order domain.Order
	err := row.Scan(
		&order.ID,
		&order.CustomerID,
		&order.ItemName,
		&order.Amount,
		&order.Status,
		&order.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &order, nil
}

func (r *OrderRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE orders SET status = $1 WHERE id = $2`,
		status, id,
	)
	return err
}
