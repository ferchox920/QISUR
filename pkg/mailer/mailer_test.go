package mailer

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
	"testing"
	"time"
)

// minimal SMTP test server to assert gomail sender sends mail.
func startTestSMTPServer(t *testing.T) (addr string, stop func(), received chan string) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start test smtp server: %v", err)
	}
	received = make(chan string, 1)
	stop = func() {
		_ = ln.Close()
	}

	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		reader := bufio.NewReader(conn)
		write := func(s string) {
			_, _ = conn.Write([]byte(s))
		}
		write("220 localhost ESMTP\r\n")
		var data strings.Builder
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				return
			}
			l := strings.TrimSpace(line)
			switch {
			case strings.HasPrefix(l, "EHLO") || strings.HasPrefix(l, "HELO"):
				write("250-localhost\r\n250 OK\r\n")
			case strings.HasPrefix(l, "MAIL FROM:"):
				write("250 OK\r\n")
			case strings.HasPrefix(l, "RCPT TO:"):
				write("250 OK\r\n")
			case l == "DATA":
				write("354 End data with <CR><LF>.<CR><LF>\r\n")
				for {
					dataLine, err := reader.ReadString('\n')
					if err != nil {
						return
					}
					if strings.TrimSpace(dataLine) == "." {
						break
					}
					data.WriteString(dataLine)
				}
				write("250 OK\r\n")
			case strings.HasPrefix(l, "QUIT"):
				write("221 Bye\r\n")
				received <- data.String()
				return
			default:
				write("250 OK\r\n")
			}
		}
	}()

	return ln.Addr().String(), stop, received
}

func TestGomailVerificationSender_SendsMail(t *testing.T) {
	addr, stop, received := startTestSMTPServer(t)
	defer stop()

	host, portStr, _ := net.SplitHostPort(addr)
	port, _ := strconv.Atoi(portStr)

	sender := NewGomailVerificationSender(host, port, "", "", "from@example.com", true)
	if sender == nil {
		t.Fatalf("expected sender to be created")
	}

	code := "999888"
	if err := sender.SendVerification(nil, "to@example.com", code); err != nil {
		t.Fatalf("unexpected send error: %v", err)
	}

	select {
	case body := <-received:
		if !strings.Contains(body, code) {
			t.Fatalf("expected body to contain code %s, got %s", code, body)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("timed out waiting for email body")
	}
}
