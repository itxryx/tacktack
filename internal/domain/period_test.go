package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCalcDateRange(t *testing.T) {
	now := time.Date(2026, 3, 11, 15, 30, 0, 0, time.Local)
	nowSunday := time.Date(2026, 3, 15, 15, 30, 0, 0, time.Local)
	endOfDayNow := time.Date(2026, 3, 11, 23, 59, 59, 0, time.Local)
	endOfDaySunday := time.Date(2026, 3, 15, 23, 59, 59, 0, time.Local)

	tests := []struct {
		name     string
		period   PeriodType
		now      time.Time
		wantFrom time.Time
		wantTo   time.Time
	}{
		{
			name:     "Day",
			period:   PeriodDay,
			now:      now,
			wantFrom: time.Date(2026, 3, 11, 0, 0, 0, 0, time.Local),
			wantTo:   endOfDayNow,
		},
		{
			name:     "Week_水曜",
			period:   PeriodWeek,
			now:      now,
			wantFrom: time.Date(2026, 3, 9, 0, 0, 0, 0, time.Local),
			wantTo:   endOfDayNow,
		},
		{
			name:     "Month",
			period:   PeriodMonth,
			now:      now,
			wantFrom: time.Date(2026, 3, 1, 0, 0, 0, 0, time.Local),
			wantTo:   endOfDayNow,
		},
		{
			name:     "HalfYear",
			period:   PeriodHalfYear,
			now:      now,
			wantFrom: time.Date(2025, 9, 11, 0, 0, 0, 0, time.Local),
			wantTo:   endOfDayNow,
		},
		{
			name:     "Year",
			period:   PeriodYear,
			now:      now,
			wantFrom: time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local),
			wantTo:   endOfDayNow,
		},
		{
			name:     "All",
			period:   PeriodAll,
			now:      now,
			wantFrom: time.Time{},
			wantTo:   now,
		},
		{
			name:     "Week_日曜",
			period:   PeriodWeek,
			now:      nowSunday,
			wantFrom: time.Date(2026, 3, 9, 0, 0, 0, 0, time.Local),
			wantTo:   endOfDaySunday,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// act
			from, to := CalcDateRange(tc.period, tc.now)
			if tc.period == PeriodAll {
				// assert
				assert.True(t, from.IsZero(), "PeriodAll の from はゼロ値")
				assert.Equal(t, tc.wantTo, to)
				return
			}

			// assert
			assert.Equal(t, tc.wantFrom, from)
			assert.Equal(t, tc.wantTo, to)
		})
	}
}
