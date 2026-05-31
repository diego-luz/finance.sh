// Package mailer sends transactional emails. When no SMTP host is configured
// (the development default) it does not deliver anything: instead it logs the
// recipient, subject and body at info level so links can still be used during
// development. PII is masked in logs.
package mailer

import (
	"fmt"
	"log/slog"
	"net/smtp"
	"strings"

	"github.com/finance-sh/finance-sh/internal/config"
	"github.com/finance-sh/finance-sh/pkg/logger"
)

// Mailer delivers emails via SMTP, or logs them when SMTP is not configured.
type Mailer struct {
	cfg config.SMTPConfig
	log *slog.Logger
}

// New builds a Mailer from SMTP config.
func New(cfg config.SMTPConfig, log *slog.Logger) *Mailer {
	if log == nil {
		log = slog.Default()
	}
	return &Mailer{cfg: cfg, log: log}
}

// Send delivers a plain-text email. When SMTP_HOST is empty it only logs the
// message (no real delivery). Errors from real delivery are returned so callers
// can decide whether to surface them (most flows ignore them to avoid leaking
// account existence).
func (m *Mailer) Send(to, subject, body string) error {
	if m == nil || strings.TrimSpace(m.cfg.Host) == "" {
		// Development / no-SMTP mode: log instead of sending.
		m.log.Info("mailer: email not sent (SMTP disabled), logging instead",
			"to", logger.MaskEmail(to),
			"subject", subject,
			"body", body,
		)
		return nil
	}

	addr := m.cfg.Host + ":" + m.cfg.Port
	msg := buildMessage(m.cfg.From, to, subject, body)

	var auth smtp.Auth
	if m.cfg.User != "" {
		auth = smtp.PlainAuth("", m.cfg.User, m.cfg.Pass, m.cfg.Host)
	}

	if err := smtp.SendMail(addr, auth, m.cfg.From, []string{to}, msg); err != nil {
		m.log.Error("mailer: failed to send email", "to", logger.MaskEmail(to), "error", err)
		return err
	}
	m.log.Info("mailer: email sent", "to", logger.MaskEmail(to), "subject", subject)
	return nil
}

func buildMessage(from, to, subject, body string) []byte {
	headers := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=\"UTF-8\"\r\n\r\n",
		from, to, subject,
	)
	return []byte(headers + body)
}
