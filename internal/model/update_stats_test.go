package model

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/itxryx/tacktack/internal/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateStats_Esc(t *testing.T) {
	// arrange
	m := newTestModel(t)
	m.mode = modeStats

	// act
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m2 := updated.(Model)

	// assert
	assert.Equal(t, modeList, m2.mode)
}

func TestUpdateStats_CursorMove(t *testing.T) {
	// arrange
	m := newTestModel(t)
	m.mode = modeStats
	m.nullLogs = []db.TimeLog{
		{ID: 1, TaskID: 1},
		{ID: 2, TaskID: 2},
	}

	// act + assert
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	m2 := updated.(Model)
	assert.Equal(t, 1, m2.statsCursor)

	updated2, _ := m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	m3 := updated2.(Model)
	assert.Equal(t, 1, m3.statsCursor, "末尾でクランプ")

	updated3, _ := m3.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	m4 := updated3.(Model)
	assert.Equal(t, 0, m4.statsCursor)

	updated4, _ := m4.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	m5 := updated4.(Model)
	assert.Equal(t, 0, m5.statsCursor, "先頭でクランプ")
}

func TestUpdateStats_EditFromStats(t *testing.T) {
	// arrange
	m := newTestModel(t)
	task := createTestTask(t, m, "異常タスク")
	require.NoError(t, m.refreshTasks())

	m.mode = modeStats
	m.nullLogs = []db.TimeLog{{ID: 1, TaskID: task.ID}}

	// act
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	m2 := updated.(Model)

	// assert
	assert.Equal(t, modeEditDetail, m2.mode)
	require.NotNil(t, m2.editTask)
	assert.Equal(t, task.ID, m2.editTask.ID)
}

func TestUpdateStats_EditFromStats_EmptyNullLogs(t *testing.T) {
	// arrange
	m := newTestModel(t)
	m.mode = modeStats

	// act
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	m2 := updated.(Model)

	// assert
	assert.Equal(t, modeStats, m2.mode)
}

func TestUpdateStats_PeriodSwitch(t *testing.T) {
	// arrange
	m := newTestModel(t)
	m.mode = modeStats
	m.statsTagPeriodIdx = 2

	// act + assert
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")})
	m2 := updated.(Model)
	assert.Equal(t, 3, m2.statsTagPeriodIdx, "l で期間インデックスが増加")

	m2.statsTagPeriodIdx = len(statsTagPeriods) - 1
	updated2, _ := m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")})
	m3 := updated2.(Model)
	assert.Equal(t, len(statsTagPeriods)-1, m3.statsTagPeriodIdx, "末尾でクランプ")

	m3.statsTagPeriodIdx = 2
	updated3, _ := m3.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")})
	m4 := updated3.(Model)
	assert.Equal(t, 1, m4.statsTagPeriodIdx, "h で期間インデックスが減少")

	m4.statsTagPeriodIdx = 0
	updated4, _ := m4.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")})
	m5 := updated4.(Model)
	assert.Equal(t, 0, m5.statsTagPeriodIdx, "先頭でクランプ")
}
