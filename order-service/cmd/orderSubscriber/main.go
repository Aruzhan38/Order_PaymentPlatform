package main

import (
	"context"
	"io"
	"log"

	orderpb "github.com/Aruzhan38/order-payment-generated/proto/order"
	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial("127.0.0.1:50052", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := orderpb.NewOrderServiceClient(conn)

	stream, err := client.SubscribeToOrderUpdates(context.Background(), &orderpb.OrderRequest{
		OrderId: "demo-order-2",
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Subscribed... waiting for updates")

	for {
		update, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("UPDATE → id=%s status=%s message=%s\n",
			update.OrderId,
			update.Status,
			update.Message,
		)
	}
}
