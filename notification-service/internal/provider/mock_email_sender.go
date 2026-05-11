package provider

import (
	"context"
	"errors"
	"log"
	"math/rand"
	"time"
)

type MockEmailSender struct{}

func NewMockEmailSender() *MockEmailSender {
	return &MockEmailSender{}
}

func (m *MockEmailSender) Send(ctx context.Context, to string, subject string, body string) error {
	time.Sleep(1 * time.Second)

	if rand.Intn(4) == 0 {
		return errors.New("simulated email provider failure")
	}

	log.Printf("[EmailProvider] Sent email to=%s subject=%s body=%s", to, subject, body)
	return nil
}
