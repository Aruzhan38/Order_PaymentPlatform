package main

import (
	"database/sql"
	"log"
	"net"
	"os"
	"payment-service/internal/repository"
	"payment-service/internal/usecase"

	_ "github.com/lib/pq"
	"google.golang.org/grpc"

	paymentpb "github.com/Aruzhan38/order-payment-generated/proto/payment"

	grpcTransport "payment-service/internal/transport/grpc"
)

func main() {
	dsn := "host=localhost port=5432 user=postgres password=0000 dbname=payment_db sslmode=disable"

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

	grpcAddr := os.Getenv("PAYMENT_GRPC_ADDR")
	if grpcAddr == "" {
		grpcAddr = ":50051"
	}

	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatal("failed to listen: ", err)
	}

	grpcServer := grpc.NewServer()
	paymentpb.RegisterPaymentServiceServer(
		grpcServer,
		grpcTransport.NewPaymentServer(paymentUC),
	)

	log.Println("Payment gRPC server running on", grpcAddr)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal("failed to serve gRPC: ", err)
	}
}
