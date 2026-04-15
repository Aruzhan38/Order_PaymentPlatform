package clients

import (
	"context"
	"time"

	paymentpb "github.com/Aruzhan38/order-payment-generated/proto/payment"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type PaymentClient struct {
	conn   *grpc.ClientConn
	client paymentpb.PaymentServiceClient
}

func NewPaymentClient(addr string) (*PaymentClient, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	c := paymentpb.NewPaymentServiceClient(conn)

	return &PaymentClient{
		conn:   conn,
		client: c,
	}, nil
}

func (p *PaymentClient) Close() error {
	if p.conn != nil {
		return p.conn.Close()
	}
	return nil
}

func (p *PaymentClient) ProcessPayment(orderID string, amount int64) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := p.client.ProcessPayment(ctx, &paymentpb.PaymentRequest{
		OrderId: orderID,
		Amount:  amount,
	})
	if err != nil {
		return "", err
	}

	return resp.Status, nil
}
