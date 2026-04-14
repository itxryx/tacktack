package model

import (
	"strings"
	"testing"
	"time"

	"github.com/itxryx/tacktack/internal/db"
	"github.com/stretchr/testify/assert"
)

func TestViewTimeline_Empty(t *testing.T) {
	// arrange
	m := newTestModel(t)
	m.mode = modeTimeline
	m.timelineDate = time.Date(2026, 3, 9, 0, 0, 0, 0, time.Local)

	// act
	output := m.viewTimeline()

	// assert
	assert.NotEmpty(t, output)
	assert.Contains(t, output, "タイムライン")
	assert.Contains(t, output, "2026-03-09")
	assert.Contains(t, output, "合計")
	assert.Contains(t, output, "─", "区切り線が表示されること")
}

func TestViewTimeline_WithLogs(t *testing.T) {
	// arrange
	m := newTestModel(t)
	m.mode = modeTimeline
	date := time.Date(2026, 3, 9, 0, 0, 0, 0, time.Local)
	m.timelineDate = date
	start := time.Date(2026, 3, 9, 9, 0, 0, 0, time.Local)
	end := time.Date(2026, 3, 9, 10, 0, 0, 0, time.Local)
	m.tasks = []db.Task{
		{
			ID:    1,
			Title: "テストタスク",
			Tags:  []db.Tag{{Type: "project", Name: "work"}},
			TimeLogs: []db.TimeLog{
				{ID: 1, TaskID: 1, StartAt: start, EndAt: &end},
			},
		},
	}

	// act
	output := m.viewTimeline()

	// assert
	assert.Contains(t, output, "テストタスク")
	assert.Contains(t, output, "+work")
	assert.Contains(t, output, "09:00")
	assert.Contains(t, output, "1時間 00分 00秒")
}

func TestViewTimeline_ActiveLog(t *testing.T) {
	// arrange
	m := newTestModel(t)
	m.mode = modeTimeline
	date := time.Now()
	m.timelineDate = date
	start := time.Now().Add(-30 * time.Minute)
	activeLog := &db.TimeLog{ID: 1, TaskID: 1, StartAt: start}
	m.activeLog = activeLog
	m.tasks = []db.Task{
		{
			ID:    1,
			Title: "計測中タスク",
			TimeLogs: []db.TimeLog{
				{ID: 1, TaskID: 1, StartAt: start, EndAt: nil},
			},
		},
	}

	// act
	output := m.viewTimeline()

	// assert
	assert.Contains(t, output, "計測中タスク")
	assert.Contains(t, output, "計測中")
}

func TestViewTimeline_TodayLabel(t *testing.T) {
	// arrange
	m := newTestModel(t)
	m.mode = modeTimeline
	m.timelineDate = time.Now()

	// act
	output := m.viewTimeline()

	// assert
	assert.Contains(t, output, time.Now().Format("2006-01-02 (Mon)"))
}

func TestViewTimeline_ZeroDate(t *testing.T) {
	// arrange
	m := newTestModel(t)
	m.mode = modeTimeline
	m.timelineDate = time.Time{}

	// act
	output := m.viewTimeline()

	// assert
	assert.NotEmpty(t, output)
	assert.Contains(t, output, time.Now().Format("2006-01-02 (Mon)"))
}

func TestViewTimeline_MultipleTasksInSlot(t *testing.T) {
	// arrange
	m := newTestModel(t)
	m.mode = modeTimeline
	date := time.Date(2026, 3, 9, 0, 0, 0, 0, time.Local)
	m.timelineDate = date
	start := time.Date(2026, 3, 9, 9, 0, 0, 0, time.Local)
	end := time.Date(2026, 3, 9, 9, 29, 0, 0, time.Local)
	m.tasks = []db.Task{
		{
			ID:    1,
			Title: "タスクA",
			TimeLogs: []db.TimeLog{
				{ID: 1, TaskID: 1, StartAt: start, EndAt: &end},
			},
		},
		{
			ID:    2,
			Title: "タスクB",
			TimeLogs: []db.TimeLog{
				{ID: 2, TaskID: 2, StartAt: start, EndAt: &end},
			},
		},
	}

	// act
	output := m.viewTimeline()

	// assert
	assert.Contains(t, output, "タスクA")
	assert.Contains(t, output, "タスクB")
}

