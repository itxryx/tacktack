package model

import (
	"testing"
	"time"

	"github.com/itxryx/tacktack/internal/db"
	"github.com/stretchr/testify/assert"
)

func TestFormatElapsed(t *testing.T) {
	tests := []struct {
		name        string
		offset      time.Duration
		wantEqual   string
		wantContain string
	}{
		{
			name:      "1時間超",
			offset:    -90*time.Minute - 5*time.Second,
			wantEqual: "01:30:05",
		},
		{
			name:        "1時間未満",
			offset:      -5 * time.Minute,
			wantContain: "00:05:",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// act
			got := formatElapsed(time.Now().Add(tc.offset))

			// assert
			if tc.wantEqual != "" {
				assert.Equal(t, tc.wantEqual, got)
			}
			if tc.wantContain != "" {
				assert.Contains(t, got, tc.wantContain)
			}
		})
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxWidth int
		want     string
	}{
		{name: "制限内", input: "hello", maxWidth: 10, want: "hello"},
		{name: "ちょうど制限", input: "hello", maxWidth: 5, want: "hello"},
		{name: "超過", input: "hello world", maxWidth: 6, want: "hello…"},
		{name: "マルチバイト", input: "あいうえお", maxWidth: 3, want: "あ…"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// act + assert
			assert.Equal(t, tc.want, truncateString(tc.input, tc.maxWidth))
		})
	}
}

func TestRenderTaskRow_Basic(t *testing.T) {
	// arrange
	m := newTestModel(t)
	task := db.Task{ID: 1, Title: "テストタスク"}

	// act
	row := m.renderTaskRow(0, task)

	// assert
	assert.NotContains(t, row, "x ")
	assert.Contains(t, row, "テストタスク")
}

func TestRenderTaskRow_Cursor(t *testing.T) {
	// arrange
	m := newTestModel(t)
	m.cursor = 0
	task := db.Task{ID: 1, Title: "カーソルタスク"}

	// act
	row := m.renderTaskRow(0, task)

	// assert
	assert.Contains(t, row, "▶")
}

func TestRenderTaskRow_NoCursor(t *testing.T) {
	// arrange
	m := newTestModel(t)
	m.cursor = 1
	task := db.Task{ID: 1, Title: "タスク"}

	// act
	row := m.renderTaskRow(0, task)

	// assert
	assert.NotContains(t, row, "▶")
}

func TestRenderTaskRow_WithPriority(t *testing.T) {
	// arrange
	m := newTestModel(t)
	task := db.Task{ID: 1, Title: "タスク", Priority: "A"}

	// act
	row := m.renderTaskRow(0, task)

	// assert
	assert.Contains(t, row, "(A)")
}

func TestRenderTaskRow_WithDueDate(t *testing.T) {
	// arrange
	m := newTestModel(t)
	due := time.Now().Add(24 * time.Hour)
	task := db.Task{ID: 1, Title: "タスク", DueDate: &due}

	// act
	row := m.renderTaskRow(0, task)

	// assert
	assert.Contains(t, row, "due:")
}

func TestRenderTaskRow_WithTags(t *testing.T) {
	// arrange
	m := newTestModel(t)
	task := db.Task{
		ID:    1,
		Title: "タスク",
		Tags: []db.Tag{
			{Type: "project", Name: "work"},
			{Type: "context", Name: "office"},
		},
	}

	// act
	row := m.renderTaskRow(0, task)

	// assert
	assert.Contains(t, row, "+work")
	assert.Contains(t, row, "@office")
}

func TestRenderTaskRow_Completed(t *testing.T) {
	// arrange
	m := newTestModel(t)
	end := time.Now()
	start := end.Add(-30 * time.Minute)
	task := db.Task{
		ID:          1,
		Title:       "完了タスク",
		IsCompleted: true,
		TimeLogs:    []db.TimeLog{{StartAt: start, EndAt: &end}},
	}

	// act
	row := m.renderTaskRow(0, task)

	// assert
	assert.Contains(t, row, "x ")
	assert.Contains(t, row, "秒", "累積時間が表示されること")
}

func TestRenderTaskRow_Completed_NoLogs(t *testing.T) {
	// arrange
	m := newTestModel(t)
	task := db.Task{ID: 1, Title: "ログなし完了タスク", IsCompleted: true}

	// act
	row := m.renderTaskRow(0, task)

	// assert
	assert.Contains(t, row, "x ")
}

func TestRenderTaskRow_Tracking(t *testing.T) {
	// arrange
	m := newTestModel(t)
	m.activeLog = &db.TimeLog{
		ID:      1,
		TaskID:  1,
		StartAt: time.Now().Add(-5 * time.Minute),
	}
	task := db.Task{ID: 1, Title: "計測中タスク"}

	// act
	row := m.renderTaskRow(0, task)

	// assert
	assert.Contains(t, row, "計測中")
	assert.Contains(t, row, "00:05:", "現在セッション経過時間がHH:MM:SS形式で表示されること")
}

func TestRenderTaskRow_Tracking_WithPastLogs(t *testing.T) {
	// arrange
	m := newTestModel(t)
	pastEnd := time.Now().Add(-10 * time.Minute)
	pastStart := pastEnd.Add(-60 * time.Minute)
	m.activeLog = &db.TimeLog{
		ID:      2,
		TaskID:  1,
		StartAt: time.Now().Add(-5 * time.Minute),
	}
	task := db.Task{
		ID:    1,
		Title: "計測中タスク",
		TimeLogs: []db.TimeLog{
			{StartAt: pastStart, EndAt: &pastEnd},
			{ID: 2, StartAt: m.activeLog.StartAt},
		},
	}

	// act
	row := m.renderTaskRow(0, task)

	// assert
	assert.Contains(t, row, "計測中")
	assert.Contains(t, row, "01:", "1時間以上の累積がHH:MM:SS形式で表示されること")
}

func TestRenderTaskRow_InProgressWithLogs(t *testing.T) {
	// arrange
	m := newTestModel(t)
	end := time.Now().Add(-10 * time.Minute)
	start := end.Add(-30 * time.Minute)
	task := db.Task{
		ID:       1,
		Title:    "過去ログあり未計測タスク",
		TimeLogs: []db.TimeLog{{StartAt: start, EndAt: &end}},
	}

	// act
	row := m.renderTaskRow(0, task)

	// assert
	assert.Contains(t, row, "秒", "累積時間が表示されること")
	assert.NotContains(t, row, "計測中")
}

func TestViewList_EmptyList(t *testing.T) {
	// arrange
	m := newTestModel(t)

	// act
	output := m.viewList()

	// assert
	assert.Contains(t, output, "タスクがありません")
}

func TestViewList_WithTasks(t *testing.T) {
	// arrange
	m := newTestModel(t)
	m.tasks = []db.Task{
		{ID: 1, Title: "タスク1"},
		{ID: 2, Title: "タスク2"},
	}

	// act
	output := m.viewList()

	// assert
	assert.Contains(t, output, "タスク1")
	assert.Contains(t, output, "タスク2")
}
