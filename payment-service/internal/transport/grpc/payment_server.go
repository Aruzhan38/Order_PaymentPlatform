package grpc

import (
	"context"
	"log"

	paymentpb "github.com/Aruzhan38/order-payment-generated/proto/payment"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

	if req.OrderId == "" {
		return nil, status.Error(codes.InvalidArgument, "order_id is required")
	}
	if req.Amount <= 0 {
		return nil, status.Error(codes.InvalidArgument, "amount must be greater than 0")
	}

	payment, err := s.uc.CreatePayment(ctx, req.OrderId, req.Amount)
	if err != nil {
		log.Printf("failed to process payment: %v", err)
		return nil, status.Error(codes.Internal, "failed to process payment")
	}

	log.Printf("Payment processed: status=%s transaction_id=%s", payment.Status, payment.TransactionID)

	return &paymentpb.PaymentResponse{
		Status:        payment.Status,
		TransactionId: payment.TransactionID,
	}, nil
}
