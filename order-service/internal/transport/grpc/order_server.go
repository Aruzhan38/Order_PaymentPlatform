package grpc

import (
	orderpb "github.com/Aruzhan38/order-payment-generated/proto/order"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"order-service/internal/stream"
)

type OrderServer struct {
	orderpb.UnimplementedOrderServiceServer
	manager *stream.OrderStreamManager
}

func NewOrderServer(manager *stream.OrderStreamManager) *OrderServer {
	return &OrderServer{manager: manager}
}

func (s *OrderServer) SubscribeToOrderUpdates(
	req *orderpb.OrderRequest,
	streamSrv orderpb.OrderService_SubscribeToOrderUpdatesServer,
) error {
	if req.OrderId == "" {
		return status.Error(codes.InvalidArgument, "order_id is required")
	}

	sub := s.manager.Subscribe(req.OrderId)
	defer s.manager.Unsubscribe(req.OrderId, sub)

	for {
		select {
		case <-streamSrv.Context().Done():
			return nil
		case update, ok := <-sub:
			if !ok {
				return nil
			}

			err := streamSrv.Send(&orderpb.OrderStatusUpdate{
				OrderId: update.OrderID,
				Status:  update.Status,
				Message: update.Message,
			})
			if err != nil {
				return status.Error(codes.Internal, "failed to send order update")
			}
		}
	}
}
