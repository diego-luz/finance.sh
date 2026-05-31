// Package money holds the supported-currency table for organization-level
// currency configuration. The app is single-currency per organization (no FX /
// conversion): this package only declares which ISO-4217 codes are accepted and
// exposes their pt-BR display name and symbol.
package money

import "strings"

// Currency is one supported ISO-4217 currency. Name is in pt-BR.
type Currency struct {
	Code   string `json:"code"`
	Name   string `json:"name"`
	Symbol string `json:"symbol"`
}

// supported is the static table of currencies the app accepts. Codes are
// uppercase ISO-4217. Keep this list curated; there is no FX between them.
var supported = []Currency{
	{Code: "BRL", Name: "Real brasileiro", Symbol: "R$"},
	{Code: "USD", Name: "Dólar americano", Symbol: "US$"},
	{Code: "EUR", Name: "Euro", Symbol: "€"},
	{Code: "GBP", Name: "Libra esterlina", Symbol: "£"},
	{Code: "ARS", Name: "Peso argentino", Symbol: "$"},
	{Code: "CLP", Name: "Peso chileno", Symbol: "$"},
	{Code: "COP", Name: "Peso colombiano", Symbol: "$"},
	{Code: "MXN", Name: "Peso mexicano", Symbol: "$"},
	{Code: "PEN", Name: "Sol peruano", Symbol: "S/"},
	{Code: "UYU", Name: "Peso uruguaio", Symbol: "$U"},
	{Code: "PYG", Name: "Guarani paraguaio", Symbol: "₲"},
	{Code: "JPY", Name: "Iene japonês", Symbol: "¥"},
	{Code: "CAD", Name: "Dólar canadense", Symbol: "C$"},
	{Code: "AUD", Name: "Dólar australiano", Symbol: "A$"},
	{Code: "CHF", Name: "Franco suíço", Symbol: "CHF"},
}

// byCode is an index for O(1) lookups, built once at init.
var byCode = func() map[string]Currency {
	m := make(map[string]Currency, len(supported))
	for _, c := range supported {
		m[c.Code] = c
	}
	return m
}()

// DefaultCode is the fallback currency used when none is provided.
const DefaultCode = "BRL"

// IsSupported reports whether code (case-insensitive, ISO-4217) is one of the
// supported currencies.
func IsSupported(code string) bool {
	_, ok := byCode[Normalize(code)]
	return ok
}

// Normalize trims and uppercases a currency code for comparison/storage.
func Normalize(code string) string {
	return strings.ToUpper(strings.TrimSpace(code))
}

// Supported returns a copy of the supported-currency table (display order).
func Supported() []Currency {
	out := make([]Currency, len(supported))
	copy(out, supported)
	return out
}
