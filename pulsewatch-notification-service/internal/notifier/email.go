package notifier

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"net/smtp"
	"os"
	"strconv"
	"strings"
	"time"
)

type EmailSender struct {
	host     string
	port     int
	user     string
	password string
	from     string
}

func NewEmailSender() *EmailSender {
	port, err := strconv.Atoi(getEnv("SMTP_PORT", "1025"))
	if err != nil {
		port = 1025
	}
	return &EmailSender{
		host:     getEnv("SMTP_HOST", "localhost"),
		port:     port,
		user:     getEnv("SMTP_USER", ""),
		password: getEnv("SMTP_PASSWORD", ""),
		from:     getEnv("SMTP_FROM", "PulseWatch <alerts@pulsewatch.dev>"),
	}
}

func (s *EmailSender) Send(to, subject, textBody, htmlBody string) error {
	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	var auth smtp.Auth
	if s.user != "" {
		auth = smtp.PlainAuth("", s.user, s.password, s.host)
	}

	boundary := fmt.Sprintf("boundary_%d", time.Now().UnixNano())

	var msg bytes.Buffer
	msg.WriteString(fmt.Sprintf("From: %s\r\n", s.from))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", to))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=\"%s\"\r\n\r\n", boundary))

	// Plain text part
	msg.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	msg.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n")
	msg.WriteString("Content-Transfer-Encoding: 7bit\r\n\r\n")
	msg.WriteString(textBody)
	msg.WriteString("\r\n\r\n")

	// HTML part
	msg.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	msg.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n")
	msg.WriteString("Content-Transfer-Encoding: 7bit\r\n\r\n")
	msg.WriteString(htmlBody)
	msg.WriteString("\r\n\r\n")

	msg.WriteString(fmt.Sprintf("--%s--\r\n", boundary))

	fromAddr := s.from
	if start := strings.Index(fromAddr, "<"); start != -1 {
		if end := strings.Index(fromAddr, ">"); end != -1 && end > start {
			fromAddr = fromAddr[start+1 : end]
		}
	}

	// Port 465 uses direct/implicit TLS (e.g. Resend, Gmail SSL)
	if s.port == 465 {
		tlsConfig := &tls.Config{
			ServerName: s.host,
		}
		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return fmt.Errorf("failed to dial SMTP over TLS: %w", err)
		}
		defer conn.Close()

		client, err := smtp.NewClient(conn, s.host)
		if err != nil {
			return fmt.Errorf("failed to create SMTP client: %w", err)
		}
		defer client.Close()

		if auth != nil {
			if err = client.Auth(auth); err != nil {
				return fmt.Errorf("failed to authenticate SMTP: %w", err)
			}
		}

		if err = client.Mail(fromAddr); err != nil {
			return fmt.Errorf("failed to set sender: %w", err)
		}
		if err = client.Rcpt(to); err != nil {
			return fmt.Errorf("failed to set recipient: %w", err)
		}

		w, err := client.Data()
		if err != nil {
			return fmt.Errorf("failed to open data writer: %w", err)
		}
		if _, err = w.Write(msg.Bytes()); err != nil {
			return fmt.Errorf("failed to write email body: %w", err)
		}
		return w.Close()
	}

	return smtp.SendMail(addr, auth, fromAddr, []string{to}, msg.Bytes())
}

func (s *EmailSender) SendIncidentOpened(to, monitorName, monitorURL, cause string, startedAt time.Time) error {
	subject := fmt.Sprintf("🚨 Outage Alert: %s is DOWN", monitorName)
	timeFormatted := startedAt.UTC().Format(time.RFC1123)

	textBody := fmt.Sprintf("Service Outage Detected!\n\nMonitor: %s\nTarget URL: %s\nDetected at: %s\nCause: %s\n\nPulseWatch is monitoring for recovery.",
		monitorName, monitorURL, timeFormatted, cause)

	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html>
<body style="font-family: -apple-system, sans-serif; background: #0b0f19; color: #f3f4f6; padding: 30px;">
  <div style="max-width: 520px; margin: 0 auto; background: #111827; border: 1px solid #dc2626; border-radius: 10px; padding: 24px;">
    <div style="font-size: 20px; font-weight: bold; color: #ef4444; margin-bottom: 12px;">🚨 Service Outage Detected</div>
    <p style="font-size: 15px; color: #d1d5db;">Monitor <strong>%s</strong> is currently DOWN.</p>
    <div style="background: #1f2937; padding: 12px; border-radius: 6px; font-family: monospace; font-size: 13px; margin: 16px 0;">
      <div><strong>Target:</strong> %s</div>
      <div><strong>Started:</strong> %s</div>
      <div><strong>Reason:</strong> %s</div>
    </div>
    <p style="font-size: 13px; color: #9ca3af;">PulseWatch probe workers will automatically alert you when this service recovers.</p>
  </div>
</body>
</html>`, monitorName, monitorURL, timeFormatted, cause)

	return s.Send(to, subject, textBody, htmlBody)
}

func (s *EmailSender) SendIncidentResolved(to, monitorName, monitorURL string, durationSeconds int, resolvedAt time.Time) error {
	subject := fmt.Sprintf("✅ Incident Resolved: %s is UP", monitorName)
	timeFormatted := resolvedAt.UTC().Format(time.RFC1123)
	mins := durationSeconds / 60
	secs := durationSeconds % 60
	durationStr := fmt.Sprintf("%dm %ds", mins, secs)

	textBody := fmt.Sprintf("Incident Resolved!\n\nMonitor: %s\nTarget URL: %s\nResolved at: %s\nDowntime Duration: %s\n\nService has met recovery thresholds.",
		monitorName, monitorURL, timeFormatted, durationStr)

	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html>
<body style="font-family: -apple-system, sans-serif; background: #0b0f19; color: #f3f4f6; padding: 30px;">
  <div style="max-width: 520px; margin: 0 auto; background: #111827; border: 1px solid #16a34a; border-radius: 10px; padding: 24px;">
    <div style="font-size: 20px; font-weight: bold; color: #22c55e; margin-bottom: 12px;">✅ Incident Resolved</div>
    <p style="font-size: 15px; color: #d1d5db;">Monitor <strong>%s</strong> has recovered and is now UP.</p>
    <div style="background: #1f2937; padding: 12px; border-radius: 6px; font-family: monospace; font-size: 13px; margin: 16px 0;">
      <div><strong>Target:</strong> %s</div>
      <div><strong>Resolved:</strong> %s</div>
      <div><strong>Total Downtime:</strong> %s</div>
    </div>
    <p style="font-size: 13px; color: #9ca3af;">Consecutive successful checks have met your configured recovery threshold.</p>
  </div>
</body>
</html>`, monitorName, monitorURL, timeFormatted, durationStr)

	return s.Send(to, subject, textBody, htmlBody)
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
