package model

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/itxryx/tacktack/internal/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func enterEditModeWithTask(t *testing.T, m Model, title string) (Model, db.Task) {
	t.Helper()
	task := createTestTask(t, m, title)
	require.NoError(t, m.refreshTasks())
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	return updated.(Model), task
}

func createCompletedLog(t *testing.T, m Model, taskID uint) {
	t.Helper()
	log, err := m.timeLogRepo.Start(taskID)
	require.NoError(t, err)
	require.NoError(t, m.timeLogRepo.Stop(log.ID))
}

func TestUpdateEdit_EmptyList_NoCrash(t *testing.T) {
	// arrange
	m := newTestModel(t)

	// act
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	m2 := updated.(Model)

	// assert
	assert.Equal(t, modeList, m2.mode)
}

func TestUpdateEdit_InitEditDetail(t *testing.T) {
	// arrange
	m := newTestModel(t)
	m2, task := enterEditModeWithTask(t, m, "テストタスク")

	// assert
	assert.Equal(t, modeEditDetail, m2.mode)
	require.NotNil(t, m2.editTask)
	assert.Equal(t, task.ID, m2.editTask.ID)
	assert.Equal(t, "テストタスク", m2.editInputs[editFieldTitle].Value())
	assert.Equal(t, editFieldTitle, m2.editField)
	assert.False(t, m2.editLogFocus)
}

func TestUpdateEdit_Esc_FromTaskArea(t *testing.T) {
	// arrange
	m := newTestModel(t)
	m2, _ := enterEditModeWithTask(t, m, "タスク")

	// act
	updated, _ := m2.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m3 := updated.(Model)

	// assert
	assert.Equal(t, modeList, m3.mode)
}

func TestUpdateEdit_SaveEditDetail(t *testing.T) {
	// arrange
	m := newTestModel(t)
	m2, task := enterEditModeWithTask(t, m, "元タイトル")
	m2.editInputs[editFieldTitle].SetValue("新タイトル")

	// act
	updated, _ := m2.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m3 := updated.(Model)

	// assert
	assert.Equal(t, modeList, m3.mode)
	assert.Nil(t, m3.lastErr)
	got, err := m.taskRepo.FindByID(task.ID)
	require.NoError(t, err)
	assert.Equal(t, "新タイトル", got.Title)
}

func TestUpdateEdit_SaveEditDetail_WithCtrlS(t *testing.T) {
	// arrange
	m := newTestModel(t)
	m2, task := enterEditModeWithTask(t, m, "元タイトル")
	m2.editInputs[editFieldTitle].SetValue("Ctrl+S保存")

	// act
	updated, _ := m2.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	m3 := updated.(Model)

	// assert
	assert.Equal(t, modeList, m3.mode)
	got, err := m.taskRepo.FindByID(task.ID)
	require.NoError(t, err)
	assert.Equal(t, "Ctrl+S保存", got.Title)
}

func TestUpdateEdit_SaveEditDetail_EmptyTitle(t *testing.T) {
	// arrange
	m := newTestModel(t)
	m2, _ := enterEditModeWithTask(t, m, "タスク")
	m2.editInputs[editFieldTitle].SetValue("")

	// act
	updated, _ := m2.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m3 := updated.(Model)

	// assert
	assert.Equal(t, modeEditDetail, m3.mode, "エラーで留まること")
	assert.NotNil(t, m3.lastErr)
}

func TestUpdateEdit_SaveEditDetail_InvalidPriority(t *testing.T) {
	// arrange
	m := newTestModel(t)
	m2, _ := enterEditModeWithTask(t, m, "タスク")
	m2.editInputs[editFieldPriority].SetValue("1")

	// act
	updated, _ := m2.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m3 := updated.(Model)

	// assert
	assert.Equal(t, modeEditDetail, m3.mode)
	assert.NotNil(t, m3.lastErr)
}

func TestUpdateEdit_SaveEditDetail_InvalidDueDate(t *testing.T) {
	// arrange
	m := newTestModel(t)
	m2, _ := enterEditModeWithTask(t, m, "タスク")
	m2.editInputs[editFieldDueDate].SetValue("not-a-date")

	// act
	updated, _ := m2.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m3 := updated.(Model)

	// assert
	assert.Equal(t, modeEditDetail, m3.mode)
	assert.NotNil(t, m3.lastErr)
}

