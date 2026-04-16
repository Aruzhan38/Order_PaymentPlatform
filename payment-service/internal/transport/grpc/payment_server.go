package grpc

import (
	"context"
	paymentpb "github.com/Aruzhan38/order-payment-generated/proto/payment"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
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
		return nil, status.Error(codes.Internal, "failed to process payment")
	}

	return &paymentpb.PaymentResponse{
		Status:        payment.Status,
		TransactionId: payment.TransactionID,
	}, nil
}

func (s *PaymentServer) GetPaymentStats(
	ctx context.Context,
	req *paymentpb.GetPaymentStatsRequest,
) (*paymentpb.PaymentStats, error) {

	total, authorized, declined, amount, err := s.uc.GetStats(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get stats")
	}

	return &paymentpb.PaymentStats{
		TotalCount:      total,
		AuthorizedCount: authorized,
		DeclinedCount:   declined,
		TotalAmount:     amount,
	}, nil
}
