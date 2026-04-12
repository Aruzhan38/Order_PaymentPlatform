package main

import (
	"database/sql"
	"log"
	"net"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	orderpb "github.com/Aruzhan38/order-payment-generated/proto/order"
	"google.golang.org/grpc"

	"order-service/internal/clients"
	"order-service/internal/repository"
	"order-service/internal/stream"
	grpcTransport "order-service/internal/transport/grpc"
	httpTransport "order-service/internal/transport/http"
	"order-service/internal/usecase"
)

func main() {
	dsn := "host=localhost port=5432 user=postgres password=0000 dbname=order_db sslmode=disable"

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("failed to connect to database:", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal("database not reachable:", err)
	}

	log.Println("Connected to PostgreSQL")

	err = godotenv.Load()
	if err != nil {
		log.Println("No .env file found")
	}

	orderRepo := repository.NewOrderRepository(db)

	grpcAddr := os.Getenv("PAYMENT_GRPC_ADDR")
	if grpcAddr == "" {
		log.Fatal("PAYMENT_GRPC_ADDR not set")
	}

	paymentClient, err := clients.NewPaymentClient(grpcAddr)
	if err != nil {
		log.Fatal("failed to connect to payment gRPC service:", err)
	}

	streamManager := stream.NewOrderStreamManager()

	orderUC := usecase.NewOrderUsecase(orderRepo, paymentClient, streamManager)
	orderHandler := httpTransport.NewOrderHandler(orderUC)

	orderGrpcServer := grpcTransport.NewOrderServer(streamManager)

	go func() {
		lis, err := net.Listen("tcp", ":50052")
		if err != nil {
			log.Fatal("failed to listen for gRPC:", err)
		}

		grpcServer := grpc.NewServer()
		orderpb.RegisterOrderServiceServer(grpcServer, orderGrpcServer)

		log.Println("Order gRPC streaming server running on :50052")

		if err := grpcServer.Serve(lis); err != nil {
			log.Fatal("failed to serve gRPC:", err)
		}
	}()

	r := gin.Default()

	r.POST("/orders", orderHandler.CreateOrder)
	r.GET("/orders/:id", orderHandler.GetOrder)
	r.PATCH("/orders/:id/cancel", orderHandler.CancelOrder)

	log.Println("Order Service running on :8080")

	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
