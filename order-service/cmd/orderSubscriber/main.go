package main

import (
	"context"
	orderpb "github.com/Aruzhan38/order-payment-generated/proto/order"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"io"
	"log"
	"os"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	orderGrpcAddr := os.Getenv("ORDER_GRPC_ADDR")
	if orderGrpcAddr == "" {
		log.Fatal("ORDER_GRPC_ADDR is not set")
	}

	if len(os.Args) < 2 {
		log.Fatal("usage: go run cmd/orderSubscriber/main.go <order id>")
	}
	orderID := os.Args[1]

	conn, err := grpc.Dial(orderGrpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to order gRPC server: %v", err)
	}
	defer conn.Close()

	client := orderpb.NewOrderServiceClient(conn)

	ctx := context.Background()

	stream, err := client.SubscribeToOrderUpdates(ctx, &orderpb.OrderRequest{
		OrderId: orderID,
	})
	if err != nil {
		log.Fatalf("failed to subscribe to order updates: %v", err)
	}

	log.Printf("Subscribed to updates for order: %s", orderID)

	for {
		update, err := stream.Recv()
		if err == io.EOF {
			log.Println("stream closed by server")
			return
		}
		if err != nil {
			log.Fatalf("stream receive error: %v", err)
		}

		log.Printf("Order update received: order_id=%s status=%s message=%s",
			update.OrderId,
			update.Status,
			update.Message,
		)
	}
}
