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

// initInputMode はタスク追加モードを初期化して遷移する。
func initInputMode(m Model) Model {
	inputs := make([]textinput.Model, editFieldCount)
	for i := range inputs {
		ti := textinput.New()
		ti.CharLimit = 200
		inputs[i] = ti
	}
	inputs[editFieldTitle].Placeholder = "タイトルを入力"
	inputs[editFieldPriority].Placeholder = "A-Z (空=優先度なし)"
	inputs[editFieldPriority].CharLimit = 1
	inputs[editFieldDueDate].Placeholder = "YYYY-MM-DD or today/tomorrow"
	inputs[editFieldTitle].Focus()

	m.editInputs = inputs
	m.editField = editFieldTitle
	m.editTagCursor = 0
	m.editLogFocus = false
	m.selectedTags = nil
	m.lastErr = nil
	m.mode = modeInput
	return m
}

func (m Model) updateInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter", "ctrl+s":
		return m.submitTaskInput()
	case "esc":
		m.selectedTags = nil
		m.lastErr = nil
		m.mode = modeList
		return m, nil
	case "tab":
		// Title→Priority→DueDate→Tags→Title のサイクル
		if m.editField < editFieldTags {
			m.editInputs[m.editField].Blur()
			m.editField++
			if m.editField < editFieldTags {
				m.editInputs[m.editField].Focus()
			}
		} else {
			// Tags → Title に戻る
			m.editField = editFieldTitle
			m.editInputs[editFieldTitle].Focus()
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
			m.tagSelectPrevMode = modeInput
			return m.enterTagSelect()
		}
	case "@":
		if m.editField == editFieldTags {
			m.pendingTagType = "context"
			m.tagSelectPrevMode = modeInput
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
	}

	// フォーカス中の textinput に転送
	if m.editField < editFieldTags {
		var cmd tea.Cmd
		m.editInputs[m.editField], cmd = m.editInputs[m.editField].Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m Model) submitTaskInput() (tea.Model, tea.Cmd) {
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

	// selectedTags の同 ID 重複を排除
	tags := dedupTags(m.selectedTags)

	task := &db.Task{
		Priority: priority,
		Title:    title,
		DueDate:  dueDate,
		Tags:     tags,
	}
	if err := m.taskRepo.Create(task); err != nil {
		m.lastErr = fmt.Errorf("タスク作成エラー: %w", err)
		return m, nil
	}

	m.selectedTags = nil
	m.lastErr = nil
	if err := m.refreshTasks(); err != nil {
		m.lastErr = err
	}
	m.clampCursor() // M3: タスク追加でリスト順序が変わった場合に cursor を有効範囲に収める
	m.mode = modeList
	return m, nil
}

// enterTagSelect はタグ選択モードへの遷移処理を行う。
func (m Model) enterTagSelect() (tea.Model, tea.Cmd) {
	tags, err := m.tagRepo.FindByType(m.pendingTagType)
	if err != nil {
		m.lastErr = fmt.Errorf("タグ取得エラー: %w", err)
		return m, nil
	}
	m.tagList = tags
	m.tagCursor = 0
	m.tagInput = ""
	m.mode = modeTagSelect
	return m, nil
}
