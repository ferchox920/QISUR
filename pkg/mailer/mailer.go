package mailer

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"

	mail "github.com/wneessen/go-mail"
)

// MailVerificationSender implementa identity.VerificationSender usando SMTP.
type MailVerificationSender struct {
	client *mail.Client
	from   string
}

// NewMailVerificationSender construye un sender de verificacion; devuelve nil si falta host.
func NewMailVerificationSender(host string, port int, username, password, from string, skipTLSVerify bool) *MailVerificationSender {
	if host == "" || from == "" {
		return nil
	}
	opts := []mail.Option{
		mail.WithPort(port),
		mail.WithTLSPolicy(mail.TLSOpportunistic),
	}
	if username != "" || password != "" {
		opts = append(opts,
			mail.WithUsername(username),
			mail.WithPassword(password),
			mail.WithSMTPAuth(mail.SMTPAuthPlain),
		)
	}
	if skipTLSVerify {
		opts = append(opts,
			mail.WithTLSPolicy(mail.NoTLS),
			mail.WithTLSConfig(&tls.Config{InsecureSkipVerify: true}),
		)
	}

	c, err := mail.NewClient(host, opts...)
	if err != nil {
		return nil
	}
	return &MailVerificationSender{client: c, from: from}
}

// SendVerification envia un correo de texto plano con el codigo de verificacion.
func (s *MailVerificationSender) SendVerification(ctx context.Context, email, code string) error {
	if s == nil || s.client == nil {
		return errors.New("mail sender no configurado")
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
