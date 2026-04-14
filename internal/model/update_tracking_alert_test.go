package model

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTwoTasksWithTracking(t *testing.T) (Model, uint) {
	t.Helper()
	m := newTestModel(t)
	createTestTask(t, m, "タスク1")
	createTestTask(t, m, "タスク2")
	require.NoError(t, m.refreshTasks())

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m2 := updated.(Model)
	require.NotNil(t, m2.activeLog)
	task1ID := m2.tasks[0].ID

	updated2, _ := m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	m3 := updated2.(Model)

	updated3, _ := m3.Update(tea.KeyMsg{Type: tea.KeySpace})
	m4 := updated3.(Model)
	require.Equal(t, modeTrackingAlert, m4.mode)

	return m4, task1ID
}

func TestTrackingAlert_SwitchTracking(t *testing.T) {
	// arrange
	m, _ := setupTwoTasksWithTracking(t)

	// act
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")})
	m2 := updated.(Model)

	// assert
	assert.Equal(t, modeList, m2.mode)
	assert.NotNil(t, m2.activeLog)
	assert.Equal(t, m2.tasks[1].ID, m2.activeLog.TaskID, "タスク2が計測中")
	assert.NotNil(t, cmd, "tickCmd が返されること")
	logs, err := m.timeLogRepo.FindNullEndAt()
	require.NoError(t, err)
	assert.Len(t, logs, 1, "アクティブセッションはタスク2の1件のみ")
}

func TestTrackingAlert_Cancel(t *testing.T) {
	// arrange
	m, _ := setupTwoTasksWithTracking(t)

	// act
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
	m2 := updated.(Model)

	// assert
	assert.Equal(t, modeList, m2.mode)
	assert.NotNil(t, m2.activeLog, "元の計測は継続中")
	assert.Equal(t, m2.tasks[0].ID, m2.activeLog.TaskID, "タスク1が計測中のまま")
}

func TestUpdateTrackingAlert_Enter(t *testing.T) {
	// arrange
	m, _ := setupTwoTasksWithTracking(t)

	// act
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m2 := updated.(Model)

	// assert
	assert.Equal(t, modeList, m2.mode)
	assert.NotNil(t, m2.activeLog, "タスク2の計測が開始されること")
	assert.NotNil(t, cmd, "tickCmd が返されること")
}

func TestUpdateTrackingAlert_Esc(t *testing.T) {
	// arrange
	m, _ := setupTwoTasksWithTracking(t)

	// act
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m2 := updated.(Model)

	// assert
	assert.Equal(t, modeList, m2.mode)
	assert.NotNil(t, m2.activeLog, "元の計測は継続中")
	assert.Equal(t, m2.tasks[0].ID, m2.activeLog.TaskID, "タスク1が計測中のまま")
}

func TestUpdateTrackingAlert_YWithNilActiveLog(t *testing.T) {
	// arrange
	m := newTestModel(t)
	m.mode = modeTrackingAlert
	m.activeLog = nil

	// act
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")})
	m2 := updated.(Model)

	// assert
	assert.Equal(t, modeList, m2.mode, "activeLog が nil でもパニックせず modeList に戻る")
}

func TestUpdateTrackingAlert_YWithEmptyTasks(t *testing.T) {
	// arrange
	m, _ := setupTwoTasksWithTracking(t)
	require.NotNil(t, m.activeLog, "activeLog は設定されていること")
	m.tasks = nil

	// act
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")})
	m2 := updated.(Model)

	// assert
	assert.Equal(t, modeList, m2.mode, "tasks が空でもパニックせず modeList に戻る")
}
