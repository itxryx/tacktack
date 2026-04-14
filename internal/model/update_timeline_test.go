package model

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestUpdateTimeline_Esc(t *testing.T) {
	// arrange
	m := newTestModel(t)
	m.mode = modeTimeline
	m.timelineDate = time.Now()

	// act
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m2 := updated.(Model)

	// assert
	assert.Equal(t, modeList, m2.mode)
}

func TestUpdateTimeline_TodayReset(t *testing.T) {
	// arrange
	m := newTestModel(t)
	m.mode = modeTimeline
	m.height = 24
	m.timelineDate = time.Now().AddDate(0, 0, -5)
	m.timelineScroll = 10

	// act
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("t")})
	m2 := updated.(Model)

	// assert
	assert.Equal(t, 24, m2.timelineScroll, "t でタスクなし日は12:00にスクロール")
	assert.True(t, truncateDay(m2.timelineDate).Equal(truncateDay(time.Now())), "t で今日に戻る")
}

func TestUpdateTimeline_Scroll(t *testing.T) {
	// arrange
	m := newTestModel(t)
	m.mode = modeTimeline
	m.height = 24
	m.timelineDate = time.Now()
	m.timelineScroll = 0

	// act + assert
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	m2 := updated.(Model)
	assert.Equal(t, 1, m2.timelineScroll)

	updated2, _ := m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	m3 := updated2.(Model)
	assert.Equal(t, 0, m3.timelineScroll)

	updated3, _ := m3.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	m4 := updated3.(Model)
	assert.Equal(t, 0, m4.timelineScroll, "先頭でクランプ")

	m5 := m4
	m5.timelineScroll = timelineMaxScroll(48, m5.height)
	updated5, _ := m5.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	m6 := updated5.(Model)
	assert.Equal(t, timelineMaxScroll(48, m5.height), m6.timelineScroll, "末尾でクランプ")
}

func TestUpdateTimeline_DateMove(t *testing.T) {
	// arrange
	m := newTestModel(t)
	m.mode = modeTimeline
	m.timelineDate = time.Now()

	// act + assert
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")})
	m2 := updated.(Model)
	assert.True(t, truncateDay(m2.timelineDate).Equal(truncateDay(time.Now())), "今日が上限")

	updated2, _ := m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")})
	m3 := updated2.(Model)
	assert.True(t, truncateDay(m3.timelineDate).Equal(truncateDay(time.Now().AddDate(0, 0, -1))), "前日に移動")

	updated3, _ := m3.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")})
	m4 := updated3.(Model)
	assert.True(t, truncateDay(m4.timelineDate).Equal(truncateDay(time.Now())), "今日に戻る")
}

func TestUpdateTimeline_DateLimits_OldestBound(t *testing.T) {
	// arrange
	m := newTestModel(t)
	m.mode = modeTimeline
	m.timelineDate = time.Now().AddDate(0, -1, 0)

	// act
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")})
	m2 := updated.(Model)

	// assert
	limit := time.Now().AddDate(0, -1, 0)
	assert.False(t, m2.timelineDate.Before(truncateDay(limit)), "1ヶ月以上前には移動できない")
}

func TestModel_TabCycle(t *testing.T) {
	// arrange
	m := newTestModel(t)
	assert.Equal(t, modeList, m.mode)

	// act + assert
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m2 := updated.(Model)
	assert.Equal(t, modeTimeline, m2.mode)

	updated2, _ := m2.Update(tea.KeyMsg{Type: tea.KeyTab})
	m3 := updated2.(Model)
	assert.Equal(t, modeStats, m3.mode)

	updated3, _ := m3.Update(tea.KeyMsg{Type: tea.KeyTab})
	m4 := updated3.(Model)
	assert.Equal(t, modeList, m4.mode)
}
