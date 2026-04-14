package model

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/itxryx/tacktack/internal/db"
	"github.com/itxryx/tacktack/internal/domain"
)

const (
	editFieldTitle    = 0
	editFieldPriority = 1
	editFieldDueDate  = 2
	editFieldTags     = 3
	editFieldComplete = 4
	editFieldCount    = 5
)

// initEditDetail は詳細編集モードを初期化して遷移する。
func (m Model) initEditDetail(taskID uint) (tea.Model, tea.Cmd) {
	task, err := m.taskRepo.FindByID(taskID)
	if err != nil {
		m.lastErr = fmt.Errorf("タスク取得エラー: %w", err)
		return m, nil
	}

	m.editTask = task

	// textinput を初期化
	inputs := make([]textinput.Model, editFieldCount)
	for i := range inputs {
		ti := textinput.New()
		ti.CharLimit = 200
		inputs[i] = ti
	}
	inputs[editFieldTitle].SetValue(task.Title)
	inputs[editFieldTitle].Placeholder = "タイトル"
	inputs[editFieldPriority].SetValue(task.Priority)
	inputs[editFieldPriority].Placeholder = "A-Z (空=優先度なし)"
	inputs[editFieldPriority].CharLimit = 1
	if task.DueDate != nil {
		inputs[editFieldDueDate].SetValue(task.DueDate.Format("2006-01-02"))
	}
	inputs[editFieldDueDate].Placeholder = "YYYY-MM-DD or today/tomorrow"
	inputs[editFieldTitle].Focus()

	m.editInputs = inputs
	m.editField = editFieldTitle
	m.editTagCursor = 0
	m.editLogCursor = 0
	m.editLogFocus = false
	m.editLogEditing = -1
	m.editLogInputs = nil
	m.lastErr = nil
	m.editStoppedLog = nil // 遅延 Stop フラグを初期化
	m.selectedTags = append([]db.Tag{}, task.Tags...)
	m.editPrevMode = m.mode // M2: 遷移元モードを記憶して Esc で戻れるようにする
	m.mode = modeEditDetail
	return m, nil
}

func (m Model) updateEdit(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.editLogFocus {
		return m.updateEditLogArea(msg)
	}
	return m.updateEditTaskArea(msg)
}

