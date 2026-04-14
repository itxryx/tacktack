package model

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/itxryx/tacktack/internal/db"
	"github.com/itxryx/tacktack/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestTask(t *testing.T, m Model, title string) db.Task {
	t.Helper()
	task := &db.Task{Title: title}
	require.NoError(t, m.taskRepo.Create(task))
	return *task
}

func TestUpdateList_CursorMove(t *testing.T) {
	// arrange
	m := newTestModel(t)
	createTestTask(t, m, "タスク1")
	createTestTask(t, m, "タスク2")
	createTestTask(t, m, "タスク3")
	require.NoError(t, m.refreshTasks())

	// act + assert
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	m2 := updated.(Model)
	assert.Equal(t, 1, m2.cursor)

	updated, _ = m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	updated, _ = updated.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	m3 := updated.(Model)
	assert.Equal(t, 2, m3.cursor, "末尾でクランプ")

	updated, _ = m3.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	m4 := updated.(Model)
	assert.Equal(t, 1, m4.cursor)

	updated, _ = m4.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	updated, _ = updated.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	m5 := updated.(Model)
	assert.Equal(t, 0, m5.cursor, "先頭でクランプ")
}

func TestUpdateList_CursorMove_EmptyList(t *testing.T) {
	// arrange
	m := newTestModel(t)

	// act
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})

	// assert
	assert.NotPanics(t, func() {
		_ = updated.(Model)
	})
}

func TestUpdateList_ToggleComplete(t *testing.T) {
	// arrange
	m := newTestModel(t)
	task := createTestTask(t, m, "タスク")
	require.NoError(t, m.refreshTasks())

	// act
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	_ = updated

	// assert
	got, err := m.taskRepo.FindByID(task.ID)
	require.NoError(t, err)
	assert.True(t, got.IsCompleted)
	assert.NotNil(t, got.CompletedAt)
}

func TestUpdateList_DeleteModal(t *testing.T) {
	// arrange
	m := newTestModel(t)
	createTestTask(t, m, "タスク")
	require.NoError(t, m.refreshTasks())

	// act
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")})
	m2 := updated.(Model)

	// assert
	assert.Equal(t, modeDeleteConfirm, m2.mode)
	view := m2.viewContent()
	assert.Contains(t, view, "タスクを削除しますか")
}

func TestUpdateList_DeleteConfirm_Yes(t *testing.T) {
	// arrange
	m := newTestModel(t)
	task := createTestTask(t, m, "削除タスク")
	require.NoError(t, m.refreshTasks())

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")})
	m2 := updated.(Model)

	// act
	updated2, _ := m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")})
	m3 := updated2.(Model)

	// assert
	assert.Equal(t, modeList, m3.mode)
	_, err := m.taskRepo.FindByID(task.ID)
	assert.Error(t, err, "削除されていること")
}

func TestUpdateList_DeleteConfirm_No(t *testing.T) {
	// arrange
	m := newTestModel(t)
	task := createTestTask(t, m, "削除タスク")
	require.NoError(t, m.refreshTasks())

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")})
	m2 := updated.(Model)

	// act
	updated2, _ := m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
	m3 := updated2.(Model)

	// assert
	assert.Equal(t, modeList, m3.mode)
	_, err := m.taskRepo.FindByID(task.ID)
	assert.NoError(t, err, "削除されていないこと")
}

func TestUpdateList_EmptyList_NoCrashOnDelete(t *testing.T) {
	// arrange
	m := newTestModel(t)

	// act
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")})
	m2 := updated.(Model)

	// assert
	assert.Equal(t, modeList, m2.mode, "空リストで削除モードにならないこと")
}

func TestUpdateList_TrackingStart(t *testing.T) {
	// arrange
	m := newTestModel(t)
	createTestTask(t, m, "タスク")
	require.NoError(t, m.refreshTasks())

	// act
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m2 := updated.(Model)

	// assert
	assert.NotNil(t, m2.activeLog)
	assert.NotNil(t, cmd, "tickCmd が返されること")
	logs, err := m.timeLogRepo.FindNullEndAt()
	require.NoError(t, err)
	assert.Len(t, logs, 1)
}

func TestUpdateList_TrackingStop(t *testing.T) {
	// arrange
	m := newTestModel(t)
	createTestTask(t, m, "タスク")
	require.NoError(t, m.refreshTasks())

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m2 := updated.(Model)

	// act
	updated2, cmd := m2.Update(tea.KeyMsg{Type: tea.KeySpace})
	m3 := updated2.(Model)

	// assert
	assert.Nil(t, m3.activeLog)
	assert.Nil(t, cmd, "tickCmd が止まること")
	logs, err := m.timeLogRepo.FindNullEndAt()
	require.NoError(t, err)
	assert.Empty(t, logs)
}

