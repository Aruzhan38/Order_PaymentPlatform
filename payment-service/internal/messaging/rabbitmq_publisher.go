package messaging

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
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

type RabbitMQPublisher struct {
	conn  *amqp.Connection
	ch    *amqp.Channel
	queue string
}

func NewRabbitMQPublisher(url, queue string) (*RabbitMQPublisher, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	dlqName := queue + ".dlq"

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
		queue,
		true,
		false,
		false,
		false,
		amqp.Table{
			"x-dead-letter-exchange":    "",
			"x-dead-letter-routing-key": dlqName,
		},
	)

	return &RabbitMQPublisher{
		conn:  conn,
		ch:    ch,
		queue: queue,
	}, nil
}

func (p *RabbitMQPublisher) PublishPaymentCompleted(ctx context.Context, orderID string, amount int64, customerEmail string, status string) error {
	event := PaymentCompletedEvent{
		EventID:       uuid.NewString(),
		OrderID:       orderID,
		Amount:        amount,
		CustomerEmail: customerEmail,
		Status:        status,
		CreatedAt:     time.Now().Format(time.RFC3339),
	}

	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return p.ch.PublishWithContext(
		ctx,
		"",
		p.queue,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
		},
	)
}

func (p *RabbitMQPublisher) Close() {
	if p.ch != nil {
		p.ch.Close()
	}
	if p.conn != nil {
		p.conn.Close()
	}
}
