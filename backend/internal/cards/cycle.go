// Package cards holds the credit-card billing-cycle math shared by the
// transaction and credit-card services. Keeping it pure (no DB, no entities)
// makes the invoice rule easy to reason about and unit-test.
package cards

import "time"

// Invoice describes the billing cycle a purchase belongs to.
//
// Rule (documented so the frontend and migrations agree):
//
//   - A purchase on date P with day(P) <= ClosingDay belongs to the invoice
//     whose CLOSING is in P's own month; otherwise it rolls to the next month.
//   - Reference is the closing year-month "YYYY-MM".
//   - PeriodStart is the day AFTER the previous month's closing date.
//   - PeriodEnd is the closing date itself.
//   - DueDate is DueDay of the closing month, advanced to the next month when
//     DueDay <= ClosingDay (the bill falls due after it closes).
//
// All day-of-month values are clamped to the last day of the target month so
// e.g. ClosingDay=31 works in February.
type Invoice struct {
	Reference   string    // "YYYY-MM" of the closing date
	PeriodStart time.Time // day after previous closing (00:00 UTC)
	PeriodEnd   time.Time // closing date (00:00 UTC)
	DueDate     time.Time // payment due date (00:00 UTC)
}

// clampDay returns a UTC midnight date for (year, month, day), clamping day to
// the last valid day of that month (so day=31 in February becomes the 28th/29th).
func clampDay(year int, month time.Month, day int) time.Time {
	if day < 1 {
		day = 1
	}
	last := lastDay(year, month)
	if day > last {
		day = last
	}
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

// lastDay returns the number of days in the given month.
func lastDay(year int, month time.Month) int {
	// Day 0 of the next month is the last day of this month.
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

// InvoiceFor computes the invoice a purchase on `purchase` belongs to, given the
// card's closingDay and dueDay. See the Invoice doc for the rule.
func InvoiceFor(closingDay, dueDay int, purchase time.Time) Invoice {
	p := purchase.UTC()
	year, month := p.Year(), p.Month()

	// The closing date for the purchase's own month.
	closing := clampDay(year, month, closingDay)

	// If the purchase happens strictly after that month's closing date, it belongs
	// to next month's invoice. We compare on the clamped day to honour edge cases.
	// Advance the (year, month) — not the date — and re-clamp so short months are
	// handled correctly (e.g. closingDay=31 -> February's closing is the 28th/29th).
	if p.Day() > closing.Day() {
		nextYear, nextMonth := year, month+1
		if nextMonth > time.December {
			nextYear, nextMonth = year+1, time.January
		}
		closing = clampDay(nextYear, nextMonth, closingDay)
	}

	return invoiceFromClosing(closing, closingDay, dueDay)
}

// InvoiceFromReference rebuilds an Invoice from its "YYYY-MM" reference. Used by
// the list/detail endpoints which already know the reference. Returns ok=false
// when the reference is malformed.
func InvoiceFromReference(closingDay, dueDay int, reference string) (Invoice, bool) {
	t, err := time.Parse("2006-01", reference)
	if err != nil {
		return Invoice{}, false
	}
	closing := clampDay(t.Year(), t.Month(), closingDay)
	return invoiceFromClosing(closing, closingDay, dueDay), true
}

// invoiceFromClosing derives the full Invoice from a (clamped) closing date.
// Month arithmetic is done on (year, month) integers — never via AddDate on a
// clamped date — so day-of-month overflow can't roll into the wrong month.
func invoiceFromClosing(closing time.Time, closingDay, dueDay int) Invoice {
	cy, cm := closing.Year(), closing.Month()

	// Previous closing: closingDay of the previous month, clamped.
	prevYear, prevMonth := cy, cm-1
	if prevMonth < time.January {
		prevYear, prevMonth = cy-1, time.December
	}
	prevClosing := clampDay(prevYear, prevMonth, closingDay)
	periodStart := prevClosing.AddDate(0, 0, 1)

	// Due date: DueDay of the closing month, rolled to next month when the due day
	// is on/before the closing day (so the due date lands after the cycle closes).
	dueYear, dueMonth := cy, cm
	if dueDay <= closingDay {
		dueMonth++
		if dueMonth > time.December {
			dueYear, dueMonth = cy+1, time.January
		}
	}
	due := clampDay(dueYear, dueMonth, dueDay)

	return Invoice{
		Reference:   closing.Format("2006-01"),
		PeriodStart: periodStart,
		PeriodEnd:   closing,
		DueDate:     due,
	}
}
