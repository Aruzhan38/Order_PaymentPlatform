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
	"payment-service/internal/messaging"
	"payment-service/internal/repository"
	grpcTransport "payment-service/internal/transport/grpc"
	"payment-service/internal/usecase"
	"time"
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

	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		log.Fatal("RABBITMQ_URL is not set")
	}

	queueName := os.Getenv("PAYMENT_EVENTS_QUEUE")
	if queueName == "" {
		log.Fatal("PAYMENT_EVENTS_QUEUE is not set")
	}

	var db *sql.DB

	for i := 0; i < 10; i++ {
		var err error

		db, err = sql.Open("postgres", dsn)
		if err == nil {
			err = db.Ping()
		}

		if err == nil {
			log.Println("Connected to PostgreSQL")
			break
		}

		log.Println("Waiting for PostgreSQL...")
		time.Sleep(2 * time.Second)

		if i == 9 {
			log.Fatal("database not reachable: ", err)
		}
	}

	var publisher *messaging.RabbitMQPublisher

	for i := 0; i < 15; i++ {
		var err error

		publisher, err = messaging.NewRabbitMQPublisher(rabbitURL, queueName)
		if err == nil {
			log.Println("Connected to RabbitMQ")
			break
		}

		log.Println("Waiting for RabbitMQ...")
		time.Sleep(2 * time.Second)

		if i == 14 {
			log.Fatal("failed to connect to RabbitMQ: ", err)
		}
	}
	defer publisher.Close()

	paymentRepo := repository.NewPaymentRepository(db)
	paymentUC := usecase.NewPaymentUsecase(paymentRepo, publisher)

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
