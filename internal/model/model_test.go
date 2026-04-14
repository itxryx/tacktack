package model

import (
	"fmt"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
	"github.com/itxryx/tacktack/internal/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)


func newTestModel(t *testing.T) Model {
	t.Helper()
	database := setupTestDB(t)
	return New(database)
}

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	database, err := db.InitDB(":memory:")
	require.NoError(t, err)
	require.NoError(t, db.Migrate(database))
	return database
}

func TestModel_Quit(t *testing.T) {
	// arrange
	m := newTestModel(t)
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(80, 24))

	// act + assert
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	tm.WaitFinished(t, teatest.WithFinalTimeout(3*time.Second))
}

func TestModel_CtrlC(t *testing.T) {
	// arrange
	m := newTestModel(t)
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(80, 24))

	// act + assert
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlC})
	tm.WaitFinished(t, teatest.WithFinalTimeout(3*time.Second))
}

func TestClampCursor_Negative(t *testing.T) {
	// arrange
	m := newTestModel(t)
	createTestTask(t, m, "タスク")
	require.NoError(t, m.refreshTasks())
	m.cursor = -1

	// act
	m.clampCursor()

	// assert
	assert.Equal(t, 0, m.cursor)
}

func TestModel_WindowSize(t *testing.T) {
	// arrange
	m := newTestModel(t)

	// act
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m2 := updated.(Model)

	// assert
	assert.Equal(t, 120, m2.width)
	assert.Equal(t, 40, m2.height)
}

func TestModel_InitDone(t *testing.T) {
	// arrange
	m := newTestModel(t)

	// act
	m2, cmd := m.Update(initDoneMsg{
		tasks:     []db.Task{{Title: "テスト"}},
		activeLog: nil,
		nullLogs:  nil,
	})

	// assert
	assert.Nil(t, cmd)
	assert.Len(t, m2.(Model).tasks, 1)
}

func TestModel_InitDone_WithError(t *testing.T) {
	// arrange
	m := newTestModel(t)
	testErr := fmt.Errorf("DB初期化エラー")

	// act
	m2, cmd := m.Update(initDoneMsg{err: testErr})

	// assert
	assert.Nil(t, cmd, "エラー時は cmd=nil")
	assert.Equal(t, testErr, m2.(Model).lastErr, "lastErr にエラーがセットされること")
}

func TestModel_InitDone_WithActiveLog(t *testing.T) {
	// arrange
	m := newTestModel(t)
	activeLog := &db.TimeLog{ID: 1, TaskID: 1, StartAt: time.Now()}

	// act
	_, cmd := m.Update(initDoneMsg{
		tasks:     []db.Task{{Title: "テスト"}},
		activeLog: activeLog,
	})

	// assert
	assert.NotNil(t, cmd, "activeLog が非nil の場合 tickCmd が返ること")
}

func TestModel_TickMsg_WithActiveLog(t *testing.T) {
	// arrange
	m := newTestModel(t)
	m.activeLog = &db.TimeLog{ID: 1, TaskID: 1, StartAt: time.Now()}

	// act
	_, cmd := m.Update(tickMsg(time.Now()))

	// assert
	assert.NotNil(t, cmd, "計測中は tickCmd が返り続けること")
}

func TestModel_TickMsg_WithoutActiveLog(t *testing.T) {
	// arrange
	m := newTestModel(t)
	m.activeLog = nil

	// act
	_, cmd := m.Update(tickMsg(time.Now()))

	// assert
	assert.Nil(t, cmd, "計測なしの場合 tickCmd は返らないこと")
}