func TestUpdateEdit_ToggleComplete(t *testing.T) {
	// arrange
	m := newTestModel(t)
	m2, _ := enterEditModeWithTask(t, m, "タスク")
	m2.editField = editFieldComplete

	// act
	updated, _ := m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" ")})
	m3 := updated.(Model)

	// assert
	assert.True(t, m3.editTask.IsCompleted)

	// act
	updated2, _ := m3.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" ")})
	m4 := updated2.(Model)

	// assert
	assert.False(t, m4.editTask.IsCompleted)
}

func TestUpdateEdit_TabFieldNavigation(t *testing.T) {
	// arrange
	m := newTestModel(t)
	m2, _ := enterEditModeWithTask(t, m, "タスク")

	// act
	updated, _ := m2.Update(tea.KeyMsg{Type: tea.KeyTab})
	m3 := updated.(Model)

	// assert
	assert.Equal(t, editFieldPriority, m3.editField)

	// act
	updated2, _ := m3.Update(tea.KeyMsg{Type: tea.KeyTab})
	m4 := updated2.(Model)

	// assert
	assert.Equal(t, editFieldDueDate, m4.editField)

	// act
	updated3, _ := m4.Update(tea.KeyMsg{Type: tea.KeyTab})
	m5 := updated3.(Model)

	// assert
	assert.Equal(t, editFieldTags, m5.editField)

	// act
	updated4, _ := m5.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	m6 := updated4.(Model)

	// assert
	assert.Equal(t, editFieldDueDate, m6.editField)
}

func TestUpdateEdit_TagUnlink(t *testing.T) {
	// arrange
	m := newTestModel(t)
	tag, err := m.tagRepo.FindOrCreate("work", "project")
	require.NoError(t, err)
	task := createTestTask(t, m, "タグ付きタスク")
	require.NoError(t, m.taskRepo.ReplaceTagsForTask(task.ID, []db.Tag{*tag}))
	require.NoError(t, m.refreshTasks())
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	m2 := updated.(Model)
	require.Len(t, m2.selectedTags, 1, "タグが1件あること")
	m2.editField = editFieldTags

	// act
	updated2, _ := m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")})
	m3 := updated2.(Model)

	// assert
	assert.Empty(t, m3.selectedTags, "タグが削除されていること")

	// act
	updated3, _ := m3.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m4 := updated3.(Model)

	// assert
	assert.Equal(t, modeList, m4.mode)
	got, err := m.taskRepo.FindByID(task.ID)
	require.NoError(t, err)
	assert.Empty(t, got.Tags, "DBのタグ紐付けが解除されていること")
}

func TestUpdateEdit_TagUnlink_CursorMove(t *testing.T) {
	// arrange
	m := newTestModel(t)
	tag1, err := m.tagRepo.FindOrCreate("work", "project")
	require.NoError(t, err)
	tag2, err := m.tagRepo.FindOrCreate("home", "project")
	require.NoError(t, err)
	task := createTestTask(t, m, "タスク")
	require.NoError(t, m.taskRepo.ReplaceTagsForTask(task.ID, []db.Tag{*tag1, *tag2}))
	require.NoError(t, m.refreshTasks())
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	m2 := updated.(Model)
	m2.editField = editFieldTags

	// act
	updated2, _ := m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	m3 := updated2.(Model)

	// assert
	assert.Equal(t, 1, m3.editTagCursor)

	// act
	updated3, _ := m3.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")})
	m4 := updated3.(Model)

	// assert
	require.Len(t, m4.selectedTags, 1)
	assert.Equal(t, "work", m4.selectedTags[0].Name)
	assert.Equal(t, 0, m4.editTagCursor)
}

func TestUpdateEdit_TagUnlink_Empty_NoCrash(t *testing.T) {
	// arrange
	m := newTestModel(t)
	m2, _ := enterEditModeWithTask(t, m, "タスク")
	m2.editField = editFieldTags

	// act
	updated, _ := m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")})
	m3 := updated.(Model)

	// assert
	assert.Empty(t, m3.selectedTags)
}

func TestUpdateEdit_TagSelection_Project(t *testing.T) {
	// arrange
	m := newTestModel(t)
	m2, _ := enterEditModeWithTask(t, m, "タスク")
	m2.editField = editFieldTags

	// act
	updated, _ := m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("+")})
	m3 := updated.(Model)

	// assert
	assert.Equal(t, modeTagSelect, m3.mode)
	assert.Equal(t, "project", m3.pendingTagType)
	assert.Equal(t, modeEditDetail, m3.tagSelectPrevMode)
}

