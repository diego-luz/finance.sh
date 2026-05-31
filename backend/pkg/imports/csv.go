package imports

import (
	"bufio"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

// CSVOptions controls how a CSV statement is parsed. Zero values trigger
// auto-detection where sensible.
type CSVOptions struct {
	// Delimiter is the field separator. When 0, it is auto-detected (';' or ',').
	Delimiter rune
	// HasHeader indicates the first record is a header row (used for auto-detect
	// of column positions and skipped from the data).
	HasHeader bool
	// DateCol/DescCol/AmountCol are 0-based column indexes. -1 means auto-detect
	// from the header name (requires HasHeader); when auto-detect fails sensible
	// positional defaults are used (0=date, 1=description, last=amount).
	DateCol   int
	DescCol   int
	AmountCol int
	// DateFormat is the Go reference layout for the date column. When empty a set
	// of common BR/ISO layouts is tried.
	DateFormat string
	// DecimalSep is the decimal separator of the amount column: ',' (pt-BR) or
	// '.'. When 0 it is inferred per value.
	DecimalSep rune
}

// defaultCSVOptions returns the zero-value options with auto-detect sentinels.
func defaultCSVOptions() CSVOptions {
	return CSVOptions{DateCol: -1, DescCol: -1, AmountCol: -1}
}

// ParseCSV reads a CSV bank statement into ParsedRows. It auto-detects the
// delimiter and (when a header is present) the column positions from common
// pt-BR/English column names. BR money ("1.234,56", "-1.234,56") and plain
// decimals are both understood. The ExternalID of each row is a short sha256
// hash of date+amount+description so re-importing the same file deduplicates.
//
// Parsing is defensive: malformed lines are skipped, a handful of error
// messages are collected (and surfaced via the returned error only when NO good
// rows were produced), and the function never panics.
func ParseCSV(r io.Reader, opts CSVOptions) ([]ParsedRow, error) {
	// Read all bytes so we can sniff the delimiter before configuring the reader.
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	text := stripBOM(string(data))
	if strings.TrimSpace(text) == "" {
		return nil, fmt.Errorf("arquivo CSV vazio")
	}

	delim := opts.Delimiter
	if delim == 0 {
		delim = detectDelimiter(text)
	}

	reader := csv.NewReader(strings.NewReader(text))
	reader.Comma = delim
	reader.FieldsPerRecord = -1 // tolerate ragged rows
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true

	records, err := reader.ReadAll()
	if err != nil {
		// csv.ReadAll aborts on the first malformed record; fall back to a
		// line-by-line tolerant read so a single bad line does not kill the import.
		records = tolerantRead(text, delim)
	}
	if len(records) == 0 {
		return nil, fmt.Errorf("nenhuma linha encontrada no CSV")
	}

	dateCol, descCol, amountCol := opts.DateCol, opts.DescCol, opts.AmountCol
	start := 0
	if opts.HasHeader {
		header := records[0]
		start = 1
		if dateCol < 0 {
			dateCol = findColumn(header, dateHeaders)
		}
		if descCol < 0 {
			descCol = findColumn(header, descHeaders)
		}
		if amountCol < 0 {
			amountCol = findColumn(header, amountHeaders)
		}
	}
	// Positional fallbacks when columns are still unknown.
	if dateCol < 0 {
		dateCol = 0
	}
	if amountCol < 0 {
		amountCol = lastIndex(records)
	}
	if descCol < 0 {
		descCol = 1
	}

	rows := make([]ParsedRow, 0, len(records))
	var errs []string
	for i := start; i < len(records); i++ {
		rec := records[i]
		if isBlankRecord(rec) {
			continue
		}
		row, perr := buildCSVRow(rec, dateCol, descCol, amountCol, opts)
		if perr != "" {
			if len(errs) < 5 {
				errs = append(errs, fmt.Sprintf("linha %d: %s", i+1, perr))
			}
			continue
		}
		rows = append(rows, row)
	}

	if len(rows) == 0 {
		if len(errs) > 0 {
			return nil, fmt.Errorf("nenhuma linha válida no CSV: %s", strings.Join(errs, "; "))
		}
		return nil, fmt.Errorf("nenhuma linha válida no CSV")
	}
	return rows, nil
}

// buildCSVRow maps one CSV record to a ParsedRow. Returns a non-empty error
// string describing why the row was skipped.
func buildCSVRow(rec []string, dateCol, descCol, amountCol int, opts CSVOptions) (ParsedRow, string) {
	if dateCol >= len(rec) || amountCol >= len(rec) {
		return ParsedRow{}, "colunas insuficientes"
	}

	date, ok := parseCSVDate(strings.TrimSpace(rec[dateCol]), opts.DateFormat)
	if !ok {
		return ParsedRow{}, "data inválida"
	}

	cents, neg, ok := parseSignedDecimal(strings.TrimSpace(rec[amountCol]), opts.DecimalSep)
	if !ok {
		return ParsedRow{}, "valor inválido"
	}

	desc := ""
	if descCol >= 0 && descCol < len(rec) {
		desc = strings.TrimSpace(rec[descCol])
	}
	if desc == "" {
		desc = "Lançamento importado"
	}

	return ParsedRow{
		Date:        date,
		Description: desc,
		AmountCents: cents,
		Type:        signType(neg),
		ExternalID:  csvHash(date, cents, neg, desc),
		Raw:         strings.Join(rec, string(opts.Delimiter)),
	}, ""
}

// csvHash builds a short, stable dedup key from the row's date, signed amount
// and description. Re-importing the same file yields the same key.
func csvHash(date time.Time, cents int64, neg bool, desc string) string {
	sign := "+"
	if neg {
		sign = "-"
	}
	seed := fmt.Sprintf("%s|%s%d|%s", date.Format("2006-01-02"), sign, cents, strings.ToLower(desc))
	sum := sha256.Sum256([]byte(seed))
	return "csv:" + hex.EncodeToString(sum[:8])
}

// dateHeaders, descHeaders and amountHeaders are the lowercase header names that
// identify each logical column for auto-detection.
var (
	dateHeaders   = []string{"data", "date", "dt", "data lançamento", "data lancamento"}
	descHeaders   = []string{"descrição", "descricao", "histórico", "historico", "lançamento", "lancamento", "memo", "description", "detalhe", "histórico/descrição"}
	amountHeaders = []string{"valor", "amount", "montante", "valor (r$)", "value", "vlr"}
)

// findColumn returns the index of the first header cell whose normalized text
// matches one of the candidate names (exact, then contains). -1 when none match.
func findColumn(header []string, candidates []string) int {
	norm := make([]string, len(header))
	for i, h := range header {
		norm[i] = strings.ToLower(strings.TrimSpace(h))
	}
	// Exact match first.
	for i, h := range norm {
		for _, c := range candidates {
			if h == c {
				return i
			}
		}
	}
	// Then substring match.
	for i, h := range norm {
		for _, c := range candidates {
			if h != "" && strings.Contains(h, c) {
				return i
			}
		}
	}
	return -1
}

// parseCSVDate parses a date cell. When format is supplied it is tried first;
// otherwise a set of common BR/ISO layouts is attempted.
func parseCSVDate(raw, format string) (time.Time, bool) {
	if raw == "" {
		return time.Time{}, false
	}
	layouts := []string{"02/01/2006", "2006-01-02", "02/01/06", "02-01-2006", "2006/01/02", "01/02/2006"}
	if format != "" {
		layouts = append([]string{format}, layouts...)
	}
	for _, l := range layouts {
		if t, err := time.Parse(l, raw); err == nil {
			return t.UTC(), true
		}
	}
	return time.Time{}, false
}

// detectDelimiter sniffs ';' vs ',' from the first non-empty line by counting
// occurrences. Defaults to ',' on a tie or when neither appears.
func detectDelimiter(text string) rune {
	sc := bufio.NewScanner(strings.NewReader(text))
	sc.Buffer(make([]byte, 1024*1024), 1024*1024)
	for sc.Scan() {
		line := sc.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}
		semi := strings.Count(line, ";")
		comma := strings.Count(line, ",")
		tab := strings.Count(line, "\t")
		if tab > semi && tab > comma {
			return '\t'
		}
		if semi > comma {
			return ';'
		}
		return ','
	}
	return ','
}

