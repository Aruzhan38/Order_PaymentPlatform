package clients

import (
	"context"
	"time"

	"google.golang.org/grpc"

	paymentpb "github.com/Aruzhan38/order-payment-generated/proto/payment"
)

type PaymentClient struct {
	client paymentpb.PaymentServiceClient
}

func NewPaymentClient(addr string) (*PaymentClient, error) {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	c := paymentpb.NewPaymentServiceClient(conn)

	return &PaymentClient{client: c}, nil
}

func (p *PaymentClient) ProcessPayment(orderID string, amount int64) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := p.client.ProcessPayment(ctx, &paymentpb.PaymentRequest{
		OrderId: orderID,
		Amount:  int64(amount),
	})

	if err != nil {
		return "", err
	}

	return resp.Status, nil
}
