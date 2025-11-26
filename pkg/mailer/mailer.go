package mailer

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"

	mail "github.com/wneessen/go-mail"
)

// MailVerificationSender implements identity.VerificationSender using SMTP via go-mail.
type MailVerificationSender struct {
	client *mail.Client
	from   string
}

// NewMailVerificationSender builds a verification sender; returns nil if host is empty.
func NewMailVerificationSender(host string, port int, username, password, from string, skipTLSVerify bool) *MailVerificationSender {
	if host == "" || from == "" {
		return nil
	}

	opts := []mail.Option{
		mail.WithPort(port),
		mail.WithTLSPolicy(mail.TLSOpportunistic),
	}
	if username != "" {
		opts = append(opts, mail.WithUsername(username))
	}
	if password != "" {
		opts = append(opts, mail.WithPassword(password))
	}
	if skipTLSVerify {
		opts = append(opts, mail.WithTLSConfig(&tls.Config{InsecureSkipVerify: true}))
	}

	client, err := mail.NewClient(host, opts...)
	if err != nil {
		return nil
	}

	return &MailVerificationSender{client: client, from: from}
}

// SendVerification dispatches a simple text email with the verification code.
func (s *MailVerificationSender) SendVerification(ctx context.Context, email, code string) error {
	if s == nil || s.client == nil {
		return errors.New("mail sender not configured")
	}

	msg := mail.NewMsg()
	if err := msg.From(s.from); err != nil {
		return err
	}
	if err := msg.To(email); err != nil {
		return err
	}
	msg.Subject("Verifica tu cuenta")
	msg.SetBodyString(mail.TypeTextPlain, fmt.Sprintf("Tu codigo de verificacion es: %s", code))

	return s.client.DialAndSendWithContext(ctx, msg)
}
