package model

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/itxryx/tacktack/internal/domain"
)

func (m Model) updateTimeline(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	now := time.Now()
	oldestAllowed := now.AddDate(0, -1, 0) // 過去1ヶ月まで（m.tasks のデータ範囲）

	switch msg.String() {
	case "h", "left":
		prev := m.timelineDate.AddDate(0, 0, -1)
		if !prev.Before(truncateDay(oldestAllowed)) {
			m.timelineDate = prev
			m.timelineScroll = firstTaskScroll(m.tasks, m.timelineDate, m.activeLog, m.height)
		}
	case "l", "right":
		next := m.timelineDate.AddDate(0, 0, 1)
		if !truncateDay(next).After(truncateDay(now)) {
			m.timelineDate = next
			m.timelineScroll = firstTaskScroll(m.tasks, m.timelineDate, m.activeLog, m.height)
		}
	case "j", "down":
		slots := domain.BuildTimelineSlots(m.tasks, m.timelineDate, m.activeLog)
		if m.timelineScroll < timelineMaxScroll(countFlatRows(slots), m.height) {
			m.timelineScroll++
		}
	case "k", "up":
		if m.timelineScroll > 0 {
			m.timelineScroll--
		}
	case "t":
		m.timelineDate = now
		m.timelineScroll = firstTaskScroll(m.tasks, now, m.activeLog, m.height)
	case "esc":
		m.mode = modeList
	}
	return m, nil
}

// truncateDay は時刻を当日 0:00:00 ローカル時間に切り詰める。
func truncateDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
}
