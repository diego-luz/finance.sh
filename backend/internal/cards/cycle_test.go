package cards

import (
	"testing"
	"time"
)

func date(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

// TestInvoiceForShortMonths covers the day-of-month clamp: a card that closes on
// the 29th/30th/31st still resolves to a valid closing date in months that are
// shorter than that. These are the cases that justify accepting closing/due days
// above 28 at the API boundary.
func TestInvoiceForShortMonths(t *testing.T) {
	tests := []struct {
		name        string
		closingDay  int
		dueDay      int
		purchase    time.Time
		wantRef     string
		wantStart   time.Time
		wantEnd     time.Time
		wantDueDate time.Time
	}{
		{
			name:       "closes on the 31st, purchase in February (28 days)",
			closingDay: 31, dueDay: 10,
			purchase:  date(2026, time.February, 5),
			wantRef:   "2026-02",
			wantStart: date(2026, time.February, 1), // day after 31 Jan
			wantEnd:   date(2026, time.February, 28),
			// dueDay <= closingDay, so the bill falls due the month after closing.
			wantDueDate: date(2026, time.March, 10),
		},
		{
			name:       "closes on the 31st, purchase on the clamped closing day itself",
			closingDay: 31, dueDay: 10,
			purchase:  date(2026, time.February, 28),
			wantRef:   "2026-02",
			wantStart: date(2026, time.February, 1),
			wantEnd:   date(2026, time.February, 28),
			// Purchase lands ON the clamped closing date: still this cycle.
			wantDueDate: date(2026, time.March, 10),
		},
		{
			name:       "closes on the 31st, purchase in a leap February",
			closingDay: 31, dueDay: 10,
			purchase:    date(2028, time.February, 10),
			wantRef:     "2028-02",
			wantStart:   date(2028, time.February, 1),
			wantEnd:     date(2028, time.February, 29), // 2028 is a leap year
			wantDueDate: date(2028, time.March, 10),
		},
		{
			name:       "closes on the 30th, purchase on 31 Jan rolls to February",
			closingDay: 30, dueDay: 10,
			purchase:    date(2026, time.January, 31),
			wantRef:     "2026-02",
			wantStart:   date(2026, time.January, 31), // day after 30 Jan
			wantEnd:     date(2026, time.February, 28),
			wantDueDate: date(2026, time.March, 10),
		},
		{
			name:       "closes on the 29th, purchase in a non-leap February",
			closingDay: 29, dueDay: 5,
			purchase:  date(2026, time.February, 20),
			wantRef:   "2026-02",
			wantStart: date(2026, time.January, 30), // day after 29 Jan
			wantEnd:   date(2026, time.February, 28),
			// January closes on the 29th but February clamps to the 28th, so this
			// cycle is 30 days long. Cycles stay contiguous: the next one starts
			// on 1 March.
			wantDueDate: date(2026, time.March, 5),
		},
		{
			name:       "closes on the 31st, December purchase rolls the due date into the next year",
			closingDay: 31, dueDay: 10,
			purchase:    date(2026, time.December, 15),
			wantRef:     "2026-12",
			wantStart:   date(2026, time.December, 1),
			wantEnd:     date(2026, time.December, 31),
			wantDueDate: date(2027, time.January, 10),
		},
		{
			name:       "due day above 28 is clamped inside the closing month",
			closingDay: 5, dueDay: 31,
			purchase:  date(2026, time.February, 3),
			wantRef:   "2026-02",
			wantStart: date(2026, time.January, 6),
			wantEnd:   date(2026, time.February, 5),
			// dueDay > closingDay, so it stays in the closing month, clamped to 28.
			wantDueDate: date(2026, time.February, 28),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := InvoiceFor(tt.closingDay, tt.dueDay, tt.purchase)
			if got.Reference != tt.wantRef {
				t.Errorf("Reference = %q, want %q", got.Reference, tt.wantRef)
			}
			if !got.PeriodStart.Equal(tt.wantStart) {
				t.Errorf("PeriodStart = %s, want %s", got.PeriodStart.Format("2006-01-02"), tt.wantStart.Format("2006-01-02"))
			}
			if !got.PeriodEnd.Equal(tt.wantEnd) {
				t.Errorf("PeriodEnd = %s, want %s", got.PeriodEnd.Format("2006-01-02"), tt.wantEnd.Format("2006-01-02"))
			}
			if !got.DueDate.Equal(tt.wantDueDate) {
				t.Errorf("DueDate = %s, want %s", got.DueDate.Format("2006-01-02"), tt.wantDueDate.Format("2006-01-02"))
			}
		})
	}
}

// TestInvoiceForCycleBoundary pins the rule that decides which invoice a purchase
// belongs to: on/before the closing day stays in the current cycle, after it rolls
// to the next one.
func TestInvoiceForCycleBoundary(t *testing.T) {
	const closingDay, dueDay = 20, 5

	before := InvoiceFor(closingDay, dueDay, date(2026, time.September, 15))
	if before.Reference != "2026-09" {
		t.Errorf("purchase before closing: Reference = %q, want %q", before.Reference, "2026-09")
	}
	if want := date(2026, time.October, 5); !before.DueDate.Equal(want) {
		t.Errorf("purchase before closing: DueDate = %s, want %s", before.DueDate.Format("2006-01-02"), want.Format("2006-01-02"))
	}

	onClosing := InvoiceFor(closingDay, dueDay, date(2026, time.September, 20))
	if onClosing.Reference != "2026-09" {
		t.Errorf("purchase on the closing day: Reference = %q, want %q", onClosing.Reference, "2026-09")
	}

	after := InvoiceFor(closingDay, dueDay, date(2026, time.September, 25))
	if after.Reference != "2026-10" {
		t.Errorf("purchase after closing: Reference = %q, want %q", after.Reference, "2026-10")
	}
	if want := date(2026, time.November, 5); !after.DueDate.Equal(want) {
		t.Errorf("purchase after closing: DueDate = %s, want %s", after.DueDate.Format("2006-01-02"), want.Format("2006-01-02"))
	}
}

// TestInvoiceFromReference checks that rebuilding an invoice from its "YYYY-MM"
// reference yields the same cycle as computing it from a purchase inside it, and
// that a malformed reference is rejected rather than silently defaulted.
func TestInvoiceFromReference(t *testing.T) {
	fromPurchase := InvoiceFor(31, 10, date(2026, time.February, 5))

	fromRef, ok := InvoiceFromReference(31, 10, "2026-02")
	if !ok {
		t.Fatal("InvoiceFromReference(31, 10, \"2026-02\") returned ok = false, want true")
	}
	if fromRef != fromPurchase {
		t.Errorf("rebuilt invoice = %+v, want %+v", fromRef, fromPurchase)
	}

	if _, ok := InvoiceFromReference(31, 10, "nonsense"); ok {
		t.Error("InvoiceFromReference with a malformed reference returned ok = true, want false")
	}
}
