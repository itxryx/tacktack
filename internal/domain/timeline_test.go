package domain

import (
	"testing"
	"time"

	"github.com/itxryx/tacktack/internal/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildTimelineSlots_Structure(t *testing.T) {
	// arrange
	date := time.Date(2026, 3, 9, 0, 0, 0, 0, time.Local)

	// act
	slots := BuildTimelineSlots(nil, date, nil)

	// assert
	assert.Len(t, slots, 48, "48スロット（30分×48=24時間）")
	assert.Equal(t, 0, slots[0].Hour)
	assert.Equal(t, 0, slots[0].Minute)
	assert.Equal(t, 9, slots[18].Hour)
	assert.Equal(t, 0, slots[18].Minute)
	assert.Equal(t, 23, slots[47].Hour)
	assert.Equal(t, 30, slots[47].Minute)
}

func TestBuildTimelineSlots_WithLog(t *testing.T) {
	// arrange
	date := time.Date(2026, 3, 9, 0, 0, 0, 0, time.Local)
	start := time.Date(2026, 3, 9, 9, 0, 0, 0, time.Local)
	end := time.Date(2026, 3, 9, 10, 30, 0, 0, time.Local)
	tasks := []db.Task{
		{
			ID:    1,
			Title: "タスクA",
			Tags:  []db.Tag{{Type: "project", Name: "work"}},
			TimeLogs: []db.TimeLog{
				{ID: 1, TaskID: 1, StartAt: start, EndAt: &end},
			},
		},
	}

	// act
	slots := BuildTimelineSlots(tasks, date, nil)

	// assert
	require.Len(t, slots, 48)
	assert.Len(t, slots[18].Tasks, 1, "9:00スロットにタスクあり")
	assert.Len(t, slots[19].Tasks, 1, "9:30スロットにタスクあり")
	assert.Len(t, slots[20].Tasks, 1, "10:00スロットにタスクあり")
	assert.Empty(t, slots[21].Tasks, "10:30スロットはタスクなし")
	assert.Equal(t, "タスクA", slots[18].Tasks[0].Title)
	assert.Equal(t, []string{"+work"}, slots[18].Tasks[0].Tags)
	assert.False(t, slots[18].Tasks[0].IsActive)
}

func TestBuildTimelineSlots_DifferentDayExcluded(t *testing.T) {
	// arrange
	date := time.Date(2026, 3, 9, 0, 0, 0, 0, time.Local)
	start := time.Date(2026, 3, 10, 9, 0, 0, 0, time.Local)
	end := time.Date(2026, 3, 10, 10, 0, 0, 0, time.Local)
	tasks := []db.Task{
		{
			ID:    1,
			Title: "タスクA",
			TimeLogs: []db.TimeLog{
				{ID: 1, TaskID: 1, StartAt: start, EndAt: &end},
			},
		},
	}

	// act
	slots := BuildTimelineSlots(tasks, date, nil)

	// assert
	for i, slot := range slots {
		assert.Empty(t, slot.Tasks, "スロット%dは空であるべき", i)
	}
}

func TestBuildTimelineSlots_ActiveLog(t *testing.T) {
	// arrange
	now := time.Now()
	date := now
	start := now.Add(-2 * time.Hour)
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	if start.Before(dayStart) {
		start = dayStart
	}
	activeLog := &db.TimeLog{ID: 1, TaskID: 1, StartAt: start}
	tasks := []db.Task{
		{
			ID:    1,
			Title: "計測中タスク",
			TimeLogs: []db.TimeLog{
				{ID: 1, TaskID: 1, StartAt: start, EndAt: nil},
			},
		},
	}

	// act
	slots := BuildTimelineSlots(tasks, date, activeLog)

	// assert
	hasActiveTask := false
	for _, slot := range slots {
		for _, task := range slot.Tasks {
			if task.IsActive {
				hasActiveTask = true
				break
			}
		}
	}
	assert.True(t, hasActiveTask, "計測中タスクが少なくとも1つのスロットに表示される")
}

func TestBuildTimelineSlots_NullLogWithoutActiveLog(t *testing.T) {
	// arrange
	date := time.Date(2026, 3, 9, 0, 0, 0, 0, time.Local)
	start := time.Date(2026, 3, 9, 9, 0, 0, 0, time.Local)
	otherActiveLog := &db.TimeLog{ID: 99, TaskID: 99}
	tasks := []db.Task{
		{
			ID:    1,
			Title: "タスクA",
			TimeLogs: []db.TimeLog{
				{ID: 1, TaskID: 1, StartAt: start, EndAt: nil},
			},
		},
	}

	// act
	slots := BuildTimelineSlots(tasks, date, otherActiveLog)

	// assert
	for _, slot := range slots {
		assert.Empty(t, slot.Tasks, "activeLog IDが異なるのでスキップされるべき")
	}
}

func TestBuildTimelineSlots_CrossDayFromPrevDay(t *testing.T) {
	// arrange
	date := time.Date(2026, 3, 9, 0, 0, 0, 0, time.Local)
	start := time.Date(2026, 3, 8, 23, 0, 0, 0, time.Local)
	end := time.Date(2026, 3, 9, 1, 0, 0, 0, time.Local)
	tasks := []db.Task{
		{
			ID:    1,
			Title: "深夜またぎタスク",
			TimeLogs: []db.TimeLog{
				{ID: 1, TaskID: 1, StartAt: start, EndAt: &end},
			},
		},
	}

	// act
	slots := BuildTimelineSlots(tasks, date, nil)

	// assert
	assert.Len(t, slots[0].Tasks, 1, "スロット0(00:00)にタスクあり")
	assert.Len(t, slots[1].Tasks, 1, "スロット1(00:30)にタスクあり")
	assert.Empty(t, slots[2].Tasks, "スロット2(01:00)はタスクなし")
}

func TestBuildTimelineSlots_CrossDayToNextDay(t *testing.T) {
	// arrange
	date := time.Date(2026, 3, 9, 0, 0, 0, 0, time.Local)
	start := time.Date(2026, 3, 9, 23, 0, 0, 0, time.Local)
	end := time.Date(2026, 3, 10, 1, 0, 0, 0, time.Local)
	tasks := []db.Task{
		{
			ID:    1,
			Title: "深夜またぎタスク",
			TimeLogs: []db.TimeLog{
				{ID: 1, TaskID: 1, StartAt: start, EndAt: &end},
			},
		},
	}

	// act
	slots := BuildTimelineSlots(tasks, date, nil)

	// assert
	assert.Len(t, slots[46].Tasks, 1, "スロット46(23:00)にタスクあり")
	assert.Len(t, slots[47].Tasks, 1, "スロット47(23:30)にタスクあり")
}

func TestBuildTimelineSlots_MultipleTasksSameSlot(t *testing.T) {
	// arrange
	date := time.Date(2026, 3, 9, 0, 0, 0, 0, time.Local)
	start := time.Date(2026, 3, 9, 9, 0, 0, 0, time.Local)
	end := time.Date(2026, 3, 9, 9, 29, 0, 0, time.Local)
	tasks := []db.Task{
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
	slots := BuildTimelineSlots(tasks, date, nil)

	// assert
	assert.Len(t, slots[18].Tasks, 2, "同一スロットに2タスク出現")
	taskIDs := []uint{slots[18].Tasks[0].TaskID, slots[18].Tasks[1].TaskID}
	assert.Contains(t, taskIDs, uint(1), "タスクAが含まれること")
	assert.Contains(t, taskIDs, uint(2), "タスクBが含まれること")
}
