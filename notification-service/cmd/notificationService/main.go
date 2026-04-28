package main

import (
	"context"
	"log"
	"notification-service/internal/messaging"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://guest:guest@rabbitmq:5672/"
	}

	queueName := os.Getenv("PAYMENT_EVENTS_QUEUE")
	if queueName == "" {
		queueName = "payment.completed"
	}

	var consumer *messaging.RabbitMQConsumer

	for i := 0; i < 15; i++ {
		var err error

		consumer, err = messaging.NewRabbitMQConsumer(rabbitURL, queueName)
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
	defer consumer.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Println("Notification Service started")

	if err := consumer.Start(ctx); err != nil {
		log.Fatal("consumer stopped with error: ", err)
	}
}
