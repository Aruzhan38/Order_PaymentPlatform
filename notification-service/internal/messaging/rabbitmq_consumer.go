package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"

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

type RabbitMQConsumer struct {
	conn      *amqp.Connection
	ch        *amqp.Channel
	queueName string

	mu        sync.Mutex
	processed map[string]bool
}

func NewRabbitMQConsumer(url, queueName string) (*RabbitMQConsumer, error) {
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
		conn:      conn,
		ch:        ch,
		queueName: queueName,
		processed: make(map[string]bool),
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

			if err := c.handleMessage(msg); err != nil {
				log.Printf("failed to process message: %v", err)

				if nackErr := msg.Nack(false, false); nackErr != nil {
					log.Printf("failed to nack message: %v", nackErr)
				}
				continue
			}

			if err := msg.Ack(false); err != nil {
				log.Printf("failed to ack message: %v", err)
			}
		}
	}
}

func (c *RabbitMQConsumer) handleMessage(msg amqp.Delivery) error {
	var event PaymentCompletedEvent

	if err := json.Unmarshal(msg.Body, &event); err != nil {
		return err
	}

	if event.Amount == 10000 {
		return fmt.Errorf("simulated permanent error for DLQ")
	}

	c.mu.Lock()
	if c.processed[event.EventID] {
		c.mu.Unlock()
		log.Printf("[Notification] Duplicate event ignored: event_id=%s", event.EventID)
		return nil
	}
	c.processed[event.EventID] = true
	c.mu.Unlock()

	log.Printf(
		"[Notification] Sent email to %s for Order #%s. Amount: %.2f Status: %s",
		event.CustomerEmail,
		event.OrderID,
		float64(event.Amount)/100,
		event.Status,
	)

	return nil
}

func (c *RabbitMQConsumer) Close() {
	if c.ch != nil {
		c.ch.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
}
