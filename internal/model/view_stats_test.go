package model

import (
	"testing"
	"time"

	"github.com/itxryx/tacktack/internal/db"
	"github.com/stretchr/testify/assert"
)

func TestFormatSeconds(t *testing.T) {
	tests := []struct {
		sec  int
		want string
	}{
		{0, "0秒"},
		{1, "1秒"},
		{59, "59秒"},
		{60, "1分 00秒"},
		{61, "1分 01秒"},
		{90, "1分 30秒"},
		{3600, "1時間 00分 00秒"},
		{3661, "1時間 01分 01秒"},
		{5400, "1時間 30分 00秒"},
		{7200, "2時間 00分 00秒"},
		{7325, "2時間 02分 05秒"},
	}
	for _, tt := range tests {
		// act + assert
		got := formatSeconds(tt.sec)
		assert.Equal(t, tt.want, got, "formatSeconds(%d)", tt.sec)
	}
}

func TestViewStats_NoData(t *testing.T) {
	// arrange
	m := newTestModel(t)

	// act
	output := m.viewStats()

	// assert
	assert.Contains(t, output, "統計・セッション情報")
	assert.Contains(t, output, "タイムトラッキング")
	assert.Contains(t, output, "セッション数")
	assert.Contains(t, output, "異常なセッションはありません")
}

func TestViewStats_WithNullLogs(t *testing.T) {
	// arrange
	m := newTestModel(t)
	m.tasks = []db.Task{{ID: 1, Title: "異常タスク"}}
	m.nullLogs = []db.TimeLog{
		{ID: 1, TaskID: 1, StartAt: time.Now()},
	}

	// act
	output := m.viewStats()

	// assert
	assert.Contains(t, output, "1件の異常なセッション")
	assert.Contains(t, output, "異常タスク")
}

func TestViewStats_WithTagSeconds(t *testing.T) {
	// arrange
	now := time.Now()
	end := now.Add(90 * time.Minute)
	m := newTestModel(t)
	m.statsTagPeriodIdx = 0
	logEntry := db.TimeLog{StartAt: now.UTC(), EndAt: &end}
	m.tasks = []db.Task{
		{
			ID:    1,
			Title: "タグ付きタスク",
			Tags:  []db.Tag{{Type: "project", Name: "mywork"}},
			TimeLogs: []db.TimeLog{logEntry},
		},
	}
	m.statsTasks = m.tasks

	// act
	output := m.viewStats()

	// assert
	assert.Contains(t, output, "プロジェクト/コンテキスト別計測時間")
	assert.Contains(t, output, "+mywork")
	assert.Contains(t, output, "h/l: 期間変更")
}

func TestViewStats_TagPeriodLabel(t *testing.T) {
	// arrange
	m := newTestModel(t)

	labels := []string{"今日", "今週", "今月", "半年", "1年"}
	for i, label := range labels {
		m.statsTagPeriodIdx = i

		// act
		output := m.viewStats()

		// assert
		assert.Contains(t, output, label, "期間インデックス%dのラベル", i)
	}
}
