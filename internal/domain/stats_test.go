package domain

import (
	"testing"
	"time"

	"github.com/itxryx/tacktack/internal/db"
	"github.com/stretchr/testify/assert"
)

func makeLog(start, end string) db.TimeLog {
	s, _ := time.Parse("2006-01-02 15:04", start)
	e, _ := time.Parse("2006-01-02 15:04", end)
	return db.TimeLog{StartAt: s, EndAt: &e}
}

func makeOpenLog(start string) db.TimeLog {
	s, _ := time.Parse("2006-01-02 15:04", start)
	return db.TimeLog{StartAt: s, EndAt: nil}
}

func makeLogLocal(start, end string) db.TimeLog {
	s, _ := time.ParseInLocation("2006-01-02 15:04", start, time.Local)
	e, _ := time.ParseInLocation("2006-01-02 15:04", end, time.Local)
	return db.TimeLog{StartAt: s, EndAt: &e}
}

func TestCalcTotalSeconds(t *testing.T) {
	tests := []struct {
		name string
		logs []db.TimeLog
		want int
	}{
		{
			name: "通常の合計",
			logs: []db.TimeLog{
				makeLog("2026-03-09 09:00", "2026-03-09 10:30"),
				makeLog("2026-03-09 14:00", "2026-03-09 15:00"),
			},
			want: 9000,
		},
		{
			name: "NullEndAtは除外",
			logs: []db.TimeLog{
				makeLog("2026-03-09 09:00", "2026-03-09 10:00"),
				makeOpenLog("2026-03-09 11:00"),
			},
			want: 3600,
		},
		{
			name: "空スライス",
			logs: nil,
			want: 0,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// act
			got := CalcTotalSeconds(tc.logs)

			// assert
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestHasNullEndAt(t *testing.T) {
	tests := []struct {
		name string
		logs []db.TimeLog
		want bool
	}{
		{
			name: "NullEndAtあり",
			logs: []db.TimeLog{makeOpenLog("2026-03-09 09:00")},
			want: true,
		},
		{
			name: "NullEndAtなし",
			logs: []db.TimeLog{makeLog("2026-03-09 09:00", "2026-03-09 10:00")},
			want: false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// act
			got := HasNullEndAt(tc.logs)

			// assert
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestCalcTagTimeStats_Basic(t *testing.T) {
	// arrange
	from := mustParseTime("2026-03-09 00:00")
	to := mustParseTime("2026-03-09 23:59")
	tasks := []db.Task{
		{
			Tags:     []db.Tag{{Type: "project", Name: "work"}},
			TimeLogs: []db.TimeLog{makeLog("2026-03-09 09:00", "2026-03-09 10:00")},
		},
		{
			Tags:     []db.Tag{{Type: "context", Name: "office"}},
			TimeLogs: []db.TimeLog{makeLog("2026-03-09 14:00", "2026-03-09 15:00")},
		},
	}

	// act
	stats := CalcTagTimeStats(tasks, from, to)

	// assert
	if len(stats) != 2 {
		t.Fatalf("expected 2 stats, got %d", len(stats))
	}
	assert.Equal(t, 3600, stats[0].TotalSec)
	assert.InDelta(t, 0.5, stats[0].Percentage, 0.001)
	assert.Equal(t, 3600, stats[1].TotalSec)
	assert.InDelta(t, 0.5, stats[1].Percentage, 0.001)
}

func TestCalcTagTimeStats_PeriodClamp(t *testing.T) {
	// arrange
	from := mustParseTime("2026-03-09 10:00")
	to := mustParseTime("2026-03-09 11:00")
	tasks := []db.Task{
		{
			Tags:     []db.Tag{{Type: "project", Name: "work"}},
			TimeLogs: []db.TimeLog{makeLog("2026-03-09 09:00", "2026-03-09 12:00")},
		},
	}

	// act
	stats := CalcTagTimeStats(tasks, from, to)

	// assert
	if len(stats) != 1 {
		t.Fatalf("expected 1 stat, got %d", len(stats))
	}
	assert.Equal(t, 3600, stats[0].TotalSec)
}

func TestCalcTagTimeStats_NoTags(t *testing.T) {
	// arrange
	from := mustParseTime("2026-03-09 00:00")
	to := mustParseTime("2026-03-09 23:59")
	tasks := []db.Task{
		{
			Tags:     []db.Tag{},
			TimeLogs: []db.TimeLog{makeLog("2026-03-09 09:00", "2026-03-09 10:00")},
		},
	}

	// act
	stats := CalcTagTimeStats(tasks, from, to)

	// assert
	assert.Empty(t, stats)
}

func TestCalcTagTimeStats_DeterministicOrder(t *testing.T) {
	// arrange
	from := mustParseTime("2026-03-09 00:00")
	to := mustParseTime("2026-03-09 23:59")
	tasks := []db.Task{
		{
			Tags: []db.Tag{
				{Type: "project", Name: "work"},
				{Type: "context", Name: "office"},
			},
			TimeLogs: []db.TimeLog{makeLog("2026-03-09 09:00", "2026-03-09 10:00")},
		},
	}

	for i := 0; i < 20; i++ {
		// act
		stats := CalcTagTimeStats(tasks, from, to)

		// assert
		if len(stats) != 2 {
			t.Fatalf("iter %d: expected 2 stats, got %d", i, len(stats))
		}
		assert.Equal(t, "context", stats[0].TagType, "iter %d: context が先頭であること", i)
		assert.Equal(t, "office", stats[0].TagName, "iter %d", i)
		assert.Equal(t, "project", stats[1].TagType, "iter %d: project が2番目であること", i)
		assert.Equal(t, "work", stats[1].TagName, "iter %d", i)
	}
}

func TestCalcTagTimeStats_OutOfRange(t *testing.T) {
	// arrange
	from := mustParseTime("2026-03-10 00:00")
	to := mustParseTime("2026-03-10 23:59")
	tasks := []db.Task{
		{
			Tags:     []db.Tag{{Type: "project", Name: "work"}},
			TimeLogs: []db.TimeLog{makeLog("2026-03-09 09:00", "2026-03-09 10:00")},
		},
	}

	// act
	stats := CalcTagTimeStats(tasks, from, to)

	// assert
	assert.Empty(t, stats)
}

func TestCalcTagTimeStats_ZeroTotalSec(t *testing.T) {
	// arrange
	from := mustParseTime("2026-03-11 00:00")
	to := mustParseTime("2026-03-11 23:59")
	tasks := []db.Task{
		{
			Tags:     []db.Tag{{Type: "project", Name: "work"}},
			TimeLogs: []db.TimeLog{makeLog("2026-03-09 09:00", "2026-03-09 10:00")},
		},
	}

	// act
	stats := CalcTagTimeStats(tasks, from, to)

	// assert
	assert.Empty(t, stats, "全ログが期間外の場合は空結果")
}

func mustParseTime(s string) time.Time {
	t, err := time.Parse("2006-01-02 15:04", s)
	if err != nil {
		panic(err)
	}
	return t
}

func TestCalcDailyStats(t *testing.T) {
	// arrange
	logs := []db.TimeLog{
		makeLog("2026-03-09 09:00", "2026-03-09 10:30"),
		makeLog("2026-03-09 14:00", "2026-03-09 15:00"),
		makeLog("2026-03-10 10:00", "2026-03-10 11:00"),
		makeOpenLog("2026-03-10 12:00"),
	}

	// act
	stats := CalcDailyStats(logs)

	// assert
	assert.Len(t, stats, 2)
	assert.Equal(t, 9000, stats[0].TotalSeconds)
	assert.Equal(t, 3600, stats[1].TotalSeconds)
}

func TestCalcDailyStats_MidnightSpanning(t *testing.T) {
	// arrange
	logs := []db.TimeLog{
		makeLogLocal("2026-03-09 23:50", "2026-03-10 00:05"),
	}

	// act
	stats := CalcDailyStats(logs)

	// assert
	assert.Len(t, stats, 2, "深夜またぎログは2日分の統計になること")
	assert.Equal(t, time.Date(2026, 3, 9, 0, 0, 0, 0, time.Local), stats[0].Date)
	assert.Equal(t, 600, stats[0].TotalSeconds, "9日分は23:50〜24:00の600秒")
	assert.Equal(t, time.Date(2026, 3, 10, 0, 0, 0, 0, time.Local), stats[1].Date)
	assert.Equal(t, 300, stats[1].TotalSeconds, "10日分は0:00〜0:05の300秒")
}

func TestCalcDailyStats_ThreeDaySpan(t *testing.T) {
	// arrange
	logs := []db.TimeLog{
		makeLogLocal("2026-03-09 23:00", "2026-03-11 01:00"),
	}

	// act
	stats := CalcDailyStats(logs)

	// assert
	assert.Len(t, stats, 3, "3日またぎログは3日分の統計になること")
	assert.Equal(t, time.Date(2026, 3, 9, 0, 0, 0, 0, time.Local), stats[0].Date)
	assert.Equal(t, 3600, stats[0].TotalSeconds, "3/9分は23:00〜24:00の1時間")
	assert.Equal(t, time.Date(2026, 3, 10, 0, 0, 0, 0, time.Local), stats[1].Date)
	assert.Equal(t, 86400, stats[1].TotalSeconds, "3/10分は0:00〜24:00の24時間")
	assert.Equal(t, time.Date(2026, 3, 11, 0, 0, 0, 0, time.Local), stats[2].Date)
	assert.Equal(t, 3600, stats[2].TotalSeconds, "3/11分は0:00〜1:00の1時間")
}

func TestCalcDailyStats_TaskCount(t *testing.T) {
	// arrange
	logs := []db.TimeLog{
		makeLog("2026-03-09 09:00", "2026-03-09 10:00"),
		makeLog("2026-03-09 14:00", "2026-03-09 15:00"),
	}

	// act
	stats := CalcDailyStats(logs)

	// assert
	assert.Len(t, stats, 1)
	assert.Equal(t, 2, stats[0].TaskCount, "同一日の2ログはTaskCount==2")
}

func TestCalcTotalSecondsInRange(t *testing.T) {
	tests := []struct {
		name string
		logs []db.TimeLog
		from time.Time
		to   time.Time
		want int
	}{
		{
			name: "期間内の通常合計",
			from: mustParseTime("2026-03-09 09:00"),
			to:   mustParseTime("2026-03-09 17:00"),
			logs: []db.TimeLog{
				makeLog("2026-03-09 09:00", "2026-03-09 10:00"),
				makeLog("2026-03-09 16:00", "2026-03-09 17:00"),
			},
			want: 7200,
		},
		{
			name: "深夜またぎ_当日クランプ",
			from: mustParseTime("2026-03-09 00:00"),
			to:   mustParseTime("2026-03-10 00:00"),
			logs: []db.TimeLog{makeLog("2026-03-09 23:50", "2026-03-10 00:05")},
			want: 600,
		},
		{
			name: "深夜またぎ_翌日クランプ",
			from: mustParseTime("2026-03-10 00:00"),
			to:   mustParseTime("2026-03-11 00:00"),
			logs: []db.TimeLog{makeLog("2026-03-09 23:50", "2026-03-10 00:05")},
			want: 300,
		},
		{
			name: "NullEndAtは除外",
			from: mustParseTime("2026-03-09 00:00"),
			to:   mustParseTime("2026-03-09 23:59"),
			logs: []db.TimeLog{
				makeLog("2026-03-09 09:00", "2026-03-09 10:00"),
				makeOpenLog("2026-03-09 11:00"),
			},
			want: 3600,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// act
			got := CalcTotalSecondsInRange(tc.logs, tc.from, tc.to)

			// assert
			assert.Equal(t, tc.want, got)
		})
	}
}
