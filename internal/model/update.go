package model

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Update は Bubble Tea の Update ループ。メッセージを受け取り状態を更新する。
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case initDoneMsg:
		if msg.err != nil {
			m.lastErr = msg.err
			return m, nil
		}
		m.tasks = msg.tasks
		m.activeLog = msg.activeLog
		m.nullLogs = msg.nullLogs
		if m.activeLog != nil {
			return m, tickCmd()
		}
		return m, nil

	case tickMsg:
		if m.activeLog != nil {
			return m, tickCmd()
		}
		return m, nil

	case tea.KeyMsg:
		// 共通キー（全モードで有効）
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "q":
			if m.mode == modeList || m.mode == modeStats || m.mode == modeTimeline {
				return m, tea.Quit
			}
		case "tab":
			if m.mode == modeList {
				// リストから遷移するたびに今日の日付にリセット
				m.timelineDate = time.Now()
				m.timelineScroll = firstTaskScroll(m.tasks, m.timelineDate, m.activeLog, m.height)
				m.mode = modeTimeline
				return m, nil
			}
			if m.mode == modeTimeline {
				// 統計ビューへ。初回遷移時に全タスクデータを取得
				if m.statsTasks == nil {
					tasks, err := m.taskRepo.FindAllWithTimeLogs()
					if err != nil {
						m.lastErr = err
					} else {
						m.statsTasks = tasks
					}
				}
				m.mode = modeStats
				return m, nil
			}
			if m.mode == modeStats {
				m.mode = modeList
				return m, nil
			}
		}

		// モード別ハンドラへ委譲
		switch m.mode {
		case modeList:
			return m.updateList(msg)
		case modeInput:
			return m.updateInput(msg)
		case modeTagSelect:
			return m.updateTagSelect(msg)
		case modeEditDetail:
			return m.updateEdit(msg)
		case modeDeleteConfirm:
			return m.updateDeleteConfirm(msg)
		case modeTrackingAlert:
			return m.updateTrackingAlert(msg)
		case modeStats:
			return m.updateStats(msg)
		case modeTimeline:
			return m.updateTimeline(msg)
		}
	}

	return m, nil
}
