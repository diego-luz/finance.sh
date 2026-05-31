// Package recurrence holds the pure scheduling math for the recurring-transaction
// engine. It has no database or entity dependencies so it can be unit-tested in
// isolation and reused from both the service and the worker.
package recurrence

import "time"

// Next returns the date that follows `from` by one step of the given frequency
// and interval. It is total: an unknown frequency falls back to a monthly step,
// and an interval < 1 is clamped to 1, so callers never have to pre-validate
// (validation lives in the service, which rejects bad input before persisting).
//
//	daily   -> +interval days
//	weekly  -> +interval*7 days
//	monthly -> +interval months (day-of-month clamped to the target month length)
//	yearly  -> +interval years  (Feb 29 clamps to Feb 28 on non-leap years)
//
// Go's time.AddDate normalises overflow (e.g. Jan 31 + 1 month -> Mar 3), which
// would silently drift a "last day of month" schedule. To keep monthly/yearly
// recurrences anchored, the day is clamped to the last valid day of the target
// month BEFORE building the date.
func Next(from time.Time, freq string, interval int) time.Time {
	if interval < 1 {
		interval = 1
	}
	switch freq {
	case "daily":
		return from.AddDate(0, 0, interval)
	case "weekly":
		return from.AddDate(0, 0, interval*7)
	case "yearly":
		return addMonthsClamped(from, interval*12)
	default: // monthly (and any unknown frequency)
		return addMonthsClamped(from, interval)
	}
}

// addMonthsClamped adds `months` calendar months to t, clamping the day of month
// to the number of days in the target month so the date never rolls over into the
// next month (which time.AddDate would do for e.g. Jan 31 + 1 month).
func addMonthsClamped(t time.Time, months int) time.Time {
	// Decompose into year/month, advance the month index, then re-clamp the day.
	y, m, d := t.Date()
	// Zero-based month arithmetic to roll years cleanly.
	total := int(m) - 1 + months
	ny := y + total/12
	nm := total % 12
	if nm < 0 {
		nm += 12
		ny--
	}
	targetMonth := time.Month(nm + 1)

	if last := daysInMonth(ny, targetMonth); d > last {
		d = last
	}
	return time.Date(ny, targetMonth, d, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
}

// daysInMonth returns the number of days in the given year/month. The 0th day of
// the next month is the last day of the requested month.
func daysInMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}
