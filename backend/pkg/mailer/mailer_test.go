package mailer

import (
	"strings"
	"testing"

	"github.com/finance-sh/finance-sh/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildMessageStripsLineBreaksFromHeaders(t *testing.T) {
	msg := string(buildMessage("app@x", "a@x.local\r\nBcc: espiao@mal.com", "Oi\r\nX-Injetado: 1", "corpo"))
	cabecalho, corpo, ok := strings.Cut(msg, "\r\n\r\n")
	require.True(t, ok)
	assert.Equal(t, "corpo", corpo)
	linhas := strings.Split(cabecalho, "\r\n")
	assert.Len(t, linhas, 5, "From, To, Subject, MIME-Version, Content-Type only")
	for _, l := range linhas {
		assert.False(t, strings.HasPrefix(l, "Bcc:") || strings.HasPrefix(l, "X-Injetado:"), l)
	}
}

func TestBuildMessageEncodesUTF8Subject(t *testing.T) {
	msg := string(buildMessage("app@x", "a@x", "Redefinição de senha", "c"))
	assert.Contains(t, msg, "Subject: =?utf-8?q?")
}

func TestSendRejectsRecipientWithHeaders(t *testing.T) {
	// SMTP "configured" on a closed port: the address is refused before dialing.
	m := New(config.SMTPConfig{Host: "127.0.0.1", Port: "1", From: "app@x"}, nil)
	for _, to := range []string{"a@x.local\r\nBcc: b@y", "Fulano <a@x.local>", "nao-e-email"} {
		err := m.Send(to, "s", "b")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid recipient", to)
	}
}
