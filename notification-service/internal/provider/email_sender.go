package provider

import "context"

type EmailSender interface {
	Send(ctx context.Context, to string, subject string, body string) error
}