func TestUpdateDeleteConfirm_Enter(t *testing.T) {
	// arrange
	m := newTestModel(t)
	task := createTestTask(t, m, "削除タスク")
	require.NoError(t, m.refreshTasks())

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")})
	m2 := updated.(Model)

	// act
	updated2, _ := m2.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m3 := updated2.(Model)

	// assert
	assert.Equal(t, modeList, m3.mode)
	_, err := m.taskRepo.FindByID(task.ID)
	assert.Error(t, err, "削除されていること")
}

func TestUpdateList_ToggleComplete_WithActiveLog(t *testing.T) {
	// arrange
	m := newTestModel(t)
	createTestTask(t, m, "タスク")
	require.NoError(t, m.refreshTasks())

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m2 := updated.(Model)
	require.NotNil(t, m2.activeLog)

	// act
	updated2, _ := m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	m3 := updated2.(Model)

	// assert
	assert.Nil(t, m3.activeLog, "計測が停止されること")
	got, err := m.taskRepo.FindByID(m2.tasks[0].ID)
	require.NoError(t, err)
	assert.True(t, got.IsCompleted)
}

func TestUpdateList_TrackingCompletedTask_Noop(t *testing.T) {
	// arrange
	m := newTestModel(t)
	createTestTask(t, m, "完了タスク")
	require.NoError(t, m.refreshTasks())

	require.NoError(t, m.taskRepo.ToggleComplete(m.tasks[0].ID))
	require.NoError(t, m.refreshTasks())

	// act
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m2 := updated.(Model)

	// assert
	assert.Nil(t, m2.activeLog, "完了済みタスクは計測開始されない")
}

func TestUpdateList_Space_EmptyList(t *testing.T) {
	// arrange
	m := newTestModel(t)

	// act
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m2 := updated.(Model)

	// assert
	assert.Nil(t, m2.activeLog, "空リストでは計測開始されない")
	assert.Nil(t, cmd, "tickCmdが返されないこと")
}

func TestUpdateDeleteConfirm_WithActiveLog(t *testing.T) {
	// arrange
	m := newTestModel(t)
	task := createTestTask(t, m, "タスク")
	require.NoError(t, m.refreshTasks())

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m2 := updated.(Model)
	require.NotNil(t, m2.activeLog)

	updated2, _ := m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")})
	m3 := updated2.(Model)
	require.Equal(t, modeDeleteConfirm, m3.mode)

	// act
	updated3, _ := m3.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")})
	m4 := updated3.(Model)

	// assert
	assert.Equal(t, modeList, m4.mode)
	assert.Nil(t, m4.activeLog, "計測が停止されること")
	_, err := m.taskRepo.FindByID(task.ID)
	assert.Error(t, err, "タスクが削除されていること")
}

func TestUpdateList_AKey_EntersInputMode(t *testing.T) {
	// arrange
	m := newTestModel(t)

	// act
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	m2 := updated.(Model)

	// assert
	assert.Equal(t, modeInput, m2.mode, `"a" キーで入力モードに遷移すること`)
}

func TestUpdateList_ArrowKeys(t *testing.T) {
	// arrange
	m := newTestModel(t)
	createTestTask(t, m, "タスク1")
	createTestTask(t, m, "タスク2")
	require.NoError(t, m.refreshTasks())

	// act + assert
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m2 := updated.(Model)
	assert.Equal(t, 1, m2.cursor, "↓キーでカーソルが下に移動すること")

	updated2, _ := m2.Update(tea.KeyMsg{Type: tea.KeyUp})
	m3 := updated2.(Model)
	assert.Equal(t, 0, m3.cursor, "↑キーでカーソルが上に移動すること")
}

func TestUpdateDeleteConfirm_Esc(t *testing.T) {
	// arrange
	m := newTestModel(t)
	createTestTask(t, m, "削除しないタスク")
	require.NoError(t, m.refreshTasks())

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")})
	m2 := updated.(Model)
	assert.Equal(t, modeDeleteConfirm, m2.mode)

	// act
	updated2, _ := m2.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m3 := updated2.(Model)

	// assert
	assert.Equal(t, modeList, m3.mode)
	require.NoError(t, m3.refreshTasks())
	assert.Len(t, m3.tasks, 1, "タスクが削除されていないこと")
}

func TestUpdateList_TrackingConflict(t *testing.T) {
	// arrange
	m := newTestModel(t)
	createTestTask(t, m, "タスク1")
	createTestTask(t, m, "タスク2")

	taskRepo := repository.NewTaskRepository(setupTestDB(t))
	_ = taskRepo
	require.NoError(t, m.refreshTasks())

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m2 := updated.(Model)
	assert.NotNil(t, m2.activeLog)

	updated2, _ := m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	m3 := updated2.(Model)

	// act
	updated3, _ := m3.Update(tea.KeyMsg{Type: tea.KeySpace})
	m4 := updated3.(Model)

	// assert
	assert.Equal(t, modeTrackingAlert, m4.mode)
}
