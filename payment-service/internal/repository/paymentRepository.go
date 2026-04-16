package repository

import (
	"context"
	"database/sql"
	"payment-service/internal/domain"
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

	_, err := r.db.ExecContext(
		ctx,
		query,
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

	var payment domain.Payment
	err := r.db.QueryRowContext(ctx, query, orderID).Scan(
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

func (r *PaymentRepository) GetStats(ctx context.Context) (int64, int64, int64, int64, error) {
	query := `
		SELECT
			COUNT(*) AS total_count,
			COUNT(*) FILTER (WHERE status = 'Authorized') AS authorized_count,
			COUNT(*) FILTER (WHERE status = 'Declined') AS declined_count,
			COALESCE(SUM(amount), 0) AS total_amount
		FROM payments
	`

	var totalCount int64
	var authorizedCount int64
	var declinedCount int64
	var totalAmount int64

	err := r.db.QueryRowContext(ctx, query).Scan(
		&totalCount,
		&authorizedCount,
		&declinedCount,
		&totalAmount,
	)
	if err != nil {
		return 0, 0, 0, 0, err
	}

	return totalCount, authorizedCount, declinedCount, totalAmount, nil
}
