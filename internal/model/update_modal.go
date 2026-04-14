package model

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) updateDeleteConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "enter":
		// C1: 削除対象が計測中タスクの場合は停止と削除をアトミックに実行する
		if m.activeLog != nil && m.activeLog.TaskID == m.deleteTargetID {
			if err := m.taskRepo.StopAndDelete(m.activeLog.ID, m.deleteTargetID); err != nil {
				m.lastErr = fmt.Errorf("削除エラー: %w", err)
				m.mode = modeList
				return m, nil
			}
			m.activeLog = nil
		} else {
			if err := m.taskRepo.Delete(m.deleteTargetID); err != nil {
				m.lastErr = fmt.Errorf("削除エラー: %w", err)
				m.mode = modeList
				return m, nil
			}
		}
		m.lastErr = nil
		if err := m.refreshTasks(); err != nil {
			m.lastErr = err
		}
		// C7: 削除タスクに属する nullLogs を除去する
		if err := m.refreshNullLogs(); err != nil && m.lastErr == nil {
			m.lastErr = err
		}
		m.clampCursor()
		m.mode = modeList

	case "n", "esc":
		m.mode = modeList
	}

	return m, nil
}
