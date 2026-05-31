package imports

import (
	"io"
	"strings"
	"time"
)

// ParseOFX reads an OFX statement (both the OFX/1.x SGML form and the OFX/2.x
// XML form) and extracts each <STMTTRN>...</STMTTRN> block into a ParsedRow.
//
// The parser is deliberately tolerant: SGML OFX often omits closing tags
// (e.g. `<FITID>123` on its own line), so for each known tag we read the value
// as everything up to the next `<` or end of line. Both styles are handled by
// the same routine. Malformed blocks are skipped silently; the function never
// panics on bad input.
func ParseOFX(r io.Reader) ([]ParsedRow, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	content := string(data)

	rows := make([]ParsedRow, 0, 16)
	// Iterate over every STMTTRN block. Matching is case-insensitive because some
	// generators emit lowercase tags.
	lower := strings.ToLower(content)
	const openTag = "<stmttrn>"
	const closeTag = "</stmttrn>"

	pos := 0
	for {
		start := strings.Index(lower[pos:], openTag)
		if start < 0 {
			break
		}
		start += pos
		blockStart := start + len(openTag)
		end := strings.Index(lower[blockStart:], closeTag)
		var block string
		if end < 0 {
			// No closing tag: take the rest of the document for this (last) block.
			block = content[blockStart:]
			pos = len(content)
		} else {
			block = content[blockStart : blockStart+end]
			pos = blockStart + end + len(closeTag)
		}

		if row, ok := parseSTMTTRN(block); ok {
			rows = append(rows, row)
		}
	}
	return rows, nil
}

// parseSTMTTRN turns a single transaction block into a ParsedRow. Returns false
// when the block lacks the minimum required fields (date + amount).
func parseSTMTTRN(block string) (ParsedRow, bool) {
	dtRaw := ofxTag(block, "DTPOSTED")
	amtRaw := ofxTag(block, "TRNAMT")
	if dtRaw == "" || amtRaw == "" {
		return ParsedRow{}, false
	}

	date, ok := parseOFXDate(dtRaw)
	if !ok {
		return ParsedRow{}, false
	}

	cents, neg, ok := parseSignedDecimal(amtRaw, '.')
	if !ok {
		return ParsedRow{}, false
	}

	// Description prefers NAME, falling back to MEMO, then PAYEE name-ish tags.
	desc := firstNonEmpty(ofxTag(block, "NAME"), ofxTag(block, "MEMO"), ofxTag(block, "PAYEE"))
	desc = strings.TrimSpace(desc)
	if desc == "" {
		desc = "Lançamento importado"
	}

	row := ParsedRow{
		Date:        date,
		Description: desc,
		AmountCents: cents,
		Type:        signType(neg),
		ExternalID:  strings.TrimSpace(ofxTag(block, "FITID")),
		Raw:         strings.TrimSpace(block),
	}
	return row, true
}

// ofxTag extracts the value of <TAG> from an OFX block. It handles both
// closed (`<NAME>foo</NAME>`) and unclosed SGML (`<NAME>foo` then newline or
// next tag) forms: the value runs from after the opening tag up to the next
// `<` or the end of the line, whichever comes first. Matching is
// case-insensitive. Returns "" when the tag is absent.
func ofxTag(block, tag string) string {
	lower := strings.ToLower(block)
	open := "<" + strings.ToLower(tag) + ">"
	idx := strings.Index(lower, open)
	if idx < 0 {
		return ""
	}
	valStart := idx + len(open)
	rest := block[valStart:]

	// Value ends at the first '<' (next tag / closing tag) or newline.
	endLT := strings.IndexByte(rest, '<')
	endNL := strings.IndexAny(rest, "\r\n")
	end := len(rest)
	if endLT >= 0 && endLT < end {
		end = endLT
	}
	if endNL >= 0 && endNL < end {
		end = endNL
	}
	return strings.TrimSpace(rest[:end])
}

// parseOFXDate parses the OFX DTPOSTED format: a YYYYMMDD prefix optionally
// followed by HHMMSS, a fractional part `[.xxx]` and a timezone `[-3:BRT]`.
// Only the date portion (first 8 digits) is used. Returns false on failure.
func parseOFXDate(raw string) (time.Time, bool) {
	raw = strings.TrimSpace(raw)
	if len(raw) < 8 {
		return time.Time{}, false
	}
	datePart := raw[:8]
	t, err := time.Parse("20060102", datePart)
	if err != nil {
		return time.Time{}, false
	}
	return t.UTC(), true
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

// signType maps a "is negative" flag to the transaction type. A negative amount
// on a statement is money leaving the account (expense); positive is income.
func signType(negative bool) string {
	if negative {
		return "expense"
	}
	return "income"
}