func TestUpdateEdit_TagSelection_Context(t *testing.T) {
	// arrange
	m := newTestModel(t)
	m2, _ := enterEditModeWithTask(t, m, "タスク")
	m2.editField = editFieldTags

	// act
	updated, _ := m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("@")})
	m3 := updated.(Model)

	// assert
	assert.Equal(t, modeTagSelect, m3.mode)
	assert.Equal(t, "context", m3.pendingTagType)
}

func TestUpdateEdit_LogArea_ShiftTab(t *testing.T) {
	// arrange
	m := newTestModel(t)
	m2, _ := enterEditModeWithTask(t, m, "タスク")
	m2.editLogFocus = true

	// act
	updated, _ := m2.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	m3 := updated.(Model)

	// assert
	assert.False(t, m3.editLogFocus)
	assert.Equal(t, editFieldComplete, m3.editField)
}

func TestUpdateEdit_LogArea_Esc(t *testing.T) {
	// arrange
	m := newTestModel(t)
	m2, _ := enterEditModeWithTask(t, m, "タスク")
	m2.editLogFocus = true
	m2.editLogEditing = -1

	// act
	updated, _ := m2.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m3 := updated.(Model)

	// assert
	assert.Equal(t, modeList, m3.mode)
}

func TestUpdateEdit_LogArea_EmptyLogs_NoCrash(t *testing.T) {
	// arrange
	m := newTestModel(t)
	m2, _ := enterEditModeWithTask(t, m, "タスク")
	m2.editLogFocus = true

	// act
	updated, _ := m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	m3 := updated.(Model)

	// assert
	assert.Equal(t, -1, m3.editLogEditing)

	// act
	updated2, _ := m3.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")})
	m4 := updated2.(Model)

	// assert
	assert.Equal(t, modeEditDetail, m4.mode)
}

func TestUpdateEdit_LogArea_CursorMove(t *testing.T) {
	// arrange
	m := newTestModel(t)
	createTestTask(t, m, "タスク")
	require.NoError(t, m.refreshTasks())
	createCompletedLog(t, m, m.tasks[0].ID)
	createCompletedLog(t, m, m.tasks[0].ID)
	require.NoError(t, m.refreshTasks())
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	m2 := updated.(Model)
	m2.editLogFocus = true

	// act
	updated2, _ := m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	m3 := updated2.(Model)

	// assert
	assert.Equal(t, 1, m3.editLogCursor)

	// act
	updated3, _ := m3.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	m4 := updated3.(Model)

	// assert
	assert.Equal(t, 1, m4.editLogCursor, "末尾でクランプ")

	// act
	updated4, _ := m4.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	m5 := updated4.(Model)

	// assert
	assert.Equal(t, 0, m5.editLogCursor)

	// act
	updated5, _ := m5.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	m6 := updated5.(Model)

	// assert
	assert.Equal(t, 0, m6.editLogCursor, "先頭でクランプ")
}

func TestUpdateEdit_LogArea_DeleteLog(t *testing.T) {
	// arrange
	m := newTestModel(t)
	createTestTask(t, m, "タスク")
	require.NoError(t, m.refreshTasks())
	createCompletedLog(t, m, m.tasks[0].ID)
	require.NoError(t, m.refreshTasks())
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	m2 := updated.(Model)
	require.Len(t, m2.editTask.TimeLogs, 1)
	m2.editLogFocus = true

	// act
	updated2, _ := m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")})
	m3 := updated2.(Model)

	// assert
	assert.Empty(t, m3.editTask.TimeLogs, "ログが削除されていること")
	assert.Nil(t, m3.lastErr)
}

func TestUpdateEdit_LogArea_EditLog_Enter(t *testing.T) {
	// arrange
	m := newTestModel(t)
	createTestTask(t, m, "タスク")
	require.NoError(t, m.refreshTasks())
	createCompletedLog(t, m, m.tasks[0].ID)
	require.NoError(t, m.refreshTasks())
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	m2 := updated.(Model)
	m2.editLogFocus = true
	updated2, _ := m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	m3 := updated2.(Model)
	assert.Equal(t, 0, m3.editLogEditing)
	require.Len(t, m3.editLogInputs, 2)
	m3.editLogInputs[0].SetValue("2026-01-01 09:00:00")
	m3.editLogInputs[1].SetValue("2026-01-01 10:30:00")

	// act
	updated3, _ := m3.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m4 := updated3.(Model)

	// assert
	assert.Equal(t, -1, m4.editLogEditing, "編集完了後は -1 に戻ること")
	assert.Nil(t, m4.lastErr)
}

