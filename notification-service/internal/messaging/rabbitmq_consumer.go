package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"notification-service/internal/provider"

	amqp "github.com/rabbitmq/amqp091-go"
)

type PaymentCompletedEvent struct {
	EventID       string `json:"event_id"`
	OrderID       string `json:"order_id"`
	Amount        int64  `json:"amount"`
	CustomerEmail string `json:"customer_email"`
	Status        string `json:"status"`
	CreatedAt     string `json:"created_at"`
}

type IdempotencyStore interface {
	IsProcessed(ctx context.Context, key string) (bool, error)
	MarkProcessed(ctx context.Context, key string, ttl time.Duration) error
}

type RabbitMQConsumer struct {
	conn      *amqp.Connection
	ch        *amqp.Channel
	queueName string

	emailSender      provider.EmailSender
	idempotencyStore IdempotencyStore
	maxRetries       int
	processedTTL     time.Duration
}

func NewRabbitMQConsumer(
	url string,
	queueName string,
	emailSender provider.EmailSender,
	idempotencyStore IdempotencyStore,
	maxRetries int,
	processedTTL time.Duration,
) (*RabbitMQConsumer, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	dlqName := queueName + ".dlq"

	_, err = ch.QueueDeclare(
		dlqName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, err
	}

	_, err = ch.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		amqp.Table{
			"x-dead-letter-exchange":    "",
			"x-dead-letter-routing-key": dlqName,
		},
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, err
	}

	return &RabbitMQConsumer{
		conn:             conn,
		ch:               ch,
		queueName:        queueName,
		emailSender:      emailSender,
		idempotencyStore: idempotencyStore,
		maxRetries:       maxRetries,
		processedTTL:     processedTTL,
	}, nil
}

func (c *RabbitMQConsumer) Start(ctx context.Context) error {
	msgs, err := c.ch.Consume(
		c.queueName,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	log.Println("Notification Service is waiting for messages...")

	for {
		select {
		case <-ctx.Done():
			log.Println("Notification Service stopped")
			return nil

		case msg, ok := <-msgs:
			if !ok {
				return nil
			}

			go func(delivery amqp.Delivery) {
				if err := c.handleMessage(ctx, delivery); err != nil {
					log.Printf("failed to process message: %v", err)

					if nackErr := delivery.Nack(false, false); nackErr != nil {
						log.Printf("failed to nack message: %v", nackErr)
					}
					return
				}

				if err := delivery.Ack(false); err != nil {
					log.Printf("failed to ack message: %v", err)
				}
			}(msg)
		}
	}
}

func (c *RabbitMQConsumer) handleMessage(ctx context.Context, msg amqp.Delivery) error {
	var event PaymentCompletedEvent

	if err := json.Unmarshal(msg.Body, &event); err != nil {
		return err
	}

	idempotencyKey := fmt.Sprintf("notification:processed:%s", event.EventID)

	processed, err := c.idempotencyStore.IsProcessed(ctx, idempotencyKey)
	if err != nil {
		return err
	}

	if processed {
		log.Printf("[Notification] Duplicate event ignored: event_id=%s", event.EventID)
		return nil
	}

	if event.Amount == 10000 {
		return fmt.Errorf("simulated permanent error for DLQ")
	}

	subject := "Payment completed"
	body := fmt.Sprintf(
		"Your payment for order %s was completed. Amount: %.2f Status: %s",
		event.OrderID,
		float64(event.Amount)/100,
		event.Status,
	)

	if err := c.sendWithRetry(ctx, event.OrderID, event.EventID, event.CustomerEmail, subject, body); err != nil {
		return err
	}

	if err := c.idempotencyStore.MarkProcessed(ctx, idempotencyKey, c.processedTTL); err != nil {
		return err
	}

	log.Printf(
		"[Notification] Sent email to %s for Order #%s. Amount: %.2f Status: %s",
		event.CustomerEmail,
		event.OrderID,
		float64(event.Amount)/100,
		event.Status,
	)

	return nil
}

func (c *RabbitMQConsumer) sendWithRetry(ctx context.Context, orderID string, eventID string, to string, subject string, body string) error {
	var lastErr error

	for attempt := 1; attempt <= c.maxRetries; attempt++ {
		err := c.emailSender.Send(ctx, to, subject, body)
		if err == nil {
			return nil
		}

		lastErr = err

		backoff := time.Duration(1<<attempt) * time.Second
		log.Printf(
			"email send failed: order_id=%s event_id=%s attempt=%d backoff=%s error=%v",
			orderID,
			eventID,
			attempt,
			backoff,
			err,
		)

		select {
		case <-time.After(backoff):
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return fmt.Errorf("email sending failed after %d retries: %w", c.maxRetries, lastErr)
}

func (c *RabbitMQConsumer) Close() {
	if c.ch != nil {
		c.ch.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
}
