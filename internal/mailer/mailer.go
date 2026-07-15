package mailer

import (
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"mime"
	"net/smtp"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Mailer sends transactional emails (OTP codes, etc.).
type Mailer interface {
	// Send delivers a plain-text email.
	Send(to, subject, body string) error
	// SendHTML delivers a multipart/alternative email carrying both an HTML
	// body and a plain-text fallback. Clients that cannot render HTML (or the
	// user's preference) fall back to the text part.
	SendHTML(to, subject, htmlBody, textBody string) error
}

// Config holds the SMTP connection details, read from the "email" block of config.json.
type Config struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	FromName string
}

type smtpMailer struct {
	cfg Config
	log *logrus.Logger
}

// NewMailer builds a Mailer from viper config. The relevant config block:
//
//	"email": {
//	    "host": "localhost", "port": 25,
//	    "username": "admin@arsiva.id", "password": "...",
//	    "from": "admin@arsiva.id", "from_name": "Arsiva"
//	}
//
// If "from" is empty it falls back to "username".
func NewMailer(config *viper.Viper, log *logrus.Logger) Mailer {
	from := config.GetString("email.from")
	if from == "" {
		from = config.GetString("email.username")
	}
	fromName := config.GetString("email.from_name")
	if fromName == "" {
		fromName = "Arsiva"
	}
	return &smtpMailer{
		cfg: Config{
			Host:     config.GetString("email.host"),
			Port:     config.GetInt("email.port"),
			Username: config.GetString("email.username"),
			Password: config.GetString("email.password"),
			From:     from,
			FromName: fromName,
		},
		log: log,
	}
}

// Send delivers a plain-text email. It is intentionally tolerant of the local
// relay setup (host=localhost, port 25): STARTTLS is used only if the server
// advertises it, and AUTH only if the server advertises it AND a username is
// configured. This lets the same code work against an authenticated public
// relay and an unauthenticated localhost relay without changes.
func (m *smtpMailer) Send(to, subject, body string) error {
	return m.deliver(to, m.buildMessage(to, subject, body))
}

// SendHTML delivers a multipart/alternative email (HTML + plain-text fallback).
func (m *smtpMailer) SendHTML(to, subject, htmlBody, textBody string) error {
	return m.deliver(to, m.buildMultipartMessage(to, subject, htmlBody, textBody))
}

// deliver opens the SMTP conversation and writes an already-assembled RFC 5322
// message. Transport concerns (STARTTLS/AUTH negotiation) live here so Send and
// SendHTML share one code path.
func (m *smtpMailer) deliver(to string, msg []byte) error {
	if m.cfg.Host == "" {
		return fmt.Errorf("mailer: email.host is not configured")
	}

	addr := fmt.Sprintf("%s:%d", m.cfg.Host, m.cfg.Port)

	c, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("mailer: dial %s: %w", addr, err)
	}
	defer c.Close()

	if err := c.Hello("localhost"); err != nil {
		return fmt.Errorf("mailer: HELO: %w", err)
	}

	// Upgrade to TLS only if the server offers it.
	if ok, _ := c.Extension("STARTTLS"); ok {
		// ServerName must match the cert; InsecureSkipVerify tolerates the
		// self-signed cert a localhost relay typically presents.
		tlsCfg := &tls.Config{ServerName: m.cfg.Host, InsecureSkipVerify: true}
		if err := c.StartTLS(tlsCfg); err != nil {
			return fmt.Errorf("mailer: STARTTLS: %w", err)
		}
	}

	// Authenticate only if the server supports AUTH and we have credentials.
	if m.cfg.Username != "" {
		if ok, _ := c.Extension("AUTH"); ok {
			auth := smtp.PlainAuth("", m.cfg.Username, m.cfg.Password, m.cfg.Host)
			if err := c.Auth(auth); err != nil {
				return fmt.Errorf("mailer: AUTH: %w", err)
			}
		}
	}

	if err := c.Mail(m.cfg.From); err != nil {
		return fmt.Errorf("mailer: MAIL FROM: %w", err)
	}
	if err := c.Rcpt(to); err != nil {
		return fmt.Errorf("mailer: RCPT TO: %w", err)
	}

	w, err := c.Data()
	if err != nil {
		return fmt.Errorf("mailer: DATA: %w", err)
	}
	if _, err := w.Write(msg); err != nil {
		return fmt.Errorf("mailer: write body: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("mailer: close body: %w", err)
	}

	return c.Quit()
}

// buildMessage assembles RFC 5322 headers + a plain-text body. CRLF line
// endings are required by SMTP.
func (m *smtpMailer) buildMessage(to, subject, body string) []byte {
	var b strings.Builder
	m.writeCommonHeaders(&b, to, subject)
	b.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n")
	b.WriteString("\r\n")
	b.WriteString(toCRLF(body))
	return []byte(b.String())
}

// buildMultipartMessage assembles a multipart/alternative message so clients can
// pick the HTML or plain-text representation. Per RFC 2046 the parts are ordered
// least-preferred first, so the text part precedes the HTML part.
func (m *smtpMailer) buildMultipartMessage(to, subject, htmlBody, textBody string) []byte {
	boundary := newBoundary()

	var b strings.Builder
	m.writeCommonHeaders(&b, to, subject)
	fmt.Fprintf(&b, "Content-Type: multipart/alternative; boundary=\"%s\"\r\n", boundary)
	b.WriteString("\r\n")

	// Plain-text part (fallback).
	fmt.Fprintf(&b, "--%s\r\n", boundary)
	b.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n")
	b.WriteString("Content-Transfer-Encoding: 8bit\r\n")
	b.WriteString("\r\n")
	b.WriteString(toCRLF(textBody))
	b.WriteString("\r\n")

	// HTML part (preferred).
	fmt.Fprintf(&b, "--%s\r\n", boundary)
	b.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n")
	b.WriteString("Content-Transfer-Encoding: 8bit\r\n")
	b.WriteString("\r\n")
	b.WriteString(toCRLF(htmlBody))
	b.WriteString("\r\n")

	fmt.Fprintf(&b, "--%s--\r\n", boundary)
	return []byte(b.String())
}

// writeCommonHeaders writes From/To/Subject/MIME-Version shared by both message
// builders. The subject is RFC 2047 encoded so non-ASCII characters survive.
func (m *smtpMailer) writeCommonHeaders(b *strings.Builder, to, subject string) {
	fmt.Fprintf(b, "From: %s <%s>\r\n", mime.QEncoding.Encode("UTF-8", m.cfg.FromName), m.cfg.From)
	fmt.Fprintf(b, "To: %s\r\n", to)
	fmt.Fprintf(b, "Subject: %s\r\n", mime.QEncoding.Encode("UTF-8", subject))
	b.WriteString("MIME-Version: 1.0\r\n")
}

// toCRLF normalises line endings to CRLF as required by SMTP.
func toCRLF(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, "\r\n", "\n"), "\n", "\r\n")
}

// newBoundary returns a random MIME multipart boundary.
func newBoundary() string {
	var buf [16]byte
	if _, err := rand.Read(buf[:]); err != nil {
		// rand.Read essentially never fails; a fixed fallback keeps callers simple.
		return "arsiva-boundary-fallback"
	}
	return "arsiva_" + hex.EncodeToString(buf[:])
}
