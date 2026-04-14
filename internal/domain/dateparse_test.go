package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testNow = time.Date(2026, 3, 9, 12, 0, 0, 0, time.Local)

func TestParseDueDate(t *testing.T) {
	date := func(year int, month time.Month, day int) *time.Time {
		d := time.Date(year, month, day, 0, 0, 0, 0, time.Local)
		return &d
	}
	wednesday := time.Date(2026, 3, 11, 12, 0, 0, 0, time.Local)

	tests := []struct {
		name     string
		input    string
		now      time.Time
		wantDate *time.Time
		wantErr  bool
	}{
		{name: "today", input: "today", now: testNow, wantDate: date(2026, 3, 9)},
		{name: "tomorrow", input: "tomorrow", now: testNow, wantDate: date(2026, 3, 10)},
		{name: "曜日_当日", input: "mon", now: testNow, wantDate: date(2026, 3, 9)},
		{name: "曜日_未来", input: "fri", now: testNow, wantDate: date(2026, 3, 13)},
		{name: "next_同曜日", input: "next mon", now: testNow, wantDate: date(2026, 3, 16)},
		{name: "next_過去曜日", input: "next sun", now: testNow, wantDate: date(2026, 3, 15)},
		{name: "next_未来曜日", input: "next fri", now: testNow, wantDate: date(2026, 3, 20)},
		{name: "明示的日付", input: "2026-03-15", now: testNow, wantDate: date(2026, 3, 15)},
		{name: "不正文字列", input: "invalid", now: testNow, wantErr: true},
		{name: "next_不正曜日", input: "next xyz", now: testNow, wantErr: true},
		{name: "過去曜日_水曜基準", input: "mon", now: wednesday, wantDate: date(2026, 3, 16)},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// act
			d, err := ParseDueDate(tc.input, tc.now)
			if tc.wantErr {
				// assert
				assert.Error(t, err)
				return
			}

			// assert
			require.NoError(t, err)
			require.NotNil(t, d)
			assert.Equal(t, *tc.wantDate, *d)
		})
	}
}