func (m Model) updateEditTaskArea(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter", "ctrl+s":
		return m.saveEditDetail()
	case "esc":
		m.editStoppedLog = nil // 遅延 Stop フラグをキャンセル
		m.mode = m.editPrevMode // M2: 遷移元モードへ戻る
		return m, nil
	case "tab":
		if m.editField < editFieldTags {
			m.editInputs[m.editField].Blur()
			m.editField++
			if m.editField < editFieldTags {
				m.editInputs[m.editField].Focus()
			}
		} else if m.editField == editFieldTags {
			m.editField = editFieldComplete
		} else {
			// TimeLog エリアへ移動
			m.editLogFocus = true
		}
	case "shift+tab":
		if m.editField > editFieldTitle {
			if m.editField < editFieldTags {
				m.editInputs[m.editField].Blur()
			}
			m.editField--
			if m.editField < editFieldTags {
				m.editInputs[m.editField].Focus()
			}
		}
	case "+":
		if m.editField == editFieldTags {
			m.pendingTagType = "project"
			m.tagSelectPrevMode = modeEditDetail
			return m.enterTagSelect()
		}
	case "@":
		if m.editField == editFieldTags {
			m.pendingTagType = "context"
			m.tagSelectPrevMode = modeEditDetail
			return m.enterTagSelect()
		}
	case "j", "down":
		if m.editField == editFieldTags && len(m.selectedTags) > 0 {
			if m.editTagCursor < len(m.selectedTags)-1 {
				m.editTagCursor++
			}
			return m, nil
		}
	case "k", "up":
		if m.editField == editFieldTags && len(m.selectedTags) > 0 {
			if m.editTagCursor > 0 {
				m.editTagCursor--
			}
			return m, nil
		}
	case "d":
		if m.editField == editFieldTags && len(m.selectedTags) > 0 {
			m.selectedTags = append(m.selectedTags[:m.editTagCursor], m.selectedTags[m.editTagCursor+1:]...)
			if m.editTagCursor >= len(m.selectedTags) && m.editTagCursor > 0 {
				m.editTagCursor--
			}
			return m, nil
		}
	case " ":
		if m.editField == editFieldComplete && m.editTask != nil {
			m.editTask.IsCompleted = !m.editTask.IsCompleted
			if m.editTask.IsCompleted {
				// C2: 完了時に CompletedAt を設定する
				now := time.Now()
				m.editTask.CompletedAt = &now
				// C3: 計測中のこのタスクを保存時に停止するため遅延 Stop フラグを立てる
				// Esc でキャンセルされた場合は editStoppedLog を破棄して何もしない
				if m.activeLog != nil && m.activeLog.TaskID == m.editTask.ID {
					m.editStoppedLog = m.activeLog
				}
			} else {
				// C2: 未完了に戻す場合は CompletedAt と遅延 Stop フラグをクリア
				m.editTask.CompletedAt = nil
				m.editStoppedLog = nil
			}
			return m, nil
		}
	}

	// フォーカス中の textinput に転送
	if m.editField < editFieldTags {
		var cmd tea.Cmd
		m.editInputs[m.editField], cmd = m.editInputs[m.editField].Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m Model) updateEditLogArea(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.editTask == nil {
		return m, nil
	}
	logs := m.editTask.TimeLogs

	// TimeLog 編集中
	if m.editLogEditing >= 0 {
		return m.updateEditLogInput(msg)
	}

	switch msg.String() {
	case "j", "down":
		if m.editLogCursor < len(logs)-1 {
			m.editLogCursor++
		}
	case "k", "up":
		if m.editLogCursor > 0 {
			m.editLogCursor--
		}
	case "e":
		if len(logs) == 0 {
			return m, nil
		}
		m.editLogEditing = m.editLogCursor
		// textinput を初期化
		startInput := textinput.New()
		startInput.SetValue(logs[m.editLogCursor].StartAt.In(time.Local).Format("2006-01-02 15:04:05"))
		startInput.Placeholder = "YYYY-MM-DD HH:MM:SS"
		startInput.CharLimit = 19
		startInput.Focus()

		endInput := textinput.New()
		if logs[m.editLogCursor].EndAt != nil {
			endInput.SetValue(logs[m.editLogCursor].EndAt.In(time.Local).Format("2006-01-02 15:04:05"))
		}
		endInput.Placeholder = "YYYY-MM-DD HH:MM:SS (空=未終了)"
		endInput.CharLimit = 19

		m.editLogInputs = []textinput.Model{startInput, endInput}
	case "d":
		if len(logs) == 0 {
			return m, nil
		}
		logID := logs[m.editLogCursor].ID
		// C5: 削除対象がアクティブなログかどうかを先に確認する
		isActiveLog := m.activeLog != nil && m.activeLog.ID == logID
		if err := m.timeLogRepo.Delete(logID); err != nil {
			m.lastErr = fmt.Errorf("ログ削除エラー: %w", err)
			return m, nil
		}
		// C5: アクティブなログが削除された場合は m.activeLog をクリア
		if isActiveLog {
			m.activeLog = nil
		}
		m.statsTasks = nil // C6: TimeLog 変更により統計キャッシュを無効化
		// C7: nullLogs を更新
		if err := m.refreshNullLogs(); err != nil {
			m.lastErr = err
		}
		// editTask を更新
		task, err := m.taskRepo.FindByID(m.editTask.ID)
		if err != nil {
			m.lastErr = err
			return m, nil
		}
		m.editTask = task
		if m.editLogCursor >= len(m.editTask.TimeLogs) {
			m.editLogCursor = max(0, len(m.editTask.TimeLogs)-1)
		}
	case "shift+tab":
		m.editLogFocus = false
		m.editField = editFieldComplete
	case "esc":
		m.editStoppedLog = nil // 遅延 Stop フラグをキャンセル
		m.mode = m.editPrevMode // M2: 遷移元モードへ戻る
	}

	return m, nil
}

func (m Model) updateEditLogInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		return m.saveLogEdit()
	case "esc":
		m.editLogEditing = -1
		m.editLogInputs = nil
		return m, nil
	case "tab":
		// start → end の切り替え
		if len(m.editLogInputs) == 2 {
			if m.editLogInputs[0].Focused() {
				m.editLogInputs[0].Blur()
				m.editLogInputs[1].Focus()
			} else {
				m.editLogInputs[1].Blur()
				m.editLogInputs[0].Focus()
			}
		}
		return m, nil
	default:
		var cmd tea.Cmd
		for i := range m.editLogInputs {
			if m.editLogInputs[i].Focused() {
				m.editLogInputs[i], cmd = m.editLogInputs[i].Update(msg)
				break
			}
		}
		return m, cmd
	}
}

