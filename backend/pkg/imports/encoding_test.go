package imports

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/text/encoding/charmap"
)

func cp1252(t *testing.T, s string) string {
	b, err := charmap.Windows1252.NewEncoder().String(s)
	require.NoError(t, err)
	return b
}

func TestParseOFXWindows1252(t *testing.T) {
	ofx := "OFXHEADER:100\nCHARSET:1252\n<OFX><BANKTRANLIST>" +
		"<STMTTRN><TRNTYPE>DEBIT<DTPOSTED>20260901<TRNAMT>-12.34<FITID>a1<NAME>Pão de Açúcar São José</STMTTRN>" +
		"<STMTTRN><TRNTYPE>CREDIT<DTPOSTED>20260902<TRNAMT>100.00<FITID>a2<MEMO>Transferência ÇÃO</STMTTRN>" +
		"</BANKTRANLIST></OFX>"
	rows, err := ParseOFX(strings.NewReader(cp1252(t, ofx)))
	require.NoError(t, err)
	require.Len(t, rows, 2)
	assert.Equal(t, "Pão de Açúcar São José", rows[0].Description)
	assert.Equal(t, int64(1234), rows[0].AmountCents)
	assert.Equal(t, "a1", rows[0].ExternalID)
	assert.Equal(t, "Transferência ÇÃO", rows[1].Description)
	assert.Equal(t, "a2", rows[1].ExternalID)
}

func TestParseCSVWindows1252(t *testing.T) {
	csv := cp1252(t, "01/09/2026;Padaria São João;-5,50\n")
	rows, err := ParseCSV(strings.NewReader(csv), CSVOptions{Delimiter: ';', DecimalSep: ',', DateFormat: "02/01/2006", DateCol: -1, DescCol: -1, AmountCol: -1})
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, "Padaria São João", rows[0].Description)
}

func TestToUTF8KeepsUTF8(t *testing.T) {
	assert.Equal(t, "ação", string(ToUTF8([]byte("ação"))))
}

func TestDecimalToCents(t *testing.T) {
	cases := map[string]int64{"12.34": 1234, "0.1": 10, "5": 500, "1234567.89": 123456789, "0.005": 1, "0.004": 0, "19.999": 2000, ".5": 50}
	for in, want := range cases {
		got, ok := decimalToCents(in)
		assert.True(t, ok, in)
		assert.Equal(t, want, got, in)
	}
	for _, in := range []string{"99999999999999999999.00", "1.2.3", "1a.00"} {
		_, ok := decimalToCents(in)
		assert.False(t, ok, in)
	}
	cents, neg, ok := parseSignedDecimal("-R$ 1.234,56", ',')
	assert.True(t, ok)
	assert.True(t, neg)
	assert.Equal(t, int64(123456), cents)
}
