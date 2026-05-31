package services

import (
	"strings"
	"testing"

	"github.com/finance-sh/finance-sh/internal/entities"
	"github.com/stretchr/testify/assert"
)

func TestSlugify(t *testing.T) {
	cases := []struct {
		name    string
		in      string
		wantPre string // expected slug body before the "-<6hex>" random suffix
	}{
		{"simple", "Acme Inc", "acme-inc"},
		{"accents and symbols", "Família & Cia!!!", "fam-lia-cia"},
		{"trims dashes", "  --Hello World--  ", "hello-world"},
		{"empty falls back to org", "   ", "org"},
		{"only symbols falls back to org", "@@@###", "org"},
		{"numbers kept", "Org 2024", "org-2024"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := slugify(tc.in)
			// Format: "<wantPre>-<6 hex chars>".
			assert.True(t, strings.HasPrefix(got, tc.wantPre+"-"),
				"slug %q must start with %q-", got, tc.wantPre)

			suffix := strings.TrimPrefix(got, tc.wantPre+"-")
			assert.Len(t, suffix, 6, "random suffix is 6 hex chars (got slug %q)", got)
			for _, r := range suffix {
				assert.True(t, (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f'),
					"suffix %q must be lowercase hex", suffix)
			}
		})
	}
}

func TestSlugifyUniqueSuffix(t *testing.T) {
	a := slugify("Acme")
	b := slugify("Acme")
	assert.NotEqual(t, a, b, "random suffix should make repeated slugs differ")
}

func TestReais(t *testing.T) {
	cases := []struct {
		cents int64
		want  string
	}{
		{0, "0,00"},
		{5, "0,05"},
		{50, "0,50"},
		{100, "1,00"},
		{123456, "1234,56"},
		{-1, "-0,01"},
		{-123456, "-1234,56"},
	}
	for _, tc := range cases {
		assert.Equal(t, tc.want, reais(tc.cents), "reais(%d)", tc.cents)
	}
}

func TestTranslateType(t *testing.T) {
	assert.Equal(t, "receita", translateType(entities.TxIncome))
	assert.Equal(t, "despesa", translateType(entities.TxExpense))
	assert.Equal(t, "transferência", translateType(entities.TxTransfer))
	// Unknown falls back to lower-cased raw value.
	assert.Equal(t, "unknown", translateType(entities.TransactionType("UNKNOWN")))
}