func (m Model) saveLogEdit() (tea.Model, tea.Cmd) {
	if m.editLogEditing < 0 || len(m.editLogInputs) < 2 {
		return m, nil
	}
	startStr := strings.TrimSpace(m.editLogInputs[0].Value())
	endStr := strings.TrimSpace(m.editLogInputs[1].Value())

	startAt, err := time.ParseInLocation("2006-01-02 15:04:05", startStr, time.Local)
	if err != nil {
		m.lastErr = fmt.Errorf("開始時刻の形式が正しくありません (YYYY-MM-DD HH:MM:SS)")
		return m, nil
	}

	var endAt *time.Time
	if endStr != "" {
		t, err := time.ParseInLocation("2006-01-02 15:04:05", endStr, time.Local)
		if err != nil {
			m.lastErr = fmt.Errorf("終了時刻の形式が正しくありません (YYYY-MM-DD HH:MM:SS)")
			return m, nil
		}
		if !t.After(startAt) {
			m.lastErr = fmt.Errorf("終了時刻は開始時刻より後でなければなりません")
			return m, nil
		}
		endAt = &t
	}

	log := m.editTask.TimeLogs[m.editLogEditing]

	// C4: endAt を nil にしようとしているが、このログ以外に計測中セッションが存在する場合は拒否する
	if endAt == nil && m.activeLog != nil && m.activeLog.ID != log.ID {
		m.lastErr = fmt.Errorf("既に別のタスクが計測中です。終了時刻を設定してください")
		return m, nil
	}

	log.StartAt = startAt
	log.EndAt = endAt
	if err := m.timeLogRepo.Update(&log); err != nil {
		m.lastErr = fmt.Errorf("ログ更新エラー: %w", err)
		return m, nil
	}

	var cmd tea.Cmd
	if endAt != nil {
		// C5: アクティブなログに終了時刻が設定された場合は m.activeLog をクリア
		if m.activeLog != nil && m.activeLog.ID == log.ID {
			m.activeLog = nil
		}
	} else {
		// C5: 終了時刻が nil の場合（= 計測中）、m.activeLog をこのログに設定しタイマーを開始
		// startAt が編集された場合も m.activeLog を更新することで経過時間表示を正しく保つ
		updatedLog := log
		m.activeLog = &updatedLog
		cmd = tickCmd()
	}

	m.statsTasks = nil // C6: TimeLog 変更により統計キャッシュを無効化
	// C7: nullLogs を更新
	if err := m.refreshNullLogs(); err != nil {
		m.lastErr = err
	}

	// editTask を再ロード
	task, err := m.taskRepo.FindByID(m.editTask.ID)
	if err != nil {
		m.lastErr = err
		return m, nil
	}
	m.editTask = task
	m.editLogEditing = -1
	m.editLogInputs = nil
	m.lastErr = nil
	return m, cmd
}

func (m Model) saveEditDetail() (tea.Model, tea.Cmd) {
	if m.editTask == nil {
		return m, nil
	}

	title := strings.TrimSpace(m.editInputs[editFieldTitle].Value())
	if title == "" {
		m.lastErr = fmt.Errorf("タイトルを入力してください")
		return m, nil
	}

	priority := strings.ToUpper(strings.TrimSpace(m.editInputs[editFieldPriority].Value()))
	if priority != "" && (len([]rune(priority)) != 1 || priority < "A" || priority > "Z") {
		m.lastErr = fmt.Errorf("優先度は A〜Z の1文字、または空にしてください")
		return m, nil
	}

	dueDateStr := strings.TrimSpace(m.editInputs[editFieldDueDate].Value())
	var dueDate *time.Time
	if dueDateStr != "" {
		d, err := domain.ParseDueDate(dueDateStr, time.Now())
		if err != nil {
			m.lastErr = fmt.Errorf("締切日の形式が正しくありません: %w", err)
			return m, nil
		}
		dueDate = d
	}

	m.editTask.Title = title
	m.editTask.Priority = priority
	m.editTask.DueDate = dueDate

	// C3+H1: 遅延 Stop がある場合は Stop・タスク保存・タグ置換を1トランザクションでアトミックに実行する
	if m.editStoppedLog != nil {
		if err := m.taskRepo.StopAndSaveWithTags(m.editStoppedLog.ID, m.editTask, m.selectedTags); err != nil {
			m.lastErr = fmt.Errorf("タスク保存エラー: %w", err)
			return m, nil
		}
		m.activeLog = nil
		m.editStoppedLog = nil
	} else {
		// H1: タスク保存とタグ置換をひとつのトランザクションでアトミックに実行する
		if err := m.taskRepo.SaveWithTags(m.editTask, m.selectedTags); err != nil {
			m.lastErr = fmt.Errorf("タスク保存エラー: %w", err)
			return m, nil
		}
	}

	m.lastErr = nil
	if err := m.refreshTasks(); err != nil {
		m.lastErr = err
	}
	m.clampCursor() // M1: 完了フラグ変更でタスクがリストから消えた場合の cursor 越境を防ぐ
	m.mode = modeList
	return m, nil
}

