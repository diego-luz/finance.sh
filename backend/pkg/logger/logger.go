package logger

import (
	"log/slog"
	"os"
	"strings"
)

// New returns a structured slog logger. JSON in production, text otherwise.
func New(env string) *slog.Logger {
	var handler slog.Handler
	opts := &slog.HandlerOptions{Level: slog.LevelInfo}
	if env == "production" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		opts.Level = slog.LevelDebug
		handler = slog.NewTextHandler(os.Stdout, opts)
	}
	l := slog.New(handler)
	slog.SetDefault(l)
	return l
}

// MaskEmail returns a privacy-preserving representation of an email so it can be
// logged without exposing the full PII. It keeps the first character of the
// local part and the domain, masking the rest: "diego@finance.sh" -> "d***@finance.sh".
// Non-email or empty inputs are reduced to a generic placeholder.
func MaskEmail(email string) string {
	email = strings.TrimSpace(email)
	if email == "" {
		return ""
	}
	at := strings.LastIndex(email, "@")
	if at <= 0 {
		// Not an email shape: mask everything but the first char.
		return mask(email)
	}
	local, domain := email[:at], email[at+1:]
	return mask(local) + "@" + domain
}

// mask keeps the first rune and replaces the remainder with "***".
func mask(s string) string {
	if s == "" {
		return "***"
	}
	r := []rune(s)
	return string(r[:1]) + "***"
}
