package model

import (
	"fmt"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/itxryx/tacktack/internal/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestViewContent_AllModes(t *testing.T) {
	// arrange
	m := newTestModel(t)
	createTestTask(t, m, "タスク")
	require.NoError(t, m.refreshTasks())

	modes := []mode{modeList, modeInput, modeTagSelect, modeDeleteConfirm,
		modeTrackingAlert, modeStats, modeTimeline}
	for _, mode := range modes {
		m.mode = mode

		// act
		output := m.viewContent()

		// assert
		assert.NotEmpty(t, output, "mode=%v", mode)
	}
}

func TestViewContent_ModeEditDetail(t *testing.T) {
	// arrange
	m := newTestModel(t)
	m2, _ := enterEditModeWithTask(t, m, "タスク")

	// act
	output := m2.viewContent()

	// assert
	assert.NotEmpty(t, output)
	assert.Contains(t, output, "タスク詳細編集")
}

func TestViewHeader_WithNullLogs(t *testing.T) {
	// arrange
	m := newTestModel(t)
	m.nullLogs = []db.TimeLog{{ID: 1}}

	// act
	header := m.viewHeader()

	// assert
	assert.Contains(t, header, "異常な計測セッション")
}

func TestViewHeader_WithoutNullLogs(t *testing.T) {
	// arrange
	m := newTestModel(t)

	// act
	header := m.viewHeader()

	// assert
	assert.Contains(t, header, "tacktack")
	assert.NotContains(t, header, "異常")
}

func TestViewFooter_AllModes(t *testing.T) {
	// arrange
	m := newTestModel(t)
	modes := []mode{modeList, modeInput, modeTagSelect, modeEditDetail,
		modeDeleteConfirm, modeTrackingAlert, modeStats, modeTimeline}
	for _, md := range modes {
		m.mode = md

		// act
		footer := m.viewFooter()

		// assert
		assert.NotEmpty(t, footer, "mode=%v", md)
	}
}

func TestViewFooter_List_SpaceHint_HiddenForCompleted(t *testing.T) {
	// arrange
	m := newTestModel(t)
	m.mode = modeList
	createTestTask(t, m, "完了タスク")
	require.NoError(t, m.refreshTasks())
	require.NoError(t, m.taskRepo.ToggleComplete(m.tasks[0].ID))
	require.NoError(t, m.refreshTasks())

	// act
	footer := m.viewFooter()

	// assert
	assert.NotContains(t, footer, "Space: 計測")
}

func TestViewFooter_List_SpaceHint_ShownForActive(t *testing.T) {
	// arrange
	m := newTestModel(t)
	m.mode = modeList
	createTestTask(t, m, "未完了タスク")
	require.NoError(t, m.refreshTasks())

	// act
	footer := m.viewFooter()

	// assert
	assert.Contains(t, footer, "Space: 計測")
}

func TestViewInput_NoTags(t *testing.T) {
	// arrange
	m := newTestModel(t)
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("i")})
	m2 := updated.(Model)

	// act
	output := m2.viewInput()

	// assert
	assert.Contains(t, output, "新しいタスクを追加")
	assert.Contains(t, output, "(なし)")
}

func TestViewInput_WithTags(t *testing.T) {
	// arrange
	m := newTestModel(t)
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("i")})
	m2 := updated.(Model)
	m2.selectedTags = []db.Tag{
		{Type: "project", Name: "work"},
		{Type: "context", Name: "home"},
	}

	// act
	output := m2.viewInput()

	// assert
	assert.Contains(t, output, "タグ")
	assert.Contains(t, output, "+work")
	assert.Contains(t, output, "@home")
}

func TestViewTagSelect_Project(t *testing.T) {
	// arrange
	m := newTestModel(t)
	tag, err := m.tagRepo.FindOrCreate("mywork", "project")
	require.NoError(t, err)
	m.tagList = []db.Tag{*tag}
	m.pendingTagType = "project"
	m.tagInput = ""
	m.tagCursor = 0

	// act
	output := m.viewTagSelect()

	// assert
	assert.Contains(t, output, "プロジェクトを選択してください")
	assert.Contains(t, output, "+mywork")
}

func TestViewTagSelect_Context(t *testing.T) {
	// arrange
	m := newTestModel(t)
	tag, err := m.tagRepo.FindOrCreate("office", "context")
	require.NoError(t, err)
	m.tagList = []db.Tag{*tag}
	m.pendingTagType = "context"
	m.tagInput = ""
	m.tagCursor = 0

	// act
	output := m.viewTagSelect()

	// assert
	assert.Contains(t, output, "@office")
}

func TestViewTagSelect_NewEntrySelected(t *testing.T) {
	// arrange
	m := newTestModel(t)
	m.tagList = []db.Tag{}
	m.pendingTagType = "project"
	m.tagInput = "newproject"
	m.tagCursor = 0

	// act
	output := m.viewTagSelect()

	// assert
	assert.Contains(t, output, "新規作成: newproject")
}

