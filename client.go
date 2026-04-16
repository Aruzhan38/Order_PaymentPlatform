package main

import (
	"context"
	"fmt"
	"log"
	"time"

	paymentpb "github.com/Aruzhan38/order-payment-generated/proto/payment"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := paymentpb.NewPaymentServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.GetPaymentStats(ctx, &paymentpb.GetPaymentStatsRequest{})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Total:", resp.TotalCount)
	fmt.Println("Authorized:", resp.AuthorizedCount)
	fmt.Println("Declined:", resp.DeclinedCount)
	fmt.Println("Amount:", resp.TotalAmount)
}
