package model

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/itxryx/tacktack/internal/db"
)

func (m Model) updateTagSelect(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// タグ削除確認中
	if m.tagDeleteTarget != nil {
		return m.updateTagDeleteConfirm(msg)
	}

	filtered := m.filteredTags()
	total := len(filtered) + 1 // +1 for "新規作成" entry

	switch msg.String() {
	case "down":
		if m.tagCursor < total-1 {
			m.tagCursor++
		}
	case "up":
		if m.tagCursor > 0 {
			m.tagCursor--
		}
	case "enter":
		return m.selectTag(filtered)
	case "D":
		return m.deleteTag(filtered)
	case "esc":
		m.tagInput = ""
		m.tagCursor = 0
		m.mode = m.tagSelectPrevMode
		return m, nil
	case "backspace":
		if len(m.tagInput) > 0 {
			m.tagInput = m.tagInput[:len(m.tagInput)-1]
			m.tagCursor = 0
		}
	default:
		if len(msg.Runes) == 1 {
			m.tagInput += string(msg.Runes)
			m.tagCursor = 0
		}
	}

	return m, nil
}

func (m Model) selectTag(filtered []db.Tag) (tea.Model, tea.Cmd) {
	var tag db.Tag
	// "新規作成" エントリは末尾
	if m.tagCursor < len(filtered) {
		// 既存タグを選択
		tag = filtered[m.tagCursor]
	} else {
		// 新規タグを作成
		name := m.tagInput
		if name == "" {
			// 入力がない場合は何もしない
			m.tagInput = ""
			m.tagCursor = 0
			m.mode = m.tagSelectPrevMode
			return m, nil
		}
		created, err := m.tagRepo.FindOrCreate(name, m.pendingTagType)
		if err != nil {
			m.lastErr = fmt.Errorf("%s作成エラー: %w", tagTypeLabel(m.pendingTagType), err)
			return m, nil
		}
		tag = *created
	}
	// 同じタグ（同 ID）が selectedTags に存在しない場合のみ追加
	if !containsTagID(m.selectedTags, tag.ID) {
		m.selectedTags = append(m.selectedTags, tag)
	}
	m.tagInput = ""
	m.tagCursor = 0
	m.mode = m.tagSelectPrevMode
	return m, nil
}

// deleteTag はカーソル位置のタグの削除確認状態に移行する。
func (m Model) deleteTag(filtered []db.Tag) (tea.Model, tea.Cmd) {
	if m.tagCursor >= len(filtered) {
		// 「新規作成」エントリは削除不可
		return m, nil
	}
	target := filtered[m.tagCursor]
	m.tagDeleteTarget = &target
	return m, nil
}

// updateTagDeleteConfirm はタグ削除確認中のキー入力を処理する。
func (m Model) updateTagDeleteConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "enter":
		target := m.tagDeleteTarget
		m.tagDeleteTarget = nil
		if err := m.tagRepo.Delete(target.ID); err != nil {
			m.lastErr = fmt.Errorf("%s削除エラー: %w", tagTypeLabel(target.Type), err)
			return m, nil
		}
		// tagList から削除
		newTagList := m.tagList[:0]
		for _, t := range m.tagList {
			if t.ID != target.ID {
				newTagList = append(newTagList, t)
			}
		}
		m.tagList = newTagList
		// selectedTags からも削除
		newSelected := m.selectedTags[:0]
		for _, t := range m.selectedTags {
			if t.ID != target.ID {
				newSelected = append(newSelected, t)
			}
		}
		m.selectedTags = newSelected
		// カーソルをクランプ
		newFiltered := m.filteredTags()
		if m.tagCursor >= len(newFiltered) && m.tagCursor > 0 {
			m.tagCursor--
		}
		m.lastErr = nil
	case "n", "esc":
		m.tagDeleteTarget = nil
	}
	return m, nil
}

// filteredTags は tagInput によるフィルタリング済みタグリストを返す。
func (m Model) filteredTags() []db.Tag {
	if m.tagInput == "" {
		return m.tagList
	}
	var result []db.Tag
	for _, tag := range m.tagList {
		if strings.Contains(tag.Name, m.tagInput) {
			result = append(result, tag)
		}
	}
	return result
}
