package main

import (
	"database/sql"
	"log"
	"net"
	"os"

	orderpb "github.com/Aruzhan38/order-payment-generated/proto/order"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"

	"order-service/internal/clients"
	"order-service/internal/repository"
	"order-service/internal/stream"
	grpcTransport "order-service/internal/transport/grpc"
	httpTransport "order-service/internal/transport/http"
	"order-service/internal/usecase"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	dsn := os.Getenv("ORDER_DB_DSN")
	if dsn == "" {
		log.Fatal("ORDER_DB_DSN is not set")
	}

	paymentGrpcAddr := os.Getenv("PAYMENT_GRPC_ADDR")
	if paymentGrpcAddr == "" {
		log.Fatal("PAYMENT_GRPC_ADDR is not set")
	}

	orderGrpcAddr := os.Getenv("ORDER_GRPC_ADDR")
	if orderGrpcAddr == "" {
		log.Fatal("ORDER_GRPC_ADDR is not set")
	}

	orderHTTPAddr := os.Getenv("ORDER_HTTP_ADDR")
	if orderHTTPAddr == "" {
		log.Fatal("ORDER_HTTP_ADDR is not set")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("failed to connect to database:", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal("database not reachable:", err)
	}

	log.Println("Connected to PostgreSQL")

	orderRepo := repository.NewOrderRepository(db)

	paymentClient, err := clients.NewPaymentClient(paymentGrpcAddr)
	if err != nil {
		log.Fatal("failed to connect to payment gRPC service:", err)
	}
	defer paymentClient.Close()

	streamManager := stream.NewOrderStreamManager()

	orderUC := usecase.NewOrderUsecase(orderRepo, paymentClient, streamManager)
	orderHandler := httpTransport.NewOrderHandler(orderUC)
	orderGrpcServer := grpcTransport.NewOrderServer(streamManager)

	go func() {
		lis, err := net.Listen("tcp", orderGrpcAddr)
		if err != nil {
			log.Fatal("failed to listen for gRPC:", err)
		}

		grpcServer := grpc.NewServer()
		orderpb.RegisterOrderServiceServer(grpcServer, orderGrpcServer)

		log.Println("Order gRPC streaming server running on", orderGrpcAddr)

		if err := grpcServer.Serve(lis); err != nil {
			log.Fatal("failed to serve gRPC:", err)
		}
	}()

	r := gin.Default()
	r.POST("/orders", orderHandler.CreateOrder)
	r.GET("/orders/:id", orderHandler.GetOrder)
	r.PATCH("/orders/:id/cancel", orderHandler.CancelOrder)

	log.Println("Order Service running on", orderHTTPAddr)

	if err := r.Run(orderHTTPAddr); err != nil {
		log.Fatal(err)
	}
}
