package grpc

import (
	"context"
	"log"

	paymentpb "github.com/Aruzhan38/order-payment-generated/proto/payment"
	"payment-service/internal/usecase"
)

type PaymentServer struct {
	paymentpb.UnimplementedPaymentServiceServer
	uc *usecase.PaymentUsecase
}

func NewPaymentServer(uc *usecase.PaymentUsecase) *PaymentServer {
	return &PaymentServer{uc: uc}
}

func (s *PaymentServer) ProcessPayment(ctx context.Context, req *paymentpb.PaymentRequest) (*paymentpb.PaymentResponse, error) {
	log.Printf("gRPC ProcessPayment called: order_id=%s amount=%d", req.OrderId, req.Amount)

	payment, err := s.uc.CreatePayment(ctx, req.OrderId, req.Amount)
	if err != nil {
		return nil, err
	}

	log.Printf("Payment processed: status=%s transaction_id=%s", payment.Status, payment.TransactionID)

	return &paymentpb.PaymentResponse{
		Status:        payment.Status,
		TransactionId: payment.TransactionID,
	}, nil
}
