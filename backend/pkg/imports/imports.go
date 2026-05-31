// Package imports parses bank statements (OFX and CSV) into a normalized set of
// rows the import service can preview and commit. It depends only on the stdlib:
// the OFX parser is a tolerant hand-written scanner (no external OFX library) and
// CSV uses encoding/csv. Parsing is defensive: malformed rows/blocks are skipped
// and a few error messages are collected rather than aborting the whole file.
package imports

import "time"

// ParsedRow is one normalized statement line. Amounts are always positive cents
// (AmountCents); the sign of the source value is encoded in Type. ExternalID is
// the dedup key (OFX FITID, or a short hash for CSV). Raw carries the original
// fragment for debugging and is not persisted.
type ParsedRow struct {
	Date        time.Time `json:"date"`
	Description string    `json:"description"`
	AmountCents int64     `json:"amount"`
	Type        string    `json:"type"` // "income" | "expense"
	ExternalID  string    `json:"external_id,omitempty"`
	Raw         string    `json:"-"`
}