func TestViewEdit_NilTask(t *testing.T) {
	// arrange
	m := newTestModel(t)
	m.mode = modeEditDetail
	m.editTask = nil

	// act
	output := m.viewEdit()

	// assert
	assert.Contains(t, output, "読み込み中")
}

func TestViewEdit_WithTask(t *testing.T) {
	// arrange
	m := newTestModel(t)
	m2, _ := enterEditModeWithTask(t, m, "編集対象タスク")

	// act
	output := m2.viewEdit()

	// assert
	assert.Contains(t, output, "タスク詳細編集")
	assert.Contains(t, output, "タイトル")
	assert.Contains(t, output, "計測ログ")
}

func TestViewEdit_WithTimeLogs(t *testing.T) {
	// arrange
	m := newTestModel(t)
	createTestTask(t, m, "タスク")
	require.NoError(t, m.refreshTasks())
	createCompletedLog(t, m, m.tasks[0].ID)
	require.NoError(t, m.refreshTasks())
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	m2 := updated.(Model)

	// act
	output := m2.viewEdit()

	// assert
	assert.Contains(t, output, "合計:")
}

func TestViewEdit_WithNullEndAtLog(t *testing.T) {
	// arrange
	m := newTestModel(t)
	m2, _ := enterEditModeWithTask(t, m, "タスク")
	m2.editTask.TimeLogs = []db.TimeLog{
		{ID: 1, TaskID: m2.editTask.ID, StartAt: time.Now().Add(-time.Hour), EndAt: nil},
	}

	// act
	output := m2.viewEdit()

	// assert
	assert.Contains(t, output, "[未終了]")
}

func TestViewEdit_WithError(t *testing.T) {
	// arrange
	m := newTestModel(t)
	m2, _ := enterEditModeWithTask(t, m, "タスク")
	m2.lastErr = fmt.Errorf("テストエラー")

	// act
	output := m2.viewEdit()

	// assert
	assert.Contains(t, output, "テストエラー")
}

func TestViewEdit_TagFocused_WithTags(t *testing.T) {
	// arrange
	m := newTestModel(t)
	tag, err := m.tagRepo.FindOrCreate("work", "project")
	require.NoError(t, err)
	m2, _ := enterEditModeWithTask(t, m, "タスク")
	m2.selectedTags = []db.Tag{*tag}
	m2.editField = editFieldTags
	m2.editLogFocus = false

	// act
	output := m2.viewEdit()

	// assert
	assert.Contains(t, output, "+work")
	assert.Contains(t, output, "j/k: 移動")
}

func TestViewEdit_LogFocused(t *testing.T) {
	// arrange
	m := newTestModel(t)
	m2, _ := enterEditModeWithTask(t, m, "タスク")
	createCompletedLog(t, m, m2.editTask.ID)
	task, err := m.taskRepo.FindByID(m2.editTask.ID)
	require.NoError(t, err)
	m2.editTask = task
	m2.editLogFocus = true
	m2.editLogCursor = 0

	// act
	output := m2.viewEdit()

	// assert
	assert.Contains(t, output, "▶ ")
}

func TestViewFooter_EditDetail_LogFocus(t *testing.T) {
	// arrange
	m := newTestModel(t)
	m.mode = modeEditDetail
	m.editLogFocus = true
	m.editLogEditing = -1

	// act
	footer := m.viewFooter()

	// assert
	assert.Contains(t, footer, "e: 編集")
	assert.Contains(t, footer, "Shift+Tab")
}

func TestViewFooter_EditDetail_LogEditing(t *testing.T) {
	// arrange
	m := newTestModel(t)
	m.mode = modeEditDetail
	m.editLogFocus = true
	m.editLogEditing = 0

	// act
	footer := m.viewFooter()

	// assert
	assert.Contains(t, footer, "Tab: 項目切替")
	assert.Contains(t, footer, "Enter: 保存")
}

func TestViewFooter_WithLastErr(t *testing.T) {
	// arrange
	m := newTestModel(t)
	m.mode = modeList
	m.lastErr = fmt.Errorf("テストエラー")

	// act
	footer := m.viewFooter()

	// assert
	assert.Contains(t, footer, "テストエラー")
	assert.Contains(t, footer, "エラー:")
}

func TestViewTrackingAlert(t *testing.T) {
	// arrange
	m := newTestModel(t)
	createTestTask(t, m, "タスク1")
	createTestTask(t, m, "タスク2")
	require.NoError(t, m.refreshTasks())
	m.mode = modeTrackingAlert
	m.conflictTaskID = m.tasks[0].ID
	m.cursor = 1

	// act
	output := m.viewTrackingAlert()

	// assert
	assert.Contains(t, output, "別のタスクを計測中です")
	assert.Contains(t, output, "タスク1")
	assert.Contains(t, output, "タスク2")
}
