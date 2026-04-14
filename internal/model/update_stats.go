package model

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) updateStats(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "j", "down":
		if m.statsCursor < len(m.nullLogs)-1 {
			m.statsCursor++
		}
	case "k", "up":
		if m.statsCursor > 0 {
			m.statsCursor--
		}
	case "h", "left":
		if m.statsTagPeriodIdx > 0 {
			m.statsTagPeriodIdx--
		}
	case "l", "right":
		if m.statsTagPeriodIdx < len(statsTagPeriods)-1 {
			m.statsTagPeriodIdx++
		}
	case "e":
		if len(m.nullLogs) == 0 {
			return m, nil
		}
		// 異常セッションに対応するタスクの詳細編集へ
		logTaskID := m.nullLogs[m.statsCursor].TaskID
		return m.initEditDetail(logTaskID)
	case "esc":
		m.mode = modeList
	}
	return m, nil
}
