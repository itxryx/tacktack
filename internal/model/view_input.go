package model

import (
	"fmt"
	"strings"

	"github.com/itxryx/tacktack/internal/ui"
)

func (m Model) viewInput() string {
	if m.editInputs == nil {
		return "読み込み中..."
	}

	var sb strings.Builder
	sb.WriteString("新しいタスクを追加（Enter: 保存 / Esc: キャンセル）\n\n")

	// タスク属性フィールド
	fields := []struct {
		label string
		idx   int
	}{
		{"タイトル ", editFieldTitle},
		{"優先度  ", editFieldPriority},
		{"締切日  ", editFieldDueDate},
	}
	for _, f := range fields {
		prefix := "  "
		if m.editField == f.idx {
			prefix = "▶ "
		}
		fmt.Fprintf(&sb, "%s%s: %s\n", prefix, f.label, m.editInputs[f.idx].View())
	}

	// タグフィールド
	tagPrefix := "  "
	tagFocused := m.editField == editFieldTags
	if tagFocused {
		tagPrefix = "▶ "
	}
	var tagStrs []string
	for i, tag := range m.selectedTags {
		p := "+"
		if tag.Type == "context" {
			p = "@"
		}
		name := p + tag.Name
		if tagFocused && i == m.editTagCursor {
			name = "[" + name + "]"
		}
		tagStrs = append(tagStrs, name)
	}
	tagDisplay := strings.Join(tagStrs, " ")
	if tagDisplay == "" {
		tagDisplay = "(なし)"
	}
	tagHint := "[+: プロジェクト  @: コンテキスト]"
	if tagFocused && len(m.selectedTags) > 0 {
		tagHint = "[+: プロジェクト  @: コンテキスト  j/k: 移動  d: 削除]"
	}
	fmt.Fprintf(&sb, "%sタグ      : %s  %s\n", tagPrefix, tagDisplay, tagHint)

	if m.lastErr != nil {
		sb.WriteString("\n" + ui.StyleError.Render("✗ "+m.lastErr.Error()) + "\n")
	}

	return sb.String()
}
