package model

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) updateList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "j", "down":
		if len(m.tasks) > 0 && m.cursor < len(m.tasks)-1 {
			m.cursor++
		}
	case "k", "up":
		if m.cursor > 0 {
			m.cursor--
		}
	case "x":
		if len(m.tasks) == 0 {
			return m, nil
		}
		task := m.tasks[m.cursor]
		// 計測中のタスクを完了にする場合は停止とトグルをアトミックに実行
		if m.activeLog != nil && m.activeLog.TaskID == task.ID {
			if err := m.taskRepo.StopAndToggleComplete(m.activeLog.ID, task.ID); err != nil {
				m.lastErr = fmt.Errorf("完了切替エラー: %w", err)
				return m, nil
			}
			m.activeLog = nil
		} else {
			if err := m.taskRepo.ToggleComplete(task.ID); err != nil {
				m.lastErr = fmt.Errorf("完了切替エラー: %w", err)
				return m, nil
			}
		}
		m.lastErr = nil
		if err := m.refreshTasks(); err != nil {
			m.lastErr = err
			return m, nil
		}
		m.clampCursor()

	case "d", "backspace":
		if len(m.tasks) == 0 {
			return m, nil
		}
		m.deleteTargetID = m.tasks[m.cursor].ID
		m.mode = modeDeleteConfirm

	case "i", "a":
		m = initInputMode(m)

	case "e":
		if len(m.tasks) == 0 {
			return m, nil
		}
		return m.initEditDetail(m.tasks[m.cursor].ID)

	case " ":
		return m.handleSpaceKey()

	}

	return m, nil
}

// handleSpaceKey はスペースキーによるタイムトラッキング開始/停止を処理する。
func (m Model) handleSpaceKey() (tea.Model, tea.Cmd) {
	if len(m.tasks) == 0 {
		return m, nil
	}

	currentTask := m.tasks[m.cursor]

	// 完了済みタスクは計測不可
	if currentTask.IsCompleted {
		return m, nil
	}

	// 同じタスクが計測中 → 停止
	if m.activeLog != nil && m.activeLog.TaskID == currentTask.ID {
		if err := m.timeLogRepo.Stop(m.activeLog.ID); err != nil {
			m.lastErr = fmt.Errorf("計測停止エラー: %w", err)
			return m, nil
		}
		m.activeLog = nil
		m.lastErr = nil
		if err := m.refreshTasks(); err != nil {
			m.lastErr = err
		}
		return m, nil // Tick を止める（tickCmd を返さない）
	}

	// 別のタスクが計測中 → 衝突アラート
	if m.activeLog != nil && m.activeLog.TaskID != currentTask.ID {
		m.conflictTaskID = m.activeLog.TaskID
		m.mode = modeTrackingAlert
		return m, nil
	}

	// 計測なし → 開始
	log, err := m.timeLogRepo.Start(currentTask.ID)
	if err != nil {
		m.lastErr = fmt.Errorf("計測開始エラー: %w", err)
		return m, nil
	}
	m.activeLog = log
	m.lastErr = nil
	if err := m.refreshTasks(); err != nil {
		m.lastErr = err
	}
	return m, tickCmd()
}

// formatElapsed は計測開始時刻から現在時刻までの経過時間を HH:MM:SS 形式で返す。
func formatElapsed(startAt time.Time) string {
	d := time.Since(startAt)
	h := int(d.Hours())
	mins := int(d.Minutes()) % 60
	secs := int(d.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d:%02d", h, mins, secs)
}