// tolerantRead is a fallback splitter used when csv.ReadAll fails on a malformed
// record. It splits each line on the delimiter without quote handling.
func tolerantRead(text string, delim rune) [][]string {
	var out [][]string
	sc := bufio.NewScanner(strings.NewReader(text))
	sc.Buffer(make([]byte, 1024*1024), 1024*1024)
	for sc.Scan() {
		line := sc.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}
		out = append(out, strings.Split(line, string(delim)))
	}
	return out
}

// lastIndex returns the index of the last column of the widest record (used as
// the positional fallback for the amount column).
func lastIndex(records [][]string) int {
	max := 0
	for _, r := range records {
		if len(r) > max {
			max = len(r)
		}
	}
	if max == 0 {
		return 0
	}
	return max - 1
}

func isBlankRecord(rec []string) bool {
	for _, c := range rec {
		if strings.TrimSpace(c) != "" {
			return false
		}
	}
	return true
}

func stripBOM(s string) string {
	return strings.TrimPrefix(s, "\ufeff")
}

// parseSignedDecimal parses a money string into positive cents plus a negative
// flag. It accepts BR ("1.234,56", "-1.234,56", "R$ 1.234,56") and plain
// decimals ("-12.34", "12.34"), parentheses for negatives ("(12,34)") and a
// trailing "D"/"C" debit/credit marker. decimalSep forces the decimal mark when
// non-zero; otherwise it is inferred. Returns (cents, negative, ok).
func parseSignedDecimal(raw string, decimalSep rune) (int64, bool, bool) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return 0, false, false
	}

	negative := false
	// Parentheses denote a negative value in some exports.
	if strings.HasPrefix(s, "(") && strings.HasSuffix(s, ")") {
		negative = true
		s = strings.TrimSuffix(strings.TrimPrefix(s, "("), ")")
	}

	// Trailing debit/credit marker.
	upper := strings.ToUpper(s)
	if strings.HasSuffix(upper, "D") {
		negative = true
		s = strings.TrimSpace(s[:len(s)-1])
	} else if strings.HasSuffix(upper, "C") {
		s = strings.TrimSpace(s[:len(s)-1])
	}

	// Strip currency symbols and spaces.
	s = strings.NewReplacer("R$", "", "BRL", "", "$", "", " ", "", " ", "").Replace(s)

	if strings.HasPrefix(s, "+") {
		s = s[1:]
	}
	if strings.HasPrefix(s, "-") {
		negative = true
		s = s[1:]
	}
	if s == "" {
		return 0, false, false
	}

	// Normalize the decimal/thousand separators into a plain "<int>.<frac>" form.
	normalized, ok := normalizeNumber(s, decimalSep)
	if !ok {
		return 0, false, false
	}

	f, err := strconv.ParseFloat(normalized, 64)
	if err != nil {
		return 0, false, false
	}
	// Round to cents to avoid binary float drift (e.g. 12.34 -> 1234).
	cents := int64(f*100 + boolSign(f))
	if cents < 0 {
		cents = -cents
	}
	return cents, negative, true
}

// boolSign returns 0.5 for non-negative and -0.5 for negative, used to round to
// the nearest cent in either direction.
func boolSign(f float64) float64 {
	if f < 0 {
		return -0.5
	}
	return 0.5
}

// normalizeNumber converts a numeric string using either an explicit decimal
// separator or an inferred one into a canonical "1234.56" form (thousands
// separators removed). Returns false when the string contains unexpected
// characters.
func normalizeNumber(s string, decimalSep rune) (string, bool) {
	hasComma := strings.ContainsRune(s, ',')
	hasDot := strings.ContainsRune(s, '.')

	var dec rune
	switch {
	case decimalSep == ',' || decimalSep == '.':
		dec = decimalSep
	case hasComma && hasDot:
		// The separator that appears last is the decimal mark.
		if strings.LastIndexByte(s, ',') > strings.LastIndexByte(s, '.') {
			dec = ','
		} else {
			dec = '.'
		}
	case hasComma:
		dec = ','
	default:
		dec = '.'
	}

	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == dec:
			b.WriteRune('.')
		case r == ',' || r == '.':
			// thousands separator — drop it
		default:
			return "", false
		}
	}
	out := b.String()
	if out == "" || out == "." {
		return "", false
	}
	return out, true
}
