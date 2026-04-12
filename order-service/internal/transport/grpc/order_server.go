package grpc

import (
	orderpb "github.com/Aruzhan38/order-payment-generated/proto/order"
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
	sub := s.manager.Subscribe(req.OrderId)
	defer s.manager.Unsubscribe(req.OrderId, sub)

	for {
		select {
		case <-streamSrv.Context().Done():
			return streamSrv.Context().Err()
		case update := <-sub:
			err := streamSrv.Send(&orderpb.OrderStatusUpdate{
				OrderId: update.OrderID,
				Status:  update.Status,
				Message: update.Message,
			})
			if err != nil {
				return err
			}
		}
	}
}
