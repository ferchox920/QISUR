package mailer

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"

	"gopkg.in/gomail.v2"
)

// GomailVerificationSender implements identity.VerificationSender using SMTP via gomail.
type GomailVerificationSender struct {
	dialer *gomail.Dialer
	from   string
}

// NewGomailVerificationSender builds a verification sender; returns nil if host is empty.
func NewGomailVerificationSender(host string, port int, username, password, from string, skipTLSVerify bool) *GomailVerificationSender {
	if host == "" || from == "" {
		return nil
	}
	d := gomail.NewDialer(host, port, username, password)
	if skipTLSVerify {
		d.TLSConfig = &tls.Config{InsecureSkipVerify: true} // only for dev/test
	}
	return &GomailVerificationSender{dialer: d, from: from}
}

// SendVerification dispatches a simple text email with the verification code.
func (s *GomailVerificationSender) SendVerification(ctx context.Context, email, code string) error {
	if s == nil || s.dialer == nil {
		return errors.New("gomail sender not configured")
	}
	msg := gomail.NewMessage()
	msg.SetHeader("From", s.from)
	msg.SetHeader("To", email)
	msg.SetHeader("Subject", "Verifica tu cuenta")
	msg.SetBody("text/plain", fmt.Sprintf("Tu código de verificación es: %s", code))
	return s.dialer.DialAndSend(msg)
}
