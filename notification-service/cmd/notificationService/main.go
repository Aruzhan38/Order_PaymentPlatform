package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"notification-service/internal/messaging"
	"notification-service/internal/provider"
	"notification-service/internal/store"
)

func main() {
	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://guest:guest@localhost:5672/"
	}

	queueName := os.Getenv("PAYMENT_EVENTS_QUEUE")
	if queueName == "" {
		queueName = "payment.completed"
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	providerMode := os.Getenv("PROVIDER_MODE")
	if providerMode == "" {
		providerMode = "SIMULATED"
	}

	maxRetries := 3
	if value := os.Getenv("MAX_RETRIES"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil {
			log.Fatal("invalid MAX_RETRIES:", err)
		}
		maxRetries = parsed
	}

	processedTTLSeconds := 3600
	if value := os.Getenv("PROCESSED_TTL_SECONDS"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil {
			log.Fatal("invalid PROCESSED_TTL_SECONDS:", err)
		}
		processedTTLSeconds = parsed
	}

	idempotencyStore := store.NewRedisIdempotencyStore(redisAddr)
	defer idempotencyStore.Close()

	for i := 0; i < 10; i++ {
		err := idempotencyStore.Ping(context.Background())
		if err == nil {
			log.Println("Connected to Redis")
			break
		}

		log.Println("Waiting for Redis...")
		time.Sleep(2 * time.Second)

		if i == 9 {
			log.Fatal("failed to connect to Redis:", err)
		}
	}

	var emailSender provider.EmailSender

	switch providerMode {
	case "SIMULATED":
		emailSender = provider.NewMockEmailSender()
	default:
		log.Fatal("unsupported PROVIDER_MODE:", providerMode)
	}

	var consumer *messaging.RabbitMQConsumer

	for i := 0; i < 15; i++ {
		var err error

		consumer, err = messaging.NewRabbitMQConsumer(
			rabbitURL,
			queueName,
			emailSender,
			idempotencyStore,
			maxRetries,
			time.Duration(processedTTLSeconds)*time.Second,
		)

		if err == nil {
			log.Println("Connected to RabbitMQ")
			break
		}

		log.Println("Waiting for RabbitMQ...")
		time.Sleep(2 * time.Second)

		if i == 14 {
			log.Fatal("failed to connect to RabbitMQ:", err)
		}
	}

	defer consumer.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Println("Notification Service started")
	log.Println("Provider mode:", providerMode)

	if err := consumer.Start(ctx); err != nil {
		log.Fatal("consumer stopped with error:", err)
	}
}
