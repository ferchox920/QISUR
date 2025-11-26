package mailer

import (
	"context"
	"log/slog"
)

// NoopVerificationSender registra el intento de envio pero no envia correo.
type NoopVerificationSender struct {
	Logr *slog.Logger
}

func (s *NoopVerificationSender) SendVerification(ctx context.Context, email, code string) error {
	if s.Logr != nil {
		s.Logr.Info("verification email noop sender", "email", email)
	}
	return nil
}
