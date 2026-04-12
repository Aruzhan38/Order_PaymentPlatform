package repository

import (
	"context"
	"database/sql"

	"Order_PaymentPlatform/internal/domain"
)

type PaymentRepository struct {
	db *sql.DB
}

func NewPaymentRepository(db *sql.DB) *PaymentRepository {
	return &PaymentRepository{db: db}
}

func (r *PaymentRepository) Create(ctx context.Context, payment *domain.Payment) error {
	query := `
		INSERT INTO payments (id, order_id, transaction_id, amount, status)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.ExecContext(ctx, query,
		payment.ID,
		payment.OrderID,
		payment.TransactionID,
		payment.Amount,
		payment.Status,
	)
	return err
}

func (r *PaymentRepository) GetByOrderID(ctx context.Context, orderID string) (*domain.Payment, error) {
	query := `
		SELECT id, order_id, transaction_id, amount, status
		FROM payments
		WHERE order_id = $1
	`

	row := r.db.QueryRowContext(ctx, query, orderID)

	var payment domain.Payment
	err := row.Scan(
		&payment.ID,
		&payment.OrderID,
		&payment.TransactionID,
		&payment.Amount,
		&payment.Status,
	)
	if err != nil {
		return nil, err
	}

	return &payment, nil
}