func TestViewTimeline_SameTaskMultipleLogsInSlot(t *testing.T) {
	// arrange
	m := newTestModel(t)
	m.mode = modeTimeline
	date := time.Date(2026, 3, 9, 0, 0, 0, 0, time.Local)
	m.timelineDate = date
	m.height = 24
	log1Start := time.Date(2026, 3, 9, 17, 0, 0, 0, time.Local)
	log1End := time.Date(2026, 3, 9, 17, 10, 0, 0, time.Local)
	log2Start := time.Date(2026, 3, 9, 17, 12, 0, 0, time.Local)
	log2End := time.Date(2026, 3, 9, 17, 20, 0, 0, time.Local)
	log3Start := time.Date(2026, 3, 9, 17, 22, 0, 0, time.Local)
	log3End := time.Date(2026, 3, 9, 17, 29, 0, 0, time.Local)
	m.tasks = []db.Task{
		{
			ID:    1,
			Title: "test",
			Tags:  []db.Tag{{Type: "project", Name: "test"}},
			TimeLogs: []db.TimeLog{
				{ID: 1, TaskID: 1, StartAt: log1Start, EndAt: &log1End},
				{ID: 2, TaskID: 1, StartAt: log2Start, EndAt: &log2End},
				{ID: 3, TaskID: 1, StartAt: log3Start, EndAt: &log3End},
			},
		},
	}
	m.timelineScroll = 34

	// act
	output := m.viewTimeline()

	// assert
	assert.Contains(t, output, "test")
	count := strings.Count(output, "+test")
	assert.Equal(t, 1, count, "同一タスクのラベルは同一スロット内で1回のみ表示される")
}

func TestViewTimeline_MidnightSpanningLog_TotalClampedToDay(t *testing.T) {
	// arrange
	m := newTestModel(t)
	m.mode = modeTimeline
	date := time.Date(2026, 3, 9, 0, 0, 0, 0, time.Local)
	m.timelineDate = date
	start := time.Date(2026, 3, 9, 23, 50, 0, 0, time.Local)
	end := time.Date(2026, 3, 10, 0, 5, 0, 0, time.Local)
	m.tasks = []db.Task{
		{
			ID:    1,
			Title: "深夜またぎタスク",
			TimeLogs: []db.TimeLog{
				{ID: 1, TaskID: 1, StartAt: start, EndAt: &end},
			},
		},
	}

	// act
	output := m.viewTimeline()

	// assert
	assert.Contains(t, output, "10分 00秒", "2026-03-09 の合計は当日分の10分のみ表示されること")
	assert.NotContains(t, output, "15分", "翌日分も含めた15分は表示されないこと")
}

func TestViewTimeline_MidnightSpanningLog_NextDay(t *testing.T) {
	// arrange
	m := newTestModel(t)
	m.mode = modeTimeline
	date := time.Date(2026, 3, 10, 0, 0, 0, 0, time.Local)
	m.timelineDate = date
	start := time.Date(2026, 3, 9, 23, 50, 0, 0, time.Local)
	end := time.Date(2026, 3, 10, 0, 5, 0, 0, time.Local)
	m.tasks = []db.Task{
		{
			ID:    1,
			Title: "深夜またぎタスク",
			TimeLogs: []db.TimeLog{
				{ID: 1, TaskID: 1, StartAt: start, EndAt: &end},
			},
		},
	}

	// act
	output := m.viewTimeline()

	// assert
	assert.Contains(t, output, "5分 00秒", "2026-03-10 の合計は翌日分の5分のみ表示されること")
	assert.NotContains(t, output, "15分", "前日分も含めた15分は表示されないこと")
}

func TestCollectTimeLogsForDate_ExcludesLogsAfterDay(t *testing.T) {
	// arrange
	date := time.Date(2026, 3, 9, 0, 0, 0, 0, time.Local)
	nextDay := time.Date(2026, 3, 10, 12, 0, 0, 0, time.Local)
	nextDayEnd := time.Date(2026, 3, 10, 13, 0, 0, 0, time.Local)
	tasks := []db.Task{
		{
			ID: 1,
			TimeLogs: []db.TimeLog{
				{ID: 1, TaskID: 1, StartAt: nextDay, EndAt: &nextDayEnd},
			},
		},
	}

	// act
	logs := collectTimeLogsForDate(tasks, date)

	// assert
	assert.Empty(t, logs, "対象日と重ならないログは除外される")
}
