package recurrence

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAddMonthsClamped(t *testing.T) {
	d := func(y int, m time.Month, dia int) time.Time { return time.Date(y, m, dia, 12, 0, 0, 0, time.UTC) }
	cases := []struct {
		de    time.Time
		meses int
		quer  time.Time
	}{
		{d(2027, 1, 31), 1, d(2027, 2, 28)},
		{d(2028, 1, 31), 1, d(2028, 2, 29)}, // leap year
		{d(2027, 1, 31), 2, d(2027, 3, 31)},
		{d(2027, 3, 31), 1, d(2027, 4, 30)},
		{d(2027, 11, 30), 3, d(2028, 2, 29)},
		{d(2027, 12, 15), 1, d(2028, 1, 15)},
		{d(2027, 1, 15), 0, d(2027, 1, 15)},
	}
	for _, c := range cases {
		assert.Equal(t, c.quer, AddMonthsClamped(c.de, c.meses), "%s + %d", c.de.Format("2006-01-02"), c.meses)
	}
}

func TestNthDoesNotDrift(t *testing.T) {
	inicio := time.Date(2027, 1, 31, 0, 0, 0, 0, time.UTC)
	var datas []string
	for n := 0; n < 4; n++ {
		datas = append(datas, Nth(inicio, "monthly", 1, n).Format("2006-01-02"))
	}
	assert.Equal(t, []string{"2027-01-31", "2027-02-28", "2027-03-31", "2027-04-30"}, datas)
	assert.Equal(t, "2027-01-14", Nth(time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC), "weekly", 2, 1).Format("2006-01-02"))
}