func TestUpdateEdit_LogArea_EditLog_EmptyEnd(t *testing.T) {
	// arrange
	m := newTestModel(t)
	createTestTask(t, m, "タスク")
	require.NoError(t, m.refreshTasks())
	createCompletedLog(t, m, m.tasks[0].ID)
	require.NoError(t, m.refreshTasks())
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	m2 := updated.(Model)
	m2.editLogFocus = true
	updated2, _ := m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	m3 := updated2.(Model)
	m3.editLogInputs[0].SetValue("2026-01-01 09:00:00")
	m3.editLogInputs[1].SetValue("")

	// act
	updated3, _ := m3.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m4 := updated3.(Model)

	// assert
	assert.Equal(t, -1, m4.editLogEditing)
	assert.Nil(t, m4.lastErr)
}

func TestUpdateEdit_LogArea_EditLog_InvalidStartAt(t *testing.T) {
	// arrange
	m := newTestModel(t)
	createTestTask(t, m, "タスク")
	require.NoError(t, m.refreshTasks())
	createCompletedLog(t, m, m.tasks[0].ID)
	require.NoError(t, m.refreshTasks())
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	m2 := updated.(Model)
	m2.editLogFocus = true
	updated2, _ := m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	m3 := updated2.(Model)
	m3.editLogInputs[0].SetValue("invalid-time")

	// act
	updated3, _ := m3.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m4 := updated3.(Model)

	// assert
	assert.NotNil(t, m4.lastErr)
	assert.Equal(t, 0, m4.editLogEditing, "エラー後は editing のまま")
}

func TestUpdateEdit_LogArea_EditLog_InvalidEndAt(t *testing.T) {
	// arrange
	m := newTestModel(t)
	createTestTask(t, m, "タスク")
	require.NoError(t, m.refreshTasks())
	createCompletedLog(t, m, m.tasks[0].ID)
	require.NoError(t, m.refreshTasks())
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	m2 := updated.(Model)
	m2.editLogFocus = true
	updated2, _ := m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	m3 := updated2.(Model)
	m3.editLogInputs[0].SetValue("2026-01-01 09:00:00")
	m3.editLogInputs[1].SetValue("not-a-time")

	// act
	updated3, _ := m3.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m4 := updated3.(Model)

	// assert
	assert.NotNil(t, m4.lastErr)
}

func TestUpdateEdit_LogArea_EditLog_EndBeforeStart(t *testing.T) {
	// arrange
	m := newTestModel(t)
	createTestTask(t, m, "タスク")
	require.NoError(t, m.refreshTasks())
	createCompletedLog(t, m, m.tasks[0].ID)
	require.NoError(t, m.refreshTasks())
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	m2 := updated.(Model)
	m2.editLogFocus = true
	updated2, _ := m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	m3 := updated2.(Model)
	m3.editLogInputs[0].SetValue("2026-01-01 10:00:00")
	m3.editLogInputs[1].SetValue("2026-01-01 09:00:00")

	// act
	updated3, _ := m3.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m4 := updated3.(Model)

	// assert
	assert.NotNil(t, m4.lastErr)
}

func TestUpdateEdit_LogArea_EditLog_Esc(t *testing.T) {
	// arrange
	m := newTestModel(t)
	createTestTask(t, m, "タスク")
	require.NoError(t, m.refreshTasks())
	createCompletedLog(t, m, m.tasks[0].ID)
	require.NoError(t, m.refreshTasks())
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	m2 := updated.(Model)
	m2.editLogFocus = true
	updated2, _ := m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	m3 := updated2.(Model)
	assert.Equal(t, 0, m3.editLogEditing)

	// act
	updated3, _ := m3.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m4 := updated3.(Model)

	// assert
	assert.Equal(t, -1, m4.editLogEditing, "Escでキャンセル")
	assert.Equal(t, modeEditDetail, m4.mode, "modeはEditDetailのまま")
}

func TestUpdateEdit_ToggleComplete_WithActiveLog_SetsStoppedLog(t *testing.T) {
	// arrange
	m := newTestModel(t)
	m2, _ := enterEditModeWithTask(t, m, "タスク")
	log, err := m.timeLogRepo.Start(m2.editTask.ID)
	require.NoError(t, err)
	m2.activeLog = log
	m2.editField = editFieldComplete

	// act
	updated, _ := m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" ")})
	m3 := updated.(Model)

	// assert
	assert.True(t, m3.editTask.IsCompleted)
	assert.NotNil(t, m3.editStoppedLog, "計測中タスクの場合は editStoppedLog が設定されること")
	assert.Equal(t, log.ID, m3.editStoppedLog.ID, "アクティブなログが editStoppedLog に設定されること")

	// act
	updated2, _ := m3.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" ")})
	m4 := updated2.(Model)

	// assert
	assert.False(t, m4.editTask.IsCompleted)
	assert.Nil(t, m4.editStoppedLog, "未完了に戻した場合は editStoppedLog がクリアされること")
}

