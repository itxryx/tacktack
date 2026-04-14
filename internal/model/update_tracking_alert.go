package model

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) updateTrackingAlert(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "enter":
		// 現在計測中タスクを停止 → 選択タスクの計測を開始（アトミック）
		if m.activeLog == nil {
			m.mode = modeList
			return m, nil
		}
		if len(m.tasks) == 0 {
			m.mode = modeList
			return m, nil
		}
		currentTask := m.tasks[m.cursor]
		log, err := m.timeLogRepo.StopAndStart(m.activeLog.ID, currentTask.ID)
		if err != nil {
			m.lastErr = fmt.Errorf("計測切替エラー: %w", err)
			m.mode = modeList
			return m, nil
		}
		m.activeLog = log
		m.lastErr = nil
		if err := m.refreshTasks(); err != nil {
			m.lastErr = err
		}
		m.mode = modeList
		return m, tickCmd()

	case "n", "esc":
		m.mode = modeList
	}

	return m, nil
}
