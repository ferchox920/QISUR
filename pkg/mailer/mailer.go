package mailer

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"

	"gopkg.in/gomail.v2"
)

// MailVerificationSender implementa identity.VerificationSender usando SMTP con gomail.
type MailVerificationSender struct {
	dialer *gomail.Dialer
	from   string
}

// NewMailVerificationSender construye un sender de verificacion; devuelve nil si falta host.
func NewMailVerificationSender(host string, port int, username, password, from string, skipTLSVerify bool) *MailVerificationSender {
	if host == "" || from == "" {
		return nil
	}
	d := gomail.NewDialer(host, port, username, password)
	if skipTLSVerify {
		d.TLSConfig = &tls.Config{InsecureSkipVerify: true} // solo dev/test
	}
	return &MailVerificationSender{dialer: d, from: from}
}

// SendVerification envia un correo de texto plano con el codigo de verificacion.
func (s *MailVerificationSender) SendVerification(ctx context.Context, email, code string) error {
	if s == nil || s.dialer == nil {
		return errors.New("mail sender no configurado")
	}
	msg := gomail.NewMessage()
	msg.SetHeader("From", s.from)
	msg.SetHeader("To", email)
	msg.SetHeader("Subject", "Verifica tu cuenta")
	msg.SetBody("text/plain", fmt.Sprintf("Tu codigo de verificacion es: %s", code))
	return s.dialer.DialAndSend(msg)
}