func TestUpdateEdit_SaveEditDetail_WithStoppedLog(t *testing.T) {
	// arrange
	m := newTestModel(t)
	m2, _ := enterEditModeWithTask(t, m, "タスク")
	log, err := m.timeLogRepo.Start(m2.editTask.ID)
	require.NoError(t, err)
	m2.activeLog = log
	m2.editField = editFieldComplete
	updated, _ := m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" ")})
	m3 := updated.(Model)
	require.NotNil(t, m3.editStoppedLog, "editStoppedLog が設定されていること")

	// act
	updated2, _ := m3.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m4 := updated2.(Model)

	// assert
	assert.Equal(t, modeList, m4.mode)
	assert.Nil(t, m4.activeLog, "計測が停止されること")
	assert.Nil(t, m4.editStoppedLog, "editStoppedLog がクリアされること")
	nullLogs, err := m.timeLogRepo.FindNullEndAt()
	require.NoError(t, err)
	assert.Empty(t, nullLogs, "アクティブなログがなくなること")
}

func TestUpdateEdit_LogArea_EditLog_DefaultInput(t *testing.T) {
	// arrange
	m := newTestModel(t)
	createTestTask(t, m, "タスク")
	require.NoError(t, m.refreshTasks())
	createCompletedLog(t, m, m.tasks[0].ID)
	require.NoError(t, m.refreshTasks())
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	m2 := updated.(Model)
	m2.editLogFocus = true
	updated2, _ := m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	m3 := updated2.(Model)
	require.Equal(t, 0, m3.editLogEditing)
	m3.editLogInputs[0].SetValue("")

	// act
	updated3, _ := m3.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("2")})
	m4 := updated3.(Model)

	// assert
	assert.Equal(t, 0, m4.editLogEditing, "編集継続中であること")
	assert.Contains(t, m4.editLogInputs[0].Value(), "2", "入力が転送されること")
}

func TestUpdateEdit_LogArea_EditLog_NilEndConflict(t *testing.T) {
	// arrange
	m := newTestModel(t)
	createTestTask(t, m, "タスク1")
	createTestTask(t, m, "タスク2")
	require.NoError(t, m.refreshTasks())
	createCompletedLog(t, m, m.tasks[0].ID)
	require.NoError(t, m.refreshTasks())
	activeLog, err := m.timeLogRepo.Start(m.tasks[1].ID)
	require.NoError(t, err)
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	m2 := updated.(Model)
	m2.activeLog = activeLog
	m2.editLogFocus = true
	updated2, _ := m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	m3 := updated2.(Model)
	require.Equal(t, 0, m3.editLogEditing)
	m3.editLogInputs[0].SetValue("2026-01-01 09:00:00")
	m3.editLogInputs[1].SetValue("")

	// act
	updated3, _ := m3.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m4 := updated3.(Model)

	// assert
	assert.NotNil(t, m4.lastErr, "別タスク計測中なのでエラーになること")
	assert.Equal(t, 0, m4.editLogEditing, "エラー後は editing のまま")
}

func TestUpdateEdit_LogArea_EditLog_Tab(t *testing.T) {
	// arrange
	m := newTestModel(t)
	createTestTask(t, m, "タスク")
	require.NoError(t, m.refreshTasks())
	createCompletedLog(t, m, m.tasks[0].ID)
	require.NoError(t, m.refreshTasks())
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	m2 := updated.(Model)
	m2.editLogFocus = true
	updated2, _ := m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	m3 := updated2.(Model)
	assert.True(t, m3.editLogInputs[0].Focused())

	// act
	updated3, _ := m3.Update(tea.KeyMsg{Type: tea.KeyTab})
	m4 := updated3.(Model)

	// assert
	assert.False(t, m4.editLogInputs[0].Focused())
	assert.True(t, m4.editLogInputs[1].Focused())

	// act
	updated4, _ := m4.Update(tea.KeyMsg{Type: tea.KeyTab})
	m5 := updated4.(Model)

	// assert
	assert.True(t, m5.editLogInputs[0].Focused())
}
