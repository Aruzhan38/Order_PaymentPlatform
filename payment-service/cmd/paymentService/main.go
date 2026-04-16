package main

import (
	"database/sql"
	paymentpb "github.com/Aruzhan38/order-payment-generated/proto/payment"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"payment-service/internal/repository"
	grpcTransport "payment-service/internal/transport/grpc"
	"payment-service/internal/usecase"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	dsn := os.Getenv("PAYMENT_DB_DSN")
	if dsn == "" {
		log.Fatal("PAYMENT_DB_DSN is not set")
	}

	grpcAddr := os.Getenv("PAYMENT_GRPC_ADDR")
	if grpcAddr == "" {
		log.Fatal("PAYMENT_GRPC_ADDR is not set")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("failed to connect to database: ", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal("database not reachable: ", err)
	}

	log.Println("Connected to PostgreSQL")

	paymentRepo := repository.NewPaymentRepository(db)
	paymentUC := usecase.NewPaymentUsecase(paymentRepo)

	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatal("failed to listen: ", err)
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(grpcTransport.LoggingInterceptor),
	)
	paymentpb.RegisterPaymentServiceServer(
		grpcServer,
		grpcTransport.NewPaymentServer(paymentUC),
	)

	log.Println("Payment gRPC server running on", grpcAddr)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal("failed to serve gRPC: ", err)
	}
}
